package list

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	mocks "goph_keeper/internal/client/interfaces/gen"
	"goph_keeper/internal/shared/models"

	"go.uber.org/mock/gomock"
)

// Тест успешного вывода списка записей
func TestNewListCmd_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	limit = 100 // Сброс глобального флага к значению по умолчанию

	mockService := mocks.NewMockTransportService(ctrl)
	ctx := context.Background()

	// Формируем фейковый список записей от сервиса
	mockResponse := []models.RecordMeta{
		{Name: "yandex_auth", DataType: "text"},
		{Name: "mastercard", DataType: "card"},
	}

	// Ожидаем, что метод ListRecords вызовется с переданным лимитом
	mockService.EXPECT().
		ListRecords(gomock.Any(), 10).
		Return(mockResponse, nil).
		Times(1)

	cmd := NewListCmd(mockService)
	cmd.SetContext(ctx)
	
	// Передаем кастомный лимит через аргументы
	cmd.SetArgs([]string{"--limit", "10"})

	// Перехватываем os.Stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := cmd.Execute()

	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	// Проверяем наличие заголовка и переданных мета-данных в таблице
	if !strings.Contains(output, "YOUR SAVED RECORDS") {
		t.Error("expected output header to be printed")
	}
	if !strings.Contains(output, "yandex_auth") || !strings.Contains(output, "text") {
		t.Error("first record metadata not found in table output")
	}
	if !strings.Contains(output, "mastercard") || !strings.Contains(output, "card") {
		t.Error("second record metadata not found in table output")
	}
}

// Тест обработки ошибки транспортного слоя
func TestNewListCmd_ServiceError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	limit = 100

	mockService := mocks.NewMockTransportService(ctrl)
	mockService.EXPECT().
		ListRecords(gomock.Any(), gomock.Any()).
		Return(nil, errors.New("grpc stream error")).
		Times(1)

	cmd := NewListCmd(mockService)
	cmd.SetArgs([]string{}) // используем лимит по умолчанию

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected wrapping service error, got nil")
	}

	if !strings.Contains(err.Error(), "list command:") {
		t.Errorf("expected wrapped error prefix 'list command:', got: %v", err)
	}
}
