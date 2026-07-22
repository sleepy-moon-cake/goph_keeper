package delivery_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"goph_keeper/internal/server/delivery"
	"goph_keeper/internal/server/interfaces"
	mocks "goph_keeper/internal/server/interfaces/gen"
	"goph_keeper/internal/shared/models"

	"go.uber.org/mock/gomock"
)

// Тест: Успешная регистрация по HTTP
func TestHTTPHandler_Register_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := mocks.NewMockRepositoryDb(ctrl)
	handler := delivery.NewHTTPHandler(mockDb, testSecretKey)

	// Формируем валидное тело запроса
	reqBody := models.AuthRequest{
		Name:         "new_http_user",
		PasswordHash: "secure_hash",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	// Ожидаем вызов сохранения пользователя в БД
	mockDb.EXPECT().
		AddUser(gomock.Any(), "new_http_user", "secure_hash").
		Return(nil).
		Times(1)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(bodyBytes))
	w := httptest.NewRecorder()

	handler.Register(w, req)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", res.StatusCode)
	}

	var resp models.AuthResponse
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Session == "" {
		t.Error("expected generated session token, got empty string")
	}
}

// Тест: Попытка регистрации с занятым username (Status Conflict)
func TestHTTPHandler_Register_Conflict(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := mocks.NewMockRepositoryDb(ctrl)
	handler := delivery.NewHTTPHandler(mockDb, testSecretKey)

	reqBody := models.AuthRequest{Name: "duplicate_user", PasswordHash: "hash"}
	bodyBytes, _ := json.Marshal(reqBody)

	// Имитируем ошибку уникальности в БД
	mockDb.EXPECT().
		AddUser(gomock.Any(), "duplicate_user", "hash").
		Return(interfaces.ErrUserAlreadyExists).
		Times(1)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(bodyBytes))
	w := httptest.NewRecorder()

	handler.Register(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusConflict {
		t.Errorf("expected status 409, got %d", res.StatusCode)
	}
}

// Тест: Успешный POST запрос на SaveRecord (проверка контекста)
func TestHTTPHandler_SaveRecord_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := mocks.NewMockRepositoryDb(ctrl)
	handler := delivery.NewHTTPHandler(mockDb, testSecretKey)

	record := models.EncryptedRecord{
		Name:     "credentials",
		DataType: "text",
		Payload:  []byte("crypted"),
		Nonce:    []byte("123"),
	}
	bodyBytes, _ := json.Marshal(record)

	// Наполняем контекст именем пользователя, имитируя работу middleware
	ctx := models.WithUserName(t.Context(), "alice")
	req := httptest.NewRequest(http.MethodPost, "/api/v1/records", bytes.NewReader(bodyBytes)).WithContext(ctx)
	w := httptest.NewRecorder()

	mockDb.EXPECT().
		SaveRecord(gomock.Any(), "alice", gomock.Eq(record)).
		Return(nil).
		Times(1)

	handler.SaveRecord(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusOK {
		t.Errorf("expected status 200 OK, got %d", res.StatusCode)
	}
}

// Тест: Вызов SaveRecord без авторизованного пользователя в контексте
func TestHTTPHandler_SaveRecord_Unauthorized(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := mocks.NewMockRepositoryDb(ctrl)
	handler := delivery.NewHTTPHandler(mockDb, testSecretKey)

	// Контекст пустой — middleware токена не отработало или отработало с ошибкой
	req := httptest.NewRequest(http.MethodPost, "/api/v1/records", bytes.NewReader([]byte("{}")))
	w := httptest.NewRecorder()

	handler.SaveRecord(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected status 401 Unauthorized, got %d", res.StatusCode)
	}
}

// Тест: Успешное получение записи по имени через HTTP GET
func TestHTTPHandler_GetRecord_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := mocks.NewMockRepositoryDb(ctrl)
	handler := delivery.NewHTTPHandler(mockDb, testSecretKey)

	ctx := models.WithUserName(t.Context(), "bob")
	// strings.TrimPrefix отсекает префикс, оставляя "my_passport"
	req := httptest.NewRequest(http.MethodGet, "/api/v1/records/my_passport", nil).WithContext(ctx)
	req.SetPathValue("name", "my_passport")
	w := httptest.NewRecorder()

	expectedRecord := models.EncryptedRecord{
		Name:     "my_passport",
		DataType: "file",
	}

	mockDb.EXPECT().
		GetRecord(gomock.Any(), "bob", "my_passport").
		Return(expectedRecord, nil).
		Times(1)

	handler.GetRecord(w, req)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Errorf("expected status 200 OK, got %d", res.StatusCode)
	}

	var result models.EncryptedRecord
	_ = json.NewDecoder(res.Body).Decode(&result)
	if result.Name != "my_passport" {
		t.Errorf("expected record name 'my_passport', got '%s'", result.Name)
	}
}
