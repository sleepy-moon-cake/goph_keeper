package interfaces

import (
	"context"
	"goph_keeper/internal/shared/models"
)

//go:generate mockgen -source=cache.go -destination=gen/mock_cache.go -package=mocks
type CacheService interface {
	GetRecords(ctx context.Context) ([]models.RecordMeta, error)
	UpdateRecords(ctx context.Context, records []models.RecordMeta) error

	GetRecordByName(ctx context.Context, name string) (*models.EncryptedRecord, error)
	UpdateSingleRecord(ctx context.Context, record *models.EncryptedRecord) error
}
