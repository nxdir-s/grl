package domain

import (
	"context"

	"github.com/google/uuid"
	"github.com/nxdir-s/grl/internal/core/entity"
	"github.com/nxdir-s/grl/internal/ports"
)

type ErrRequestNotFound struct{}

func (e *ErrRequestNotFound) Error() string {
	return "request not found in collection"
}

type Collections struct {
	service ports.CollectionService
}

func NewCollections(service ports.CollectionService) *Collections {
	return &Collections{
		service: service,
	}
}

func (d *Collections) Create(ctx context.Context, name string) (*entity.Collection, error) {
	collection := &entity.Collection{
		ID:   d.generateID(),
		Name: name,
	}

	if err := d.service.Save(ctx, collection); err != nil {
		return nil, err
	}

	return collection, nil
}

func (d *Collections) List(ctx context.Context) ([]entity.Collection, error) {
	return d.service.List(ctx)
}

func (d *Collections) Delete(ctx context.Context, id string) error {
	return d.service.Delete(ctx, id)
}

func (d *Collections) AddRequest(ctx context.Context, collectionID string, req *entity.Request) error {
	collection, err := d.service.Load(ctx, collectionID)
	if err != nil {
		return err
	}

	if len(req.ID) == 0 {
		req.ID = d.generateID()
	}

	if len(req.Name) == 0 {
		req.Name = req.Method.String() + " " + req.URL
	}

	collection.Requests = append(collection.Requests, *req)

	return d.service.Save(ctx, collection)
}

func (d *Collections) RemoveRequest(ctx context.Context, collectionID string, requestID string) error {
	collection, err := d.service.Load(ctx, collectionID)
	if err != nil {
		return err
	}

	for i, r := range collection.Requests {
		if r.ID == requestID {
			collection.Requests = append(collection.Requests[:i], collection.Requests[i+1:]...)

			return d.service.Save(ctx, collection)
		}
	}

	return &ErrRequestNotFound{}
}

func (d *Collections) generateID() string {
	return uuid.New().String()
}
