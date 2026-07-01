package transport

import (
	"fmt"
	"goph_keeper/internal/client/interfaces"
)

type TransportConfig struct {
	AddrGRPC string
	AddrHTTP string
}

func NewTransportService(cfg *TransportConfig) (interfaces.TransportService, error) {
	if cfg.AddrGRPC != "" {
		return NewGRPCTransportService(cfg.AddrGRPC), nil
	}

	if cfg.AddrHTTP != "" {
		return NewHttpTransportService(cfg.AddrHTTP), nil
	}

	return nil, fmt.Errorf("addrGRPC and addrHTTP cant be empty")
}
