package model

import "time"

// Subscription — подписка пользователя Telegram на авто-рассылку
type Subscription struct {
	ID             int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	ChatID         int64     `gorm:"not bull;uniqueIndex:idx_chat_crypto" json:"chat_id"`
	Cryptocurrency string    `gorm:"type:varchar(20);not null;default:'all';uniqueIndex:idx_chat_crypto" json:"cryptocurrency"`
	IntervalMin    int       `gorm:"not null;default:10" json:"interval_min"`
	IsActive       bool      `gorm:"not null;default:true;index:idx_subscriptions_active" json:"is_active"`
	CreatedAt      time.Time `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt      time.Time `gorm:"not null;default:now()" json:"updated_at"`
}

func (Subscription) TableName() string {
	return "subscriptions"
}
