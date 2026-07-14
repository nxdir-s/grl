package secondary

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nxdir-s/grl/internal/core/domain"
	"github.com/nxdir-s/grl/internal/core/entity"
	"github.com/nxdir-s/grl/internal/core/service"
	"github.com/nxdir-s/grl/internal/core/valobj"
)

// countingStorage is a ports.Storage fake that records how many times each
// inner operation runs
type countingStorage struct {
	mu     sync.Mutex
	counts map[string]int

	cfg         *valobj.Config
	envs        map[string]*entity.Environment
	collections []entity.Collection
	history     []entity.HistoryEntry
}

func newCountingStorage() *countingStorage {
	return &countingStorage{
		counts: make(map[string]int),
		envs:   make(map[string]*entity.Environment),
	}
}

func (s *countingStorage) count(op string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.counts[op]++
}

func (s *countingStorage) calls(op string) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.counts[op]
}

func (s *countingStorage) LoadCollection(ctx context.Context, id string) (*entity.Collection, error) {
	s.count("LoadCollection")

	for i := range s.collections {
		if s.collections[i].ID == id {
			collection := s.collections[i]
			return &collection, nil
		}
	}

	return nil, &ErrLoadCollection{fmt.Errorf("not found")}
}

func (s *countingStorage) SaveCollection(ctx context.Context, collection *entity.Collection) error {
	s.count("SaveCollection")
	return nil
}

func (s *countingStorage) ListCollections(ctx context.Context) ([]entity.Collection, error) {
	s.count("ListCollections")
	return s.collections, nil
}

func (s *countingStorage) DeleteCollection(ctx context.Context, id string) error {
	s.count("DeleteCollection")
	return nil
}

func (s *countingStorage) LoadHistory(ctx context.Context) ([]entity.HistoryEntry, error) {
	s.count("LoadHistory")
	return s.history, nil
}

func (s *countingStorage) SaveHistory(ctx context.Context, history []entity.HistoryEntry) error {
	s.count("SaveHistory")
	s.history = history

	return nil
}

func (s *countingStorage) LoadEnvironment(ctx context.Context, id string) (*entity.Environment, error) {
	s.count("LoadEnvironment")

	env, ok := s.envs[id]
	if !ok {
		return nil, &ErrLoadEnvironment{fmt.Errorf("not found")}
	}

	return env, nil
}

func (s *countingStorage) SaveEnvironment(ctx context.Context, env *entity.Environment) error {
	s.count("SaveEnvironment")

	if env == nil {
		return &ErrNilEnvironment{}
	}

	s.envs[env.ID] = env

	return nil
}

func (s *countingStorage) ListEnvironments(ctx context.Context) ([]entity.Environment, error) {
	s.count("ListEnvironments")

	envs := make([]entity.Environment, 0, len(s.envs))
	for _, env := range s.envs {
		envs = append(envs, *env)
	}

	return envs, nil
}

func (s *countingStorage) DeleteEnvironment(ctx context.Context, id string) error {
	s.count("DeleteEnvironment")
	delete(s.envs, id)

	return nil
}

func (s *countingStorage) SaveConfig(ctx context.Context, cfg *valobj.Config) error {
	s.count("SaveConfig")

	if cfg == nil {
		return &ErrNilConfig{}
	}

	s.cfg = cfg

	return nil
}

func (s *countingStorage) LoadConfig(ctx context.Context) (*valobj.Config, error) {
	s.count("LoadConfig")
	return s.cfg, nil
}

func TestCachedStorageLoadConfig(t *testing.T) {
	ctx := context.Background()
	inner := newCountingStorage()
	inner.cfg = &valobj.Config{ActiveEnvID: "env-1", DefaultMethod: "GET"}

	cache := NewCachedStorage(inner)

	first, err := cache.LoadConfig(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "env-1", first.ActiveEnvID)

	second, err := cache.LoadConfig(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "env-1", second.ActiveEnvID)

	assert.Equal(t, 1, inner.calls("LoadConfig"), "second load should be served from cache")

	// mutating a returned config must not poison the cache
	second.ActiveEnvID = "mutated"

	third, err := cache.LoadConfig(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "env-1", third.ActiveEnvID)
}

func TestCachedStorageLoadConfigMissing(t *testing.T) {
	ctx := context.Background()
	inner := newCountingStorage()

	cache := NewCachedStorage(inner)

	cfg, err := cache.LoadConfig(ctx)
	assert.NoError(t, err)
	assert.Nil(t, cfg)

	cfg, err = cache.LoadConfig(ctx)
	assert.NoError(t, err)
	assert.Nil(t, cfg)

	assert.Equal(t, 1, inner.calls("LoadConfig"), "missing config should be cached as absent")
}

