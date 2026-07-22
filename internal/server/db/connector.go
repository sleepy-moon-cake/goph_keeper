package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewConnector(ctx context.Context, connStr string) (*pgxpool.Pool, error) {
	ctxWithTime, cancel := context.WithTimeout(ctx, 5*time.Second)

	defer cancel()

	pool, err := pgxpool.New(ctxWithTime, connStr)

	if err != nil {
		return nil, fmt.Errorf("connection error: %w", err)
	}

	if err := pool.Ping(ctxWithTime); err != nil {
		pool.Close()
		return nil, fmt.Errorf("data ping error: %w", err)
	}
	return pool, nil
}
