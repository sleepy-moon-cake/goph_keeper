package cache

import (
	"errors"
	"testing"

	mocks "goph_keeper/internal/client/interfaces/gen" // Проверьте правильность пути к папке gen
	"goph_keeper/internal/shared/models"

	"go.uber.org/mock/gomock"
)

// Тест успешного получения записи по имени
func TestCacheService_GetRecordByName_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCache := mocks.NewMockCacheService(ctrl)
	ctx := t.Context()
	recordName := "secret_key"

	expectedRecord := &models.EncryptedRecord{
		// Заполните поля вашей структуры models.EncryptedRecord для теста
	}

	// Настраиваем ожидание: при вызове с recordName вернуть объект и nil
	mockCache.EXPECT().
		GetRecordByName(ctx, recordName).
		Return(expectedRecord, nil).
		Times(1)

	// Вызов тестируемого метода
	res, err := mockCache.GetRecordByName(ctx, recordName)

	// Проверки
	if err != nil {
		t.Fatalf("ожидалась ошибка nil, получена: %v", err)
	}
	if res != expectedRecord {
		t.Errorf("ожидался рекорд %v, получен %v", expectedRecord, res)
	}
}

// Тест возврата ошибки при получении записи
func TestCacheService_GetRecordByName_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCache := mocks.NewMockCacheService(ctrl)
	ctx := t.Context()
	expectedErr := errors.New("cache miss")

	// Используем gomock.Any(), чтобы сработало на любую строку
	mockCache.EXPECT().
		GetRecordByName(ctx, gomock.Any()).
		Return(nil, expectedErr).
		Times(1)

	_, err := mockCache.GetRecordByName(ctx, "any_name")

	if !errors.Is(err, expectedErr) {
		t.Errorf("ожидалась ошибка %v, получена %v", expectedErr, err)
	}
}

// Тест обновления списка записей
func TestCacheService_UpdateRecords(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCache := mocks.NewMockCacheService(ctrl)
	ctx := t.Context()

	fakeRecords := []models.RecordMeta{
		// Добавьте фейковые данные если необходимо
	}

	mockCache.EXPECT().
		UpdateRecords(ctx, fakeRecords).
		Return(nil).
		Times(1)

	err := mockCache.UpdateRecords(ctx, fakeRecords)
	if err != nil {
		t.Errorf("не ожидалось ошибки, получена: %v", err)
	}
}
