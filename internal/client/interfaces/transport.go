package interfaces

import (
	"context"
	"goph_keeper/internal/shared/models"
)

type TransportService interface {
	SaveText(ctx context.Context, data models.TextData) error
	SaveCard(ctx context.Context, data models.CardData) error
	SaveFile(ctx context.Context, data models.BinaryData) error
	DeleteEntityByName(ctx context.Context, name string) error
	GetEntityByName(ctx context.Context, name string) (*models.EncryptedRecord, error)
	ListRecords(ctx context.Context, limit int) ([]models.RecordMeta, error)
	Login(ctx context.Context, name string, password string) (string, error)
	Register(ctx context.Context, name string, password string) (string, error)
}
