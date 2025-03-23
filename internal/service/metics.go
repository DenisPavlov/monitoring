package metrics

import (
	"errors"
	"github.com/DenisPavlov/monitoring/internal/models"
	"github.com/DenisPavlov/monitoring/internal/storage"
	"strconv"
)

var ErrWrongValue = errors.New("wrong value")
var ErrUnknownMetricType = errors.New("unknown metric type")

const (
	GaugeMetricName   = "gauge"
	CounterMetricName = "counter"
)

func Save(metrics *models.Metrics, storage storage.Storage) error {
	switch metrics.MType {
	case GaugeMetricName:
		if metrics.Value == nil {
			return ErrWrongValue
		}
		storage.AddGauge(metrics.ID, *metrics.Value)
		updatedValue, _ := storage.Gauge(metrics.ID)
		*metrics.Value = updatedValue
	case CounterMetricName:
		if metrics.Delta == nil {
			return ErrWrongValue
		}
		storage.AddCounter(metrics.ID, *metrics.Delta)
		updatedValue, _ := storage.Counter(metrics.ID)
		*metrics.Delta = updatedValue
	default:
		return ErrUnknownMetricType
	}
	return nil
}

func Get(id string, mType string, storage storage.Storage) *models.Metrics {

	var metrics = models.Metrics{
		ID:    id,
		MType: mType,
	}

	switch mType {
	case GaugeMetricName:
		metricValue, ok := storage.Gauge(id)
		if ok {
			metrics.Value = &metricValue
		} else {
			return nil
		}
	case CounterMetricName:
		metricValue, ok := storage.Counter(id)
		if ok {
			metrics.Delta = &metricValue
		} else {
			return nil
		}
	default:
		return nil
	}

	return &metrics
}

func CreateMetrics(id string, mType string, mValue string) (*models.Metrics, error) {
	var metrics = models.Metrics{
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
		return nil, ErrUnknownMetricType
	}
	return &metrics, nil
}
