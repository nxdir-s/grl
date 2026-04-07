package main

import (
	"context"
	"crypto/tls"
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

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	home, err := os.UserHomeDir()
	if err != nil {
		logger.Error("failed to get current home directory", slog.String("err", err.Error()))
		os.Exit(1)
	}

	dataDir := filepath.Join(home, ".config", "grl")

	collectionsDir := filepath.Join(dataDir, "collections")
	if err := os.MkdirAll(collectionsDir, 0o755); err != nil {
		logger.Error("failed to create collections directory", slog.String("err", err.Error()))
		os.Exit(1)
	}

	var http ports.Http
	var storage ports.Storage

	var requestService ports.RequestService
	var collectionService ports.CollectionService
	var historyService ports.HistoryService

	var requests ports.Requests
	var collections ports.Collections
	var history ports.History

	httpCfg := &secondary.HttpConfig{
		TlsConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	}

	http = secondary.NewHttpAdapter(httpCfg, logger)
	storage = secondary.NewJSONAdapter(dataDir, collectionsDir)

	requestService = service.NewRequestService(http)
	collectionService = service.NewCollectionService(storage)
	historyService = service.NewHistoryService(storage)

	requests = domain.NewRequests(requestService)
	collections = domain.NewCollections(collectionService)
	history = domain.NewHistory(historyService)

	var adapter ports.CLI

	adapter = primary.NewCLIAdapter(logger,
		primary.WithRequests(requests),
		primary.WithCollections(collections),
		primary.WithHistory(history),
	)

	app := tui.New(adapter)

	if err := app.Run(ctx); err != nil {
		logger.Error("failed to run", slog.String("err", err.Error()))
		os.Exit(1)
	}
}
