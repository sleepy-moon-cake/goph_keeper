package migrations

import (
	"database/sql"
	"testing"
)

// Тест успешного наката миграций на чистую БД в памяти
func TestRunMigrations_Success(t *testing.T) {
	// 1. Открываем временную БД в оперативной памяти
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open sqlite in-memory db: %v", err)
	}
	defer db.Close()

	// 2. Запускаем миграции первый раз (накат таблиц)
	err = RunMigrations(db)
	if err != nil {
		t.Fatalf("expected no error during migrations up, got: %v", err)
	}

	// 3. Проверяем, что таблица из SQL-файла действительно создалась
	var tableName string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='users';").Scan(&tableName)
	if err != nil {
		t.Fatalf("failed to query sqlite_master, table 'users' probably not created: %v", err)
	}

	if tableName != "users" {
		t.Errorf("expected table 'users' to exist, got '%s'", tableName)
	}

	// 4. Запускаем повторно, чтобы проверить обработку migrate.ErrNoChange
	err = RunMigrations(db)
	if err != nil {
		t.Errorf("expected no error when running migrations on already updated DB, got: %v", err)
	}
}

// Тест поведения при передаче закрытого подключения
func TestRunMigrations_ClosedDB_Error(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	// Сразу закрываем соединение
	db.Close()

	// Вызов должен завершиться ошибкой драйвера
	err = RunMigrations(db)
	if err == nil {
		t.Error("expected error when running migrations on a closed database connection, got nil")
	}
}