func TestCachedStorageSaveConfigWriteThrough(t *testing.T) {
	ctx := context.Background()
	inner := newCountingStorage()

	cache := NewCachedStorage(inner)

	assert.NoError(t, cache.SaveConfig(ctx, &valobj.Config{ActiveEnvID: "env-2"}))

	cfg, err := cache.LoadConfig(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "env-2", cfg.ActiveEnvID)

	assert.Equal(t, 1, inner.calls("SaveConfig"))
	assert.Equal(t, 0, inner.calls("LoadConfig"), "load after save should be served from cache")
}

func TestCachedStorageLoadEnvironment(t *testing.T) {
	ctx := context.Background()
	inner := newCountingStorage()
	inner.envs["env-1"] = &entity.Environment{
		ID:        "env-1",
		Name:      "staging",
		Variables: map[string]string{"host": "example.com"},
	}

	cache := NewCachedStorage(inner)

	first, err := cache.LoadEnvironment(ctx, "env-1")
	assert.NoError(t, err)
	assert.Equal(t, "example.com", first.Variables["host"])

	// mutating the returned variables must not poison the cache
	first.Variables["host"] = "mutated"

	second, err := cache.LoadEnvironment(ctx, "env-1")
	assert.NoError(t, err)
	assert.Equal(t, "example.com", second.Variables["host"])

	assert.Equal(t, 1, inner.calls("LoadEnvironment"))
}

func TestCachedStorageSaveEnvironmentInvalidates(t *testing.T) {
	ctx := context.Background()
	inner := newCountingStorage()
	inner.envs["env-1"] = &entity.Environment{
		ID:        "env-1",
		Name:      "staging",
		Variables: map[string]string{"host": "old.example.com"},
	}

	cache := NewCachedStorage(inner)

	_, err := cache.LoadEnvironment(ctx, "env-1")
	assert.NoError(t, err)

	_, err = cache.ListEnvironments(ctx)
	assert.NoError(t, err)

	err = cache.SaveEnvironment(ctx, &entity.Environment{
		ID:        "env-1",
		Name:      "staging",
		Variables: map[string]string{"host": "new.example.com"},
	})
	assert.NoError(t, err)

	env, err := cache.LoadEnvironment(ctx, "env-1")
	assert.NoError(t, err)
	assert.Equal(t, "new.example.com", env.Variables["host"], "load after save must reflect the write")
	assert.Equal(t, 1, inner.calls("LoadEnvironment"), "saved environment should be cached")

	_, err = cache.ListEnvironments(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 2, inner.calls("ListEnvironments"), "save should invalidate the environment list")
}

func TestCachedStorageDeleteEnvironmentInvalidates(t *testing.T) {
	ctx := context.Background()
	inner := newCountingStorage()
	inner.envs["env-1"] = &entity.Environment{ID: "env-1", Name: "staging"}

	cache := NewCachedStorage(inner)

	_, err := cache.LoadEnvironment(ctx, "env-1")
	assert.NoError(t, err)

	assert.NoError(t, cache.DeleteEnvironment(ctx, "env-1"))

	_, err = cache.LoadEnvironment(ctx, "env-1")
	assert.Error(t, err, "deleted environment must not be served from cache")
}

func TestCachedStorageListCollections(t *testing.T) {
	ctx := context.Background()
	inner := newCountingStorage()
	inner.collections = []entity.Collection{
		{
			ID:   "col-1",
			Name: "api",
			Requests: []entity.Request{
				{ID: "req-1", Name: "get users", URL: "https://example.com/users"},
			},
		},
	}

	cache := NewCachedStorage(inner)

	first, err := cache.ListCollections(ctx)
	assert.NoError(t, err)
	assert.Len(t, first, 1)

	// mutating a returned collection must not poison the cache
	first[0].Requests[0].Name = "mutated"

	second, err := cache.ListCollections(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "get users", second[0].Requests[0].Name)

	assert.Equal(t, 1, inner.calls("ListCollections"))

	assert.NoError(t, cache.SaveCollection(ctx, &entity.Collection{ID: "col-2", Name: "other"}))

	_, err = cache.ListCollections(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 2, inner.calls("ListCollections"), "save should invalidate the collection list")
}

func TestCachedStorageHistory(t *testing.T) {
	ctx := context.Background()
	inner := newCountingStorage()
	inner.history = []entity.HistoryEntry{
		{ID: "h1", Request: &entity.Request{URL: "https://example.com/1"}},
		{ID: "h2", Request: &entity.Request{URL: "https://example.com/2"}},
	}

	cache := NewCachedStorage(inner)

	first, err := cache.LoadHistory(ctx)
	assert.NoError(t, err)
	assert.Len(t, first, 2)

	// reordering the returned slice (History.Load reverses in place) must
	// not poison the cache
	first[0], first[1] = first[1], first[0]

	second, err := cache.LoadHistory(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "h1", second[0].ID)
	assert.Equal(t, 1, inner.calls("LoadHistory"))

	appended := append(second, entity.HistoryEntry{ID: "h3", Request: &entity.Request{URL: "https://example.com/3"}})
	assert.NoError(t, cache.SaveHistory(ctx, appended))

	third, err := cache.LoadHistory(ctx)
	assert.NoError(t, err)
	assert.Len(t, third, 3)
	assert.Equal(t, 1, inner.calls("LoadHistory"), "load after save should be served from cache")
	assert.Equal(t, 1, inner.calls("SaveHistory"))
}

