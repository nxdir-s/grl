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

	err = &ErrHistory{}
	if len(err.Error()) == 0 {
		t.Error("missing error message for ErrHistory")
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
