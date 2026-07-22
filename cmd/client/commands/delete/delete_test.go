package delete // Тот же пакет, что и у кода команды

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	mocks "goph_keeper/internal/client/interfaces/gen" // Путь к вашему MockTransportService

	"go.uber.org/mock/gomock"
)

// Тест успешного удаления сущности при передаче флага --name
func TestNewDeleteCmd_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Сбрасываем глобальную переменную пакета перед тестом
	name = ""

	mockService := mocks.NewMockTransportService(ctrl)

	// Ожидаем вызов метода DeleteEntityByName с правильным именем
	mockService.EXPECT().
		DeleteEntityByName(gomock.Any(), "my_secret_passport").
		Return(nil).
		Times(1)

	cmd := NewDeleteCmd(mockService)

	// Изолируем вывод команды
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetContext(t.Context())

	// Передаем аргументы флага
	cmd.SetArgs([]string{"--name", "my_secret_passport"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error from delete command: %v", err)
	}
}

// Тест: Ошибка, если обязательный флаг --name не передан
func TestNewDeleteCmd_MissingName_Error(t *testing.T) {
	// Сбрасываем глобальную переменную пакета перед тестом
	name = ""

	mockService := mocks.NewMockTransportService(nil) // Мок не будет вызываться

	cmd := NewDeleteCmd(mockService)

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetContext(t.Context())

	// Запускаем без флагов
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error due to missing name flag, got nil")
	}

	expectedErr := "name is required param"
	if !strings.Contains(err.Error(), expectedErr) {
		t.Errorf("expected error to contain '%s', got: %v", expectedErr, err)
	}
}

// Тест: Ошибка, если транспортный сервис вернул ошибку при удалении
func TestNewDeleteCmd_ServiceError(t *testing.T) {
	name = ""

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockTransportService(ctrl)

	// Настраиваем мок на возврат ошибки сети/бд
	mockService.EXPECT().
		DeleteEntityByName(gomock.Any(), "any_name").
		Return(errors.New("db connection failure")).
		Times(1)

	cmd := NewDeleteCmd(mockService)

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetContext(t.Context())
	cmd.SetArgs([]string{"--name", "any_name"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected service error, got nil")
	}

	if !strings.Contains(err.Error(), "delete entity:") {
		t.Errorf("expected wrapped error with prefix 'delete entity:', got: %v", err)
	}
}
