package login

import (
	"testing"

	mocks "goph_keeper/internal/client/interfaces/gen"
)

// Тест проверяет, что команда успешно создается и имеет правильные параметры
func TestNewLoginCmd_Creation(t *testing.T) {
	mockService := mocks.NewMockTransportService(nil)
	saveSessionFunc := func(name, key, token string) error { return nil }

	// Просто создаем команду
	cmd := NewLoginCmd(mockService, saveSessionFunc)

	// Проверяем базовые свойства Cobra-команды
	if cmd.Use != "login" {
		t.Errorf("ожидалось имя команды 'login', получено: %s", cmd.Use)
	}

	if cmd.Short != "authorization" {
		t.Errorf("ожидалось описание 'authorization', получено: %s", cmd.Short)
	}
}
