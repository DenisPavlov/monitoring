package storage

import (
	"context"
	"github.com/DenisPavlov/monitoring/internal/models"
)

type MetricsStorage interface {
	Save(ctx context.Context, metric *models.Metric) error
	SaveAll(ctx context.Context, metrics []models.Metric) error
	GetByTypeAndID(ctx context.Context, ID, mType string) (models.Metric, error)
	GetAllByType(ctx context.Context, mType string) ([]models.Metric, error)
}
