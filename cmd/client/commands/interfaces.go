package commands

import (
	"context"
	"goph_keeper/internal/shared/models"
)

type Service interface {
	SaveText(ctx context.Context, data models.TextData) error
	SaveCard(ctx context.Context, data models.CardData) error
	SaveFile(ctx context.Context, data models.BinaryData) error
}
