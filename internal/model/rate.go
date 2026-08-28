package model

import "time"

type Rate struct {
	ID             int64     `json:"id"`
	Cryptocurrency string    `json:"cryptocurrency"`
	PriceUSD       float64   `json:"price_usd"`
	Timestamp      time.Time `json:"timestamp"`
}

type RateStats struct {
	Cryptocurrency  string    `json:"cryptocurrency"`
	CurrentPrice    float64   `json:"current_price"`
	MinPrice24h     float64   `json:"min_price_24h"`
	MaxPrice24h     float64   `json:"max_price_24h"`
	ChangePercent1h float64   `json:"change_percent_1h"`
	LastUpdated     time.Time `json:"last_updated"`
}

// точка для графика курса(цена и время)
type PricePoint struct {
	Price     float64   `json:"price"`
	Timestamp time.Time `json:"timestamp"`
}
