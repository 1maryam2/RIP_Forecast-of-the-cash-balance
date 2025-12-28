package ds

import (
	"time"
)

type FundsApplicationItem struct {
	ID                 uint      `gorm:"primaryKey" json:"id"`
	FundsApplicationID uint      `gorm:"not null" json:"funds_application_id"`
	AccountID          uint      `gorm:"not null" json:"account_id"`
	Amount             float64   `gorm:"type:decimal(15,2);not null" json:"amount"`
	Comment            string    `gorm:"type:text" json:"comment"`
	CreatedAt          time.Time `json:"-"`
	UpdatedAt          time.Time `json:"-"`

	FundsApplication FundsApplication `gorm:"foreignKey:FundsApplicationID" json:"-"`
	Account          Account          `gorm:"foreignKey:AccountID" json:"account"`
}
