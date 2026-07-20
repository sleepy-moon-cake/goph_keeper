package delivery_test

import (
	"errors"
	"testing"

	"goph_keeper/internal/server/delivery"
	mocks "goph_keeper/internal/server/interfaces/gen" // Сверьте путь к вашим мокам repository_db
	"goph_keeper/internal/shared/models"
	"goph_keeper/internal/shared/pb"

	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const testSecretKey = "super-secret-jwt-key-32-bytes!!"

// Тест успешной регистрации нового пользователя
func TestGRPCTransportServer_Register_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := mocks.NewMockRepositoryDb(ctrl)
	server := delivery.NewGRPCHandler(mockDb, testSecretKey)

	req := &pb.AuthRequest{}
	req.SetUsername("test_user")
	req.SetPasswordHash("hashed_password")

	// Настраиваем мок: ожидаем вызов AddUser и возвращаем отсутствие ошибки
	mockDb.EXPECT().
		AddUser(gomock.Any(), "test_user", "hashed_password").
		Return(nil).
		Times(1)

	resp, err := server.Register(t.Context(), req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.GetToken() == "" {
		t.Error("expected generated JWT token, got empty string")
	}
}

// Тест валидации пустых полей при регистрации
func TestGRPCTransportServer_Register_InvalidArgument(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := mocks.NewMockRepositoryDb(ctrl)
	server := delivery.NewGRPCHandler(mockDb, testSecretKey)

	req := &pb.AuthRequest{} // Пустой запрос

	_, err := server.Register(t.Context(), req)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.InvalidArgument {
		t.Errorf("expected gRPC code %v, got %v", codes.InvalidArgument, st.Code())
	}
}

// Тест успешной аутентификации (Login)
func TestGRPCTransportServer_Login_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := mocks.NewMockRepositoryDb(ctrl)
	server := delivery.NewGRPCHandler(mockDb, testSecretKey)

	req := &pb.AuthRequest{}
	req.SetUsername("existing_user")
	req.SetPasswordHash("correct_hash")

	// База возвращает верный хэш
	mockDb.EXPECT().
		GetUserPassword(gomock.Any(), "existing_user").
		Return("correct_hash", nil).
		Times(1)

	resp, err := server.Login(t.Context(), req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.GetToken() == "" {
		t.Error("expected token, got empty string")
	}
}

// Тест ошибки авторизации при неверном пароле
func TestGRPCTransportServer_Login_InvalidCredentials(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := mocks.NewMockRepositoryDb(ctrl)
	server := delivery.NewGRPCHandler(mockDb, testSecretKey)

	req := &pb.AuthRequest{}
	req.SetUsername("user")
	req.SetPasswordHash("wrong_hash")

	mockDb.EXPECT().
		GetUserPassword(gomock.Any(), "user").
		Return("actual_hash_in_db", nil).
		Times(1)

	_, err := server.Login(t.Context(), req)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.Unauthenticated {
		t.Errorf("expected gRPC code %v, got %v", codes.Unauthenticated, st.Code())
	}
}

// Тест SaveRecord без авторизационного контекста пользователя
func TestGRPCTransportServer_SaveRecord_Unauthenticated(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := mocks.NewMockRepositoryDb(ctrl)
	server := delivery.NewGRPCHandler(mockDb, testSecretKey)

	req := &pb.Record{}
	req.SetName("secret")

	// Вызов с пустым контекстом (без модели UserContextKey)
	_, err := server.SaveRecord(t.Context(), req)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.Unauthenticated {
		t.Errorf("expected gRPC code %v, got %v", codes.Unauthenticated, st.Code())
	}
}

// Тест успешного получения записи по имени через GetRecord
func TestGRPCTransportServer_GetRecord_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := mocks.NewMockRepositoryDb(ctrl)
	server := delivery.NewGRPCHandler(mockDb, testSecretKey)

	// Добавляем имя пользователя в контекст, имитируя работу middleware авторизации
	ctx := models.WithUserName(t.Context(), "authorized_user")
	req := &pb.GetRecordRequest{}
	req.SetName("bank_card")

	dbRecord := models.EncryptedRecord{
		Name:     "bank_card",
		DataType: "card",
		Payload:  []byte("crypted-bytes"),
		Nonce:    []byte("nonce123"),
	}

	// Сгенерированный метод возвращает объект по значению
	mockDb.EXPECT().
		GetRecord(ctx, "authorized_user", "bank_card").
		Return(dbRecord, nil).
		Times(1)

	resp, err := server.GetRecord(ctx, req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.GetName() != "bank_card" || resp.GetDataType() != "card" {
		t.Errorf("response data mismatch: got name=%s, type=%s", resp.GetName(), resp.GetDataType())
	}
}

// Тест поведения GetRecord, если запись не найдена в БД
func TestGRPCTransportServer_GetRecord_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := mocks.NewMockRepositoryDb(ctrl)
	server := delivery.NewGRPCHandler(mockDb, testSecretKey)

	ctx := models.WithUserName(t.Context(), "authorized_user")

	req := &pb.GetRecordRequest{}
	req.SetName("lost_item")

	mockDb.EXPECT().
		GetRecord(ctx, "authorized_user", "lost_item").
		Return(models.EncryptedRecord{}, errors.New("sql: no rows in result set")).
		Times(1)

	_, err := server.GetRecord(ctx, req)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.NotFound {
		t.Errorf("expected gRPC code %v, got %v", codes.NotFound, st.Code())
	}
}
