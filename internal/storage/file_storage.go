package storage

import (
	"context"
	"encoding/json"
	"github.com/DenisPavlov/monitoring/internal/logger"
	"github.com/DenisPavlov/monitoring/internal/models"
	"os"
)

type FileStorage struct {
	*MemStorage
	needToSaveSync bool
	filename       string
}

// jsonMetrics need to safe metrics to json file
type jsonMetrics struct {
	Metrics map[string]models.Metrics
}

func NewFileStorage(needToSaveSync bool, filename string) *FileStorage {
	return &FileStorage{
		MemStorage:     NewMemStorage(),
		needToSaveSync: needToSaveSync,
		filename:       filename,
	}
}

func InitFromFile(needToSaveSync bool, filename string) (*FileStorage, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	jsonMetrics := jsonMetrics{}
	err = json.Unmarshal(data, &jsonMetrics)
	if err != nil {
		return nil, err
	}

	storage := FileStorage{
		MemStorage:     &MemStorage{metrics: jsonMetrics.Metrics},
		needToSaveSync: needToSaveSync,
		filename:       filename,
	}
	return &storage, nil
}

func (s *FileStorage) SaveToFile() error {
	data, err := json.MarshalIndent(jsonMetrics{Metrics: s.metrics}, "", "   ")
	if err != nil {
		logger.Log.Error("cannot create byte data from storage", err)
		return err
	}
	err = os.WriteFile(s.filename, data, 0666)
	if err != nil {
		logger.Log.Error("cannot save to file", err)
		return err
	}
	return os.WriteFile(s.filename, data, 0666)
}

func (s *FileStorage) Save(ctx context.Context, metric *models.Metrics) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		err := s.MemStorage.Save(ctx, metric)
		if err != nil {
			return err
		}
		if s.needToSaveSync {
			err := s.SaveToFile()
			if err != nil {
				return err
			}
		}
	}
	return nil
}
