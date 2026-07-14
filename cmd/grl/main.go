package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"runtime/pprof"

	"github.com/nxdir-s/grl/internal/adapters/primary"
	"github.com/nxdir-s/grl/internal/adapters/secondary"
	"github.com/nxdir-s/grl/internal/config"
	"github.com/nxdir-s/grl/internal/core/domain"
	"github.com/nxdir-s/grl/internal/core/service"
	"github.com/nxdir-s/grl/internal/logs"
	"github.com/nxdir-s/grl/internal/ports"
	"github.com/nxdir-s/grl/internal/tui"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if path := os.Getenv("GRL_CPUPROFILE"); len(path) != 0 {
		cpuFile, err := os.Create(path)
		if err != nil {
			fmt.Fprintf(os.Stdout, "failed to create cpu profile: %s\n", err.Error())
			os.Exit(1)
		}

		if err := pprof.StartCPUProfile(cpuFile); err != nil {
			fmt.Fprintf(os.Stdout, "failed to start cpu profile: %s\n", err.Error())
			os.Exit(1)
		}

		defer func() {
			pprof.StopCPUProfile()

			if err := cpuFile.Close(); err != nil {
				fmt.Fprintf(os.Stdout, "failed to close cpu profile: %s\n", err.Error())
			}
		}()
	}

	if path := os.Getenv("GRL_MEMPROFILE"); len(path) != 0 {
		defer func() {
			memFile, err := os.Create(path)
			if err != nil {
				fmt.Fprintf(os.Stdout, "failed to create mem profile: %s\n", err.Error())
				return
			}

			defer memFile.Close()

			runtime.GC()

			if err := pprof.WriteHeapProfile(memFile); err != nil {
				fmt.Fprintf(os.Stdout, "failed to write mem profile: %s\n", err.Error())
			}
		}()
	}

	cfg, err := config.New(ctx, config.WithCredentials())
	if err != nil {
		fmt.Fprintf(os.Stdout, "failed to load config: %s\n", err.Error())
		os.Exit(1)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stdout, "failed to get current home directory: %s\n", err.Error())
		os.Exit(1)
	}

	dataDir := filepath.Join(home, ".config", "grl")

	collectionsDir := filepath.Join(dataDir, "collections")
	if err := os.MkdirAll(collectionsDir, 0o755); err != nil {
		fmt.Fprintf(os.Stdout, "failed to create collections directory: %s\n", err.Error())
		os.Exit(1)
	}

	environmentsDir := filepath.Join(dataDir, "environments")
	if err := os.MkdirAll(environmentsDir, 0o755); err != nil {
		fmt.Fprintf(os.Stdout, "failed to create environments directory: %s\n", err.Error())
		os.Exit(1)
	}

	logFile, err := os.Create(filepath.Join(dataDir, "out.log"))
	if err != nil {
		fmt.Fprintf(os.Stdout, "failed to create log file: %s\n", err.Error())
		os.Exit(1)
	}

	defer func() {
		if err := logFile.Close(); err != nil {
			fmt.Fprintf(os.Stdout, "failed to close log file: %s\n", err.Error())
		}
	}()

	logLevel := slog.LevelInfo
	if os.Getenv("GRL_LOG") == "debug" {
		logLevel = slog.LevelDebug
	}

	logger := slog.New(logs.NewHandler(slog.NewTextHandler(logFile, &slog.HandlerOptions{
		Level: logLevel,
	})))
	slog.SetDefault(logger)

	var http ports.Http
	var storage ports.Storage

	var requestService ports.RequestService
	var collectionService ports.CollectionService
	var historyService ports.HistoryService
	var environmentService ports.EnvironmentService
	var configService ports.ConfigService

	var requests ports.Requests
	var collections ports.Collections
	var history ports.History
	var configs ports.Configs
	var environments ports.Environments
	var substitutions ports.Substitutions
	var formatter ports.Formatter
	var clipboard ports.Clipboard
	var auth ports.Auth

	var adapter ports.TUI

	storage = secondary.NewCachedStorage(secondary.NewJSONAdapter(logger, dataDir, collectionsDir, environmentsDir))

	collectionService = service.NewCollectionService(storage)
	historyService = service.NewHistoryService(storage)
	environmentService = service.NewEnvironmentService(storage)
	configService = service.NewConfigService(storage)

	configs = domain.NewConfigs(configService)

	// loading once here also warms the storage cache for the send path
	userCfg, err := configs.Get(ctx)
	if err != nil {
		logger.Error("failed to load config, using defaults", slog.String("err", err.Error()))
		userCfg = configs.Defaults()
	}

	httpCfg := &secondary.HttpConfig{
		TlsConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
		Timeout:         userCfg.TimeoutSeconds,
		FollowRedirects: userCfg.FollowRedirects,
	}

	httpOpts := make([]secondary.HttpOpt, 0)
	if cfg.Credentials {
		httpOpts = append(httpOpts,
			secondary.WithCredentials(ctx, cfg.ClientId, cfg.ClientSecret, cfg.OAuthURL, cfg.OAuthScopes...),
		)
	}

	http = secondary.NewHttpAdapter(httpCfg, logger, httpOpts...)

	requestService = service.NewRequestService(http)

	auth = domain.NewAuth()
	formatter = domain.NewFormatter()
	clipboard = domain.NewClipboard()
	environments = domain.NewEnvironments(environmentService, configs)
	substitutions = domain.NewSubstitutions()
	collections = domain.NewCollections(collectionService)
	history = domain.NewHistory(historyService)
	requests = domain.NewRequests(requestService, environments, substitutions, auth)

	adapter = primary.NewTUIAdapter(logger,
		primary.WithRequests(requests),
		primary.WithCollections(collections),
		primary.WithHistory(history),
		primary.WithEnvironments(environments),
		primary.WithFormatter(formatter),
		primary.WithClipboard(clipboard),
		primary.WithConfigs(configs),
	)

	app := tui.New(adapter, tui.WithContext(ctx))

	if err := app.Run(ctx); err != nil {
		logger.Error("failed to run", slog.String("err", err.Error()))
		os.Exit(1)
	}
}
