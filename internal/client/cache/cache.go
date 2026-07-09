package cache

import (
	"context"
	"database/sql"
	"fmt"
	"goph_keeper/internal/client/db"
	"goph_keeper/internal/shared/models"
	"log/slog"
)

type CacheService struct {
	queries *db.Queries
}

func NewCacheService(sqliteDB *sql.DB) *CacheService {
	return &CacheService{
		queries: db.New(sqliteDB),
	}
}

func (c *CacheService) GetRecords(ctx context.Context) ([]models.RecordMeta, error) {
	userName, ok := ctx.Value(models.UserContextKey).(string)
	if !ok || userName == "" {
		return nil, fmt.Errorf("cache: user name missing in context")
	}

	cacheResp, err := c.queries.GetAllRecordsByUsername(ctx, userName)
	if err != nil {
		return nil, fmt.Errorf("cache: failed to read from sqlite: %w", err)
	}

	records := make([]models.RecordMeta, 0, len(cacheResp))
	for _, v := range cacheResp {
		records = append(records, models.RecordMeta{
			Name:     v.RecordName,
			DataType: v.DataType,
		})
	}

	return records, nil
}

func (c *CacheService) UpdateRecords(ctx context.Context, records []models.RecordMeta) error {
	userName, ok := ctx.Value(models.UserContextKey).(string)
	if !ok || userName == "" {
		return fmt.Errorf("cache: user name missing in context")
	}

	for _, v := range records {
		_, err := c.queries.SaveRecord(ctx, db.SaveRecordParams{
			UserName:   userName,
			RecordName: v.Name,
			SyncStatus: "synced",
		})
		if err != nil {
			slog.Error("cache: failed to save record", "name", v.Name, "error", err)
		}
	}

	return nil
}

func (c *CacheService) GetRecordByName(ctx context.Context, name string) (*models.EncryptedRecord, error) {
	userName, ok := ctx.Value(models.UserContextKey).(string)
	if !ok || userName == "" {
		return nil, fmt.Errorf("cache: user name missing in context")
	}

	v, err := c.queries.GetRecordByUniqueKey(ctx, db.GetRecordByUniqueKeyParams{
		UserName:   userName,
		RecordName: name,
	})
	if err != nil {
		return nil, fmt.Errorf("cache: failed to read single record: %w", err)
	}

	return &models.EncryptedRecord{
		Name:     v.RecordName,
		DataType: v.DataType,
		Payload:  v.Payload,
		Nonce:    v.Nonce,
	}, nil
}

func (c *CacheService) UpdateSingleRecord(ctx context.Context, record *models.EncryptedRecord) error {
	userName, ok := ctx.Value(models.UserContextKey).(string)
	if !ok || userName == "" {
		return fmt.Errorf("cache: user name missing in context")
	}

	_, err := c.queries.SaveRecord(ctx, db.SaveRecordParams{
		UserName:   userName,
		RecordName: record.Name,
		DataType:   record.DataType,
		Payload:    record.Payload,
		Nonce:      record.Nonce,
		SyncStatus: "synced",
	})

	if err != nil {
		return fmt.Errorf("cache: failed to update single record: %w", err)
	}

	return nil
}
