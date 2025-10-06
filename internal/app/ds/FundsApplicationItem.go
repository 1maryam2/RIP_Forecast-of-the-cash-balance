package ds

import (
	"time"
)

type FundsApplicationItem struct {
	ID                 uint    `gorm:"primaryKey"`
	FundsApplicationID uint    `gorm:"not null;uniqueIndex:idx_funds_application_account"`
	AccountID          uint    `gorm:"not null;uniqueIndex:idx_funds_application_account"`
	Amount             float64 `gorm:"type:decimal(15,2);not null"`
	Comment            string  `gorm:"type:text"`
	CreatedAt          time.Time
	UpdatedAt          time.Time

	FundsApplication FundsApplication `gorm:"foreignKey:FundsApplicationID"`
	Account          Account          `gorm:"foreignKey:AccountID"`
}
