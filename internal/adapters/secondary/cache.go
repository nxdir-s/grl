package secondary

import (
	"context"
	"maps"
	"slices"
	"sync"

	"github.com/nxdir-s/grl/internal/core/entity"
	"github.com/nxdir-s/grl/internal/core/valobj"
	"github.com/nxdir-s/grl/internal/ports"
)

// CachedStorage decorates a ports.Storage with an in-memory write-through
// cache so hot paths (request sends, sidebar refreshes) avoid disk reads.
//
// Values are cloned on every read and write: callers mutate what they get
// back (config defaults, history reordering), so sharing interior pointers
// with the cache would race. LoadCollection passes through uncached because
// its result is mutated and re-saved by the collections domain.
type CachedStorage struct {
	inner ports.Storage

	mu sync.RWMutex

	cfg       *valobj.Config
	cfgLoaded bool

	envs          map[string]*entity.Environment
	envList       []entity.Environment
	envListLoaded bool

	collectionList       []entity.Collection
	collectionListLoaded bool

	history       []entity.HistoryEntry
	historyLoaded bool
}

func NewCachedStorage(inner ports.Storage) *CachedStorage {
	return &CachedStorage{
		inner: inner,
		envs:  make(map[string]*entity.Environment),
	}
}

func (a *CachedStorage) LoadCollection(ctx context.Context, id string) (*entity.Collection, error) {
	return a.inner.LoadCollection(ctx, id)
}

func (a *CachedStorage) SaveCollection(ctx context.Context, collection *entity.Collection) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.collectionList = nil
	a.collectionListLoaded = false

	return a.inner.SaveCollection(ctx, collection)
}

func (a *CachedStorage) ListCollections(ctx context.Context) ([]entity.Collection, error) {
	a.mu.RLock()
	if a.collectionListLoaded {
		cached := cloneCollections(a.collectionList)
		a.mu.RUnlock()

		return cached, nil
	}
	a.mu.RUnlock()

	a.mu.Lock()
	defer a.mu.Unlock()

	if a.collectionListLoaded {
		return cloneCollections(a.collectionList), nil
	}

	collections, err := a.inner.ListCollections(ctx)
	if err != nil {
		return nil, err
	}

	a.collectionList = cloneCollections(collections)
	a.collectionListLoaded = true

	return collections, nil
}

func (a *CachedStorage) DeleteCollection(ctx context.Context, id string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.collectionList = nil
	a.collectionListLoaded = false

	return a.inner.DeleteCollection(ctx, id)
}

func (a *CachedStorage) LoadHistory(ctx context.Context) ([]entity.HistoryEntry, error) {
	a.mu.RLock()
	if a.historyLoaded {
		cached := slices.Clone(a.history)
		a.mu.RUnlock()

		return cached, nil
	}
	a.mu.RUnlock()

	a.mu.Lock()
	defer a.mu.Unlock()

	if a.historyLoaded {
		return slices.Clone(a.history), nil
	}

	history, err := a.inner.LoadHistory(ctx)
	if err != nil {
		return nil, err
	}

	a.history = slices.Clone(history)
	a.historyLoaded = true

	return history, nil
}

func (a *CachedStorage) SaveHistory(ctx context.Context, history []entity.HistoryEntry) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if err := a.inner.SaveHistory(ctx, history); err != nil {
		a.history = nil
		a.historyLoaded = false

		return err
	}

	a.history = slices.Clone(history)
	a.historyLoaded = true

	return nil
}

func (a *CachedStorage) LoadEnvironment(ctx context.Context, id string) (*entity.Environment, error) {
	a.mu.RLock()
	if env, ok := a.envs[id]; ok {
		cached := cloneEnvironment(env)
		a.mu.RUnlock()

		return cached, nil
	}
	a.mu.RUnlock()

	a.mu.Lock()
	defer a.mu.Unlock()

	if env, ok := a.envs[id]; ok {
		return cloneEnvironment(env), nil
	}

	env, err := a.inner.LoadEnvironment(ctx, id)
	if err != nil {
		return nil, err
	}

	a.envs[id] = cloneEnvironment(env)

	return env, nil
}

