package utils

import (
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

const testSecretKey = "my-super-secret-key-for-jwt-signing"

// Тест успешного цикла создания и парсинга токена
func Test_GenerateAndParseToken_Success(t *testing.T) {
	expectedUser := "alex_cloud"

	// 1. Генерируем токен
	tokenStr, err := GenerateToken(expectedUser, testSecretKey)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}
	if tokenStr == "" {
		t.Fatal("generated token is empty")
	}

	// 2. Парсим токен
	parsedUser, err := ParseToken(tokenStr, testSecretKey)
	if err != nil {
		t.Fatalf("failed to parse valid token: %v", err)
	}

	// 3. Проверяем имя пользователя
	if parsedUser != expectedUser {
		t.Errorf("expected username '%s', got '%s'", expectedUser, parsedUser)
	}
}

// Тест: Попытка распарсить токен с неверным секретным ключом
func Test_ParseToken_WrongKey_Error(t *testing.T) {
	tokenStr, _ := GenerateToken("user1", testSecretKey)
	wrongKey := "completely-different-secret-key"

	_, err := ParseToken(tokenStr, wrongKey)
	if err == nil {
		t.Fatal("expected error when parsing token with a wrong key, got nil")
	}

	if !errors.Is(err, ErrParseToken) {
		t.Errorf("expected error to wrap ErrParseToken, got: %v", err)
	}
}

// Тест: Попытка распарсить сломанную строку вместо JWT
func Test_ParseToken_CorruptedString_Error(t *testing.T) {
	_, err := ParseToken("not.a.valid.jwt.string", testSecretKey)
	if err == nil {
		t.Fatal("expected error for malformed jwt string, got nil")
	}

	if !errors.Is(err, ErrParseToken) {
		t.Errorf("expected error to wrap ErrParseToken, got: %v", err)
	}
}

// Тест: Проверка автоматической валидации истекшего по времени токена
func Test_ParseToken_Expired_Error(t *testing.T) {
	// Создаем токен вручную, который уже истек (минус 1 час назад)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
		},
		UserName: "expired_user",
	})

	expiredTokenStr, err := token.SignedString([]byte(testSecretKey))
	if err != nil {
		t.Fatalf("failed to sign expired token: %v", err)
	}

	// Пытаемся распарсить — библиотека jwt должна выбросить ошибку времени действия
	_, err = ParseToken(expiredTokenStr, testSecretKey)
	if err == nil {
		t.Fatal("expected error for expired token, got nil")
	}

	if !errors.Is(err, ErrParseToken) {
		t.Errorf("expected error to wrap ErrParseToken, got: %v", err)
	}
}
