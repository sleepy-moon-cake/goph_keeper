package db

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func NewConnector(ctx context.Context) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "client_db.db")

	if err != nil {
		return nil, fmt.Errorf("client db: %w", err)
	}

	ctwWithTime, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctwWithTime); err != nil {
		db.Close()
		return nil, fmt.Errorf("client db: ping: %w", err)
	}

	slog.Info("HAS CONNECTED TO INNER CACHE DB")

	return db, nil
}
