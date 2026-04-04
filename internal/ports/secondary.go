package ports

import (
	"context"

	"github.com/nxdir-s/grl/internal/core/entity"
)

type Http interface {
	Send(ctx context.Context, req *entity.Request) (*entity.Response, error)
}

type Storage interface {
	LoadCollection(ctx context.Context, id string) (*entity.Collection, error)
	SaveCollection(ctx context.Context, collection *entity.Collection) error
	ListCollections(ctx context.Context) ([]entity.Collection, error)
	DeleteCollection(ctx context.Context, id string) error

	LoadHistory(ctx context.Context) ([]entity.HistoryEntry, error)
	SaveHistory(ctx context.Context, history []entity.HistoryEntry) error
}
