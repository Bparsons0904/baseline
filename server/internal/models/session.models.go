package models

import (
	"server/internal/utils"
	"time"
)

const (
	SESSION_COOKIE_KEY = "sessionID"
)

type Session struct {
	ID        string    `gorm:"-" json:"id"`
	UserID    string    `gorm:"-" json:"userId"`
	Token     string    `gorm:"-" json:"token"`
	ExpiresAt time.Time `gorm:"-" json:"expiresAt"`
	RefreshAt time.Time `gorm:"-" json:"refreshAt"`
}

type TokenClaims utils.TokenClaims

