package metrics

//import (
//	"errors"
//	"github.com/DenisPavlov/monitoring/internal/models"
//	"github.com/DenisPavlov/monitoring/internal/storage"
//	"strconv"
//)
//
//// todo - вынести эти константы
//var ErrWrongValue = errors.New("wrong value")
//var ErrUnknownMetricType = errors.New("unknown metric type")
//
//func Save(metrics *models.Metrics, storage storage.Storage) error {
//	switch metrics.MType {
//	case GaugeMetricName:
//		if metrics.Value == nil {
//			return ErrWrongValue
//		}
//		if err := storage.AddGauge(metrics.ID, *metrics.Value); err != nil {
//			return err
//		}
//		updatedValue, _ := storage.Gauge(metrics.ID)
//		*metrics.Value = updatedValue
//	case CounterMetricName:
//		if metrics.Delta == nil {
//			return ErrWrongValue
//		}
//		if err := storage.AddCounter(metrics.ID, *metrics.Delta); err != nil {
//			return err
//		}
//		updatedValue, _ := storage.Counter(metrics.ID)
//		*metrics.Delta = updatedValue
//	default:
//		return ErrUnknownMetricType
//	}
//	return nil
//}
//
//func Get(id string, mType string, storage storage.Storage) *models.Metrics {
//
//	var metrics = models.Metrics{
//		ID:    id,
//		MType: mType,
//	}
//
//	switch mType {
//	case GaugeMetricName:
//		metricValue, ok := storage.Gauge(id)
//		if ok {
//			metrics.Value = &metricValue
//		} else {
//			return nil
//		}
//	case CounterMetricName:
//		metricValue, ok := storage.Counter(id)
//		if ok {
//			metrics.Delta = &metricValue
//		} else {
//			return nil
//		}
//	default:
//		return nil
//	}
//
//	return &metrics
//}
//
