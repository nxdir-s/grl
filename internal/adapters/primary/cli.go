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

func WithEnvironments(domain ports.Environments) CLIOpts {
	return func(a *CLIAdapter) {
		a.environments = domain
	}
}

type CLIAdapter struct {
	logger       *slog.Logger
	requests     ports.Requests
	collections  ports.Collections
	history      ports.History
	environments ports.Environments
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

	if req == nil {
		return nil, &ErrNilRequest{}
	}

	resp, err := a.requests.Send(ctx, req)
	if err != nil {
		return nil, &ErrSendRequest{err}
	}

	return resp, nil
}

func (a *CLIAdapter) RecordHistory(ctx context.Context, req *entity.Request, resp *entity.Response) error {
	if a.history == nil {
		return &ErrNilDomain{HistoryDomain}
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

func (a *CLIAdapter) GetHistory(ctx context.Context, limit int) ([]entity.HistoryEntry, error) {
	if a.history == nil {
		return nil, &ErrNilDomain{HistoryDomain}
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

func (a *CLIAdapter) CreateEnvironment(ctx context.Context, name string) (*entity.Environment, error) {
	if a.environments == nil {
		return nil, &ErrNilDomain{EnvironmentsDomain}
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

func (a *CLIAdapter) ListEnvironments(ctx context.Context) ([]entity.Environment, error) {
	if a.environments == nil {
		return nil, &ErrNilDomain{EnvironmentsDomain}
	}

	environments, err := a.environments.List(ctx)
	if err != nil {
		return nil, err
	}

	return environments, nil
}

func (a *CLIAdapter) GetEnvironment(ctx context.Context, id string) (*entity.Environment, error) {
	if a.environments == nil {
		return nil, &ErrNilDomain{EnvironmentsDomain}
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

func (a *CLIAdapter) SaveEnvironment(ctx context.Context, env *entity.Environment) error {
	if a.environments == nil {
		return &ErrNilDomain{EnvironmentsDomain}
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

func (a *CLIAdapter) DeleteEnvironment(ctx context.Context, id string) error {
	if a.environments == nil {
		return &ErrNilDomain{EnvironmentsDomain}
	}

	if len(id) == 0 {
		return &ErrMissingEnvID{}
	}

	if err := a.environments.Delete(ctx, id); err != nil {
		return err
	}

	return nil
}

func (a *CLIAdapter) GetActiveEnvironment(ctx context.Context) (*entity.Environment, error) {
	if a.environments == nil {
		return nil, &ErrNilDomain{EnvironmentsDomain}
	}

	env, err := a.environments.GetActive(ctx)
	if err != nil {
		return nil, err
	}

	return env, nil
}

func (a *CLIAdapter) SetActiveEnvironment(ctx context.Context, id string) error {
	if a.environments == nil {
		return &ErrNilDomain{EnvironmentsDomain}
	}

	if len(id) == 0 {
		return &ErrMissingEnvID{}
	}

	if err := a.environments.SetActive(ctx, id); err != nil {
		return err
	}

	return nil
}

func (a *CLIAdapter) GetActiveEnvVars(ctx context.Context) (map[string]string, error) {
	if a.environments == nil {
		return map[string]string{}, &ErrNilDomain{EnvironmentsDomain}
	}

	return a.environments.ActiveVars(ctx), nil
}
