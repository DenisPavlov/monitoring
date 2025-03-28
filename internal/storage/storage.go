package storage

import (
	"encoding/json"
	"github.com/DenisPavlov/monitoring/internal/logger"
	"os"
)

type Storage interface {
	AddGauge(name string, value float64) error
	AddCounter(name string, value int64) error
	Gauge(name string) (float64, bool)
	Counter(name string) (int64, bool)
	AllGauges() map[string]float64
	AllCounters() map[string]int64
}

type MemStorage struct {
	Gauges         map[string]float64
	Counters       map[string]int64
	needToSaveSync bool
	filename       string
}

func NewMemStorage(needToSaveSync bool, filename string) *MemStorage {
	return &MemStorage{
		Gauges:         make(map[string]float64),
		Counters:       make(map[string]int64),
		needToSaveSync: needToSaveSync,
		filename:       filename,
	}
}

func LoadFromFile(needToSaveSync bool, filename string) (*MemStorage, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	storage := NewMemStorage(needToSaveSync, filename)
	err = json.Unmarshal(data, storage)
	if err != nil {
		return nil, err
	}
	return storage, nil
}

func (s *MemStorage) AddGauge(name string, value float64) error {
	s.Gauges[name] = value
	if s.needToSaveSync {
		err := SaveToFile(s.filename, s)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *MemStorage) AddCounter(name string, value int64) error {
	s.Counters[name] = s.Counters[name] + value
	if s.needToSaveSync {
		err := SaveToFile(s.filename, s)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *MemStorage) Gauge(name string) (float64, bool) {
	i, ok := s.Gauges[name]
	return i, ok
}

func (s *MemStorage) Counter(name string) (int64, bool) {
	i, ok := s.Counters[name]
	return i, ok
}

func (s *MemStorage) AllGauges() map[string]float64 {
	return s.Gauges
}

func (s *MemStorage) AllCounters() map[string]int64 {
	return s.Counters
}

func SaveToFile(fname string, storage interface{}) error {
	data, err := json.MarshalIndent(storage, "", "   ")
	if err != nil {
		logger.Log.Error("cannot create byte data from storage", err)
		return err
	}
	err = os.WriteFile(fname, data, 0666)
	if err != nil {
		logger.Log.Error("cannot save to file", err)
		return err
	}
	return os.WriteFile(fname, data, 0666)
}
