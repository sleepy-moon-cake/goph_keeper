package interfaces

import (
	"context"
	"errors"
	"goph_keeper/internal/shared/models"
)

//go:generate mockgen -source=repository_db.go -destination=gen/mock_repository_db.go -package=mocks
type RepositoryDb interface {
	AddUser(ctx context.Context, username, passwordHash string) error
	GetUserPassword(ctx context.Context, username string) (string, error)
	SaveRecord(ctx context.Context, username string, record models.EncryptedRecord) error
	GetRecord(ctx context.Context, username, name string) (models.EncryptedRecord, error)
	DeleteRecord(ctx context.Context, username, name string) error
	ListRecords(ctx context.Context, username string, limit int32) ([]models.RecordMeta, error)
}

var (
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrUserNotFound      = errors.New("user not found")
	ErrRecordNotFound    = errors.New("record not found")
)
