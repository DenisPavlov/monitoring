package storage

import (
	"context"
	"github.com/DenisPavlov/monitoring/internal/models"
)

type Storage interface {
	Save(ctx context.Context, metric *models.Metrics) error
	SaveAll(ctx context.Context, metrics []models.Metrics) error
	GetByTypeAndID(ctx context.Context, ID, mType string) (models.Metrics, error)
	GetAllByType(ctx context.Context, mType string) ([]models.Metrics, error)
}
