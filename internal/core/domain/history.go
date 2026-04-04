package domain

import (
	"context"

	"github.com/nxdir-s/grl/internal/core/entity"
	"github.com/nxdir-s/grl/internal/ports"
)

type History struct {
	service ports.HistoryService
}

func NewHistory(service ports.HistoryService) *History {
	return &History{
		service: service,
	}
}

func (d *History) Load(ctx context.Context, limit int) ([]entity.HistoryEntry, error) {
	history, err := d.service.Get(ctx)
	if err != nil {
		return nil, err
	}

	for i, j := 0, len(history)-1; i < j; i, j = i+1, j-1 {
		history[i], history[j] = history[j], history[i]
	}

	if limit > 0 && len(history) > limit {
		history = history[:limit]
	}

	return history, nil
}

func (d *History) Append(ctx context.Context, req *entity.Request, resp *entity.Response) error {
	history, err := d.service.Get(ctx)
	if err != nil {
		return err
	}

	history = append(history, *entity.NewHistoryEntry(req, resp))

	if len(history) > 100 {
		history = history[len(history)-100:]
	}

	if err := d.service.Save(ctx, history); err != nil {
		return err
	}

	return nil
}
