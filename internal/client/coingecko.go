package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/supercharged09/crypto-rates/internal/config"
)

// CoinGeckoClient - клиент для работы с API CoinGecko
type CoinGeckoClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// CoinGeckoAPI - интерфейс для тестирования
type CoinGeckoAPI interface {
	GetPrices() (map[string]float64, error)
	GetPrice(crypto string) (float64, error)
	GetAnyPrice(cryptoID string) (float64, error)
	GetPricesForIDs(ids []string) (map[string]float64, error)
}

// apiKeyTransport — middleware, добавляющий API key в каждый запрос
type apiKeyTransport struct {
	apiKey string
	base   http.RoundTripper
}

// RoundTrip добавляет заголовок x-cg-demo-api-key к каждому запросу
func (t *apiKeyTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.apiKey != "" {
		req.Header.Set("x-cg-demo-api-key", t.apiKey)
	}
	return t.base.RoundTrip(req)
}

// NewCoinGeckoClient создает новый экземпляр клиента
func NewCoinGeckoClient(cfg config.CoinGeckoConfig) *CoinGeckoClient {
	transport := &apiKeyTransport{
		apiKey: cfg.APIKey,
		base:   http.DefaultTransport,
	}

	return &CoinGeckoClient{
		baseURL: cfg.BaseURL,
		apiKey:  cfg.APIKey,
		httpClient: &http.Client{
			Timeout:   time.Duration(cfg.TimeoutSecond) * time.Second,
			Transport: transport,
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

	var result map[string]map[string]float64
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

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

// GetPricesForIDs получает цены для произвольного списка монет
func (c *CoinGeckoClient) GetPricesForIDs(ids []string) (map[string]float64, error) {
	if len(ids) == 0 {
		return map[string]float64{}, nil
	}

	idsParam := ""
	for i, id := range ids {
		if i > 0 {
			idsParam += ","
		}
		idsParam += id
	}

	url := fmt.Sprintf("%s/simple/price?ids=%s&vs_currencies=usd", c.baseURL, idsParam)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get prices: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result map[string]map[string]float64
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	prices := make(map[string]float64)
	for crypto, usdPrice := range result {
		prices[crypto] = usdPrice["usd"]
	}

	return prices, nil
}