func (a *CachedStorage) SaveEnvironment(ctx context.Context, env *entity.Environment) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.envList = nil
	a.envListLoaded = false

	if err := a.inner.SaveEnvironment(ctx, env); err != nil {
		if env != nil {
			delete(a.envs, env.ID)
		}

		return err
	}

	a.envs[env.ID] = cloneEnvironment(env)

	return nil
}

func (a *CachedStorage) ListEnvironments(ctx context.Context) ([]entity.Environment, error) {
	a.mu.RLock()
	if a.envListLoaded {
		cached := cloneEnvironments(a.envList)
		a.mu.RUnlock()

		return cached, nil
	}
	a.mu.RUnlock()

	a.mu.Lock()
	defer a.mu.Unlock()

	if a.envListLoaded {
		return cloneEnvironments(a.envList), nil
	}

	envs, err := a.inner.ListEnvironments(ctx)
	if err != nil {
		return nil, err
	}

	a.envList = cloneEnvironments(envs)
	a.envListLoaded = true

	return envs, nil
}

func (a *CachedStorage) DeleteEnvironment(ctx context.Context, id string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	delete(a.envs, id)
	a.envList = nil
	a.envListLoaded = false

	return a.inner.DeleteEnvironment(ctx, id)
}

func (a *CachedStorage) SaveConfig(ctx context.Context, cfg *valobj.Config) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if err := a.inner.SaveConfig(ctx, cfg); err != nil {
		a.cfg = nil
		a.cfgLoaded = false

		return err
	}

	a.cfg = cloneConfig(cfg)
	a.cfgLoaded = true

	return nil
}

func (a *CachedStorage) LoadConfig(ctx context.Context) (*valobj.Config, error) {
	a.mu.RLock()
	if a.cfgLoaded {
		cached := cloneConfig(a.cfg)
		a.mu.RUnlock()

		return cached, nil
	}
	a.mu.RUnlock()

	a.mu.Lock()
	defer a.mu.Unlock()

	if a.cfgLoaded {
		return cloneConfig(a.cfg), nil
	}

	cfg, err := a.inner.LoadConfig(ctx)
	if err != nil {
		return nil, err
	}

	// a nil config with a nil error means no config exists yet; cache the
	// absence so a missing file isn't re-read on every request send
	a.cfg = cloneConfig(cfg)
	a.cfgLoaded = true

	return cfg, nil
}

func cloneConfig(cfg *valobj.Config) *valobj.Config {
	if cfg == nil {
		return nil
	}

	out := *cfg

	return &out
}

func cloneEnvironment(env *entity.Environment) *entity.Environment {
	if env == nil {
		return nil
	}

	return &entity.Environment{
		ID:        env.ID,
		Name:      env.Name,
		Variables: maps.Clone(env.Variables),
	}
}

func cloneEnvironments(envs []entity.Environment) []entity.Environment {
	if envs == nil {
		return nil
	}

	out := make([]entity.Environment, 0, len(envs))
	for i := range envs {
		out = append(out, *cloneEnvironment(&envs[i]))
	}

	return out
}

func cloneRequest(req *entity.Request) entity.Request {
	out := *req

	out.Headers = slices.Clone(req.Headers)
	out.Params = slices.Clone(req.Params)
	out.FormFields = slices.Clone(req.FormFields)

	return out
}

func cloneCollections(collections []entity.Collection) []entity.Collection {
	if collections == nil {
		return nil
	}

	out := make([]entity.Collection, 0, len(collections))

	for i := range collections {
		collection := collections[i]

		collection.Requests = make([]entity.Request, 0, len(collections[i].Requests))
		for j := range collections[i].Requests {
			collection.Requests = append(collection.Requests, cloneRequest(&collections[i].Requests[j]))
		}

		out = append(out, collection)
	}

	return out
}
