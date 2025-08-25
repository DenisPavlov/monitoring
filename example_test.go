package handler_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DenisPavlov/monitoring/internal/handler"
	"github.com/DenisPavlov/monitoring/internal/models"
)

// MockStorage implements MetricsStorage interface for testing purposes
type MockStorage struct {
	metrics map[string]models.Metric
}

func NewMockStorage() *MockStorage {
	return &MockStorage{
		metrics: make(map[string]models.Metric),
	}
}

func (m *MockStorage) Save(ctx context.Context, metric *models.Metric) error {
	key := fmt.Sprintf("%s:%s", metric.MType, metric.ID)
	m.metrics[key] = *metric
	return nil
}

func (m *MockStorage) SaveAll(ctx context.Context, metrics []models.Metric) error {
	var err error
	for _, metric := range metrics {
		err = m.Save(ctx, &metric)
	}
	return err
}

func (m *MockStorage) GetByTypeAndID(ctx context.Context, id, mType string) (models.Metric, error) {
	key := fmt.Sprintf("%s:%s", mType, id)
	metric, exists := m.metrics[key]
	if !exists {
		return models.Metric{}, nil
	}
	return metric, nil
}

func (m *MockStorage) GetAllByType(ctx context.Context, mType string) ([]models.Metric, error) {
	var result []models.Metric
	for key, metric := range m.metrics {
		if len(key) > len(mType) && key[:len(mType)] == mType {
			result = append(result, metric)
		}
	}
	return result, nil
}

func (m *MockStorage) Ping(ctx context.Context) error {
	return nil
}