// TestCachedStorageSendPath wires the real domain layers over the cache and
// asserts that repeated request sends resolve config and environment without
// touching the inner storage again
func TestCachedStorageSendPath(t *testing.T) {
	ctx := context.Background()
	inner := newCountingStorage()
	inner.cfg = &valobj.Config{ActiveEnvID: "env-1"}
	inner.envs["env-1"] = &entity.Environment{
		ID:        "env-1",
		Name:      "staging",
		Variables: map[string]string{"host": "example.com"},
	}

	cache := NewCachedStorage(inner)

	configs := domain.NewConfigs(service.NewConfigService(cache))
	environments := domain.NewEnvironments(service.NewEnvironmentService(cache), configs)

	for i := 0; i < 100; i++ {
		vars := environments.ActiveVars(ctx)
		assert.Equal(t, "example.com", vars["host"])
	}

	assert.LessOrEqual(t, inner.calls("LoadConfig"), 1)
	assert.LessOrEqual(t, inner.calls("LoadEnvironment"), 1)
}

func TestCachedStorageConcurrent(t *testing.T) {
	ctx := context.Background()
	inner := newCountingStorage()
	inner.cfg = &valobj.Config{ActiveEnvID: "env-1"}
	inner.envs["env-1"] = &entity.Environment{
		ID:        "env-1",
		Name:      "staging",
		Variables: map[string]string{"host": "example.com"},
	}
	inner.collections = []entity.Collection{{ID: "col-1", Name: "api"}}

	cache := NewCachedStorage(inner)

	var wg sync.WaitGroup

	for i := 0; i < 20; i++ {
		wg.Add(1)

		go func(n int) {
			defer wg.Done()

			for j := 0; j < 50; j++ {
				switch n % 5 {
				case 0:
					cfg, _ := cache.LoadConfig(ctx)
					if cfg != nil {
						cfg.ActiveEnvID = "mutated"
					}
				case 1:
					env, _ := cache.LoadEnvironment(ctx, "env-1")
					if env != nil {
						env.Variables["host"] = "mutated"
					}
				case 2:
					_, _ = cache.ListCollections(ctx)
					_, _ = cache.ListEnvironments(ctx)
				case 3:
					_ = cache.SaveConfig(ctx, &valobj.Config{ActiveEnvID: "env-1"})
				case 4:
					history, _ := cache.LoadHistory(ctx)
					_ = cache.SaveHistory(ctx, append(history, entity.HistoryEntry{ID: "h"}))
				}
			}
		}(i)
	}

	wg.Wait()

	cfg, err := cache.LoadConfig(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "env-1", cfg.ActiveEnvID)
}

func BenchmarkCachedStorageLoadConfig(b *testing.B) {
	ctx := context.Background()
	adapter := newBenchJSONAdapter(b)
	cache := NewCachedStorage(adapter)

	if err := cache.SaveConfig(ctx, &valobj.Config{ActiveEnvID: "env-1", DefaultMethod: "GET"}); err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()

	for b.Loop() {
		if _, err := cache.LoadConfig(ctx); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCachedStorageLoadEnvironment(b *testing.B) {
	ctx := context.Background()
	adapter := newBenchJSONAdapter(b)
	cache := NewCachedStorage(adapter)

	env := makeEnvironment("env-1", 20)
	if err := cache.SaveEnvironment(ctx, env); err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()

	for b.Loop() {
		if _, err := cache.LoadEnvironment(ctx, "env-1"); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCachedStorageLoadHistory(b *testing.B) {
	ctx := context.Background()
	adapter := newBenchJSONAdapter(b)
	cache := NewCachedStorage(adapter)

	if err := cache.SaveHistory(ctx, makeHistory(100)); err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()

	for b.Loop() {
		if _, err := cache.LoadHistory(ctx); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCachedStorageListCollections(b *testing.B) {
	ctx := context.Background()
	adapter := newBenchJSONAdapter(b)
	cache := NewCachedStorage(adapter)

	for i := 0; i < 10; i++ {
		if err := cache.SaveCollection(ctx, makeCollection(fmt.Sprintf("col-%d", i), 10)); err != nil {
			b.Fatal(err)
		}
	}

	b.ReportAllocs()

	for b.Loop() {
		if _, err := cache.ListCollections(ctx); err != nil {
			b.Fatal(err)
		}
	}
}
