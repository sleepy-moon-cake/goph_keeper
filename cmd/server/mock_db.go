package main

import (
	"context"
	"errors"
	"strings"
	"sync"

	"goph_keeper/internal/shared/models"
)

// Глобальные ошибки для имитации БД
var (
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrUserNotFound      = errors.New("user not found")
	ErrRecordNotFound    = errors.New("record not found")
)

// MockMemoryStorage — это наша временная БД в памяти
type MockMemoryStorage struct {
	mu      sync.RWMutex
	users   map[string]string                 // username -> passwordHash
	storage map[string]models.EncryptedRecord // "username:record_name" -> record
}

func NewMockMemoryStorage() *MockMemoryStorage {
	return &MockMemoryStorage{
		users:   make(map[string]string),
		storage: make(map[string]models.EncryptedRecord),
	}
}

// 1. Имитация регистрации
func (m *MockMemoryStorage) AddUser(ctx context.Context, username, passwordHash string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.users[username]; exists {
		return ErrUserAlreadyExists
	}

	m.users[username] = passwordHash
	return nil
}

// 2. Имитация проверки пароля при входе
func (m *MockMemoryStorage) GetUserPassword(ctx context.Context, username string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	passwordHash, exists := m.users[username]
	if !exists {
		return "", ErrUserNotFound
	}

	return passwordHash, nil
}

// 3. Имитация сохранения секрета
func (m *MockMemoryStorage) SaveRecord(ctx context.Context, username string, record models.EncryptedRecord) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	userKey := username + ":" + record.Name
	m.storage[userKey] = record
	return nil
}

// 4. Имитация получения секрета
func (m *MockMemoryStorage) GetRecord(ctx context.Context, username, name string) (models.EncryptedRecord, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	userKey := username + ":" + name
	record, exists := m.storage[userKey]
	if !exists {
		return models.EncryptedRecord{}, ErrRecordNotFound
	}

	return record, nil
}

// 5. Имитация удаления
func (m *MockMemoryStorage) DeleteRecord(ctx context.Context, username, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	userKey := username + ":" + name
	if _, exists := m.storage[userKey]; !exists {
		return ErrRecordNotFound
	}

	delete(m.storage, userKey)
	return nil
}

// 6. Имитация получения списка
func (m *MockMemoryStorage) ListRecords(ctx context.Context, username string, limit int32) ([]models.RecordMeta, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []models.RecordMeta
	var count int32

	prefix := username + ":"
	for key, record := range m.storage {
		if strings.HasPrefix(key, prefix) {
			if limit > 0 && count >= limit {
				break
			}

			result = append(result, models.RecordMeta{
				Name:     record.Name,
				DataType: record.DataType,
			})
			count++
		}
	}

	return result, nil
}
