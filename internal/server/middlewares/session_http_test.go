package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"goph_keeper/internal/server/utils"
	"goph_keeper/internal/shared/models"
)

// Фейковый хэндлер для проверки успешного прохождения middleware
func mockHTTPHandler(t *testing.T, expectedUser string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Проверяем, что имя пользователя попало в контекст запроса
		username, ok := models.GetUserName(r.Context())

		if !ok || username != expectedUser {
			t.Errorf("expected user '%s' in context, got '%s'", expectedUser, username)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("success"))
	})
}

// Тест: Успешная авторизация с валидным JWT токеном в заголовке
func TestJWTSession_Success(t *testing.T) {
	username := "alice_http"
	validToken, err := utils.GenerateToken(username, testSecretKey)
	if err != nil {
		t.Fatalf("failed to generate test token: %v", err)
	}

	middleware := JWTSession(testSecretKey)
	handlerToTest := middleware(mockHTTPHandler(t, username))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/records", nil)
	req.Header.Set("Authorization", "Bearer "+validToken)
	w := httptest.NewRecorder()

	handlerToTest.ServeHTTP(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusOK {
		t.Errorf("expected status 200 OK, got %d", res.StatusCode)
	}
}

// Тест: Отсутствие заголовка Authorization возвращает 401 Unauthorized
func TestJWTSession_MissingHeader_Unauthorized(t *testing.T) {
	middleware := JWTSession(testSecretKey)

	// Если middleware сработает неверно и пропустит запрос, внутренний хэндлер упадет, так как пользователя нет в контексте
	handlerToTest := middleware(mockHTTPHandler(t, ""))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/records", nil) // Без заголовка
	w := httptest.NewRecorder()

	handlerToTest.ServeHTTP(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected status 401 Unauthorized, got %d", res.StatusCode)
	}
}

// Тест: Попытка пройти с невалидным или испорченным токеном возвращает 401 Unauthorized
func TestJWTSession_InvalidToken_Unauthorized(t *testing.T) {
	middleware := JWTSession(testSecretKey)
	handlerToTest := middleware(mockHTTPHandler(t, ""))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/records", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.payload")
	w := httptest.NewRecorder()

	handlerToTest.ServeHTTP(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected status 401 Unauthorized, got %d", res.StatusCode)
	}
}
