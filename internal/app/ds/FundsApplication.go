package ds

import (
	"time"
)

type FundsApplication struct {
	ID          uint    `gorm:"primaryKey"`
	Name        string  `gorm:"type:varchar(255);not null"`
	CompanyName string  `gorm:"type:varchar(255)"`
	INN         string  `gorm:"type:varchar(12)"`
	OGRN        string  `gorm:"type:varchar(15)"`
	InitialSum  float64 `gorm:"type:decimal(15,2)"`
	Quarter     int
	Result      float64
	IsActive    bool `gorm:"default:true"`
	CreatedAt   time.Time
	UpdatedAt   time.Time

	FundsApplicationItem []FundsApplicationItem `gorm:"foreignKey:FundsApplicationID"`
}
