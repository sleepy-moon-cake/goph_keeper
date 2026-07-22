package register

import (
	"testing"

	mocks "goph_keeper/internal/client/interfaces/gen"
)

func TestNewRegisterCmd_Creation(t *testing.T) {
	mockService := mocks.NewMockTransportService(nil)
	saveSessionFunc := func(name, key, token string) error { return nil }

	cmd := NewRegisterCmd(mockService, saveSessionFunc)

	if cmd.Use != "register" {
		t.Errorf("ожидалось имя команды 'register', получено: %s", cmd.Use)
	}
}
