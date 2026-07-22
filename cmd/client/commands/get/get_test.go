package get

import (
	"bytes"
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	mocks "goph_keeper/internal/client/interfaces/gen"
	"goph_keeper/internal/shared/models"

	"go.uber.org/mock/gomock"
)

// Тест успешного получения и отображения текстовых данных
func TestNewGetCmd_Text_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	name = "" // Сброс глобального флага

	mockService := mocks.NewMockTransportService(ctrl)

	mockResponse := &models.DecryptedRecord{
		Name:     "my_notes",
		DataType: "text",
		Data: models.TextData{
			Name: "my_notes",
			Text: "my-secret-password-123",
		},
	}

	mockService.EXPECT().
		GetEntityByName(gomock.Any(), "my_notes").
		Return(mockResponse, nil).
		Times(1)

	cmd := NewGetCmd(mockService)
	cmd.SetContext(t.Context())
	cmd.SetArgs([]string{"--name", "my_notes"})

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

	// ИСПРАВЛЕНО: Для текста проверяем только наличие самого текста
	if !strings.Contains(output, "YOUR SAVED RECORDS") {
		t.Error("expected output header to be printed")
	}
	if !strings.Contains(output, "my_notes") || !strings.Contains(output, "my-secret-password-123") {
		t.Error("expected entity details to be printed in table format")
	}
}

// Тест успешного получения и форматирования банковской карты
func TestNewGetCmd_Card_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	name = ""

	mockService := mocks.NewMockTransportService(ctrl)

	mockResponse := &models.DecryptedRecord{
		Name:     "visa_card",
		DataType: "card",
		Data: models.CardData{
			CardNumber:     "44445555",
			ExpirationDate: "12/30",
			CVV:            "999",
		},
	}

	mockService.EXPECT().
		GetEntityByName(gomock.Any(), "visa_card").
		Return(mockResponse, nil).
		Times(1)

	cmd := NewGetCmd(mockService)
	cmd.SetContext(t.Context())
	cmd.SetArgs([]string{"--name", "visa_card"})

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	_ = cmd.Execute()

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	// ИСПРАВЛЕНО: Проверки карты находятся здесь
	if !strings.Contains(output, "Номер: 44445555") {
		t.Error("card number not found in table output")
	}
	if !strings.Contains(output, "Срок: 12/30") {
		t.Error("expiration date not found in table output")
	}
	if !strings.Contains(output, "CVV: 999") {
		t.Error("CVV not found in table output")
	}
}

// Тест ошибки, если обязательный флаг --name пустой
func TestNewGetCmd_MissingName_Error(t *testing.T) {
	name = ""

	mockService := mocks.NewMockTransportService(nil)
	cmd := NewGetCmd(mockService)
	cmd.SetArgs([]string{}) // флагов нет

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error due to missing name parameter, got nil")
	}

	if !strings.Contains(err.Error(), "name is required param") {
		t.Errorf("expected error 'name is required param', got: %v", err)
	}
}

// Тест обработки ошибки транспортного слоя
func TestNewGetCmd_ServiceError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	name = ""

	mockService := mocks.NewMockTransportService(ctrl)
	mockService.EXPECT().
		GetEntityByName(gomock.Any(), gomock.Any()).
		Return(nil, errors.New("network timeout")).
		Times(1)

	cmd := NewGetCmd(mockService)
	cmd.SetArgs([]string{"--name", "any_entity"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected wrapping service error, got nil")
	}

	if !strings.Contains(err.Error(), "get entity:") {
		t.Errorf("expected wrapped error prefix 'get entity:', got: %v", err)
	}
}
