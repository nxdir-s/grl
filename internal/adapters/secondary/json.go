package secondary

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"sort"

	"github.com/nxdir-s/grl/internal/core/entity"
	"github.com/nxdir-s/grl/internal/core/valobj"
)

const (
	JSONFileExt         string = ".json"
	JSONHistoryFileName string = "history"
	JSONConfigFileName  string = "config"
)

type ErrSaveCollection struct {
	err error
}

func (e *ErrSaveCollection) Error() string {
	return "failed to save collection: " + e.err.Error()
}

type ErrNilCollection struct{}

func (e *ErrNilCollection) Error() string {
	return "collection is nil"
}

type ErrCollectionID struct{}

func (e *ErrCollectionID) Error() string {
	return "missing collection id"
}

type ErrLoadCollection struct {
	err error
}

func (e *ErrLoadCollection) Error() string {
	return "failed to load collection: " + e.err.Error()
}

type ErrListCollections struct {
	err error
}

func (e *ErrListCollections) Error() string {
	return "failed to list collections: " + e.err.Error()
}

type ErrDeleteCollection struct {
	err error
}

func (e *ErrDeleteCollection) Error() string {
	return "failed to delete collection: " + e.err.Error()
}

type ErrSaveHistory struct {
	err error
}

func (e *ErrSaveHistory) Error() string {
	return "failed to save history: " + e.err.Error()
}

type ErrLoadHistory struct {
	err error
}

func (e *ErrLoadHistory) Error() string {
	return "failed to load history: " + e.err.Error()
}

type ErrNilEnvironment struct{}

func (e *ErrNilEnvironment) Error() string {
	return "environment is nil"
}

type ErrEnvironmentID struct{}

func (e *ErrEnvironmentID) Error() string {
	return "missing environment id"
}

type ErrSaveEnvironment struct {
	err error
}

func (e *ErrSaveEnvironment) Error() string {
	return "failed to save environment: " + e.err.Error()
}

type ErrLoadEnvironment struct {
	err error
}

func (e *ErrLoadEnvironment) Error() string {
	return "failed to load environment: " + e.err.Error()
}

type ErrListEnvironments struct {
	err error
}

func (e *ErrListEnvironments) Error() string {
	return "failed to list environments: " + e.err.Error()
}

type ErrDeleteEnvironment struct {
	err error
}

func (e *ErrDeleteEnvironment) Error() string {
	return "failed to delete environment: " + e.err.Error()
}

type ErrNilConfig struct{}

func (e *ErrNilConfig) Error() string {
	return "config is nil"
}

type ErrSaveConfig struct {
	err error
}

func (e *ErrSaveConfig) Error() string {
	return "failed to save config: " + e.err.Error()
}

type ErrLoadConfig struct {
	err error
}

func (e *ErrLoadConfig) Error() string {
	return "failed to load config: " + e.err.Error()
}

type JSONOpt func(a *JSONAdapter)

func WithJSONHistoryFile(filePath string) JSONOpt {
	return func(a *JSONAdapter) {
		a.historyFile = filePath
	}
}

func WithJSONConfigFile(filePath string) JSONOpt {
	return func(a *JSONAdapter) {
		a.configFile = filePath
	}
}

type JSONAdapter struct {
	logger          *slog.Logger
	baseDir         string
	collectionsDir  string
	environmentsDir string
	historyFile     string
	configFile      string
}

func NewJSONAdapter(logger *slog.Logger, baseDir string, collectionsDir string, environmentsDir string, opts ...JSONOpt) *JSONAdapter {
	adapter := &JSONAdapter{
		logger:          logger,
		baseDir:         baseDir,
		collectionsDir:  collectionsDir,
		environmentsDir: environmentsDir,
		historyFile:     filepath.Join(baseDir, JSONHistoryFileName+JSONFileExt),
		configFile:      filepath.Join(baseDir, JSONConfigFileName+JSONFileExt),
	}

	for _, opt := range opts {
		opt(adapter)
	}

	return adapter
}

