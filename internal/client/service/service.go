package service

import (
	"context"
	"goph_keeper/internal/shared/models"
)

type TransportService interface {
	SaveText(ctx context.Context, data models.TextData) error
	SaveCard(ctx context.Context, data models.CardData) error
	SaveFile(ctx context.Context, data models.BinaryData) error
}

type ClientService struct {
	transport TransportService
}

func NewClientService(transport TransportService) *ClientService {
	return &ClientService{transport}
}

func (s *ClientService) SaveText(ctx context.Context, data models.TextData) error {
	return s.transport.SaveText(ctx, data)
}

func (s *ClientService) SaveCard(ctx context.Context, data models.CardData) error {
	return s.transport.SaveCard(ctx, data)
}

func (s *ClientService) SaveFile(ctx context.Context, data models.BinaryData) error {
	return s.transport.SaveFile(ctx, data)
}

func (s *ClientService) GetEntity() {}

func (s *ClientService) DeleteEntity() {}

func (s *ClientService) GetEntiteis() {}
