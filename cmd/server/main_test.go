package main

import (
	storage2 "github.com/DenisPavlov/monitoring/internal/storage"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSaveMetrics(t *testing.T) {
	testCases := []struct {
		method       string
		path         string
		expectedCode int
	}{
		{method: http.MethodGet, path: updateBasePath, expectedCode: http.StatusMethodNotAllowed},
		{method: http.MethodPost, path: updateBasePath + "gaugE", expectedCode: http.StatusNotFound},
		{method: http.MethodPost, path: updateBasePath + "gauge/", expectedCode: http.StatusNotFound},
		{method: http.MethodPost, path: updateBasePath + "gauge/m1/aa", expectedCode: http.StatusBadRequest},
		{method: http.MethodPost, path: updateBasePath + "counter/", expectedCode: http.StatusNotFound},
		{method: http.MethodPost, path: updateBasePath + "counter/m1/aa", expectedCode: http.StatusBadRequest},
	}

	var storage = storage2.NewMemStorage()

	for _, tc := range testCases {
		t.Run(tc.method, func(t *testing.T) {
			r := httptest.NewRequest(tc.method, tc.path, http.NoBody)
			w := httptest.NewRecorder()

			saveMetricsHandler(storage)(w, r)

			assert.Equal(t, tc.expectedCode, w.Code, "Код ответа не совпадает с ожидаемым")
		})
	}
}

func TestGaugeAdd(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, updateBasePath+"gauge/m1/1.01", http.NoBody)
	w := httptest.NewRecorder()

	var storage = storage2.NewMemStorage()

	saveMetricsHandler(storage)(w, r)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, storage.Gauge("m1"), 1.01)
}

func TestCounterAdd(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, updateBasePath+"counter/m1/5", http.NoBody)
	w := httptest.NewRecorder()

	var storage = storage2.NewMemStorage()

	saveMetricsHandler(storage)(w, r)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, storage.Counter("m1"), int64(5))
}
