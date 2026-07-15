package migrations

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed *.sql
var MigrationsFS embed.FS

func RunMigrations(db *sql.DB) error {
	sourceDriver, err := iofs.New(MigrationsFS, ".")
	if err != nil {
		slog.Error("SourceDriver error")
		return err
	}

	// Создаем драйвер для SQLite3 на основе существующего *sql.DB подключения
	dbDriver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		slog.Error("dbDriver error")
		return err
	}

	// Меняем схему "pgx" на "sqlite3"
	m, err := migrate.NewWithInstance(
		"iofs",
		sourceDriver,
		"sqlite3",
		dbDriver,
	)
	if err != nil {
		slog.Error("Migration error")
		return err
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			return nil
		}
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	return nil
}
