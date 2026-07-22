package transport_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	mocks "goph_keeper/internal/client/interfaces/gen" // Проверьте правильность пути к папке gen
	"goph_keeper/internal/client/transport"
	"goph_keeper/internal/shared/models"

	"go.uber.org/mock/gomock"
)

// Тест успешной авторизации через метод Login
func TestHttpTransportService_Login_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// 1. Создаем тестовый HTTP-сервер
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/auth/login" || r.Method != http.MethodPost {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		var req models.AuthRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(models.AuthResponse{Session: "fake-jwt-token"})
	}))
	defer server.Close()

	// 2. Инициализируем мок и наш сервис
	mockCache := mocks.NewMockCacheService(ctrl)
	svc := transport.NewHttpTransportService(server.URL, mockCache, "key")

	// 3. Вызываем тестируемый метод
	token, err := svc.Login(t.Context(), "test_user", "password123")

	// 4. Проверяем утверждения (Assertions)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if token != "fake-jwt-token" {
		t.Errorf("expected 'fake-jwt-token', got: %s", token)
	}
}

// Тест обработки ошибки сервера (например, 500 Internal Server Error)
func TestHttpTransportService_Login_BadStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	mockCache := mocks.NewMockCacheService(ctrl)
	svc := transport.NewHttpTransportService(server.URL, mockCache, "key")
	_, err := svc.Login(t.Context(), "user", "pass")

	if err == nil {
		t.Error("expected error due to bad status code, got nil")
	}
}

// Тест метода GetEntityByName при успешном ответе сети и обновлении кэша
func TestHttpTransportService_GetEntityByName_NetworkSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	fakeRecord := models.EncryptedRecord{
		Name:     "summer",
		DataType: "text",
		Payload:  []byte("encrypted-payload-bytes"),
		Nonce:    []byte("12bytesnonce"),
	}

	// Исправлено: r *http.Request вместо *http.Format
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(fakeRecord)
	}))
	defer server.Close()

	// Настраиваем Mock-кэш: ожидаем вызов UpdateSingleRecord с записью "summer"
	mockCache := mocks.NewMockCacheService(ctrl)
	mockCache.EXPECT().
		UpdateSingleRecord(gomock.Any(), gomock.Eq(&fakeRecord)).
		Return(nil).
		Times(1)

	svc := transport.NewHttpTransportService(server.URL, mockCache, "key")
	ctx := models.WithCryptedKey(t.Context(), "super_secret_hex_key")

	// Вызываем метод
	_, _ = svc.GetEntityByName(ctx, "summer")
}

// Тест автоматического переключения на локальный кэш (Fallback) при падении сети
func TestHttpTransportService_GetEntityByName_FallbackToCache(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	server.Close() // Имитируем отсутствие сети

	mockCache := mocks.NewMockCacheService(ctrl)

	// Настраиваем Mock-кэш: ожидаем вызов GetRecordByName при падении сети
	mockCache.EXPECT().
		GetRecordByName(gomock.Any(), "summer").
		Return(&models.EncryptedRecord{
			Name:     "summer",
			DataType: "text",
			Payload:  []byte{1, 2, 3},
			Nonce:    []byte{4, 5, 6},
		}, nil).
		Times(1)

	svc := transport.NewHttpTransportService(server.URL, mockCache, "key")
	ctx := models.WithCryptedKey(t.Context(), "super_secret_hex_key")

	_, _ = svc.GetEntityByName(ctx, "summer")
}
