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
		{method: http.MethodPost, path: updateBasePath + "/gaugE", expectedCode: http.StatusNotFound},
		{method: http.MethodPost, path: updateBasePath + "/gauge/", expectedCode: http.StatusNotFound},
		{method: http.MethodPost, path: updateBasePath + "/gauge/m1/aa", expectedCode: http.StatusBadRequest},
		{method: http.MethodPost, path: updateBasePath + "/counter/", expectedCode: http.StatusNotFound},
		{method: http.MethodPost, path: updateBasePath + "/counter/m1/aa", expectedCode: http.StatusBadRequest},
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
		Post(srv.URL + updateBasePath + "/gauge/m1/1.01")

	assert.NoError(t, err, "error making HTTP request")
	assert.Equal(t, http.StatusOK, resp.StatusCode())
	val, ok := storage.Gauge("m1")
	assert.True(t, ok)
	assert.Equal(t, val, 1.01)
}

func TestCounterAdd(t *testing.T) {
	var storage = storage2.NewMemStorage()
	srv := httptest.NewServer(BuildRouter(storage))
	defer srv.Close()

	resp, err := resty.New().R().
		SetHeader("Content-Type", "text/plain").
		Post(srv.URL + updateBasePath + "/counter/m1/5")

	assert.NoError(t, err, "error making HTTP request")
	assert.Equal(t, http.StatusOK, resp.StatusCode())
	val, ok := storage.Counter("m1")
	assert.True(t, ok)
	assert.Equal(t, val, int64(5))
}

func TestGet(t *testing.T) {
	var storage = storage2.NewMemStorage()
	storage.AddGauge("g1", 1.001)
	storage.AddCounter("c1", 2)

	srv := httptest.NewServer(BuildRouter(storage))
	defer srv.Close()

	resp, err := resty.New().R().
		SetHeader("Accept-Encoding", "").
		Get(srv.URL + getBasePath + "/gauge/g1")
	assert.NoError(t, err, "error making HTTP request")
	assert.Equal(t, "1.001", string(resp.Body()))

	resp, err = resty.New().R().
		SetHeader("Accept-Encoding", "").
		Get(srv.URL + getBasePath + "/counter/c1")
	assert.NoError(t, err, "error making HTTP request")
	assert.Equal(t, "2", string(resp.Body()))
}
