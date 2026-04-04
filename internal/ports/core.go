package ports

import (
	"context"

	"github.com/nxdir-s/grl/internal/core/entity"
	"github.com/nxdir-s/grl/internal/core/valobj"
)

type Requests interface {
	Send(ctx context.Context, req *entity.Request) (*entity.Response, error)
	Validate(req *entity.Request) error
	BuildURL(baseURL string, params []valobj.QueryParam) (string, error)
}

type RequestService interface {
	Send(ctx context.Context, req *entity.Request) (*entity.Response, error)
}

type Collections interface {
	Create(ctx context.Context, name string) (*entity.Collection, error)
	List(ctx context.Context) ([]entity.Collection, error)
	Delete(ctx context.Context, id string) error

	AddRequest(ctx context.Context, collectionID string, req *entity.Request) error
	RemoveRequest(ctx context.Context, collectionID string, requestID string) error
}

type CollectionService interface {
	Save(ctx context.Context, collection *entity.Collection) error
	Load(ctx context.Context, id string) (*entity.Collection, error)
	List(ctx context.Context) ([]entity.Collection, error)
	Delete(ctx context.Context, id string) error
}

type History interface {
	Load(ctx context.Context, limit int) ([]entity.HistoryEntry, error)
	Append(ctx context.Context, req *entity.Request, resp *entity.Response)
}

type HistoryService interface {
	Get(ctx context.Context) ([]entity.HistoryEntry, error)
	Save(ctx context.Context, history []entity.HistoryEntry) error
}
