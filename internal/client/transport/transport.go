package transport

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"goph_keeper/internal/client/interfaces"
	"goph_keeper/internal/shared/models"
	"io"
)

type TransportConfig struct {
	AddrGRPC   string
	AddrHTTP   string
	Cache      interfaces.CacheService
	EncryptKey string
}

func NewTransportService(cfg *TransportConfig) (interfaces.TransportService, error) {
	if cfg.AddrGRPC != "" {
		return NewGRPCTransportService(cfg.AddrGRPC, cfg.Cache, cfg.EncryptKey), nil
	}

	if cfg.AddrHTTP != "" {
		return NewHttpTransportService(cfg.AddrHTTP, cfg.Cache, cfg.EncryptKey), nil
	}

	return nil, fmt.Errorf("addrGRPC and addrHTTP cant be empty")
}

func getEncryptedData[T models.CardData | models.BinaryData | models.TextData](data T, secretKey string) ([]byte, []byte, error) {
	// Маршалинг структуры в JSON
	rawBytes, err := json.Marshal(data)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	// 1. ДЕКОДИРУЕМ HEX-СТРОКУ В БАЙТЫ (64 символа превратятся в 32 байта)
	keyBytes, err := hex.DecodeString(secretKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode hex secret key: %w", err)
	}

	// 2. Передаем правильные байты в AES
	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return nil, nil, fmt.Errorf("cipher init error: %w", err)
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, fmt.Errorf("gcm init error: %w", err)
	}

	// Генерируем СЛУЧАЙНЫЙ nonce
	nonce := make([]byte, aesgcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, fmt.Errorf("nonce generation error: %w", err)
	}

	// Шифруем данные
	ciphertext := aesgcm.Seal(nil, nonce, rawBytes, nil)

	return ciphertext, nonce, nil
}

func getDecryptedData(record *models.EncryptedRecord, cryptedKey string) (*models.DecryptedRecord, error) {
	decRecord := models.DecryptedRecord{
		Name:     record.Name,
		DataType: record.DataType,
	}

	switch record.DataType {
	case "text":
		var v models.TextData
		if err := decrypt(record.Payload, record.Nonce, cryptedKey, &v); err != nil {
			return nil, fmt.Errorf("failed to decrypt text: %w", err)
		}
		decRecord.Data = v
		return &decRecord, nil

	case "card":
		var v models.CardData
		if err := decrypt(record.Payload, record.Nonce, cryptedKey, &v); err != nil {
			return nil, fmt.Errorf("failed to decrypt card: %w", err)
		}
		decRecord.Data = v
		return &decRecord, nil

	case "file":
		var v models.BinaryData
		if err := decrypt(record.Payload, record.Nonce, cryptedKey, &v); err != nil {
			return nil, fmt.Errorf("failed to decrypt file: %w", err)
		}
		decRecord.Data = v
		return &decRecord, nil
	}

	return nil, errors.New("unknown data type")
}
func decrypt[T models.CardData | models.BinaryData | models.TextData](ciphertext, nonce []byte, secretKey string, target *T) error {
	// 1. Декодируем Hex-строку ключа в бинарные 32 байта
	keyBytes, err := hex.DecodeString(secretKey)
	if err != nil {
		return fmt.Errorf("failed to decode hex secret key: %w", err)
	}

	// 2. Инициализируем AES
	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return fmt.Errorf("cipher init error: %w", err)
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("gcm init error: %w", err)
	}

	if len(nonce) != aesgcm.NonceSize() {
		return fmt.Errorf("incorrect nonce length: got %d, want 12", len(nonce))
	}

	rawBytes, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return fmt.Errorf("failed to decrypt data (wrong key or corrupted data): %w", err)
	}

	// 4. Превращаем JSON-байты обратно в готовую структуру
	if err := json.Unmarshal(rawBytes, target); err != nil {
		return fmt.Errorf("failed to unmarshal decrypted data: %w", err)
	}

	return nil
}
