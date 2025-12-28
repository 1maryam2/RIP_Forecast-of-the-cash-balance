package ds

import (
	"lab_1/internal/app/role"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
)

type JWTClaims struct {
	jwt.StandardClaims
	UserUUID uuid.UUID `json:"user_uuid"`
	UserID   uint      `json:"user_id"`
	Role     role.Role `json:"role"`
	Scopes   []string  `json:"scopes"`
}