// Example tests for endpoint handlers
func Example_updateMetricHandler() {
	mockStorage := NewMockStorage()
	db := &sql.DB{}
	router := handler.BuildRouter(mockStorage, db, "")

	gaugeMetric := models.Metric{
		ID:    "test_gauge",
		MType: "gauge",
		Value: func() *float64 { v := 123.45; return &v }(),
	}

	body, _ := json.Marshal(gaugeMetric)
	req := httptest.NewRequest("POST", "/update/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	resp := w.Result()
	defer func() {
		_ = resp.Body.Close()
	}()

	fmt.Printf("Status: %d\n", resp.StatusCode)
	fmt.Printf("Content-Type: %s\n", resp.Header.Get("Content-Type"))

	// Output:
	// Status: 200
	// Content-Type: application/json
}

func Example_saveMetricsHandler() {
	mockStorage := NewMockStorage()
	db := &sql.DB{}
	router := handler.BuildRouter(mockStorage, db, "")

	req := httptest.NewRequest("POST", "/update/gauge/test_metric/42.5", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	resp := w.Result()
	defer func() {
		_ = resp.Body.Close()
	}()

	fmt.Printf("Status: %d\n", resp.StatusCode)

	metric, _ := mockStorage.GetByTypeAndID(context.Background(), "test_metric", "gauge")
	fmt.Printf("Metric saved: %s = %f\n", metric.ID, *metric.Value)

	// Output:
	// Status: 200
	// Metric saved: test_metric = 42.500000
}

func Example_getMetricHandler() {
	mockStorage := NewMockStorage()
	db := &sql.DB{}
	router := handler.BuildRouter(mockStorage, db, "")

	testValue := 99.9
	_ = mockStorage.Save(context.Background(), &models.Metric{
		ID:    "temperature",
		MType: "gauge",
		Value: &testValue,
	})

	req := httptest.NewRequest("GET", "/value/gauge/temperature", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	resp := w.Result()
	defer func() {
		_ = resp.Body.Close()
	}()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("Status: %d\n", resp.StatusCode)
	fmt.Printf("Response: %s\n", string(body))

	// Output:
	// Status: 200
	// Response: 99.9
}

func Example_getJSONMetricHandler() {
	mockStorage := NewMockStorage()
	db := &sql.DB{}
	router := handler.BuildRouter(mockStorage, db, "")

	testDelta := int64(100)
	_ = mockStorage.Save(context.Background(), &models.Metric{
		ID:    "requests",
		MType: "counter",
		Delta: &testDelta,
	})

	requestBody := map[string]string{
		"id":   "requests",
		"type": "counter",
	}
	body, _ := json.Marshal(requestBody)

	req := httptest.NewRequest("POST", "/value/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	resp := w.Result()
	defer func() {
		_ = resp.Body.Close()
	}()

	var responseMetric models.Metric
	_ = json.NewDecoder(resp.Body).Decode(&responseMetric)

	fmt.Printf("Status: %d\n", resp.StatusCode)
	fmt.Printf("Metric: %s = %d\n", responseMetric.ID, *(responseMetric.Delta))

	// Output:
	// Status: 200
	// Metric: requests = 100
}

func Example_updatesHandler() {
	mockStorage := NewMockStorage()
	db := &sql.DB{}
	router := handler.BuildRouter(mockStorage, db, "")

	metrics := []models.Metric{
		{
			ID:    "cpu_usage",
			MType: "gauge",
			Value: func() *float64 { v := 75.3; return &v }(),
		},
		{
			ID:    "memory_used",
			MType: "gauge",
			Value: func() *float64 { v := 2048.0; return &v }(),
		},
	}

	body, _ := json.Marshal(metrics)
	req := httptest.NewRequest("POST", "/updates/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	resp := w.Result()
	defer func() {
		_ = resp.Body.Close()
	}()

	cpuMetric, _ := mockStorage.GetByTypeAndID(context.Background(), "cpu_usage", "gauge")
	memoryMetric, _ := mockStorage.GetByTypeAndID(context.Background(), "memory_used", "gauge")

	fmt.Printf("Status: %d\n", resp.StatusCode)
	fmt.Printf("CPU: %f\n", *cpuMetric.Value)
	fmt.Printf("Memory: %f\n", *memoryMetric.Value)

	// Output:
	// Status: 200
	// CPU: 75.300000
	// Memory: 2048.000000
}

func Example_getAllMetricsHandler() {
	mockStorage := NewMockStorage()
	db := &sql.DB{}
	router := handler.BuildRouter(mockStorage, db, "")

	cpuValue := 65.5
	memValue := 1024.0
	requestCount := int64(42)

	_ = mockStorage.Save(context.Background(), &models.Metric{
		ID:    "cpu",
		MType: "gauge",
		Value: &cpuValue,
	})
	_ = mockStorage.Save(context.Background(), &models.Metric{
		ID:    "memory",
		MType: "gauge",
		Value: &memValue,
	})
	_ = mockStorage.Save(context.Background(), &models.Metric{
		ID:    "requests",
		MType: "counter",
		Delta: &requestCount,
	})

	// Request all metrics
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	resp := w.Result()
	defer func() {
		_ = resp.Body.Close()
	}()

	fmt.Printf("Status: %d\n", resp.StatusCode)
	fmt.Printf("Content-Type: %s\n", resp.Header.Get("Content-Type"))

	// Output:
	// Status: 200
	// Content-Type: text/html; charset=utf-8
}

func Example_pingDBHandler() {
	mockStorage := NewMockStorage()

	db := &sql.DB{}
	router := handler.BuildRouter(mockStorage, db, "")

	req := httptest.NewRequest("GET", "/ping", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	resp := w.Result()
	defer func() {
		_ = resp.Body.Close()
	}()

	fmt.Printf("Status: %d\n", resp.StatusCode)

	// Output:
	// Status: 500
}

// Unit tests for detailed handler testing
func TestUpdateMetricHandler(t *testing.T) {
	mockStorage := NewMockStorage()
	db := &sql.DB{}
	router := handler.BuildRouter(mockStorage, db, "")

	tests := []struct {
		name       string
		metric     models.Metric
		wantStatus int
	}{
		{
			name: "valid gauge metric",
			metric: models.Metric{
				ID:    "test_gauge",
				MType: "gauge",
				Value: func() *float64 { v := 123.45; return &v }(),
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "valid counter metric",
			metric: models.Metric{
				ID:    "test_counter",
				MType: "counter",
				Delta: func() *int64 { v := int64(100); return &v }(),
			},
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.metric)
			req := httptest.NewRequest("POST", "/update/", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			resp := w.Result()
			if resp.StatusCode != tt.wantStatus {
				t.Errorf("Expected status %d, got %d", tt.wantStatus, resp.StatusCode)
			}
			defer func() {
				_ = resp.Body.Close()
			}()

			// Verify that metric was saved
			savedMetric, err := mockStorage.GetByTypeAndID(context.Background(), tt.metric.ID, tt.metric.MType)
			if err != nil {
				t.Errorf("Error getting saved metric: %v", err)
			}

			if savedMetric.ID != tt.metric.ID {
				t.Errorf("Expected metric ID %s, got %s", tt.metric.ID, savedMetric.ID)
			}
		})
	}
}
