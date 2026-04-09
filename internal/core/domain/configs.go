package domain

import (
	"context"

	"github.com/nxdir-s/grl/internal/core/valobj"
	"github.com/nxdir-s/grl/internal/ports"
)

type Configs struct {
	service ports.ConfigService
}

func NewConfigs(service ports.ConfigService) *Configs {
	return &Configs{
		service: service,
	}
}

func (d *Configs) Get(ctx context.Context) (*valobj.Config, error) {
	return d.service.Get(ctx)
}

func (d *Configs) Save(ctx context.Context, cfg *valobj.Config) error {
	return d.service.Save(ctx, cfg)
}
