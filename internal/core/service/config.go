package service

import (
	"context"

	"github.com/nxdir-s/grl/internal/core/valobj"
	"github.com/nxdir-s/grl/internal/ports"
)

type ConfigService struct {
	storage ports.Storage
}

func NewConfigService(storage ports.Storage) *ConfigService {
	return &ConfigService{
		storage: storage,
	}
}

func (s *ConfigService) Get(ctx context.Context) (*valobj.Config, error) {
	return s.storage.LoadConfig(ctx)
}

func (s *ConfigService) Save(ctx context.Context, cfg *valobj.Config) error {
	return s.storage.SaveConfig(ctx, cfg)
}
