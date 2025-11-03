package auth

import (
	"errors"
	"fmt"
	"lab_1/internal/app/config"
	"lab_1/internal/app/ds"
	"time"

	"github.com/golang-jwt/jwt"
)

type JWTService struct {
	config *config.Config
}

func NewJWTService(cfg *config.Config) *JWTService {
	return &JWTService{
		config: cfg,
	}
}

func (s *JWTService) GenerateTokenPair(user *ds.Users) (*ds.TokenPair, error) {
	// Access token
	accessTokenClaims := &ds.JWTClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(s.config.JWT.ExpiresIn).Unix(),
			IssuedAt:  time.Now().Unix(),
			Issuer:    "cash-forecast-service",
		},
		UserUUID: user.UUID,
		UserID:   user.ID,
		Role:     user.Role,
		Scopes:   []string{"api:access"},
	}

	accessToken := jwt.NewWithClaims(s.config.JWT.SigningMethod, accessTokenClaims)
	accessTokenString, err := accessToken.SignedString([]byte(s.config.JWT.Token))
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Refresh token
	refreshTokenClaims := &ds.JWTClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(s.config.JWT.RefreshExpiry).Unix(),
			IssuedAt:  time.Now().Unix(),
			Issuer:    "cash-forecast-service",
		},
		UserUUID: user.UUID,
		UserID:   user.ID,
		Role:     user.Role,
		Scopes:   []string{"api:refresh"},
	}

	refreshToken := jwt.NewWithClaims(s.config.JWT.SigningMethod, refreshTokenClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(s.config.JWT.Token))
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &ds.TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.config.JWT.ExpiresIn / time.Second),
	}, nil
}

func (s *JWTService) ValidateToken(tokenString string) (*ds.JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &ds.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.config.JWT.Token), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*ds.JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
