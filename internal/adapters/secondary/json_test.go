package secondary

import (
	"context"
	_ "embed"
	"log/slog"
	"os"
	"strconv"
	"testing"

	"github.com/nxdir-s/grl/internal/core/entity"
)

func TestSaveCollection(t *testing.T) {
	cases := []struct {
		opts            []JSONOpt
		collection      *entity.Collection
		baseDir         string
		collectionsDir  string
		environmentsDir string
		expectedErr     error
	}{
		{
			opts:        []JSONOpt{},
			collection:  &entity.Collection{},
			expectedErr: nil,
		},
	}

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	for i, tt := range cases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			adapter := NewJSONAdapter(logger, tt.baseDir, tt.collectionsDir, tt.environmentsDir, tt.opts...)
			_ = adapter
			_ = ctx

			// err := adapter.SaveCollection(ctx, tt.collection)

			// assert.Equal(t, tt.expectedErr, err)
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
