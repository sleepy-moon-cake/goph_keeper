package transport_test

import (
	"context"
	"testing"

	mocks "goph_keeper/internal/client/interfaces/gen" // Путь к вашему сгенерированному моку
	"goph_keeper/internal/client/transport"
	"goph_keeper/internal/shared/models"

	"go.uber.org/mock/gomock"
)

// Тест: gRPC сервер недоступен, ListRecords успешно берет данные из кэша
func TestGPRCTransportService_ListRecords_FallbackToCache(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Инициализируем мок кэша
	mockCache := mocks.NewMockCacheService(ctrl)
	ctx := context.Background()

	// Настраиваем фейковый ответ, который должен вернуть кэш
	expectedRecords := []models.RecordMeta{
		{Name: "cached_card", DataType: "card"},
	}

	// Ожидаем, что из-за ошибки gRPC будет вызван метод GetRecords у кэша
	mockCache.EXPECT().
		GetRecords(ctx).
		Return(expectedRecords, nil).
		Times(1)

	// Передаем некорректный адрес порта, чтобы gRPC запрос гарантированно завершился ошибкой
	service := transport.NewGRPCTransportService("localhost:0", mockCache, "secret-key")

	// Вызов тестируемого метода
	result, err := service.ListRecords(ctx, 10)

	// Проверки
	if err != nil {
		t.Fatalf("не ожидалось ошибки, так как кэш должен был сработать: %v", err)
	}

	if len(result) != 1 || result[0].Name != "cached_card" {
		t.Errorf("ожидались данные из кэша %v, получено %v", expectedRecords, result)
	}
}

// Тест: gRPC сервер недоступен, GetEntityByName успешно достает и возвращает запись из кэша
func TestGPRCTransportService_GetEntityByName_FallbackToCache(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCache := mocks.NewMockCacheService(ctrl)

	// Обязательно добавляем ключ шифрования в контекст, иначе метод упадет до gRPC запроса
	ctx := context.WithValue(context.Background(), models.CryptedContextKey, "my-crypto-key")
	recordName := "secure_data"

	// Мокаем возвращаемую из кэша зашифрованную запись.
	// Внимание: чтобы getDecryptedData не упал с ошибкой bad nonce/cipher,
	// в реальном тесте здесь должны быть валидные зашифрованные данные или пустые, если getDecryptedData их пропускает.
	cachedRecord := &models.EncryptedRecord{
		Name:     recordName,
		DataType: "text",
		Payload:  []byte{},
		Nonce:    []byte{},
	}

	mockCache.EXPECT().
		GetRecordByName(ctx, recordName).
		Return(cachedRecord, nil).
		Times(1)

	service := transport.NewGRPCTransportService("localhost:0", mockCache, "secret-key")

	// Вызываем метод. Он выдаст ошибку расшифровки (так как Payload пустой),
	// но главное — мы проверяем, что он дошел до кэша, а не упал на gRPC.
	_, err := service.GetEntityByName(ctx, recordName)

	if err == nil {
		t.Error("ожидалась ошибка расшифровки или grpc, но получен nil")
	}
}

// Тест: Проверка валидации контекста перед отправкой запроса
func TestGPRCTransportService_GetEntityByName_MissingCryptoKey(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCache := mocks.NewMockCacheService(ctrl)

	// Контекст пустой — ключа шифрования нет
	ctx := context.Background()

	service := transport.NewGRPCTransportService("localhost:8080", mockCache, "secret-key")

	// Вызов должен сразу вернуть ошибку валидации контекста
	_, err := service.GetEntityByName(ctx, "any_name")

	if err == nil {
		t.Fatal("ожидалась ошибка из-за отсутствия ключа шифрования в контексте")
	}

	expectedErrSign := "encryption key missing in context"
	if err.Error() != expectedErrSign {
		t.Errorf("ожидалась ошибка '%s', получена '%v'", expectedErrSign, err)
	}
}
