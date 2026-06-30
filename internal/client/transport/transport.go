package transport

import (
	"context"
	"fmt"
	"goph_keeper/internal/shared/models"
)

type TransportService interface {
	SaveText(ctx context.Context, data models.TextData) error
	SaveCard(ctx context.Context, data models.CardData) error
	SaveFile(ctx context.Context, data models.BinaryData) error
}

type TransportConfig struct {
	AddrGRPC string
	AddrHTTP string
}

func NewTransportService(cfg *TransportConfig) (TransportService, error) {
	if cfg.AddrGRPC != "" {
		return NewGRPCTransportService(cfg.AddrGRPC), nil
	}

	if cfg.AddrHTTP != "" {
		return NewHttpTransportService(cfg.AddrHTTP), nil
	}

	return nil, fmt.Errorf("addrGRPC and addrHTTP cant be emty")
}
