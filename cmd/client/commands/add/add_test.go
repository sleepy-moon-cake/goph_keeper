package add

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	mocks "goph_keeper/internal/client/interfaces/gen" // Путь к вашему MockTransportService

	"go.uber.org/mock/gomock"
)

// Функция для очистки глобального состояния перед каждым тестом
func resetAddFlags() {
	isText = false
	isFile = false
	isCard = false
	name = ""
	path = ""
	value = ""
	holder = "card_holder" // значение по умолчанию из вашей привязки флага
	number = ""
	cvv = ""
	expire = ""
}

// Тест успешного добавления текстовых данных
func TestNewAddCommand_SaveText_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	resetAddFlags()

	mockService := mocks.NewMockTransportService(ctrl)
	ctx := context.Background()

	// Ожидаем вызов SaveText. Замените тип аргументов на структуру, которую генерирует ваш handleText()
	mockService.EXPECT().
		SaveText(gomock.Any(), gomock.Any()).
		Return(nil).
		Times(1)

	cmd := NewAddCommand(mockService)

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetContext(ctx)

	// Передаем аргументы для сохранения текста
	cmd.SetArgs([]string{"--text", "--name", "my_notes", "--value", "some secret text"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error from add text command: %v", err)
	}
}

// Тест: Ошибка, если не указан ни один тип данных (--text, --file, --card)
func TestNewAddCommand_MissingDataType_Error(t *testing.T) {
	resetAddFlags()

	mockService := mocks.NewMockTransportService(nil)
	cmd := NewAddCommand(mockService)

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetContext(context.Background())

	// Передаем только имя, забыв тип данных
	cmd.SetArgs([]string{"--name", "any_name"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error due to missing data type flag, got nil")
	}

	expectedErr := "assign data type (--text, --card or --file)"
	if !strings.Contains(err.Error(), expectedErr) {
		t.Errorf("expected error to contain '%s', got: %v", expectedErr, err)
	}
}

// Тест: Ошибка, если передан тип данных, но пропущено обязательное имя
func TestNewAddCommand_MissingName_Error(t *testing.T) {
	resetAddFlags()

	mockService := mocks.NewMockTransportService(nil)
	cmd := NewAddCommand(mockService)

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetContext(context.Background())

	// Передаем тип, но не передаем имя
	cmd.SetArgs([]string{"--text", "--value", "hello"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error due to missing name flag, got nil")
	}

	expectedErr := "name is required param"
	if !strings.Contains(err.Error(), expectedErr) {
		t.Errorf("expected error to contain '%s', got: %v", expectedErr, err)
	}
}

// Тест: Ошибка, если транспортный сервис вернул ошибку при сохранении карты
// Тест: Ошибка, если транспортный сервис вернул ошибку при сохранении карты
func TestNewAddCommand_SaveCard_ServiceError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	resetAddFlags()

	mockService := mocks.NewMockTransportService(ctrl)

	// Имитируем падение сети/сервера при вызове SaveCard
	mockService.EXPECT().
		SaveCard(gomock.Any(), gomock.Any()).
		Return(errors.New("grpc internal error")).
		Times(1)

	cmd := NewAddCommand(mockService)

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetContext(context.Background())

	// ДОБАВЛЕНО: флаг --expire, чтобы handleCard() пропустил валидацию
	cmd.SetArgs([]string{
		"--card",
		"--name", "visa",
		"--number", "1111222233334444",
		"--cvv", "123",
		"--expire", "12/28",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected service error, got nil")
	}

	if !strings.Contains(err.Error(), "saveCard") {
		t.Errorf("expected wrapped error with prefix 'saveCard', got: %v", err)
	}
}
