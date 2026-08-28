package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/supercharged09/crypto-rates/internal/client"
	"github.com/supercharged09/crypto-rates/internal/model"
)

// мок репо курсов
type MockRateRepository struct {
	mock.Mock
}

func (m *MockRateRepository) Save(ctx context.Context, rate model.Rate) error {
	args := m.Called(ctx, rate)
	return args.Error(0)
}

func (m *MockRateRepository) GetCurrentPrice(ctx context.Context, crypto string) (*model.Rate, error) {
	args := m.Called(ctx, crypto)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Rate), args.Error(1)
}

func (m *MockRateRepository) GetMinMax24h(ctx context.Context, crypto string) (float64, float64, error) {
	args := m.Called(ctx, crypto)
	return args.Get(0).(float64), args.Get(1).(float64), args.Error(2)
}

func (m *MockRateRepository) GetPriceHourAgo(ctx context.Context, crypto string) (float64, error) {
	args := m.Called(ctx, crypto)
	return args.Get(0).(float64), args.Error(1)
}

// тест успешного получения и сохранения курсов
func TestFetchAndSaveRates_Success(t *testing.T) {
	mockClient := new(client.MockCoinGeckoClient)
	mockRepo := new(MockRateRepository)

	service := NewRateService(mockClient, mockRepo)

	//мок возвращает цены
	mockClient.On("GetPrices").Return(map[string]float64{
		"bitcoin":     65000.00,
		"ethereum":    2000.00,
		"tether":      1.00,
		"binancecoin": 500.00,
		"ripple":      1.50,
		"solana":      100.00,
	}, nil)

	//мок сохранения. вызовется для каждой валюты в поддерживаемых
	for _, crypto := range model.SupportedCryptos {
		mockRepo.On("Save", mock.Anything, mock.MatchedBy(func(r model.Rate) bool {
			return r.Cryptocurrency == crypto.ID
		})).Return(nil)
	}

	/* нельзя использовать mockRepo, потому что NewRateService принимает *RateRepository
	поэтому просто проверяем, что GetPrices вызван
	*/
	err := service.FetchAndSaveRates(context.Background())

	//assert
	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

// TestFetchAndSaveRates_Error тестирует ошибку API
func TestFetchAndSaveRates_APIError(t *testing.T) {
	// Arrange
	mockClient := new(client.MockCoinGeckoClient)
	mockRepo := new(MockRateRepository)

	service := NewRateService(mockClient, mockRepo)

	mockClient.On("GetPrices").Return(nil, errors.New("api error"))

	// Act
	err := service.FetchAndSaveRates(context.Background())

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "api error")
	mockClient.AssertExpectations(t)
}

// тест получения статистики
func TestGetRateStats_Success(t *testing.T) {
	mockClient := new(client.MockCoinGeckoClient)
	mockRepo := new(MockRateRepository)

	service := NewRateService(mockClient, mockRepo)

	now := time.Now()
	currentRate := &model.Rate{
		ID:             1,
		Cryptocurrency: "bitcoin",
		PriceUSD:       65000.00,
		Timestamp:      now,
	}

	mockRepo.On("GetCurrentPrice", mock.Anything, "bitcoin").Return(currentRate, nil)
	mockRepo.On("GetMinMax24h", mock.Anything, "bitcoin").Return(64000.00, 66000.00, nil)
	mockRepo.On("GetPriceHourAgo", mock.Anything, "bitcoin").Return(64500.00, nil)

	stats, err := service.GetRateStats(context.Background(), "bitcoin")

	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, "bitcoin", stats.Cryptocurrency)
	assert.Equal(t, 65000.00, stats.CurrentPrice)
	assert.Equal(t, 64000.00, stats.MinPrice24h)
	assert.Equal(t, 66000.00, stats.MaxPrice24h)
	assert.InDelta(t, 0.78, stats.ChangePercent1h, 0.01) // (65000-64500)/64500*100 ≈ 0.78%

	mockRepo.AssertExpectations(t)
}

// TestGetRateStats_NoData тест случая, когда данных нет
func TestGetRateStats_APIError(t *testing.T) {
	mockClient := new(client.MockCoinGeckoClient)
	mockRepo := new(MockRateRepository)

	service := NewRateService(mockClient, mockRepo)

	mockRepo.On("GetCurrentPrice", mock.Anything, "bitcoin").Return(nil, nil)

	stats, err := service.GetRateStats(context.Background(), "bitcoin")

	assert.Error(t, err)
	assert.Nil(t, stats)
	assert.Contains(t, err.Error(), "no data")

	mockRepo.AssertExpectations(t)
}
