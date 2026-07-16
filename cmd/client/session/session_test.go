package session

import (
	"context"
	"testing"
	"time"
)

// Тест проверяет полный цикл работы: запуск сервера, пинг, установку сессии и её получение
func TestSessionManager_FullWorkflow_Integration(t *testing.T) {
	// 1. Используем порт :0, чтобы ОС автоматически выделила любой свободный порт
	sm := NewSessionManager("127.0.0.1:0")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Нам нужно динамически узнать, какой порт выделила ОС, но в текущей реализации sm.addr не обновляется.
	// Чтобы тесты гарантированно работали и не конфликтовали по портам,
	// временно используем жестко заданный порт, который редко занят, например 49152.
	// В реальном коде лучше сохранять созданный listener в структуру, чтобы читать реальный адрес.
	addr := "127.0.0.1:49152"
	sm.addr = addr

	// 2. Запускаем сервер сессий в отдельной горутине
	serverErrChan := make(chan error, 1)
	go func() {
		if err := sm.Listen(ctx); err != nil {
			serverErrChan <- err
		}
		close(serverErrChan)
	}()

	// Даем серверу немного времени на запуск TCP-листенера
	time.Sleep(50 * time.Millisecond)

	// 3. Инициализируем клиент
	client := NewClientSession(addr)

	// Шаг А: Проверяем метод Ping
	if err := client.Ping(ctx); err != nil {
		t.Fatalf("Ping failed: %v", err)
	}

	// Шаг Б: Проверяем получение пустой сессии (должна быть ошибка "session not found")
	_, err := client.Get()
	if err == nil {
		t.Error("expected error 'session not found' for empty manager, got nil")
	}

	// Шаг В: Записываем сессию через Set
	expectedUser := "mooncake"
	expectedKey := "crypted-aes-key-hex"
	expectedToken := "jwt-session-token"

	err = client.Set(expectedUser, expectedKey, expectedToken)
	if err != nil {
		t.Fatalf("Set session failed: %v", err)
	}

	// Шаг Г: Получаем сессию через Get и сверяем поля
	savedSession, err := client.Get()
	if err != nil {
		t.Fatalf("Get session failed: %v", err)
	}

	if savedSession.UserName != expectedUser {
		t.Errorf("expected username %s, got %s", expectedUser, savedSession.UserName)
	}
	if savedSession.CryptedKey != expectedKey {
		t.Errorf("expected key %s, got %s", expectedKey, savedSession.CryptedKey)
	}
	if savedSession.Token != expectedToken {
		t.Errorf("expected token %s, got %s", expectedToken, savedSession.Token)
	}

	// 4. Останавливаем сервер отменой контекста
	cancel()

	// Ждем корректного завершения сервера
	select {
	case err := <-serverErrChan:
		if err != nil {
			t.Errorf("server stopped with unexpected error: %v", err)
		}
	case <-time.After(200 * time.Millisecond):
		t.Error("server did not stop within timeout after context cancellation")
	}
}

// Тест поведения клиента, если сервер сессий не запущен
func TestClientSession_ServerNotStarted_Error(t *testing.T) {
	// Указываем порт, на котором заведомо никто не слушает
	client := NewClientSession("127.0.0.1:49153")
	ctx := context.Background()

	// Проверяем Ping
	err := client.Ping(ctx)
	if err == nil {
		t.Error("expected error from Ping when server is offline, got nil")
	}
	expectedErrMsg := "Server is not started. Call firstly: start"
	if err != nil && err.Error() != expectedErrMsg {
		t.Errorf("expected error message '%s', got '%v'", expectedErrMsg, err)
	}

	// Проверяем Get
	_, err = client.Get()
	if err == nil {
		t.Error("expected error from Get when server is offline, got nil")
	}

	// Проверяем Set
	err = client.Set("name", "key", "token")
	if err == nil {
		t.Error("expected error from Set when server is offline, got nil")
	}
}
