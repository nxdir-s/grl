package service

import (
	"context"

	"github.com/nxdir-s/grl/internal/core/entity"
	"github.com/nxdir-s/grl/internal/ports"
)

type CollectionService struct {
	storage ports.Storage
}

func NewCollectionService(storage ports.Storage) *CollectionService {
	return &CollectionService{
		storage: storage,
	}
}

func (s *CollectionService) Save(ctx context.Context, collection *entity.Collection) error {
	return s.storage.SaveCollection(ctx, collection)
}

func (s *CollectionService) Load(ctx context.Context, id string) (*entity.Collection, error) {
	return s.storage.LoadCollection(ctx, id)
}

func (s *CollectionService) List(ctx context.Context) ([]entity.Collection, error) {
	return s.storage.ListCollections(ctx)
}

func (s *CollectionService) Delete(ctx context.Context, id string) error {
	return s.storage.DeleteCollection(ctx, id)
}
