package service

import (
	"context"

	"github.com/nxdir-s/grl/internal/core/entity"
	"github.com/nxdir-s/grl/internal/ports"
)

type EnvironmentService struct {
	storage ports.Storage
}

func NewEnvironmentService(storage ports.Storage) *EnvironmentService {
	return &EnvironmentService{
		storage: storage,
	}
}

func (s *EnvironmentService) List(ctx context.Context) ([]entity.Environment, error) {
	return s.storage.ListEnvironments(ctx)
}

func (s *EnvironmentService) Get(ctx context.Context, id string) (*entity.Environment, error) {
	return s.storage.LoadEnvironment(ctx, id)
}

func (s *EnvironmentService) Save(ctx context.Context, env *entity.Environment) error {
	return s.storage.SaveEnvironment(ctx, env)
}

func (s *EnvironmentService) Delete(ctx context.Context, id string) error {
	return s.storage.DeleteEnvironment(ctx, id)
}
