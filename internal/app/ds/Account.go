package ds

import (
	"time"
)

type Account struct {
	ID          uint   `gorm:"primaryKey"`
	Code        string `gorm:"uniqueIndex;size:20"` // "90.01", "90.02"
	Title       string `gorm:"size:255"`
	Description string
	Type        string `gorm:"size:50"`
	Category    string `gorm:"size:100"`
	Image       string
	InCart      bool `gorm:"-"`
	IsActive    bool `gorm:"default:true"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
