package models

import "time"

type User struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	Username      string         `gorm:"size:32;not null;uniqueIndex:ux_users_username" json:"username"`
	AvatarURL     string         `gorm:"size:255" json:"avatar_url"`
	Email         string         `gorm:"size:255;not null;uniqueIndex:ux_users_email" json:"email"`
	PasswordHash  string         `gorm:"size:255;not null" json:"-"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	RefreshTokens []RefreshToken `json:"-"`
}

type RefreshToken struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"not null;index"`
	TokenHash string    `gorm:"size:64;not null;uniqueIndex"`
	ExpiresAt time.Time `gorm:"not null;index"`
	RevokedAt *time.Time
	CreatedAt time.Time
	User      User `gorm:"constraint:OnDelete:CASCADE;"`
}
