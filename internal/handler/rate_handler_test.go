package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/supercharged09/crypto-rates/internal/model"
)

// MockRateService - мок сервиса курсов
type MockRateService struct {
	mock.Mock
}

func (m *MockRateService) GetRateStats(ctx context.Context, crypto string) (*model.RateStats, error) {
	args := m.Called(ctx, crypto)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.RateStats), args.Error(1)
}

// TestGetAllRates_Success тестирование GET /rates
func TestGetAllRates_Success(t *testing.T) {

	mockSvc := new(MockRateService)
	handler := NewRateHandler(mockSvc, nil)

	//мок возвращает статистику для каждой валюты
	for _, crypto := range model.SupportedCryptos {
		mockSvc.On("GetRateStats", mock.Anything, crypto.ID).Return(&model.RateStats{
			Cryptocurrency:  crypto.ID,
			CurrentPrice:    100.00,
			MinPrice24h:     90.00,
			MaxPrice24h:     110.00,
			ChangePercent1h: 5.0,
			LastUpdated:     time.Now(),
		}, nil)
	}

	router := chi.NewRouter()
	router.Get("/rates", handler.GetAllRates)

	req := httptest.NewRequest("GET", "/rates", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "rates")
}

// TestGetRate_Success тест GET /rates/bitcoin
func TestGetAllRates_Error(t *testing.T) {

	mockSvc := new(MockRateService)
	handler := NewRateHandler(mockSvc, nil)

	mockSvc.On("GetRateStats", mock.Anything, "bitcoin").Return(&model.RateStats{
		Cryptocurrency: "bitcoin",
		CurrentPrice:   65000.00,
		LastUpdated:    time.Now(),
	}, nil)

	router := chi.NewRouter()
	router.Get("/rates/{cryptocurrency}", handler.GetRate)

	req := httptest.NewRequest("GET", "/rates/bitcoin", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var stats model.RateStats
	err := json.Unmarshal(w.Body.Bytes(), &stats)
	assert.NoError(t, err)
	assert.Equal(t, "bitcoin", stats.Cryptocurrency)
	assert.Equal(t, 65000.00, stats.CurrentPrice)
}

// TestGetRate_NotFound тест неподдерживаемой валюты
func TestGetRate_NotFound(t *testing.T) {

	mockSvc := new(MockRateService)
	handler := NewRateHandler(mockSvc, nil)

	router := chi.NewRouter()
	router.Get("/rates/{cryptocurrency}", handler.GetRate)

	req := httptest.NewRequest("GET", "/rates/dogecoin", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestGetRate_ServiceError тест ошибки сервиса
func TestGetRate_ServiceError(t *testing.T) {

	mockSvc := new(MockRateService)
	handler := NewRateHandler(mockSvc, nil)

	mockSvc.On("GetRateStats", mock.Anything, "bitcoin").Return(nil, errors.New("database error"))

	router := chi.NewRouter()
	router.Get("/rates/{cryptocurrency}", handler.GetRate)

	req := httptest.NewRequest("GET", "/rates/bitcoin", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}
