package model

import "time"

// Rate - запись о курсе крипты
type Rate struct {
	ID             int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Cryptocurrency string    `gorm:"type:varchar(20);not null;index:idx_rates_crypto_time,priority:1" json:"cryptocurrency"`
	PriceUSDCents  int64     `gorm:"not null" json:"price_usd_cents"`
	Timestamp      time.Time `gorm:"not bull;default:now();index:idx_rates_crypto_time,priority:2,sort:desc;index:idx_rates_timestamp" json:"timestamp"`
}

// TableName явно задает имя таблицы
func (Rate) TableName() string {
	return "rates"
}

// RateStats агрегированная стата для ответа клиенту
type RateStats struct {
	Cryptocurrency  string    `json:"cryptocurrency"`
	CurrentPrice    float64   `json:"current_price"`
	MinPrice24h     float64   `json:"min_price_24h"`
	MaxPrice24h     float64   `json:"max_price_24h"`
	ChangePercent1h float64   `json:"change_percent_1h"`
	LastUpdated     time.Time `json:"last_updated"`
}

// PricePoint точка для графика
type PricePoint struct {
	Price     float64   `json:"price"`
	Timestamp time.Time `json:"timestamp"`
}
