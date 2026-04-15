package ports

import (
	"context"

	"github.com/nxdir-s/grl/internal/core/entity"
	"github.com/nxdir-s/grl/internal/core/valobj"
)

type TUI interface {
	SendRequest(ctx context.Context, req *entity.Request) (*entity.Response, error)

	RecordHistory(ctx context.Context, req *entity.Request, resp *entity.Response) error
	GetHistory(ctx context.Context, limit int) ([]entity.HistoryEntry, error)

	CreateCollection(ctx context.Context, name string, req *entity.Request) ([]entity.Collection, error)
	ListCollections(ctx context.Context) ([]entity.Collection, error)
	SaveRequestToCollection(ctx context.Context, req *entity.Request, name string, collectionID string) ([]entity.Collection, error)

	CreateEnvironment(ctx context.Context, name string) (*entity.Environment, error)
	ListEnvironments(ctx context.Context) ([]entity.Environment, error)
	GetEnvironment(ctx context.Context, id string) (*entity.Environment, error)
	SaveEnvironment(ctx context.Context, env *entity.Environment) error
	DeleteEnvironment(ctx context.Context, id string) error

	GetActiveEnvironment(ctx context.Context) (*entity.Environment, error)
	SetActiveEnvironment(ctx context.Context, id string) error
	GetActiveEnvVars(ctx context.Context) (map[string]string, error)

	GetConfig(ctx context.Context) *valobj.Config
	SaveConfig(ctx context.Context, cfg *valobj.Config) error

	ColorizeJSON(s string) string
	CopyToClipboard(s string) error
}
