package transport

import (
	"encoding/json"
	"fmt"
	"goph_keeper/internal/client/interfaces"
	"goph_keeper/internal/shared/models"
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

func getEncryptedData[T models.CardData | models.BinaryData | models.TextData](data T) ([]byte, []byte, error) {
	rawBytes, err := json.Marshal(data)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	mockNonce := []byte("mock-nonce-12")

	return rawBytes, mockNonce, nil
}

func getDecryptedData[T models.CardData | models.BinaryData | models.TextData](payload []byte, nonce []byte, target *T) error {
	if err := json.Unmarshal(payload, target); err != nil {
		return fmt.Errorf("failed to unmarshal mock data: %w", err)
	}
	return nil
}
