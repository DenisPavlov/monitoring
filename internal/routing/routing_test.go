package routing

import (
	storage2 "github.com/DenisPavlov/monitoring/internal/storage"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
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
		{method: http.MethodPost, path: updateBasePath + gaugeMetricName + "/", expectedCode: http.StatusNotFound},
		{method: http.MethodPost, path: updateBasePath + gaugeMetricName + "/m1/aa", expectedCode: http.StatusBadRequest},
		{method: http.MethodPost, path: updateBasePath + counterMetricName + "/", expectedCode: http.StatusNotFound},
		{method: http.MethodPost, path: updateBasePath + counterMetricName + "/m1/aa", expectedCode: http.StatusBadRequest},
	}

	var storage = storage2.NewMemStorage()
	srv := httptest.NewServer(BuildRouter(storage))
	defer srv.Close()

	for _, tc := range testCases {
		t.Run(tc.method, func(t *testing.T) {
			req := resty.New().R()
			req.Method = tc.method
			req.URL = srv.URL + tc.path

			resp, err := req.Send()
			assert.NoError(t, err, "error making HTTP request")

			assert.Equal(t, tc.expectedCode, resp.StatusCode(), "Код ответа не совпадает с ожидаемым")
		})
	}
}

func TestGaugeAdd(t *testing.T) {
	var storage = storage2.NewMemStorage()
	srv := httptest.NewServer(BuildRouter(storage))
	defer srv.Close()

	resp, err := resty.New().R().
		SetHeader("Content-Type", "text/plain").
		Post(srv.URL + updateBasePath + "gauge/m1/1.01")

	assert.NoError(t, err, "error making HTTP request")
	assert.Equal(t, http.StatusOK, resp.StatusCode())
	assert.Equal(t, storage.Gauge("m1"), 1.01)
}

func TestCounterAdd(t *testing.T) {
	var storage = storage2.NewMemStorage()
	srv := httptest.NewServer(BuildRouter(storage))
	defer srv.Close()

	resp, err := resty.New().R().
		SetHeader("Content-Type", "text/plain").
		Post(srv.URL + updateBasePath + "counter/m1/5")

	assert.NoError(t, err, "error making HTTP request")
	assert.Equal(t, http.StatusOK, resp.StatusCode())
	assert.Equal(t, storage.Counter("m1"), int64(5))
}
