package secondary

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nxdir-s/grl/internal/core/entity"
	"github.com/nxdir-s/grl/internal/core/valobj"
	"github.com/stretchr/testify/assert"
)

// newBenchJSONAdapter returns an adapter backed by a temp directory unique to the benchmark
func newBenchJSONAdapter(b *testing.B) *JSONAdapter {
	dataDir := b.TempDir()

	collectionsDir := filepath.Join(dataDir, "collections")
	if err := os.MkdirAll(collectionsDir, 0o755); err != nil {
		b.Fatalf("failed to create collections directory: %s", err.Error())
	}

	environmentsDir := filepath.Join(dataDir, "environments")
	if err := os.MkdirAll(environmentsDir, 0o755); err != nil {
		b.Fatalf("failed to create environments directory: %s", err.Error())
	}

	return NewJSONAdapter(benchLogger(), dataDir, collectionsDir, environmentsDir)
}

// makeRequest returns a populated request for benchmark fixtures
func makeRequest(name string) *entity.Request {
	return &entity.Request{
		ID:     uuid.New().String(),
		Name:   name,
		Method: valobj.MethodPost,
		URL:    TestURL,
		Headers: []valobj.Header{
			{Key: ContentTypeHeader, Value: BenchContentTypeJSON, Enabled: true},
			{Key: "Accept", Value: BenchContentTypeJSON, Enabled: true},
		},
		Body: `{"key":"value"}`,
	}
}

// makeCollection returns a collection with numRequests populated requests
func makeCollection(id string, numRequests int) *entity.Collection {
	collection := &entity.Collection{
		ID:       id,
		Name:     "bench" + id,
		Requests: make([]entity.Request, 0, numRequests),
	}

	for i := 0; i < numRequests; i++ {
		collection.Requests = append(collection.Requests, *makeRequest("request" + strconv.Itoa(i)))
	}

	return collection
}

// makeHistory returns numEntries populated history entries
func makeHistory(numEntries int) []entity.HistoryEntry {
	timestamp := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)

	history := make([]entity.HistoryEntry, 0, numEntries)

	for i := 0; i < numEntries; i++ {
		history = append(history, entity.HistoryEntry{
			ID:        uuid.New().String(),
			Request:   makeRequest("request" + strconv.Itoa(i)),
			Timestamp: timestamp,
		})
	}

	return history
}

// makeEnvironment returns an environment with numVars variables
func makeEnvironment(id string, numVars int) *entity.Environment {
	env := &entity.Environment{
		ID:        id,
		Name:      "bench" + id,
		Variables: make(map[string]string, numVars),
	}

	for i := 0; i < numVars; i++ {
		env.Variables["var"+strconv.Itoa(i)] = "value" + strconv.Itoa(i)
	}

	return env
}

