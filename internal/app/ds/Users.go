package ds

import (
	"lab_1/internal/app/role"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Users struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UUID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();uniqueIndex" json:"uuid"`
	Login       string    `gorm:"type:varchar(25);unique;not null" json:"login"`
	Password    string    `gorm:"type:varchar(100);not null" json:"-"`
	Role        role.Role `gorm:"type:integer;default:0" json:"role"`
	IsModerator bool      `gorm:"type:boolean;default:false" json:"is_moderator"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (u *Users) BeforeCreate(tx *gorm.DB) error {
	if u.UUID == uuid.Nil {
		u.UUID = uuid.New()
	}
	return nil
}
