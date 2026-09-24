package model

import (
	"database/sql"
	"time"
)

// User — информация о пользователе бота
type User struct {
	ID           int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	ChatID       int64          `gorm:"not null;uniqueIndex" json:"chat_id"`
	Username     sql.NullString `gorm:"type:varchar(255)" json:"username"`
	FirstName    sql.NullString `gorm:"type:varchar(255)" json:"first_name"`
	LastName     sql.NullString `gorm:"type:varchar(255)" json:"last_name"`
	LanguageCode sql.NullString `gorm:"type:varchar(10)" json:"language_code"`
	FirstSeen    time.Time      `gorm:"not null;default:now()" json:"first_seen"`
	LastActive   time.Time      `gorm:"not null;default:now();index:idx_users_last_active,sort:desc" json:"last_active"`
	CommandCount int            `gorm:"not null;default:0" json:"command_count"`
}

func (User) TableName() string {
	return "users"
}
