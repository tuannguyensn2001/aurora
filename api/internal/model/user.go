package model

import (
	"time"

	"gorm.io/gorm"
)

// User represents a user in the system
type User struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	Email        string         `gorm:"uniqueIndex;not null" json:"email"`
	Name         string         `json:"name"`
	Picture      string         `json:"picture"`
	GoogleID     string         `gorm:"uniqueIndex;not null" json:"google_id"`
	AccessToken  string         `gorm:"type:text" json:"-"`
	RefreshToken string         `gorm:"type:text" json:"-"`
	TokenExpiry  *time.Time     `json:"-"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	LastLoginAt  *time.Time     `json:"last_login_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}
