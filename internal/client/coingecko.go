package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// CoinGeckoClient - клиент для работы с API CoinGecko
type CoinGeckoClient struct {
	baseURL    string
	httpClient *http.Client
}

// CoinGeckoAPI - интерфейс для тестирования
type CoinGeckoAPI interface {
	GetPrices() (map[string]float64, error)
	GetPrice(crypto string) (float64, error)
	GetAnyPrice(cryptoID string) (float64, error)
}

// NewCoinGeckoClient создает новый экземпляр клиента
func NewCoinGeckoClient(baseURL string) *CoinGeckoClient {
	return &CoinGeckoClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetPrices получает текущие цены для всех поддерживаемых валют
func (c *CoinGeckoClient) GetPrices() (map[string]float64, error) {
	url := fmt.Sprintf(
		"%s/simple/price?ids=bitcoin,ethereum,tether,binancecoin,ripple,solana&vs_currencies=usd",
		c.baseURL,
	)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get prices: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	//парсинг в map[string]map[string]float64
	var result map[string]map[string]float64
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	//преобразование в простую мапу map[crypto]price
	prices := make(map[string]float64)
	for crypto, usdPrice := range result {
		prices[crypto] = usdPrice["usd"]
	}

	return prices, nil
}

// GetPrice получает цену конкретной крипты
func (c *CoinGeckoClient) GetPrice(crypto string) (float64, error) {
	prices, err := c.GetPrices()
	if err != nil {
		return 0, err
	}

	price, ok := prices[crypto]

	if !ok {
		return 0, fmt.Errorf("cryptocurrency %s not found in response", crypto)
	}

	return price, nil
}

// GetAnyPrice получает цену любой монеты с CoinGecko
func (c *CoinGeckoClient) GetAnyPrice(cryptoID string) (float64, error) {
	url := fmt.Sprintf("%s/simple/price?ids=%s&vs_currencies=usd", c.baseURL, cryptoID)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return 0, fmt.Errorf("failed to get price: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return 0, fmt.Errorf("cryptocurrency %s not found", cryptoID)
	}
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result map[string]map[string]float64
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("failed to decode response: %w", err)
	}

	priceData, ok := result[cryptoID]
	if !ok {
		return 0, fmt.Errorf("cryptocurrency %s not found in response", cryptoID)
	}

	price, ok := priceData["usd"]
	if !ok {
		return 0, fmt.Errorf("USD price not found for %s", cryptoID)
	}

	return price, nil
}
