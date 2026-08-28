package client

import (
	"github.com/stretchr/testify/mock"
)

type MockCoinGeckoClient struct {
	mock.Mock
}

// GetPrices - мок метода получения всех цен
func (m *MockCoinGeckoClient) GetPrices() (map[string]float64, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]float64), args.Error(1)
}

// мок метода получения одной цены
func (m *MockCoinGeckoClient) GetPrice(crypto string) (float64, error) {
	args := m.Called(crypto)
	return args.Get(0).(float64), args.Error(1)
}
