package primary

import (
	"context"
	"log/slog"

	"github.com/nxdir-s/grl/internal/core/entity"
	"github.com/nxdir-s/grl/internal/ports"
)

const (
	RequestsDomain    string = "requests"
	CollectionsDomain string = "collections"
	HistoryDomain     string = "history"
)

type ErrNilDomain struct {
	domain string
}

func (e *ErrNilDomain) Error() string {
	return "missing required domain '" + e.domain + "'"
}

type ErrSendRequest struct {
	err error
}

func (e *ErrSendRequest) Error() string {
	return "failed to send request: " + e.err.Error()
}

type ErrRecordHistory struct {
	err error
}

func (e *ErrRecordHistory) Error() string {
	return "failed to record history: " + e.err.Error()
}

type ErrGetHistory struct {
	err error
}

func (e *ErrGetHistory) Error() string {
	return "failed to get history: " + e.err.Error()
}

type ErrListCollections struct {
	err error
}

func (e *ErrListCollections) Error() string {
	return "failed to list collections: " + e.err.Error()
}

type CLIOpts func(a *CLIAdapter)

func WithRequests(domain ports.Requests) CLIOpts {
	return func(a *CLIAdapter) {
		a.requests = domain
	}
}

func WithCollections(domain ports.Collections) CLIOpts {
	return func(a *CLIAdapter) {
		a.collections = domain
	}
}

func WithHistory(domain ports.History) CLIOpts {
	return func(a *CLIAdapter) {
		a.history = domain
	}
}

type CLIAdapter struct {
	logger      *slog.Logger
	requests    ports.Requests
	collections ports.Collections
	history     ports.History
}

func NewCLIAdapter(logger *slog.Logger, opts ...CLIOpts) *CLIAdapter {
	adapter := &CLIAdapter{
		logger: logger,
	}

	for _, opt := range opts {
		opt(adapter)
	}

	return adapter
}

func (a *CLIAdapter) SendRequest(ctx context.Context, req *entity.Request) (*entity.Response, error) {
	if a.requests == nil {
		return nil, &ErrNilDomain{RequestsDomain}
	}

	resp, err := a.requests.Send(ctx, req)
	if err != nil {
		return nil, &ErrSendRequest{err}
	}

	// a.logger.Info("recieved response", slog.Any("resp", *resp))

	return resp, nil
}

func (a *CLIAdapter) RecordHistory(ctx context.Context, req *entity.Request, resp *entity.Response) error {
	if a.history == nil {
		return &ErrNilDomain{HistoryDomain}
	}

	if err := a.history.Append(ctx, req, resp); err != nil {
		return &ErrRecordHistory{err}
	}

	return nil
}

func (a *CLIAdapter) GetHistory(ctx context.Context, limit int) ([]entity.HistoryEntry, error) {
	if a.history == nil {
		return nil, &ErrNilDomain{HistoryDomain}
	}

	history, err := a.history.Load(ctx, limit)
	if err != nil {
		return nil, &ErrGetHistory{err}
	}

	return history, nil
}

func (a *CLIAdapter) ListCollections(ctx context.Context) ([]entity.Collection, error) {
	if a.collections == nil {
		return nil, &ErrNilDomain{CollectionsDomain}
	}

	collections, err := a.collections.List(ctx)
	if err != nil {
		return nil, &ErrListCollections{err}
	}

	return collections, nil
}