func (a *JSONAdapter) SaveCollection(ctx context.Context, collection *entity.Collection) error {
	if collection == nil {
		return &ErrNilCollection{}
	}

	data, err := json.MarshalIndent(collection, "", "  ")
	if err != nil {
		return &ErrSaveCollection{err}
	}

	a.logger.Debug("saving collection",
		slog.String("id", collection.ID),
		slog.String("name", collection.Name),
		slog.Int("total_requests", len(collection.Requests)),
	)

	if err := os.WriteFile(a.collectionFile(collection.ID), data, 0o644); err != nil {
		return &ErrSaveCollection{err}
	}

	return nil
}

func (a *JSONAdapter) LoadCollection(ctx context.Context, id string) (*entity.Collection, error) {
	if len(id) == 0 {
		return nil, &ErrCollectionID{}
	}

	data, err := os.ReadFile(a.collectionFile(id))
	if err != nil {
		return nil, &ErrLoadCollection{err}
	}

	a.logger.Debug("loading collection",
		slog.String("id", id),
	)

	var collection entity.Collection
	if err := json.Unmarshal(data, &collection); err != nil {
		return nil, &ErrLoadCollection{err}
	}

	return &collection, nil
}

func (a *JSONAdapter) ListCollections(ctx context.Context) ([]entity.Collection, error) {
	a.logger.Debug("loading collections")

	entries, err := os.ReadDir(a.collectionsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}

		return nil, &ErrListCollections{err}
	}

	collections := make([]entity.Collection, 0)

	for i := range entries {
		if entries[i].IsDir() || filepath.Ext(entries[i].Name()) != JSONFileExt {
			continue
		}

		data, err := os.ReadFile(filepath.Join(a.collectionsDir, entries[i].Name()))
		if err != nil {
			continue
		}

		var collection entity.Collection
		if err := json.Unmarshal(data, &collection); err != nil {
			continue
		}

		collections = append(collections, collection)
	}

	sort.Slice(collections, func(i, j int) bool {
		return collections[i].Name < collections[j].Name
	})

	a.logger.Debug("collections found", slog.Int("total", len(collections)))

	return collections, nil
}

func (a *JSONAdapter) DeleteCollection(ctx context.Context, id string) error {
	if len(id) == 0 {
		return &ErrCollectionID{}
	}

	a.logger.Debug("deleting collection", slog.String("id", id))

	if err := os.Remove(filepath.Join(a.collectionsDir, id+JSONFileExt)); err != nil {
		return &ErrDeleteCollection{err}
	}

	return nil
}

func (a *JSONAdapter) SaveHistory(ctx context.Context, history []entity.HistoryEntry) error {
	if history == nil {
		history = make([]entity.HistoryEntry, 0)
	}

	// history is machine-only, so skip pretty-printing on the hot path
	data, err := json.Marshal(history)
	if err != nil {
		return &ErrSaveHistory{err}
	}

	a.logger.Debug("saving history", slog.Int("total_entries", len(history)))

	if err := os.WriteFile(a.historyFile, data, 0o644); err != nil {
		return &ErrSaveHistory{err}
	}

	return nil
}

func (a *JSONAdapter) LoadHistory(ctx context.Context) ([]entity.HistoryEntry, error) {
	a.logger.Debug("loading history")

	data, err := os.ReadFile(a.historyFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}

		return nil, &ErrLoadHistory{err}
	}

	var history []entity.HistoryEntry
	if err := json.Unmarshal(data, &history); err != nil {
		return nil, &ErrLoadHistory{err}
	}

	if history == nil {
		history = make([]entity.HistoryEntry, 0)
	}

	a.logger.Debug("history found", slog.Int("total_entries", len(history)))

	return history, nil
}

