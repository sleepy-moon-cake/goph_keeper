package utils

import (
	"encoding/hex"
	"testing"
)

// Тест успешной генерации ключа и проверки его свойств
func TestGenerateSecretKey_Success(t *testing.T) {
	password := "super_secret_password"
	username := "mooncake"

	keyHex := GenerateSecretKey(password, username)

	// 1. Проверяем, что ключ не пустой
	if keyHex == "" {
		t.Fatal("expected generated secret key to be not empty")
	}

	// 2. Декодируем из HEX обратно в байты, чтобы проверить длину
	keyBytes, err := hex.DecodeString(keyHex)
	if err != nil {
		t.Fatalf("failed to decode generated hex key: %v", err)
	}

	// 3. PBKDF2 настроен на 32 байта (AES-256), проверяем длину бинарного ключа
	expectedLength := 32
	if len(keyBytes) != expectedLength {
		t.Errorf("expected key length to be %d bytes, got %d", expectedLength, len(keyBytes))
	}
}

// Тест детерминированности алгоритма (одинаковые входные данные всегда дают одинаковый результат)
func TestGenerateSecretKey_Deterministic(t *testing.T) {
	password := "same_password"
	username := "same_user"

	key1 := GenerateSecretKey(password, username)
	key2 := GenerateSecretKey(password, username)

	if key1 != key2 {
		t.Errorf("expected keys to be identical for the same input, got:\nkey1: %s\nkey2: %s", key1, key2)
	}
}

// Тест того, что изменение соли (username) или пароля полностью меняет выходной ключ
func TestGenerateSecretKey_ChangesWithInput(t *testing.T) {
	password := "password"
	username1 := "user1"
	username2 := "user2"

	key1 := GenerateSecretKey(password, username1)
	key2 := GenerateSecretKey(password, username2)

	if key1 == key2 {
		t.Error("expected different keys for different usernames, but got the same output")
	}
}
