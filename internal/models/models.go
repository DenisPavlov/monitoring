package models

import (
	"errors"
	"strconv"
)

const (
	GaugeMetricName   = "gauge"
	CounterMetricName = "counter"
)

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

func CreateMetrics(id string, mType string, mValue string) (*Metrics, error) {
	var metrics = Metrics{
		ID:    id,
		MType: mType,
	}

	switch mType {
	case GaugeMetricName:
		value, err := strconv.ParseFloat(mValue, 64)
		if err != nil {
			return nil, err
		}
		metrics.Value = &value
	case CounterMetricName:
		value, err := strconv.ParseInt(mValue, 10, 64)
		if err != nil {
			return nil, err
		}
		metrics.Delta = &value
	default:
		return nil, errors.New("invalid metric type")
	}
	return &metrics, nil
}
