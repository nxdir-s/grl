package ports

import (
	"context"

	"github.com/nxdir-s/grl/internal/core/entity"
)

type CLI interface {
	SendRequest(ctx context.Context, req *entity.Request) (*entity.Response, error)
	RecordHistory(ctx context.Context, req *entity.Request, resp *entity.Response) error
	GetHistory(ctx context.Context, limit int) ([]entity.HistoryEntry, error)
	ListCollections(ctx context.Context) ([]entity.Collection, error)
}
