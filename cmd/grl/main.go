package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/nxdir-s/grl/internal/adapters/primary"
	"github.com/nxdir-s/grl/internal/adapters/secondary"
	"github.com/nxdir-s/grl/internal/core/domain"
	"github.com/nxdir-s/grl/internal/core/service"
	"github.com/nxdir-s/grl/internal/ports"
	"github.com/nxdir-s/grl/internal/tui"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

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

	logger := slog.New(slog.NewTextHandler(logFile, nil))
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

	var adapter ports.TUI

	httpCfg := &secondary.HttpConfig{
		TlsConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	}

	http = secondary.NewHttpAdapter(httpCfg, logger)
	storage = secondary.NewJSONAdapter(logger, dataDir, collectionsDir, environmentsDir)

	requestService = service.NewRequestService(http)
	collectionService = service.NewCollectionService(storage)
	historyService = service.NewHistoryService(storage)
	environmentService = service.NewEnvironmentService(storage)
	configService = service.NewConfigService(storage)

	formatter = domain.NewFormatter()
	clipboard = domain.NewClipboard()
	configs = domain.NewConfigs(configService)
	environments = domain.NewEnvironments(environmentService, configs)
	substitutions = domain.NewSubstitutions()
	collections = domain.NewCollections(collectionService)
	history = domain.NewHistory(historyService)
	requests = domain.NewRequests(requestService, environments, substitutions)

	adapter = primary.NewTUIAdapter(logger,
		primary.WithRequests(requests),
		primary.WithCollections(collections),
		primary.WithHistory(history),
		primary.WithEnvironments(environments),
		primary.WithFormatter(formatter),
		primary.WithClipboard(clipboard),
	)

	app := tui.New(adapter, tui.WithContext(ctx))

	if err := app.Run(ctx); err != nil {
		logger.Error("failed to run", slog.String("err", err.Error()))
		os.Exit(1)
	}
}
