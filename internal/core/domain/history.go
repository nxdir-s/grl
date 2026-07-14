package domain

import (
	"context"
	"slices"
	"sync"

	"github.com/nxdir-s/grl/internal/core/entity"
	"github.com/nxdir-s/grl/internal/ports"
)

const HistoryCap int = 100

type History struct {
	service ports.HistoryService

	mu sync.Mutex
}

func NewHistory(service ports.HistoryService) *History {
	return &History{
		service: service,
	}
}

// canonicalize sorts entries into the canonical on-disk order, ascending by
// timestamp, healing files persisted in display order by older versions
func canonicalize(history []entity.HistoryEntry) {
	sorted := slices.IsSortedFunc(history, func(a, b entity.HistoryEntry) int {
		return a.Timestamp.Compare(b.Timestamp)
	})

	if !sorted {
		slices.SortStableFunc(history, func(a, b entity.HistoryEntry) int {
			return a.Timestamp.Compare(b.Timestamp)
		})
	}
}

// Load returns entries newest-first without mutating the stored order
func (d *History) Load(ctx context.Context, limit int) ([]entity.HistoryEntry, error) {
	history, err := d.service.Get(ctx)
	if err != nil {
		return nil, err
	}

	canonicalize(history)

	out := make([]entity.HistoryEntry, 0, len(history))
	for i := len(history) - 1; i >= 0; i-- {
		out = append(out, history[i])
	}

	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}

	return out, nil
}

func (d *History) Append(ctx context.Context, req *entity.Request, resp *entity.Response) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	history, err := d.service.Get(ctx)
	if err != nil {
		return err
	}

	canonicalize(history)

	history = append(history, entity.NewHistoryEntry(req, resp))

	if len(history) > HistoryCap {
		history = history[len(history)-HistoryCap:]
	}

	if err := d.service.Save(ctx, history); err != nil {
		return err
	}

	return nil
}

func (d *History) DeleteEntry(ctx context.Context, id string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	history, err := d.service.Get(ctx)
	if err != nil {
		return err
	}

	canonicalize(history)

	out := make([]entity.HistoryEntry, 0, len(history))

	for i := range history {
		if history[i].ID != id {
			out = append(out, history[i])
		}
	}

	if err := d.service.Save(ctx, out); err != nil {
		return err
	}

	return nil
}
