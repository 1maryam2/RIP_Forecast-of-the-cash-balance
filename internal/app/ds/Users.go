package ds

import (
	"time"
)

type Users struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Login       string    `gorm:"type:varchar(25);unique;not null" json:"login"`
	Password    string    `gorm:"type:varchar(100);not null" json:"-"`
	IsModerator bool      `gorm:"type:boolean;default:false" json:"is_moderator"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
