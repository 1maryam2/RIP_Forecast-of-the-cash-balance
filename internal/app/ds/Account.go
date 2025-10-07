package ds

import (
	"time"
)

type Account struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	Code        string `gorm:"uniqueIndex;size:20" json:"code"`
	Title       string `gorm:"size:255" json:"title"`
	Description string `json:"description"`
	Type        string `gorm:"size:50" json:"type"`

	Category  string    `gorm:"size:100" json:"category"`
	Image     string    `json:"image"`
	InCart    bool      `gorm:"-" json:"in_cart"`
	IsActive  bool      `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
