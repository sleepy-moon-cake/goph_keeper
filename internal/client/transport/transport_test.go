package transport

import (
	"crypto/rand"
	"encoding/hex"
	"testing"

	mocks "goph_keeper/internal/client/interfaces/gen"
	"goph_keeper/internal/shared/models"

	"go.uber.org/mock/gomock"
)

// Генерирует валидный 32-байтовый HEX-ключ для AES-256
func generateTestHexKey(t *testing.T) string {
	t.Helper()
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		t.Fatalf("failed to generate random key: %v", err)
	}
	return hex.EncodeToString(bytes)
}

// Тест успешного цикла шифрования и дешифрования текстовых данных
func Test_EncryptDecrypt_TextData_Success(t *testing.T) {
	secretKey := generateTestHexKey(t)
	inputData := models.TextData{
		Name: "my_notes",
		Text: "top_secret_content",
	}

	// 1. Тестируем шифрование
	ciphertext, nonce, err := getEncryptedData(inputData, secretKey)
	if err != nil {
		t.Fatalf("failed to encrypt data: %v", err)
	}
	if len(ciphertext) == 0 || len(nonce) != 12 {
		t.Errorf("invalid ciphertext or nonce size")
	}

	// Формируем структуру зашифрованной записи
	record := &models.EncryptedRecord{
		Name:     "my_notes",
		DataType: "text",
		Payload:  ciphertext,
		Nonce:    nonce,
	}

	// 2. Тестируем дешифрование
	decryptedRecord, err := getDecryptedData(record, secretKey)
	if err != nil {
		t.Fatalf("failed to decrypt data: %v", err)
	}

	// 3. Проверяем корректность данных
	outputData, ok := decryptedRecord.Data.(models.TextData)
	if !ok {
		t.Fatalf("expected decrypted data to be models.TextData, got %T", decryptedRecord.Data)
	}

	if outputData.Text != inputData.Text || outputData.Name != inputData.Name {
		t.Errorf("decrypted data inside object mismatch: got %+v, want %+v", outputData, inputData)
	}
}

// Тест возврата ошибки при использовании невалидного (битого) ключа
func Test_Decrypt_WrongKey_Error(t *testing.T) {
	correctKey := generateTestHexKey(t)
	wrongKey := generateTestHexKey(t)

	inputData := models.TextData{Name: "test", Text: "data"}
	ciphertext, nonce, _ := getEncryptedData(inputData, correctKey)

	record := &models.EncryptedRecord{
		Name:     "test",
		DataType: "text",
		Payload:  ciphertext,
		Nonce:    nonce,
	}

	// Пробуем расшифровать другим ключом
	_, err := getDecryptedData(record, wrongKey)
	if err == nil {
		t.Error("expected encryption error when using incorrect key, got nil")
	}
}

// Тест возврата ошибки дешифрования при повреждении nonce
func Test_Decrypt_CorruptedNonce_Error(t *testing.T) {
	secretKey := generateTestHexKey(t)
	inputData := models.TextData{Name: "test", Text: "data"}
	ciphertext, nonce, _ := getEncryptedData(inputData, secretKey)

	// Ломаем длину nonce
	invalidNonce := append(nonce, 0x00)

	record := &models.EncryptedRecord{
		Name:     "test",
		DataType: "text",
		Payload:  ciphertext,
		Nonce:    invalidNonce,
	}

	_, err := getDecryptedData(record, secretKey)
	if err == nil {
		t.Error("expected error due to incorrect nonce size, got nil")
	}
}

// Тест фабрики NewTransportService на валидацию пустых адресов
func TestNewTransportService_EmptyAddresses_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCache := mocks.NewMockCacheService(ctrl)

	// Конфиг без адресов
	cfg := &TransportConfig{
		Cache:      mockCache,
		EncryptKey: "some-key",
	}

	_, err := NewTransportService(cfg)
	if err == nil {
		t.Error("expected factory error when both AddrGRPC and AddrHTTP are empty, got nil")
	}
}
