package domain

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/nxdir-s/grl/internal/core/entity"
)

// fakeHistoryService is an in-memory ports.HistoryService that mimics the
// cached storage contract: Get returns a fresh copy each call
type fakeHistoryService struct {
	mu      sync.Mutex
	history []entity.HistoryEntry
}

func (s *fakeHistoryService) Get(ctx context.Context) ([]entity.HistoryEntry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	out := make([]entity.HistoryEntry, len(s.history))
	copy(out, s.history)

	return out, nil
}

func (s *fakeHistoryService) Save(ctx context.Context, history []entity.HistoryEntry) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.history = make([]entity.HistoryEntry, len(history))
	copy(s.history, history)

	return nil
}

func (s *fakeHistoryService) stored() []entity.HistoryEntry {
	s.mu.Lock()
	defer s.mu.Unlock()

	out := make([]entity.HistoryEntry, len(s.history))
	copy(out, s.history)

	return out
}

func makeEntries(n int, ascending bool) []entity.HistoryEntry {
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	entries := make([]entity.HistoryEntry, 0, n)
	for i := 0; i < n; i++ {
		idx := i
		if !ascending {
			idx = n - 1 - i
		}

		entries = append(entries, entity.HistoryEntry{
			ID:        fmt.Sprintf("entry-%d", idx),
			Request:   &entity.Request{URL: fmt.Sprintf("https://example.com/%d", idx)},
			Timestamp: base.Add(time.Duration(idx) * time.Minute),
		})
	}

	return entries
}

func assertAscending(t *testing.T, entries []entity.HistoryEntry) {
	t.Helper()

	for i := 1; i < len(entries); i++ {
		assert.False(t, entries[i].Timestamp.Before(entries[i-1].Timestamp),
			"stored history must be ascending by timestamp")
	}
}

func TestHistoryLoadNewestFirst(t *testing.T) {
	ctx := context.Background()
	service := &fakeHistoryService{history: makeEntries(5, true)}
	history := NewHistory(service)

	out, err := history.Load(ctx, 0)
	assert.NoError(t, err)
	assert.Len(t, out, 5)
	assert.Equal(t, "entry-4", out[0].ID)
	assert.Equal(t, "entry-0", out[4].ID)

	assertAscending(t, service.stored())
}

func TestHistoryLoadLimit(t *testing.T) {
	ctx := context.Background()
	service := &fakeHistoryService{history: makeEntries(10, true)}
	history := NewHistory(service)

	out, err := history.Load(ctx, 3)
	assert.NoError(t, err)
	assert.Len(t, out, 3)
	assert.Equal(t, "entry-9", out[0].ID)
}

func TestHistoryDeleteEntryPreservesOrder(t *testing.T) {
	ctx := context.Background()
	service := &fakeHistoryService{history: makeEntries(5, true)}
	history := NewHistory(service)

	assert.NoError(t, history.DeleteEntry(ctx, "entry-2"))

	stored := service.stored()
	assert.Len(t, stored, 4)
	assertAscending(t, stored)

	assert.NoError(t, history.DeleteEntry(ctx, "entry-0"))

	stored = service.stored()
	assert.Len(t, stored, 3)
	assertAscending(t, stored)
	assert.Equal(t, "entry-1", stored[0].ID)
	assert.Equal(t, "entry-4", stored[2].ID)
}

// TestHistoryHealsReversedFile seeds storage in descending order, as written
// by older versions after a delete, and verifies one operation restores the
// canonical ascending order
func TestHistoryHealsReversedFile(t *testing.T) {
	ctx := context.Background()
	service := &fakeHistoryService{history: makeEntries(5, false)}
	history := NewHistory(service)

	assert.NoError(t, history.Append(ctx, &entity.Request{URL: "https://example.com/new"}, nil))

	stored := service.stored()
	assert.Len(t, stored, 6)
	assertAscending(t, stored)

	out, err := history.Load(ctx, 0)
	assert.NoError(t, err)
	assert.Equal(t, "https://example.com/new", out[0].Request.URL, "newest entry must come first")
}

func TestHistoryDeleteLastEntry(t *testing.T) {
	ctx := context.Background()
	service := &fakeHistoryService{history: makeEntries(1, true)}
	history := NewHistory(service)

	assert.NoError(t, history.DeleteEntry(ctx, "entry-0"))

	assert.Empty(t, service.stored())

	out, err := history.Load(ctx, 0)
	assert.NoError(t, err)
	assert.Empty(t, out)
}

func TestHistoryAppendCap(t *testing.T) {
	ctx := context.Background()
	service := &fakeHistoryService{history: makeEntries(HistoryCap, true)}
	history := NewHistory(service)

	assert.NoError(t, history.Append(ctx, &entity.Request{URL: "https://example.com/new"}, nil))

	stored := service.stored()
	assert.Len(t, stored, HistoryCap)
	assert.Equal(t, "https://example.com/new", stored[len(stored)-1].Request.URL)
	assert.Equal(t, "entry-1", stored[0].ID, "oldest entry must be evicted")
}

func TestHistoryAppendConcurrent(t *testing.T) {
	ctx := context.Background()
	service := &fakeHistoryService{}
	history := NewHistory(service)

	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(1)

		go func(n int) {
			defer wg.Done()

			err := history.Append(ctx, &entity.Request{URL: fmt.Sprintf("https://example.com/%d", n)}, nil)
			assert.NoError(t, err)
		}(i)
	}

	wg.Wait()

	assert.Len(t, service.stored(), 50, "concurrent appends must not lose entries")
}
