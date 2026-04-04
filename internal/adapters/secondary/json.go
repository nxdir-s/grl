package secondary

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"

	"github.com/nxdir-s/grl/internal/core/entity"
)

const (
	JSONFileExt         string = ".json"
	JSONHistoryFileName string = "history"
)

type ErrSaveCollection struct {
	err error
}

func (e *ErrSaveCollection) Error() string {
	return "failed to save collection: " + e.err.Error()
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

type JSONOpt func(a *JSONAdapter)

func WithJSONHistoryFile(filePath string) JSONOpt {
	return func(a *JSONAdapter) {
		a.historyFile = filePath
	}
}

type JSONAdapter struct {
	baseDir        string
	collectionsDir string
	historyFile    string
}

func NewJSONAdapter(baseDir string, collectionsDir string, opts ...JSONOpt) *JSONAdapter {
	// collectionsDir := filepath.Join(baseDir, "collections")
	// if err := os.MkdirAll(collectionsDir, 0o755); err != nil {
	// 	return nil, err
	// }

	adapter := &JSONAdapter{
		baseDir:        baseDir,
		collectionsDir: collectionsDir,
		historyFile:    filepath.Join(baseDir, JSONHistoryFileName+JSONFileExt),
	}

	for _, opt := range opts {
		opt(adapter)
	}

	return adapter
}

func (a *JSONAdapter) SaveCollection(ctx context.Context, collection *entity.Collection) error {
	data, err := json.MarshalIndent(collection, "", "  ")
	if err != nil {
		return &ErrSaveCollection{err}
	}

	if err := os.WriteFile(a.collectionFile(collection.ID), data, 0o644); err != nil {
		return &ErrSaveCollection{err}
	}

	return nil
}

func (a *JSONAdapter) LoadCollection(ctx context.Context, id string) (*entity.Collection, error) {
	data, err := os.ReadFile(a.collectionFile(id))
	if err != nil {
		return nil, &ErrLoadCollection{err}
	}

	var collection entity.Collection
	if err := json.Unmarshal(data, &collection); err != nil {
		return nil, &ErrLoadCollection{err}
	}

	return &collection, nil
}

func (a *JSONAdapter) ListCollections(ctx context.Context) ([]entity.Collection, error) {
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

	return collections, nil
}

func (a *JSONAdapter) DeleteCollection(ctx context.Context, id string) error {
	if err := os.Remove(filepath.Join(a.collectionsDir, id+JSONFileExt)); err != nil {
		return &ErrDeleteCollection{err}
	}

	return nil
}

func (a *JSONAdapter) SaveHistory(ctx context.Context, history []entity.HistoryEntry) error {
	data, err := json.MarshalIndent(history, "", "  ")
	if err != nil {
		return &ErrSaveHistory{err}
	}

	if err := os.WriteFile(a.historyFile, data, 0o644); err != nil {
		return &ErrSaveHistory{err}
	}

	return nil
}

func (a *JSONAdapter) LoadHistory(ctx context.Context) ([]entity.HistoryEntry, error) {
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

	return history, nil
}

func (a *JSONAdapter) collectionFile(id string) string {
	return filepath.Join(a.collectionsDir, id+JSONFileExt)
}
