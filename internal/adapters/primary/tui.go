package primary

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/nxdir-s/grl/internal/core/entity"
	"github.com/nxdir-s/grl/internal/core/valobj"
	"github.com/nxdir-s/grl/internal/ports"
)

const (
	RequestsDomain     string = "requests"
	CollectionsDomain  string = "collections"
	HistoryDomain      string = "history"
	EnvironmentsDomain string = "environments"
	FormatterDomain    string = "formatter"
	ClipboardDomain    string = "clipboard"
	ConfigsDomain      string = "configs"
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

type ErrNilConfig struct{}

func (e *ErrNilConfig) Error() string {
	return "config is nil"
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

type ErrSaveRequestToCollection struct {
	err error
}

func (e *ErrSaveRequestToCollection) Error() string {
	return "failed to save request to collection: " + e.err.Error()
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

type ErrCreateCollection struct {
	err error
}

func (e *ErrCreateCollection) Error() string {
	return "failed to create collection: " + e.err.Error()
}

type ErrListCollections struct {
	err error
}

func (e *ErrListCollections) Error() string {
	return "failed to list collections: " + e.err.Error()
}

type ErrCreateEnvironment struct {
	err error
}

func (e *ErrCreateEnvironment) Error() string {
	return "failed to create environment: " + e.err.Error()
}

type ErrListEnvironments struct {
	err error
}

func (e *ErrListEnvironments) Error() string {
	return "failed to list environments: " + e.err.Error()
}

type ErrGetEnvironment struct {
	err error
}

func (e *ErrGetEnvironment) Error() string {
	return "failed to get environment: " + e.err.Error()
}

type ErrSaveEnvironment struct {
	err error
}

func (e *ErrSaveEnvironment) Error() string {
	return "failed to save environment: " + e.err.Error()
}

type ErrDeleteEnvironment struct {
	err error
}

func (e *ErrDeleteEnvironment) Error() string {
	return "failed to delete environment: " + e.err.Error()
}

type ErrGetActiveEnvironment struct {
	err error
}

func (e *ErrGetActiveEnvironment) Error() string {
	return "failed to get active environment: " + e.err.Error()
}

type ErrSetActiveEnvironment struct {
	err error
}

func (e *ErrSetActiveEnvironment) Error() string {
	return "failed to set active environment: " + e.err.Error()
}

type ErrGetActiveEnvVars struct {
	err error
}

func (e *ErrGetActiveEnvVars) Error() string {
	return "failed to get active env vars: " + e.err.Error()
}

type ErrGetConfig struct {
	err error
}

func (e *ErrGetConfig) Error() string {
	return "failed to get config: " + e.err.Error()
}

type ErrSaveConfig struct {
	err error
}

func (e *ErrSaveConfig) Error() string {
	return "failed to save config: " + e.err.Error()
}

type ErrCopyToClipboard struct {
	err error
}

func (e *ErrCopyToClipboard) Error() string {
	return "failed to copy to clipboard: " + e.err.Error()
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

func WithConfigs(domain ports.Configs) TUIOpts {
	return func(a *TUIAdapter) {
		a.configs = domain
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
	logger *slog.Logger

	requests     ports.Requests
	collections  ports.Collections
	history      ports.History
	environments ports.Environments
	formatter    ports.Formatter
	clipboard    ports.Clipboard
	configs      ports.Configs
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

		return nil, &ErrSendRequest{err}
	}

	if req == nil {
		err := &ErrNilRequest{}
		a.logger.Error("failed to send request", slog.String("err", err.Error()))

		return nil, &ErrSendRequest{err}
	}

	resp, err := a.requests.Send(ctx, req)
	if err != nil {
		a.logger.Error("failed to send request", slog.String("err", err.Error()))
		return nil, &ErrSendRequest{err}
	}

	return resp, nil
}

func (a *TUIAdapter) SaveRequestToCollection(ctx context.Context, req *entity.Request, name string, collectionID string) ([]entity.Collection, error) {
	if a.collections == nil {
		err := &ErrNilDomain{CollectionsDomain}
		a.logger.Error("failed to save request to collection",
			slog.String("name", name),
			slog.String("collection_id", collectionID),
			slog.String("err", err.Error()),
		)

		return nil, &ErrSaveRequestToCollection{err}
	}

	if err := a.collections.AddRequest(ctx, collectionID, req); err != nil {
		return nil, &ErrSaveRequestToCollection{err}
	}

	collections, err := a.collections.List(ctx)
	if err != nil {
		return nil, &ErrSaveRequestToCollection{err}
	}

	return collections, nil
}

func (a *TUIAdapter) RecordHistory(ctx context.Context, req *entity.Request, resp *entity.Response) error {
	if a.history == nil {
		err := &ErrNilDomain{HistoryDomain}
		a.logger.Error("failed to record history", slog.String("err", err.Error()))

		return &ErrRecordHistory{err}
	}

	if req == nil {
		err := &ErrNilRequest{}
		a.logger.Error("failed to record history", slog.String("err", err.Error()))

		return &ErrRecordHistory{err}
	}

	if resp == nil {
		err := &ErrNilResponse{}
		a.logger.Error("failed to record history", slog.String("err", err.Error()))

		return &ErrRecordHistory{err}
	}

	if err := a.history.Append(ctx, req, resp); err != nil {
		a.logger.Error("failed to record history", slog.String("err", err.Error()))
		return &ErrRecordHistory{err}
	}

	return nil
}

func (a *TUIAdapter) GetHistory(ctx context.Context, limit int) ([]entity.HistoryEntry, error) {
	if a.history == nil {
		err := &ErrNilDomain{HistoryDomain}
		a.logger.Error("failed to get history", slog.String("err", err.Error()))

		return nil, &ErrGetHistory{err}
	}

	if limit == 0 {
		err := &ErrLimit{}
		a.logger.Error("failed to get history", slog.String("err", err.Error()))

		return nil, &ErrGetHistory{err}
	}

	history, err := a.history.Load(ctx, limit)
	if err != nil {
		a.logger.Error("failed to get history", slog.String("err", err.Error()))
		return nil, &ErrGetHistory{err}
	}

	return history, nil
}

func (a *TUIAdapter) CreateCollection(ctx context.Context, name string, req *entity.Request) ([]entity.Collection, error) {
	if a.collections == nil {
		err := &ErrNilDomain{CollectionsDomain}
		a.logger.Error("failed to create collection", slog.String("err", err.Error()))

		return nil, &ErrCreateCollection{err}
	}

	if _, err := a.collections.Create(ctx, name, req); err != nil {
		a.logger.Error("failed to create collection", slog.String("err", err.Error()))
		return nil, &ErrCreateCollection{err}
	}

	collections, err := a.collections.List(ctx)
	if err != nil {
		a.logger.Error("failed to create collection", slog.String("err", err.Error()))
		return nil, &ErrCreateCollection{err}
	}

	return collections, nil
}

func (a *TUIAdapter) ListCollections(ctx context.Context) ([]entity.Collection, error) {
	if a.collections == nil {
		err := &ErrNilDomain{CollectionsDomain}
		a.logger.Error("failed to list collections", slog.String("err", err.Error()))

		return nil, &ErrListCollections{err}
	}

	collections, err := a.collections.List(ctx)
	if err != nil {
		a.logger.Error("failed to list collections", slog.String("err", err.Error()))
		return nil, &ErrListCollections{err}
	}

	return collections, nil
}

func (a *TUIAdapter) CreateEnvironment(ctx context.Context, name string) (*entity.Environment, error) {
	if a.environments == nil {
		err := &ErrNilDomain{EnvironmentsDomain}
		a.logger.Error("failed to create environment", slog.String("err", err.Error()))

		return nil, &ErrCreateEnvironment{err}
	}

	if len(name) == 0 {
		err := &ErrMissingEnvName{}
		a.logger.Error("failed to create environment", slog.String("err", err.Error()))

		return nil, &ErrCreateEnvironment{err}
	}

	environment, err := a.environments.Create(ctx, a.generateID(), name)
	if err != nil {
		a.logger.Error("failed to create environment", slog.String("err", err.Error()))
		return nil, &ErrCreateEnvironment{err}
	}

	return environment, nil
}

func (a *TUIAdapter) ListEnvironments(ctx context.Context) ([]entity.Environment, error) {
	if a.environments == nil {
		err := &ErrNilDomain{EnvironmentsDomain}
		a.logger.Error("failed to list environments", slog.String("err", err.Error()))

		return nil, &ErrListEnvironments{err}
	}

	environments, err := a.environments.List(ctx)
	if err != nil {
		a.logger.Error("failed to list environments", slog.String("err", err.Error()))
		return nil, &ErrListEnvironments{err}
	}

	return environments, nil
}

func (a *TUIAdapter) GetEnvironment(ctx context.Context, id string) (*entity.Environment, error) {
	if a.environments == nil {
		err := &ErrNilDomain{EnvironmentsDomain}
		a.logger.Error("failed to get environment", slog.String("err", err.Error()))

		return nil, &ErrGetEnvironment{err}
	}

	if len(id) == 0 {
		err := &ErrMissingEnvID{}
		a.logger.Error("failed to get environment", slog.String("err", err.Error()))

		return nil, &ErrGetEnvironment{err}
	}

	environment, err := a.environments.Get(ctx, id)
	if err != nil {
		a.logger.Error("failed to get environment", slog.String("err", err.Error()))
		return nil, &ErrGetEnvironment{err}
	}

	return environment, nil
}

func (a *TUIAdapter) SaveEnvironment(ctx context.Context, env *entity.Environment) error {
	if a.environments == nil {
		err := &ErrNilDomain{EnvironmentsDomain}
		a.logger.Error("failed to save environment", slog.String("err", err.Error()))

		return &ErrSaveEnvironment{err}
	}

	if env == nil {
		err := &ErrNilEnv{}
		a.logger.Error("failed to save environment", slog.String("err", err.Error()))

		return &ErrSaveEnvironment{err}
	}

	if env.Variables == nil {
		env.Variables = make(map[string]string)
	}

	if err := a.environments.Save(ctx, env); err != nil {
		a.logger.Error("failed to save environment", slog.String("err", err.Error()))
		return &ErrSaveEnvironment{err}
	}

	return nil
}

func (a *TUIAdapter) DeleteEnvironment(ctx context.Context, id string) error {
	if a.environments == nil {
		err := &ErrNilDomain{EnvironmentsDomain}
		a.logger.Error("failed to delete environment", slog.String("err", err.Error()))

		return &ErrDeleteEnvironment{err}
	}

	if len(id) == 0 {
		err := &ErrMissingEnvID{}
		a.logger.Error("failed to delete environment", slog.String("err", err.Error()))

		return &ErrDeleteEnvironment{err}
	}

	if err := a.environments.Delete(ctx, id); err != nil {
		a.logger.Error("failed to delete environment", slog.String("err", err.Error()))
		return &ErrDeleteEnvironment{err}
	}

	return nil
}

func (a *TUIAdapter) GetActiveEnvironment(ctx context.Context) (*entity.Environment, error) {
	if a.environments == nil {
		err := &ErrNilDomain{EnvironmentsDomain}
		a.logger.Error("failed to get active environment", slog.String("err", err.Error()))

		return nil, &ErrGetActiveEnvironment{err}
	}

	env, err := a.environments.GetActive(ctx)
	if err != nil {
		a.logger.Error("failed to get active environment", slog.String("err", err.Error()))

		return nil, &ErrGetActiveEnvironment{err}
	}

	return env, nil
}

func (a *TUIAdapter) SetActiveEnvironment(ctx context.Context, id string) error {
	if a.environments == nil {
		err := &ErrNilDomain{EnvironmentsDomain}
		a.logger.Error("failed to set active environment", slog.String("err", err.Error()))

		return &ErrSetActiveEnvironment{err}
	}

	if len(id) == 0 {
		err := &ErrMissingEnvID{}
		a.logger.Error("failed to set active environment", slog.String("err", err.Error()))

		return &ErrSetActiveEnvironment{err}
	}

	if err := a.environments.SetActive(ctx, id); err != nil {
		a.logger.Error("failed to set active environment", slog.String("err", err.Error()))
		return &ErrSetActiveEnvironment{err}
	}

	return nil
}

func (a *TUIAdapter) GetActiveEnvVars(ctx context.Context) (map[string]string, error) {
	if a.environments == nil {
		err := &ErrNilDomain{EnvironmentsDomain}
		a.logger.Error("failed to get active env vars", slog.String("err", err.Error()))

		return map[string]string{}, &ErrGetActiveEnvVars{err}
	}

	return a.environments.ActiveVars(ctx), nil
}

func (a *TUIAdapter) GetConfig(ctx context.Context) *valobj.Config {
	if a.configs == nil {
		err := &ErrNilDomain{ConfigsDomain}
		a.logger.Error("failed to get config", slog.String("err", err.Error()))

		return &valobj.Config{}
	}

	config, err := a.configs.Get(ctx)
	if err != nil {
		a.logger.Error("failed to get config", slog.String("err", err.Error()))
	}

	if config == nil {
		return &valobj.Config{}
	}

	return config
}

func (a *TUIAdapter) SaveConfig(ctx context.Context, cfg *valobj.Config) error {
	if a.configs == nil {
		err := &ErrNilDomain{ConfigsDomain}
		a.logger.Error("failed to save config", slog.String("err", err.Error()))

		return &ErrSaveConfig{err}
	}

	if cfg == nil {
		err := &ErrNilConfig{}
		a.logger.Error("failed to save config", slog.String("err", err.Error()))

		return &ErrSaveConfig{err}
	}

	if err := a.configs.Save(ctx, cfg); err != nil {
		a.logger.Error("failed to save config", slog.String("err", err.Error()))
		return &ErrSaveConfig{err}
	}

	return nil
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

		return &ErrCopyToClipboard{err}
	}

	if err := a.clipboard.Copy(s); err != nil {
		a.logger.Error("failed to copy to clipboard", slog.String("err", err.Error()))
		return &ErrCopyToClipboard{err}
	}

	return nil
}

func (a *TUIAdapter) generateID() string {
	return uuid.New().String()
}
