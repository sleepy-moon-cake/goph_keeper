package repository

import (
	"context"
	"errors"
	"fmt"
	"goph_keeper/internal/server/db"
	"goph_keeper/internal/server/interfaces"
	"goph_keeper/internal/shared/models"
	"math"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type SQLRepository struct {
	queries *db.Queries
}

func NewSQLRepository(queries *db.Queries) *SQLRepository {
	return &SQLRepository{queries: queries}
}

func (r *SQLRepository) AddUser(ctx context.Context, username, passwordHash string) error {
	_, err := r.queries.CreateUser(ctx, db.CreateUserParams{
		UserName:     username,
		PasswordHash: passwordHash,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return interfaces.ErrUserAlreadyExists
			}
		}

		return fmt.Errorf("repo: failed to add user: %w", err)
	}
	return nil
}

func (r *SQLRepository) GetUserPassword(ctx context.Context, username string) (string, error) {
	user, err := r.queries.GetUserByUsername(ctx, username)

	if errors.Is(err, pgx.ErrNoRows) {
		return "", interfaces.ErrUserNotFound
	}

	if err != nil {
		return "", fmt.Errorf("repo: failed to get user password: %w", err)
	}

	return user.PasswordHash, nil
}

func (r *SQLRepository) SaveRecord(ctx context.Context, username string, record models.EncryptedRecord) error {
	_, err := r.queries.CreateRecord(ctx, db.CreateRecordParams{
		UserName:   username,
		RecordName: record.Name,
		DataType:   record.DataType,
		Payload:    record.Payload,
		Nonce:      record.Nonce,
	})
	if err != nil {
		return fmt.Errorf("repo: failed to save record: %w", err)
	}
	return nil
}

func (r *SQLRepository) GetRecord(ctx context.Context, username, name string) (models.EncryptedRecord, error) {
	row, err := r.queries.GetRecordByUniqueKey(ctx, db.GetRecordByUniqueKeyParams{
		UserName:   username,
		RecordName: name,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return models.EncryptedRecord{}, fmt.Errorf("repo: record not found")
	}
	if err != nil {
		return models.EncryptedRecord{}, fmt.Errorf("repo: failed to get record: %w", err)
	}

	return models.EncryptedRecord{
		Name:     row.RecordName,
		DataType: row.DataType,
		Payload:  row.Payload,
		Nonce:    row.Nonce,
	}, nil
}

func (r *SQLRepository) DeleteRecord(ctx context.Context, username, name string) error {
	err := r.queries.DeleteRecord(ctx, db.DeleteRecordParams{
		UserName:   username,
		RecordName: name,
	})
	if err != nil {
		return fmt.Errorf("repo: failed to delete record: %w", err)
	}
	return nil
}

func (r *SQLRepository) ListRecords(ctx context.Context, username string, limit int32) ([]models.RecordMeta, error) {
	dbLimit := limit
	if dbLimit <= 0 {
		dbLimit = math.MaxInt32
	}

	rows, err := r.queries.GetAllRecordsByUsername(ctx, db.GetAllRecordsByUsernameParams{
		UserName: username,
		Limit:    limit,
	})
	if err != nil {
		return nil, fmt.Errorf("repo: failed to list records: %w", err)
	}

	result := make([]models.RecordMeta, 0, len(rows))
	for _, row := range rows {
		result = append(result, models.RecordMeta{
			Name:     row.RecordName,
			DataType: row.DataType,
		})
	}

	return result, nil
}
