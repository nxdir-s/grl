package service

import (
	"context"

	"github.com/nxdir-s/grl/internal/core/entity"
	"github.com/nxdir-s/grl/internal/ports"
)

type RequestService struct {
	http ports.Http
}

func NewRequestService(http ports.Http) *RequestService {
	return &RequestService{
		http: http,
	}
}

func (s *RequestService) Send(ctx context.Context, req *entity.Request) (*entity.Response, error) {
	return s.http.Send(ctx, req)
}
