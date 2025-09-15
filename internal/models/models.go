// Package models defines data structures and utility functions
// for handling metrics in the monitoring system.
package models

import (
	"errors"
	"strconv"
)

// Constants defining supported metric types.
const (
	// GaugeMetricName represents a gauge metric type that stores float values.
	// Gauges are used for measurements that can go up and down (e.g., temperature, memory usage).
	GaugeMetricName = "gauge"

	// CounterMetricName represents a counter metric type that stores integer values.
	// Counters are used for measurements that only increase (e.g., request count, errors).
	CounterMetricName = "counter"
)

// Metric represents a monitoring metric with its metadata and value.
//
// The struct uses pointer fields for Delta and Value to distinguish between
// zero values and unset values in JSON serialization/deserialization.
//
// JSON tags are provided for compatibility with REST API operations.
type Metric struct {
	// Delta is the value for counter metrics. Only present when MType is "counter".
	// The pointer allows for proper JSON omitempty behavior.
	Delta *int64 `json:"delta,omitempty"`

	// Value is the value for gauge metrics. Only present when MType is "gauge".
	// The pointer allows for proper JSON omitempty behavior.
	Value *float64 `json:"value,omitempty"`

	// ID is the unique name identifier of the metric.
	// Example: "cpu_usage", "memory_used", "request_count"
	ID string `json:"id"`

	// MType is the metric type, either "gauge" or "counter".
	MType string `json:"type"`
}

// CreateMetric creates a new Metric instance from string parameters.
//
// This function validates the metric type and parses the value string into
// the appropriate numeric type (float64 for gauge, int64 for counter).
//
// Parameters:
//   - id: Metric name identifier (e.g., "cpu_usage")
//   - mType: Metric type, must be "gauge" or "counter"
//   - mValue: String representation of the metric value
//
// Returns:
//   - *Metric: Pointer to the created Metric instance
//   - error: Validation or parsing error if any parameter is invalid
//
// Example usage:
//
//	// Create a gauge metric
//	gauge, err := CreateMetric("temperature", "gauge", "23.5")
//
//	// Create a counter metric
//	counter, err := CreateMetric("requests", "counter", "42")
//
// Possible errors:
//   - Invalid metric type (not "gauge" or "counter")
//   - Invalid numeric format for the value
//   - Value out of range for the target type
func CreateMetric(id string, mType string, mValue string) (*Metric, error) {
	var metrics = Metric{
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
