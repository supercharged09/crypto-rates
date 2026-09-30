package model

import "time"

// TrackedCrypto — монета, которую отслеживает сервис
type TrackedCrypto struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	CoinID    string    `gorm:"type:varchar(50);not null;uniqueIndex" json:"coin_id"` // ID на CoinGecko: "dogecoin"
	Symbol    string    `gorm:"type:varchar(20);not null" json:"symbol"`              // "DOGE"
	Name      string    `gorm:"type:varchar(100);not null" json:"name"`               // "Dogecoin"
	Emoji     string    `gorm:"type:varchar(10);default:'🪙'" json:"emoji"`            // "🐕"
	AddedBy   int64     `gorm:"not null" json:"added_by"`                             // chat_id, кто первым добавил
	CreatedAt time.Time `gorm:"not null;default:now()" json:"created_at"`
}

func (TrackedCrypto) TableName() string {
	return "tracked_cryptos"
}
