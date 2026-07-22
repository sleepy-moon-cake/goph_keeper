package start

import (
	"bytes"
	"context"
	"goph_keeper/cmd/client/session"
	"strings"
	"testing"
	"time"
)

// Тест проверяет, что команда успешно запускается, реагирует на контекст и логирует старт
func TestNewStartServerCmd_Success_Workflow(t *testing.T) {
	// 1. Создаем реальный менеджер сессий на случайном порту :0 (чтобы не занимать порты в системе)
	sm := session.NewSessionManager("127.0.0.1:0")

	// 2. Создаем контекст, который мы отменим сразу после запуска, чтобы выйти из бесконечного цикла Listen
	ctx, cancel := context.WithCancel(context.Background())

	// Инициализируем команду Cobra
	cmd := NewStartServerCmd(sm)

	// Перенаправляем стандартный вывод, чтобы тест не мусорил в консоли
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetContext(ctx)
	cmd.SetArgs([]string{}) // аргументы не требуются

	// Флаг для фиксации результата выполнения горутины
	errChan := make(chan error, 1)

	// 3. Запускаем выполнение команды в отдельной горутине, так как Listen блокирует поток
	go func() {
		errChan <- cmd.Execute()
	}()

	// Даем команде 50 миллисекунд, чтобы slog.Info("Запуск сервера сессий...") и net.Listen отработали
	time.Sleep(50 * time.Millisecond)

	// 4. Отменяем контекст. Это заставит sm.Listen завершиться без ошибки (return nil)
	cancel()

	// Ждем результат выполнения команды
	select {
	case err := <-errChan:
		if err != nil {
			t.Fatalf("Команда завершилась с неожиданной ошибкой: %v", err)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Команда зависла и не завершилась после отмены контекста")
	}
}

// Тест проверяет поведение команды при ошибке старта (например, невалидный адрес)
func TestNewStartServerCmd_ListenError(t *testing.T) {
	// Передаем заведомо некорректный адрес, который net.Listen не сможет распарсить
	invalidAddr := "999.999.999.999:99999"
	sm := session.NewSessionManager(invalidAddr)

	cmd := NewStartServerCmd(sm)

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{})

	// Выполняем синхронно, так как при ошибке net.Listen метод упадет мгновенно без блокировки
	err := cmd.Execute()

	if err == nil {
		t.Fatal("Ожидалась ошибка сервера из-за некорректного адреса, но получен nil")
	}

	// Проверяем, что ошибка обернута так, как написано в вашем RunE: "server error: ..."
	if !strings.Contains(err.Error(), "server error") {
		t.Errorf("Ожидался префикс ошибки 'server error', получено: %v", err)
	}
}
