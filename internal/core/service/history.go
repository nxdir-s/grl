package service

import (
	"context"

	"github.com/nxdir-s/grl/internal/core/entity"
	"github.com/nxdir-s/grl/internal/ports"
)

type HistoryService struct {
	storage ports.Storage
}

func NewHistoryService(storage ports.Storage) *HistoryService {
	return &HistoryService{
		storage: storage,
	}
}

func (s *HistoryService) Get(ctx context.Context) ([]entity.HistoryEntry, error) {
	return s.storage.LoadHistory(ctx)
}

func (s *HistoryService) Save(ctx context.Context, history []entity.HistoryEntry) error {
	return s.storage.SaveHistory(ctx, history)
}
