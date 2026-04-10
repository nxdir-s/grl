package primary

import (
	"context"
	"log/slog"

	"github.com/nxdir-s/grl/internal/core/entity"
	"github.com/nxdir-s/grl/internal/ports"
)

const (
	RequestsDomain     string = "requests"
	CollectionsDomain  string = "collections"
	HistoryDomain      string = "history"
	EnvironmentsDomain string = "environments"
	FormatterDomain    string = "formatter"
	ClipboardDomain    string = "clipboard"
)

type ErrNilDomain struct {
	domain string
}

func (e *ErrNilDomain) Error() string {
	return "missing required domain '" + e.domain + "'"
}

type ErrNilEnv struct{}

func (e *ErrNilEnv) Error() string {
	return "environment is nil"
}

type ErrNilRequest struct{}

func (e *ErrNilRequest) Error() string {
	return "request is nil"
}

type ErrNilResponse struct{}

func (e *ErrNilResponse) Error() string {
	return "response is nil"
}

type ErrLimit struct{}

func (e *ErrLimit) Error() string {
	return "limit must be greater than 0"
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

type ErrMissingEnvName struct{}

func (e *ErrMissingEnvName) Error() string {
	return "environment name is required"
}

type ErrMissingEnvID struct{}

func (e *ErrMissingEnvID) Error() string {
	return "environment id is required"
}

type TUIOpts func(a *TUIAdapter)

func WithRequests(domain ports.Requests) TUIOpts {
	return func(a *TUIAdapter) {
		a.requests = domain
	}
}

func WithCollections(domain ports.Collections) TUIOpts {
	return func(a *TUIAdapter) {
		a.collections = domain
	}
}

func WithHistory(domain ports.History) TUIOpts {
	return func(a *TUIAdapter) {
		a.history = domain
	}
}

func WithEnvironments(domain ports.Environments) TUIOpts {
	return func(a *TUIAdapter) {
		a.environments = domain
	}
}

func WithFormatter(domain ports.Formatter) TUIOpts {
	return func(a *TUIAdapter) {
		a.formatter = domain
	}
}

func WithClipboard(domain ports.Clipboard) TUIOpts {
	return func(a *TUIAdapter) {
		a.clipboard = domain
	}
}

type TUIAdapter struct {
	logger       *slog.Logger
	requests     ports.Requests
	collections  ports.Collections
	history      ports.History
	environments ports.Environments
	formatter    ports.Formatter
	clipboard    ports.Clipboard
}

func NewTUIAdapter(logger *slog.Logger, opts ...TUIOpts) *TUIAdapter {
	adapter := &TUIAdapter{
		logger: logger,
	}

	for _, opt := range opts {
		opt(adapter)
	}

	return adapter
}

func (a *TUIAdapter) SendRequest(ctx context.Context, req *entity.Request) (*entity.Response, error) {
	if a.requests == nil {
		err := &ErrNilDomain{RequestsDomain}
		a.logger.Error("failed to send request", slog.String("err", err.Error()))

		return nil, err
	}

	if req == nil {
		return nil, &ErrNilRequest{}
	}

	resp, err := a.requests.Send(ctx, req)
	if err != nil {
		return nil, &ErrSendRequest{err}
	}

	return resp, nil
}

func (a *TUIAdapter) RecordHistory(ctx context.Context, req *entity.Request, resp *entity.Response) error {
	if a.history == nil {
		err := &ErrNilDomain{HistoryDomain}
		a.logger.Error("failed to record history", slog.String("err", err.Error()))

		return err
	}

	if req == nil {
		return &ErrNilRequest{}
	}

	if resp == nil {
		return &ErrNilResponse{}
	}

	if err := a.history.Append(ctx, req, resp); err != nil {
		return &ErrRecordHistory{err}
	}

	return nil
}

func (a *TUIAdapter) GetHistory(ctx context.Context, limit int) ([]entity.HistoryEntry, error) {
	if a.history == nil {
		err := &ErrNilDomain{HistoryDomain}
		a.logger.Error("failed to get history", slog.String("err", err.Error()))

		return nil, err
	}

	if limit == 0 {
		return nil, &ErrLimit{}
	}

	history, err := a.history.Load(ctx, limit)
	if err != nil {
		return nil, &ErrGetHistory{err}
	}

	return history, nil
}

func (a *TUIAdapter) ListCollections(ctx context.Context) ([]entity.Collection, error) {
	if a.collections == nil {
		err := &ErrNilDomain{CollectionsDomain}
		a.logger.Error("failed to list collections", slog.String("err", err.Error()))

		return nil, err
	}

	collections, err := a.collections.List(ctx)
	if err != nil {
		return nil, &ErrListCollections{err}
	}

	return collections, nil
}

func (a *TUIAdapter) CreateEnvironment(ctx context.Context, name string) (*entity.Environment, error) {
	if a.environments == nil {
		err := &ErrNilDomain{EnvironmentsDomain}
		a.logger.Error("failed to create environment", slog.String("err", err.Error()))

		return nil, err
	}

	if len(name) == 0 {
		return nil, &ErrMissingEnvName{}
	}

	environment, err := a.environments.Create(ctx, name)
	if err != nil {
		return nil, err
	}

	return environment, nil
}

func (a *TUIAdapter) ListEnvironments(ctx context.Context) ([]entity.Environment, error) {
	if a.environments == nil {
		err := &ErrNilDomain{EnvironmentsDomain}
		a.logger.Error("failed to list environments", slog.String("err", err.Error()))

		return nil, err
	}

	environments, err := a.environments.List(ctx)
	if err != nil {
		return nil, err
	}

	return environments, nil
}

func (a *TUIAdapter) GetEnvironment(ctx context.Context, id string) (*entity.Environment, error) {
	if a.environments == nil {
		err := &ErrNilDomain{EnvironmentsDomain}
		a.logger.Error("failed to get environment", slog.String("err", err.Error()))

		return nil, err
	}

	if len(id) == 0 {
		return nil, &ErrMissingEnvID{}
	}

	environment, err := a.environments.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	return environment, nil
}

func (a *TUIAdapter) SaveEnvironment(ctx context.Context, env *entity.Environment) error {
	if a.environments == nil {
		err := &ErrNilDomain{EnvironmentsDomain}
		a.logger.Error("failed to save environment", slog.String("err", err.Error()))

		return err
	}

	if env == nil {
		return &ErrNilEnv{}
	}

	if env.Variables == nil {
		env.Variables = make(map[string]string)
	}

	if err := a.environments.Save(ctx, env); err != nil {
		a.logger.Error("failed to save environment", slog.String("err", err.Error()))
		return err
	}

	return nil
}

func (a *TUIAdapter) DeleteEnvironment(ctx context.Context, id string) error {
	if a.environments == nil {
		err := &ErrNilDomain{EnvironmentsDomain}
		a.logger.Error("failed to delete environment", slog.String("err", err.Error()))

		return err
	}

	if len(id) == 0 {
		return &ErrMissingEnvID{}
	}

	if err := a.environments.Delete(ctx, id); err != nil {
		return err
	}

	return nil
}

func (a *TUIAdapter) GetActiveEnvironment(ctx context.Context) (*entity.Environment, error) {
	if a.environments == nil {
		err := &ErrNilDomain{EnvironmentsDomain}
		a.logger.Error("failed to get active environment", slog.String("err", err.Error()))

		return nil, err
	}

	env, err := a.environments.GetActive(ctx)
	if err != nil {
		return nil, err
	}

	return env, nil
}

func (a *TUIAdapter) SetActiveEnvironment(ctx context.Context, id string) error {
	if a.environments == nil {
		err := &ErrNilDomain{EnvironmentsDomain}
		a.logger.Error("failed to set active environment", slog.String("err", err.Error()))

		return err
	}

	if len(id) == 0 {
		return &ErrMissingEnvID{}
	}

	if err := a.environments.SetActive(ctx, id); err != nil {
		return err
	}

	return nil
}

func (a *TUIAdapter) GetActiveEnvVars(ctx context.Context) (map[string]string, error) {
	if a.environments == nil {
		err := &ErrNilDomain{EnvironmentsDomain}
		a.logger.Error("failed to get active env vars", slog.String("err", err.Error()))

		return map[string]string{}, err
	}

	return a.environments.ActiveVars(ctx), nil
}

func (a *TUIAdapter) ColorizeJSON(s string) string {
	if a.formatter == nil {
		err := &ErrNilDomain{FormatterDomain}
		a.logger.Error("failed to colorize JSON", slog.String("err", err.Error()))

		return s
	}

	if len(s) == 0 {
		return s
	}

	return a.formatter.ColorizeJSON(s)
}

func (a *TUIAdapter) CopyToClipboard(s string) error {
	if a.clipboard == nil {
		err := &ErrNilDomain{ClipboardDomain}
		a.logger.Error("failed to copy to clipboard", slog.String("err", err.Error()))

		return err
	}

	if err := a.clipboard.Copy(s); err != nil {
		return err
	}

	return nil
}