func BenchmarkJSONAdapterSaveCollection(b *testing.B) {
	for _, numRequests := range []int{1, 10, 100} {
		b.Run("requests="+strconv.Itoa(numRequests), func(b *testing.B) {
			adapter := newBenchJSONAdapter(b)
			collection := makeCollection(uuid.New().String(), numRequests)

			b.ReportAllocs()

			for b.Loop() {
				if err := adapter.SaveCollection(b.Context(), collection); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkJSONAdapterLoadCollection(b *testing.B) {
	for _, numRequests := range []int{1, 10, 100} {
		b.Run("requests="+strconv.Itoa(numRequests), func(b *testing.B) {
			adapter := newBenchJSONAdapter(b)

			id := uuid.New().String()
			if err := adapter.SaveCollection(b.Context(), makeCollection(id, numRequests)); err != nil {
				b.Fatal(err)
			}

			info, err := os.Stat(adapter.collectionFile(id))
			if err != nil {
				b.Fatal(err)
			}

			b.ReportAllocs()
			b.SetBytes(info.Size())

			for b.Loop() {
				if _, err := adapter.LoadCollection(b.Context(), id); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkJSONAdapterListCollections(b *testing.B) {
	for _, numCollections := range []int{1, 10, 50} {
		b.Run("collections="+strconv.Itoa(numCollections), func(b *testing.B) {
			adapter := newBenchJSONAdapter(b)

			for i := 0; i < numCollections; i++ {
				if err := adapter.SaveCollection(b.Context(), makeCollection(uuid.New().String(), 10)); err != nil {
					b.Fatal(err)
				}
			}

			collections, err := adapter.ListCollections(b.Context())
			if err != nil {
				b.Fatal(err)
			}

			if len(collections) != numCollections {
				b.Fatalf("expected %d collections, found %d", numCollections, len(collections))
			}

			b.ReportAllocs()

			for b.Loop() {
				if _, err := adapter.ListCollections(b.Context()); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkJSONAdapterSaveHistory(b *testing.B) {
	for _, numEntries := range []int{10, 100, 1000} {
		b.Run("entries="+strconv.Itoa(numEntries), func(b *testing.B) {
			adapter := newBenchJSONAdapter(b)
			history := makeHistory(numEntries)

			b.ReportAllocs()

			for b.Loop() {
				if err := adapter.SaveHistory(b.Context(), history); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkJSONAdapterLoadHistory(b *testing.B) {
	for _, numEntries := range []int{10, 100, 1000} {
		b.Run("entries="+strconv.Itoa(numEntries), func(b *testing.B) {
			adapter := newBenchJSONAdapter(b)

			if err := adapter.SaveHistory(b.Context(), makeHistory(numEntries)); err != nil {
				b.Fatal(err)
			}

			info, err := os.Stat(adapter.historyFile)
			if err != nil {
				b.Fatal(err)
			}

			b.ReportAllocs()
			b.SetBytes(info.Size())

			for b.Loop() {
				if _, err := adapter.LoadHistory(b.Context()); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkJSONAdapterEnvironment(b *testing.B) {
	b.Run("save", func(b *testing.B) {
		adapter := newBenchJSONAdapter(b)
		env := makeEnvironment(uuid.New().String(), 20)

		b.ReportAllocs()

		for b.Loop() {
			if err := adapter.SaveEnvironment(b.Context(), env); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("load", func(b *testing.B) {
		adapter := newBenchJSONAdapter(b)

		env := makeEnvironment(uuid.New().String(), 20)
		if err := adapter.SaveEnvironment(b.Context(), env); err != nil {
			b.Fatal(err)
		}

		b.ReportAllocs()

		for b.Loop() {
			if _, err := adapter.LoadEnvironment(b.Context(), env.ID); err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkJSONAdapterConfig(b *testing.B) {
	cfg := &valobj.Config{
		ActiveEnvID:     uuid.New().String(),
		DefaultMethod:   valobj.MethodGet.String(),
		TimeoutSeconds:  DefaultTimeout,
		FollowRedirects: true,
		HistoryLimit:    100,
	}

	b.Run("save", func(b *testing.B) {
		adapter := newBenchJSONAdapter(b)

		b.ReportAllocs()

		for b.Loop() {
			if err := adapter.SaveConfig(b.Context(), cfg); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("load", func(b *testing.B) {
		adapter := newBenchJSONAdapter(b)

		if err := adapter.SaveConfig(b.Context(), cfg); err != nil {
			b.Fatal(err)
		}

		b.ReportAllocs()

		for b.Loop() {
			if _, err := adapter.LoadConfig(b.Context()); err != nil {
				b.Fatal(err)
			}
		}
	})
}

const (
	TestCollectionName  string = "unittest"
	TestEnvironmentName string = "testing"
	TestURL             string = "https://example.com"
)

func TestSaveCollection(t *testing.T) {
	cases := []struct {
		opts        []JSONOpt
		collection  *entity.Collection
		expectedErr error
	}{
		{
			opts: []JSONOpt{
				WithJSONConfigFile(filepath.Join(os.TempDir(), JSONConfigFileName+JSONFileExt)),
				WithJSONHistoryFile(filepath.Join(os.TempDir(), JSONHistoryFileName+JSONFileExt)),
			},
			collection: &entity.Collection{
				ID:       uuid.New().String(),
				Name:     TestCollectionName,
				Requests: make([]entity.Request, 0),
			},
			expectedErr: nil,
		},
	}

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	tmpDir := os.TempDir()

	dataDir := filepath.Join(tmpDir, "grl")
	defer func() {
		if err := os.RemoveAll(dataDir); err != nil {
			fmt.Fprintf(os.Stdout, "failed to remove directory: %s\n", dataDir)
		}
	}()

	collectionsDir := filepath.Join(dataDir, "collections")
	if err := os.MkdirAll(collectionsDir, 0o755); err != nil {
		t.Errorf("failed to create collections directory: %s", err.Error())
	}

	environmentsDir := filepath.Join(dataDir, "environments")
	if err := os.MkdirAll(environmentsDir, 0o755); err != nil {
		t.Errorf("failed to create environments directory: %s", err.Error())
	}

	for i, tt := range cases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			adapter := NewJSONAdapter(logger, dataDir, collectionsDir, environmentsDir, tt.opts...)

			err := adapter.SaveCollection(ctx, tt.collection)

			assert.Equal(t, tt.expectedErr, err)
		})
	}
}

func TestLoadCollection(t *testing.T) {
	cases := []struct {
		opts        []JSONOpt
		ID          string
		expectedErr error
	}{
		{
			opts:        []JSONOpt{},
			ID:          uuid.New().String()[:8],
			expectedErr: nil,
		},
	}

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	tmpDir := os.TempDir()

	dataDir := filepath.Join(tmpDir, "grl")
	defer func() {
		if err := os.RemoveAll(dataDir); err != nil {
			fmt.Fprintf(os.Stdout, "failed to remove directory: %s\n", dataDir)
		}
	}()

	collectionsDir := filepath.Join(dataDir, "collections")
	if err := os.MkdirAll(collectionsDir, 0o755); err != nil {
		t.Errorf("failed to create collections directory: %s", err.Error())
	}

	environmentsDir := filepath.Join(dataDir, "environments")
	if err := os.MkdirAll(environmentsDir, 0o755); err != nil {
		t.Errorf("failed to create environments directory: %s", err.Error())
	}

	for i, tt := range cases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			adapter := NewJSONAdapter(logger, dataDir, collectionsDir, environmentsDir, tt.opts...)

			in := &entity.Collection{
				ID:       tt.ID,
				Name:     TestCollectionName + tt.ID,
				Requests: make([]entity.Request, 0),
			}

			err := adapter.SaveCollection(ctx, in)
			assert.Equal(t, tt.expectedErr, err)

			out, err := adapter.LoadCollection(ctx, tt.ID)

			assert.Equal(t, tt.expectedErr, err)
			if err == nil {
				assert.True(t, reflect.DeepEqual(in, out))
			}

			if err := os.Remove(filepath.Join(collectionsDir, in.ID+JSONFileExt)); err != nil {
				t.Errorf("failed to delete test collection: %s", err.Error())
			}
		})
	}
}

func TestListCollections(t *testing.T) {
	cases := []struct {
		opts           []JSONOpt
		numCollections int
		expectedErr    error
	}{
		{
			opts:           []JSONOpt{},
			numCollections: 0,
			expectedErr:    nil,
		},
		{
			opts:           []JSONOpt{},
			numCollections: 1,
			expectedErr:    nil,
		},
		{
			opts:           []JSONOpt{},
			numCollections: 5,
			expectedErr:    nil,
		},
		{
			opts:           []JSONOpt{},
			numCollections: 10,
			expectedErr:    nil,
		},
	}

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	tmpDir := os.TempDir()

	dataDir := filepath.Join(tmpDir, "grl")
	defer func() {
		if err := os.RemoveAll(dataDir); err != nil {
			fmt.Fprintf(os.Stdout, "failed to remove directory: %s\n", dataDir)
		}
	}()

	collectionsDir := filepath.Join(dataDir, "collections")
	if err := os.MkdirAll(collectionsDir, 0o755); err != nil {
		t.Errorf("failed to create collections directory: %s", err.Error())
	}

	environmentsDir := filepath.Join(dataDir, "environments")
	if err := os.MkdirAll(environmentsDir, 0o755); err != nil {
		t.Errorf("failed to create environments directory: %s", err.Error())
	}

	for i, tt := range cases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			adapter := NewJSONAdapter(logger, dataDir, collectionsDir, environmentsDir, tt.opts...)

			collections := make(map[string]*entity.Collection)
			for n := range tt.numCollections {
				collection := &entity.Collection{
					ID:       uuid.New().String(),
					Name:     TestCollectionName + strconv.Itoa(n),
					Requests: make([]entity.Request, 0),
				}

				err := adapter.SaveCollection(ctx, collection)
				assert.Equal(t, tt.expectedErr, err)

				collections[collection.ID] = collection
			}

			out, err := adapter.ListCollections(ctx)

			assert.Equal(t, tt.expectedErr, err)
			if err == nil {
				assert.Equal(t, tt.numCollections, len(out))

				for n := range out {
					collection, ok := collections[out[n].ID]
					if !ok {
						t.Errorf("collection not found")
					}

					assert.True(t, reflect.DeepEqual(*collection, out[n]))
					if err := os.Remove(filepath.Join(collectionsDir, collection.ID+JSONFileExt)); err != nil {
						t.Errorf("failed to delete test collection: %s", err.Error())
					}
				}
			}
		})
	}
}

func TestDeleteCollection(t *testing.T) {
	cases := []struct {
		opts        []JSONOpt
		ID          string
		expectedErr error
	}{
		{
			opts:        []JSONOpt{},
			ID:          uuid.New().String()[:8],
			expectedErr: nil,
		},
	}

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	tmpDir := os.TempDir()

	dataDir := filepath.Join(tmpDir, "grl")
	defer func() {
		if err := os.RemoveAll(dataDir); err != nil {
			fmt.Fprintf(os.Stdout, "failed to remove directory: %s\n", dataDir)
		}
	}()

	collectionsDir := filepath.Join(dataDir, "collections")
	if err := os.MkdirAll(collectionsDir, 0o755); err != nil {
		t.Errorf("failed to create collections directory: %s", err.Error())
	}

	environmentsDir := filepath.Join(dataDir, "environments")
	if err := os.MkdirAll(environmentsDir, 0o755); err != nil {
		t.Errorf("failed to create environments directory: %s", err.Error())
	}

	for i, tt := range cases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			adapter := NewJSONAdapter(logger, dataDir, collectionsDir, environmentsDir, tt.opts...)

			in := &entity.Collection{
				ID:       tt.ID,
				Name:     TestCollectionName + tt.ID,
				Requests: make([]entity.Request, 0),
			}

			err := adapter.SaveCollection(ctx, in)
			assert.Equal(t, tt.expectedErr, err)

			err = adapter.DeleteCollection(ctx, tt.ID)
			assert.Equal(t, tt.expectedErr, err)

			_, err = os.Stat(adapter.collectionFile(in.ID))
			assert.True(t, errors.Is(err, os.ErrNotExist))
		})
	}
}

func TestSaveHistory(t *testing.T) {
	cases := []struct {
		opts        []JSONOpt
		history     []entity.HistoryEntry
		expectedErr error
	}{
		{
			opts: []JSONOpt{},
			history: []entity.HistoryEntry{
				{
					ID:        uuid.New().String(),
					Request:   entity.NewRequest(),
					Timestamp: time.Now(),
				},
			},
			expectedErr: nil,
		},
	}

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	tmpDir := os.TempDir()

	dataDir := filepath.Join(tmpDir, "grl")
	defer func() {
		if err := os.RemoveAll(dataDir); err != nil {
			fmt.Fprintf(os.Stdout, "failed to remove directory: %s\n", dataDir)
		}
	}()

	collectionsDir := filepath.Join(dataDir, "collections")
	if err := os.MkdirAll(collectionsDir, 0o755); err != nil {
		t.Errorf("failed to create collections directory: %s", err.Error())
	}

	environmentsDir := filepath.Join(dataDir, "environments")
	if err := os.MkdirAll(environmentsDir, 0o755); err != nil {
		t.Errorf("failed to create environments directory: %s", err.Error())
	}

	for i, tt := range cases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			adapter := NewJSONAdapter(logger, dataDir, collectionsDir, environmentsDir, tt.opts...)

			err := adapter.SaveHistory(ctx, tt.history)

			assert.Equal(t, tt.expectedErr, err)
		})
	}
}

func TestSaveHistoryEmpty(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	dataDir := t.TempDir()
	adapter := NewJSONAdapter(logger, dataDir, dataDir, dataDir)

	err := adapter.SaveHistory(ctx, []entity.HistoryEntry{})
	assert.NoError(t, err, "deleting the last history entry must be able to persist an empty file")

	out, err := adapter.LoadHistory(ctx)
	assert.NoError(t, err)
	assert.Empty(t, out)
}

func TestLoadHistory(t *testing.T) {
	cases := []struct {
		opts        []JSONOpt
		history     []entity.HistoryEntry
		expectedErr error
	}{
		{
			opts: []JSONOpt{},
			history: []entity.HistoryEntry{
				{
					ID:        uuid.New().String(),
					Request:   entity.NewRequest(),
					Timestamp: time.Now(),
				},
			},
			expectedErr: nil,
		},
	}

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	tmpDir := os.TempDir()

	dataDir := filepath.Join(tmpDir, "grl")
	defer func() {
		if err := os.RemoveAll(dataDir); err != nil {
			fmt.Fprintf(os.Stdout, "failed to remove directory: %s\n", dataDir)
		}
	}()

	collectionsDir := filepath.Join(dataDir, "collections")
	if err := os.MkdirAll(collectionsDir, 0o755); err != nil {
		t.Errorf("failed to create collections directory: %s", err.Error())
	}

	environmentsDir := filepath.Join(dataDir, "environments")
	if err := os.MkdirAll(environmentsDir, 0o755); err != nil {
		t.Errorf("failed to create environments directory: %s", err.Error())
	}

	for i, tt := range cases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			adapter := NewJSONAdapter(logger, dataDir, collectionsDir, environmentsDir, tt.opts...)

			err := adapter.SaveHistory(ctx, tt.history)
			assert.Equal(t, tt.expectedErr, err)

			out, err := adapter.LoadHistory(ctx)

			assert.Equal(t, tt.expectedErr, err)
			if err == nil {
				assert.Equal(t, len(tt.history), len(out))
			}

			if err := os.Remove(adapter.historyFile); err != nil {
				t.Errorf("failed to delete test history: %s", err.Error())
			}
		})
	}
}

func TestSaveEnvironment(t *testing.T) {
	cases := []struct {
		opts        []JSONOpt
		env         *entity.Environment
		expectedErr error
	}{
		{
			opts: []JSONOpt{},
			env: &entity.Environment{
				ID:   uuid.New().String(),
				Name: TestEnvironmentName + uuid.New().String()[:8],
				Variables: map[string]string{
					"url": TestURL,
				},
			},
			expectedErr: nil,
		},
	}

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	tmpDir := os.TempDir()

	dataDir := filepath.Join(tmpDir, "grl")
	defer func() {
		if err := os.RemoveAll(dataDir); err != nil {
			fmt.Fprintf(os.Stdout, "failed to remove directory: %s\n", dataDir)
		}
	}()

	collectionsDir := filepath.Join(dataDir, "collections")
	if err := os.MkdirAll(collectionsDir, 0o755); err != nil {
		t.Errorf("failed to create collections directory: %s", err.Error())
	}

	environmentsDir := filepath.Join(dataDir, "environments")
	if err := os.MkdirAll(environmentsDir, 0o755); err != nil {
		t.Errorf("failed to create environments directory: %s", err.Error())
	}

	for i, tt := range cases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			adapter := NewJSONAdapter(logger, dataDir, collectionsDir, environmentsDir, tt.opts...)

			err := adapter.SaveEnvironment(ctx, tt.env)

			assert.Equal(t, tt.expectedErr, err)

			if err := os.Remove(filepath.Join(adapter.environmentsDir, tt.env.ID+".json")); err != nil {
				t.Errorf("failed to delete test environment: %s", err.Error())
			}
		})
	}
}

func TestLoadEnvironment(t *testing.T) {
	cases := []struct {
		opts        []JSONOpt
		env         *entity.Environment
		expectedErr error
	}{
		{
			opts: []JSONOpt{},
			env: &entity.Environment{
				ID:   uuid.New().String(),
				Name: TestEnvironmentName + uuid.New().String()[:8],
				Variables: map[string]string{
					"url": TestURL,
				},
			},
			expectedErr: nil,
		},
	}

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	tmpDir := os.TempDir()

	dataDir := filepath.Join(tmpDir, "grl")
	defer func() {
		if err := os.RemoveAll(dataDir); err != nil {
			fmt.Fprintf(os.Stdout, "failed to remove directory: %s\n", dataDir)
		}
	}()

	collectionsDir := filepath.Join(dataDir, "collections")
	if err := os.MkdirAll(collectionsDir, 0o755); err != nil {
		t.Errorf("failed to create collections directory: %s", err.Error())
	}

	environmentsDir := filepath.Join(dataDir, "environments")
	if err := os.MkdirAll(environmentsDir, 0o755); err != nil {
		t.Errorf("failed to create environments directory: %s", err.Error())
	}

	for i, tt := range cases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			adapter := NewJSONAdapter(logger, dataDir, collectionsDir, environmentsDir, tt.opts...)

			err := adapter.SaveEnvironment(ctx, tt.env)
			assert.Equal(t, tt.expectedErr, err)

			out, err := adapter.LoadEnvironment(ctx, tt.env.ID)

			assert.Equal(t, tt.expectedErr, err)
			if err == nil {
				assert.True(t, reflect.DeepEqual(tt.env, out))
			}

			if err := os.Remove(filepath.Join(adapter.environmentsDir, tt.env.ID+".json")); err != nil {
				t.Errorf("failed to delete test environment: %s", err.Error())
			}
		})
	}
}

func TestListEnvironments(t *testing.T) {
	cases := []struct {
		opts        []JSONOpt
		numEnv      int
		expectedErr error
	}{
		{
			opts:        []JSONOpt{},
			numEnv:      0,
			expectedErr: nil,
		},
		{
			opts:        []JSONOpt{},
			numEnv:      1,
			expectedErr: nil,
		},
		{
			opts:        []JSONOpt{},
			numEnv:      5,
			expectedErr: nil,
		},
		{
			opts:        []JSONOpt{},
			numEnv:      10,
			expectedErr: nil,
		},
	}

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	tmpDir := os.TempDir()

	dataDir := filepath.Join(tmpDir, "grl")
	defer func() {
		if err := os.RemoveAll(dataDir); err != nil {
			fmt.Fprintf(os.Stdout, "failed to remove directory: %s\n", dataDir)
		}
	}()

	collectionsDir := filepath.Join(dataDir, "collections")
	if err := os.MkdirAll(collectionsDir, 0o755); err != nil {
		t.Errorf("failed to create collections directory: %s", err.Error())
	}

	environmentsDir := filepath.Join(dataDir, "environments")
	if err := os.MkdirAll(environmentsDir, 0o755); err != nil {
		t.Errorf("failed to create environments directory: %s", err.Error())
	}

	for i, tt := range cases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			adapter := NewJSONAdapter(logger, dataDir, collectionsDir, environmentsDir, tt.opts...)

			envs := make(map[string]*entity.Environment)
			for n := range tt.numEnv {
				env := &entity.Environment{
					ID:   uuid.New().String(),
					Name: TestEnvironmentName + strconv.Itoa(n),
					Variables: map[string]string{
						"url": TestURL,
					},
				}

				err := adapter.SaveEnvironment(ctx, env)
				assert.Equal(t, tt.expectedErr, err)

				envs[env.ID] = env
			}

			out, err := adapter.ListEnvironments(ctx)

			assert.Equal(t, tt.expectedErr, err)
			if err == nil {
				assert.Equal(t, tt.numEnv, len(out))

				for n := range out {
					env, ok := envs[out[n].ID]
					if !ok {
						t.Errorf("environment not found")
					}

					assert.True(t, reflect.DeepEqual(*env, out[n]))
					if err := os.Remove(filepath.Join(adapter.environmentsDir, env.ID+".json")); err != nil {
						t.Errorf("failed to delete test environment: %s", err.Error())
					}
				}
			}
		})
	}
}

func TestDeleteEnvironment(t *testing.T) {
	cases := []struct {
		opts        []JSONOpt
		env         *entity.Environment
		expectedErr error
	}{
		{
			opts: []JSONOpt{},
			env: &entity.Environment{
				ID:   uuid.New().String(),
				Name: TestEnvironmentName + uuid.New().String()[:8],
				Variables: map[string]string{
					"url": TestURL,
				},
			},
			expectedErr: nil,
		},
	}

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	tmpDir := os.TempDir()

	dataDir := filepath.Join(tmpDir, "grl")
	defer func() {
		if err := os.RemoveAll(dataDir); err != nil {
			fmt.Fprintf(os.Stdout, "failed to remove directory: %s\n", dataDir)
		}
	}()

	collectionsDir := filepath.Join(dataDir, "collections")
	if err := os.MkdirAll(collectionsDir, 0o755); err != nil {
		t.Errorf("failed to create collections directory: %s", err.Error())
	}

	environmentsDir := filepath.Join(dataDir, "environments")
	if err := os.MkdirAll(environmentsDir, 0o755); err != nil {
		t.Errorf("failed to create environments directory: %s", err.Error())
	}

	for i, tt := range cases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			adapter := NewJSONAdapter(logger, dataDir, collectionsDir, environmentsDir, tt.opts...)

			err := adapter.SaveEnvironment(ctx, tt.env)
			assert.Equal(t, tt.expectedErr, err)

			err = adapter.DeleteEnvironment(ctx, tt.env.ID)
			assert.Equal(t, tt.expectedErr, err)

			_, err = os.Stat(filepath.Join(adapter.environmentsDir, tt.env.ID+".json"))
			assert.True(t, errors.Is(err, os.ErrNotExist))
		})
	}
}

func TestSaveConfig(t *testing.T) {
	cases := []struct {
		opts        []JSONOpt
		cfg         *valobj.Config
		expectedErr error
	}{
		{
			opts: []JSONOpt{},
			cfg: &valobj.Config{
				DefaultMethod:  valobj.MethodGet.String(),
				TimeoutSeconds: DefaultTimeout,
				HistoryLimit:   50,
			},
			expectedErr: nil,
		},
	}

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	tmpDir := os.TempDir()

	dataDir := filepath.Join(tmpDir, "grl")
	defer func() {
		if err := os.RemoveAll(dataDir); err != nil {
			fmt.Fprintf(os.Stdout, "failed to remove directory: %s\n", dataDir)
		}
	}()

	collectionsDir := filepath.Join(dataDir, "collections")
	if err := os.MkdirAll(collectionsDir, 0o755); err != nil {
		t.Errorf("failed to create collections directory: %s", err.Error())
	}

	environmentsDir := filepath.Join(dataDir, "environments")
	if err := os.MkdirAll(environmentsDir, 0o755); err != nil {
		t.Errorf("failed to create environments directory: %s", err.Error())
	}

	for i, tt := range cases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			adapter := NewJSONAdapter(logger, dataDir, collectionsDir, environmentsDir, tt.opts...)

			err := adapter.SaveConfig(ctx, tt.cfg)

			assert.Equal(t, tt.expectedErr, err)

			if err := os.Remove(adapter.configFile); err != nil {
				t.Errorf("failed to delete test config: %s", err.Error())
			}
		})
	}
}

func TestLoadConfig(t *testing.T) {
	cases := []struct {
		opts        []JSONOpt
		cfg         *valobj.Config
		expectedErr error
	}{
		{
			opts: []JSONOpt{},
			cfg: &valobj.Config{
				DefaultMethod:  valobj.MethodGet.String(),
				TimeoutSeconds: DefaultTimeout,
				HistoryLimit:   50,
			},
			expectedErr: nil,
		},
	}

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	tmpDir := os.TempDir()

	dataDir := filepath.Join(tmpDir, "grl")
	defer func() {
		if err := os.RemoveAll(dataDir); err != nil {
			fmt.Fprintf(os.Stdout, "failed to remove directory: %s\n", dataDir)
		}
	}()

	collectionsDir := filepath.Join(dataDir, "collections")
	if err := os.MkdirAll(collectionsDir, 0o755); err != nil {
		t.Errorf("failed to create collections directory: %s", err.Error())
	}

	environmentsDir := filepath.Join(dataDir, "environments")
	if err := os.MkdirAll(environmentsDir, 0o755); err != nil {
		t.Errorf("failed to create environments directory: %s", err.Error())
	}

	for i, tt := range cases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			adapter := NewJSONAdapter(logger, dataDir, collectionsDir, environmentsDir, tt.opts...)

			err := adapter.SaveConfig(ctx, tt.cfg)
			assert.Equal(t, tt.expectedErr, err)

			out, err := adapter.LoadConfig(ctx)

			assert.Equal(t, tt.expectedErr, err)
			if err == nil {
				assert.True(t, reflect.DeepEqual(tt.cfg, out))
			}

			if err := os.Remove(adapter.configFile); err != nil {
				t.Errorf("failed to delete test config: %s", err.Error())
			}
		})
	}
}