func (s *JSONAdapter) SaveEnvironment(ctx context.Context, env *entity.Environment) error {
	if env == nil {
		return &ErrNilEnvironment{}
	}

	data, err := json.MarshalIndent(env, "", "  ")
	if err != nil {
		return &ErrSaveEnvironment{err}
	}

	path := filepath.Join(s.environmentsDir, env.ID+".json")

	s.logger.Debug("saving environment",
		slog.String("id", env.ID),
		slog.String("name", env.Name),
		slog.String("filePath", path),
	)

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return &ErrSaveEnvironment{err}
	}

	return nil
}

func (a *JSONAdapter) LoadEnvironment(ctx context.Context, id string) (*entity.Environment, error) {
	if len(id) == 0 {
		return nil, &ErrEnvironmentID{}
	}

	a.logger.Debug("loading environment")

	path := filepath.Join(a.environmentsDir, id+".json")

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, &ErrLoadEnvironment{err}
	}

	var env entity.Environment
	if err := json.Unmarshal(data, &env); err != nil {
		return nil, &ErrLoadEnvironment{err}
	}

	a.logger.Debug("environment found",
		slog.String("id", env.ID),
		slog.String("name", env.ID),
	)

	return &env, nil
}

func (a *JSONAdapter) ListEnvironments(ctx context.Context) ([]entity.Environment, error) {
	a.logger.Debug("listing environments")

	entries, err := os.ReadDir(a.environmentsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}

		return nil, &ErrListEnvironments{err}
	}

	var envs []entity.Environment
	for i := range entries {
		if entries[i].IsDir() || filepath.Ext(entries[i].Name()) != ".json" {
			continue
		}

		data, err := os.ReadFile(filepath.Join(a.environmentsDir, entries[i].Name()))
		if err != nil {
			continue
		}

		var env entity.Environment
		if err := json.Unmarshal(data, &env); err != nil {
			continue
		}

		envs = append(envs, env)
	}

	if envs == nil {
		envs = make([]entity.Environment, 0)
	}

	sort.Slice(envs, func(i, j int) bool {
		return envs[i].Name < envs[j].Name
	})

	a.logger.Debug("environments found", slog.Int("total", len(envs)))

	return envs, nil
}

func (a *JSONAdapter) DeleteEnvironment(ctx context.Context, id string) error {
	if len(id) == 0 {
		return &ErrEnvironmentID{}
	}

	a.logger.Debug("deleting environment", slog.String("id", id))

	path := filepath.Join(a.environmentsDir, id+".json")

	if err := os.Remove(path); err != nil {
		return &ErrDeleteEnvironment{err}
	}

	return nil
}

func (a *JSONAdapter) SaveConfig(ctx context.Context, cfg *valobj.Config) error {
	if cfg == nil {
		return &ErrNilConfig{}
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return &ErrSaveConfig{err}
	}

	a.logger.Debug("saving config",
		slog.String("active_env", cfg.ActiveEnvID),
		slog.String("default_method", cfg.DefaultMethod),
		slog.Int("timeout", cfg.TimeoutSeconds),
		slog.Int("history_limit", cfg.HistoryLimit),
		slog.Bool("follow_redirects", cfg.FollowRedirects),
	)

	if err := os.WriteFile(a.configFile, data, 0o644); err != nil {
		return &ErrSaveConfig{err}
	}

	return nil
}

func (a *JSONAdapter) LoadConfig(ctx context.Context) (*valobj.Config, error) {
	a.logger.Debug("loading config")

	data, err := os.ReadFile(a.configFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}

		return nil, &ErrLoadConfig{err}
	}

	var cfg valobj.Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, &ErrLoadConfig{err}
	}

	a.logger.Debug("found config",
		slog.String("active_env", cfg.ActiveEnvID),
		slog.String("default_method", cfg.DefaultMethod),
		slog.Int("timeout", cfg.TimeoutSeconds),
		slog.Int("history_limit", cfg.HistoryLimit),
		slog.Bool("follow_redirects", cfg.FollowRedirects),
	)

	return &cfg, nil
}

func (a *JSONAdapter) collectionFile(id string) string {
	return filepath.Join(a.collectionsDir, id+JSONFileExt)
}
