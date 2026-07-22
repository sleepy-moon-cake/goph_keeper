package middlewares

import (
	"context"
	"testing"

	"goph_keeper/internal/server/utils"
	"goph_keeper/internal/shared/models"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const testSecretKey = "super-secret-jwt-key-32-bytes!!"

// Вспомогательный пустой хэндлер, который проверяет наличие пользователя в контексте
func mockUnaryHandler(ctx context.Context, req interface{}) (interface{}, error) {
	return ctx, nil
}

// Тест: Методы Login и Register должны пропускаться без проверки токена
func TestAuthUnaryInterceptor_BypassAuthMethods(t *testing.T) {
	interceptor := AuthUnaryInterceptor(testSecretKey)
	ctx := t.Context()

	bypassMethods := []string{"/pb.TransportService/Login", "/pb.TransportService/Register"}

	for _, method := range bypassMethods {
		info := &grpc.UnaryServerInfo{FullMethod: method}

		// Вызываем интерцептор с пустым контекстом без метаданных
		resCtx, err := interceptor(ctx, nil, info, mockUnaryHandler)

		if err != nil {
			t.Errorf("method %s should bypass token check, but got error: %v", method, err)
		}

		if resCtx == nil {
			t.Errorf("method %s returned nil context", method)
		}
	}
}

// Тест: Успешная авторизация с валидным Bearer токеном
func TestAuthUnaryInterceptor_Success(t *testing.T) {
	interceptor := AuthUnaryInterceptor(testSecretKey)

	// Генерируем валидный токен для теста
	username := "test_grpc_user"
	validToken, err := utils.GenerateToken(username, testSecretKey)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	// Создаем контекст с gRPC метаданными авторизации
	md := metadata.Pairs("authorization", "Bearer "+validToken)
	ctx := metadata.NewIncomingContext(t.Context(), md)

	info := &grpc.UnaryServerInfo{FullMethod: "/pb.TransportService/SaveRecord"}

	// Запускаем перехватчик
	resCtxInterface, err := interceptor(ctx, nil, info, mockUnaryHandler)
	if err != nil {
		t.Fatalf("expected successful authentication, got error: %v", err)
	}

	resCtx, ok := resCtxInterface.(context.Context)
	if !ok {
		t.Fatalf("handler did not return context")
	}

	// Проверяем, что имя пользователя корректно записалось в контекст для хэндлеров
	ctxUsername, ok := models.GetUserName(resCtx)
	if !ok || ctxUsername != username {
		t.Errorf("expected username '%s' in context, got '%s'", username, ctxUsername)
	}
}

// Тест: Отсутствие метаданных в запросе возвращает Unauthenticated
func TestAuthUnaryInterceptor_MissingMetadata(t *testing.T) {
	interceptor := AuthUnaryInterceptor(testSecretKey)

	info := &grpc.UnaryServerInfo{FullMethod: "/pb.TransportService/GetRecord"}

	_, err := interceptor(t.Context(), nil, info, mockUnaryHandler)
	if err == nil {
		t.Fatal("expected error due to missing metadata, got nil")
	}

	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.Unauthenticated || st.Message() != "metadata is missing" {
		t.Errorf("expected Unauthenticated status with 'metadata is missing', got: %v", err)
	}
}

// Тест: Отсутствие токена в заголовках метаданных
func TestAuthUnaryInterceptor_MissingTokenHeader(t *testing.T) {
	interceptor := AuthUnaryInterceptor(testSecretKey)

	// Метаданные есть, но нужного ключа "authorization" внутри нет
	md := metadata.Pairs("some-other-header", "value")
	ctx := metadata.NewIncomingContext(t.Context(), md)

	info := &grpc.UnaryServerInfo{FullMethod: "/pb.TransportService/GetRecord"}

	_, err := interceptor(ctx, nil, info, mockUnaryHandler)
	if err == nil {
		t.Fatal("expected error due to missing token header, got nil")
	}

	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.Unauthenticated || st.Message() != "authorization token is missing" {
		t.Errorf("expected Unauthenticated status with 'authorization token is missing', got: %v", err)
	}
}

// Тест: Попытка пройти с невалидным или испорченным токеном
func TestAuthUnaryInterceptor_InvalidToken(t *testing.T) {
	interceptor := AuthUnaryInterceptor(testSecretKey)

	md := metadata.Pairs("authorization", "Bearer corrupted.jwt.token")
	ctx := metadata.NewIncomingContext(t.Context(), md)

	info := &grpc.UnaryServerInfo{FullMethod: "/pb.TransportService/GetRecord"}

	_, err := interceptor(ctx, nil, info, mockUnaryHandler)
	if err == nil {
		t.Fatal("expected error due to invalid token, got nil")
	}

	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.Unauthenticated || st.Message() != "invalid or expired token" {
		t.Errorf("expected Unauthenticated status with 'invalid or expired token', got: %v", err)
	}
}
