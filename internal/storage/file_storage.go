package storage

import (
	"encoding/json"
	"github.com/DenisPavlov/monitoring/internal/logger"
	"os"
)

type FileStorage struct {
	Gauges         map[string]float64
	Counters       map[string]int64
	needToSaveSync bool
	filename       string
}

func NewFileStorage(needToSaveSync bool, filename string) *FileStorage {
	return &FileStorage{
		Gauges:         make(map[string]float64),
		Counters:       make(map[string]int64),
		needToSaveSync: needToSaveSync,
		filename:       filename,
	}
}

func InitFromFile(needToSaveSync bool, filename string) (*FileStorage, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	storage := NewFileStorage(needToSaveSync, filename)
	err = json.Unmarshal(data, storage)
	if err != nil {
		return nil, err
	}
	return storage, nil
}

func (s *FileStorage) SaveToFile() error {
	data, err := json.MarshalIndent(s, "", "   ")
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

func (s *FileStorage) AddGauge(name string, value float64) error {
	s.Gauges[name] = value
	if s.needToSaveSync {
		err := s.SaveToFile()
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *FileStorage) AddCounter(name string, value int64) error {
	s.Counters[name] = s.Counters[name] + value
	if s.needToSaveSync {
		err := s.SaveToFile()
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *FileStorage) Gauge(name string) (float64, bool) {
	i, ok := s.Gauges[name]
	return i, ok
}

func (s *FileStorage) Counter(name string) (int64, bool) {
	i, ok := s.Counters[name]
	return i, ok
}

func (s *FileStorage) AllGauges() map[string]float64 {
	return s.Gauges
}

func (s *FileStorage) AllCounters() map[string]int64 {
	return s.Counters
}
