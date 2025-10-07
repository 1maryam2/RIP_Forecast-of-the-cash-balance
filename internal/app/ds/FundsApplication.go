package ds

import (
	"time"
)

type FundsApplicationStatus string

const (
	Draft     FundsApplicationStatus = "draft"
	Formed    FundsApplicationStatus = "formed"
	Completed FundsApplicationStatus = "completed"
	Rejected  FundsApplicationStatus = "rejected"
	Deleted   FundsApplicationStatus = "deleted"
)

type FundsApplication struct {
	ID          uint                   `gorm:"primaryKey" json:"id"`
	Name        string                 `gorm:"type:varchar(255);not null" json:"name"`
	CompanyName string                 `gorm:"type:varchar(255)" json:"company_name"`
	INN         string                 `gorm:"type:varchar(12)" json:"inn"`
	OGRN        string                 `gorm:"type:varchar(15)" json:"ogrn"`
	InitialSum  float64                `gorm:"type:decimal(15,2)" json:"initial_sum"`
	Quarter     int                    `json:"quarter"`
	Result      float64                `gorm:"type:decimal(15,2)" json:"result"`
	IsActive    bool                   `gorm:"default:true" json:"is_active"`
	Status      FundsApplicationStatus `gorm:"type:varchar(20);default:'draft'" json:"status"`
	CreatedAt   time.Time              `json:"created_at"`
	FormedAt    *time.Time             `gorm:"null" json:"formed_at"`
	CompletedAt *time.Time             `gorm:"null" json:"completed_at"`
	UpdatedAt   time.Time              `json:"updated_at"`

	CreatorID            uint                   `gorm:"not null" json:"creator_id"`
	ModeratorID          *uint                  `gorm:"null" json:"moderator_id"`
	FundsApplicationItem []FundsApplicationItem `gorm:"foreignKey:FundsApplicationID" json:"items"`
	Creator              *Users                 `gorm:"foreignKey:CreatorID" json:"creator"`
	Moderator            *Users                 `gorm:"foreignKey:ModeratorID" json:"moderator"`
}
