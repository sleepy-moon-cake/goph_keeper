package interfaces

import (
	"context"
	"goph_keeper/internal/shared/models"
)

type CacheService interface {
	GetRecords(ctx context.Context) ([]models.RecordMeta, error)
	UpdateRecords(ctx context.Context, records []models.RecordMeta) error

	GetRecordByName(ctx context.Context, name string) (*models.EncryptedRecord, error)
	UpdateSingleRecord(ctx context.Context, record *models.EncryptedRecord) error
}