func TestJSONErrors(t *testing.T) {
	var err error

	err = &ErrSaveCollection{&ErrTest{}}
	if len(err.Error()) == 0 {
		t.Error("missing error message for ErrSaveCollection")
	}

	err = &ErrNilCollection{}
	if len(err.Error()) == 0 {
		t.Error("missing error message for ErrNilCollection")
	}

	err = &ErrCollectionID{}
	if len(err.Error()) == 0 {
		t.Error("missing error message for ErrCollectionID")
	}

	err = &ErrLoadCollection{&ErrTest{}}
	if len(err.Error()) == 0 {
		t.Error("missing error message for ErrLoadCollection")
	}

	err = &ErrListCollections{&ErrTest{}}
	if len(err.Error()) == 0 {
		t.Error("missing error message for ErrListCollections")
	}

	err = &ErrDeleteCollection{&ErrTest{}}
	if len(err.Error()) == 0 {
		t.Error("missing error message for ErrDeleteCollection")
	}

	err = &ErrSaveHistory{&ErrTest{}}
	if len(err.Error()) == 0 {
		t.Error("missing error message for ErrSaveHistory")
	}

	err = &ErrLoadHistory{&ErrTest{}}
	if len(err.Error()) == 0 {
		t.Error("missing error message for ErrLoadHistory")
	}

	err = &ErrNilEnvironment{}
	if len(err.Error()) == 0 {
		t.Error("missing error message for ErrNilEnvironment")
	}

	err = &ErrEnvironmentID{}
	if len(err.Error()) == 0 {
		t.Error("missing error message for ErrEnvironmentID")
	}

	err = &ErrSaveEnvironment{&ErrTest{}}
	if len(err.Error()) == 0 {
		t.Error("missing error message for ErrSaveEnvironment")
	}

	err = &ErrLoadEnvironment{&ErrTest{}}
	if len(err.Error()) == 0 {
		t.Error("missing error message for ErrLoadEnvironment")
	}

	err = &ErrListEnvironments{&ErrTest{}}
	if len(err.Error()) == 0 {
		t.Error("missing error message for ErrListEnvironments")
	}

	err = &ErrDeleteEnvironment{&ErrTest{}}
	if len(err.Error()) == 0 {
		t.Error("missing error message for ErrDeleteEnvironment")
	}

	err = &ErrNilConfig{}
	if len(err.Error()) == 0 {
		t.Error("missing error message for ErrNilConfig")
	}

	err = &ErrSaveConfig{&ErrTest{}}
	if len(err.Error()) == 0 {
		t.Error("missing error message for ErrSaveConfig")
	}

	err = &ErrLoadConfig{&ErrTest{}}
	if len(err.Error()) == 0 {
		t.Error("missing error message for ErrLoadConfig")
	}
}
