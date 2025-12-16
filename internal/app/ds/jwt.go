package ds

import (
	"lab_1/internal/app/role"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
)

type JWTClaims struct {
	jwt.StandardClaims
	UserUUID uuid.UUID `json:"user_uuid"`
	UserID   uint      `json:"user_id"` // ДОБАВЛЕНО
	Role     role.Role `json:"role"`    // Изменено с string на Role
	Scopes   []string  `json:"scopes"`  // ДОБАВЛЕНО
}
