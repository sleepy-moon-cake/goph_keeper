package db

import (
	"context"
	"testing"
	"time"
)

// Тест: Передача невалидного Connection String должна приводить к ошибке парсинга/подключения
func TestNewConnector_InvalidConnStr_Error(t *testing.T) {
	invalidConnStr := "postgres://invalid_user:wrong_password@localhost:54321/non_existent_db?sslmode=disable"

	// Так как адрес и порт недоступны, функция должна завершиться ошибкой на этапе Ping или подключения
	pool, err := NewConnector(t.Context(), invalidConnStr)

	// Проверяем, что пул не создался, а ошибка вернулась
	if pool != nil {
		pool.Close()
		t.Error("expected pool to be nil for invalid connection string, but got an instance")
	}

	if err == nil {
		t.Error("expected connection or ping error for invalid connection string, got nil")
	}
}

// Тест: Передача уже отмененного контекста должна сразу возвращать ошибку
func TestNewConnector_CancelledContext_Error(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel() // Отменяем контекст сразу перед вызовом

	// Передаем даже корректную синтаксически строку, но с отмененным контекстом
	pool, err := NewConnector(ctx, "postgres://user:pass@localhost:5432/db")

	if pool != nil {
		pool.Close()
		t.Error("expected pool to be nil when context is cancelled, but got an instance")
	}

	if err == nil {
		t.Error("expected error due to cancelled context, got nil")
	}
}

// Тест: Проверка таймаута (если контекст завершается быстрее, чем за секунду)
func TestNewConnector_TimeoutContext_Error(t *testing.T) {
	// Создаем контекст с экстремально коротким таймаутом в 1 наносекунду
	ctx, cancel := context.WithTimeout(t.Context(), 1*time.Nanosecond)
	defer cancel()

	pool, err := NewConnector(ctx, "postgres://user:pass@localhost:5432/db")

	if pool != nil {
		pool.Close()
		t.Error("expected pool to be nil due to immediate timeout, but got an instance")
	}

	if err == nil {
		t.Error("expected timeout error, got nil")
	}
}
