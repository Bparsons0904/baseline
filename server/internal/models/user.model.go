package models

import (
	"server/internal/logger"
	"server/internal/utils"

	"gorm.io/gorm"
)

type User struct {
	BaseModel
	FirstName string `gorm:"type:text"                      json:"firstName"`
	LastName  string `gorm:"type:text"                      json:"lastName"`
	Login     string `gorm:"type:text;uniqueIndex;not null" json:"login"`
	Password  string `gorm:"type:text;not null"             json:"-"`
	IsAdmin   bool   `gorm:"type:bool;default:false"        json:"isAdmin"`
}

type LoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.Password != "" {
		hashedPassword, err := utils.HashPassword(u.Password)
		if err != nil {
			return logger.New("models").
				File("User").
				Function("BeforeCreate").
				Err("failed to hash password", err, "user", u)
		}
		u.Password = hashedPassword
	}
	return nil
}
