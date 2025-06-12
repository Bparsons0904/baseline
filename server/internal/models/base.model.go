package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BaseModel struct {
	ID        string    `gorm:"type:text;primaryKey" json:"id"`
	CreatedAt time.Time `gorm:"autoCreateTime"       json:"createdAt"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"       json:"updatedAt"`
}

func (b *BaseModel) BeforeSave(tx *gorm.DB) error {
	if b.ID == "" {
		uuidString, _ := uuid.NewV7()
		b.ID = uuidString.String()
	}
	return nil
}
