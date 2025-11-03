package handler

import (
	"lab_1/internal/app/auth"
	"lab_1/internal/app/ds"
	"lab_1/internal/app/redis"
	"lab_1/internal/app/repository"
	"lab_1/internal/app/role"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	repo        *repository.Repository
	jwtService  *auth.JWTService
	redisClient *redis.Client
}

func NewAuthHandler(repo *repository.Repository, jwtService *auth.JWTService, redisClient *redis.Client) *AuthHandler {
	return &AuthHandler{
		repo:        repo,
		jwtService:  jwtService,
		redisClient: redisClient,
	}
}

// Register godoc
// @Summary Регистрация пользователя
// @Description Создание нового пользователя
// @Tags auth
// @Accept json
// @Produce json
// @Param request body ds.RegisterRequest true "Данные для регистрации"
// @Success 200 {object} ds.Users
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req ds.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.errorResponse(c, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}
	existingUser, _ := h.repo.GetUserByLogin(req.Login)
	if existingUser != nil {
		h.errorResponse(c, http.StatusBadRequest, "User with this login already exists")
		return
	}

	user := &ds.Users{
		Login:       req.Login,
		Password:    req.Password,
		IsModerator: req.IsModerator,
	}
	if req.IsModerator {
		user.Role = role.Manager
	} else {
		user.Role = role.Creator
	}

	if err := h.repo.CreateUser(user); err != nil {
		h.errorResponse(c, http.StatusInternalServerError, "Failed to create user: "+err.Error())
		return
	}

	user.Password = ""
	h.successResponse(c, user)
}

// Login godoc
// @Summary Авторизация пользователя
// @Description Вход в систему и получение токенов
// @Tags auth
// @Accept json
// @Produce json
// @Param request body ds.LoginRequest true "Данные для авторизации"
// @Success 200 {object} ds.LoginResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req ds.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.errorResponse(c, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	user, err := h.repo.GetUserByLogin(req.Login)
	if err != nil {
		h.errorResponse(c, http.StatusUnauthorized, "Invalid credentials or user not found")
		return
	}

	if !h.repo.CheckPasswordHash(req.Password, user.Password) {
		h.errorResponse(c, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	tokenPair, err := h.jwtService.GenerateTokenPair(user)
	if err != nil {
		h.errorResponse(c, http.StatusInternalServerError, "Failed to generate tokens: "+err.Error())
		return
	}

	err = h.redisClient.StoreRefreshToken(c.Request.Context(), user.UUID.String(), tokenPair.RefreshToken, 7*24*time.Hour)
	if err != nil {
		h.errorResponse(c, http.StatusInternalServerError, "Failed to store refresh token: "+err.Error())
		return
	}

	user.Password = ""
	h.successResponse(c, ds.LoginResponse{
		TokenPair: *tokenPair,
		User:      user,
	})
}

// Refresh godoc
// @Summary Обновление токенов
// @Description Получение новой пары access и refresh токенов
// @Tags auth
// @Accept json
// @Produce json
// @Param request body ds.RefreshTokenRequest true "Refresh token"
// @Success 200 {object} ds.TokenPair
// @Router /api/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req ds.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.errorResponse(c, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.errorResponse(c, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	claims, err := h.jwtService.ValidateToken(req.RefreshToken)
	if err != nil {
		h.errorResponse(c, http.StatusUnauthorized, "Invalid refresh token")
		return
	}

	hasRefreshScope := false
	for _, scope := range claims.Scopes {
		if scope == "api:refresh" {
			hasRefreshScope = true
			break
		}
	}

	if !hasRefreshScope {
		h.errorResponse(c, http.StatusUnauthorized, "Token is not a refresh token")
		return
	}

	user, err := h.repo.GetUserByUUID(claims.UserUUID.String())
	if err != nil {
		h.errorResponse(c, http.StatusUnauthorized, "User not found")
		return
	}

	storedToken, err := h.redisClient.GetRefreshToken(c.Request.Context(), user.UUID.String())
	if err != nil || storedToken != req.RefreshToken {
		h.errorResponse(c, http.StatusUnauthorized, "Invalid refresh token")
		return
	}

	newTokenPair, err := h.jwtService.GenerateTokenPair(user)
	if err != nil {
		h.errorResponse(c, http.StatusInternalServerError, "Failed to generate tokens: "+err.Error())
		return
	}

	err = h.redisClient.StoreRefreshToken(c.Request.Context(), user.UUID.String(), newTokenPair.RefreshToken, 7*24*time.Hour)
	if err != nil {
		h.errorResponse(c, http.StatusInternalServerError, "Failed to store refresh token: "+err.Error())
		return
	}

	h.successResponse(c, newTokenPair)
}

// Logout godoc
// @Summary Выход из системы
// @Description Добавление токена в blacklist
// @Tags auth
// @Security BearerAuth
// @Success 200 {object} map[string]string
// @Router /api/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	jwtStr := c.GetHeader("Authorization")
	if len(jwtStr) > 7 {
		jwtStr = jwtStr[7:]

		err := h.redisClient.WriteJWTToBlacklist(c.Request.Context(), jwtStr, time.Hour*24)
		if err != nil {
			h.errorResponse(c, http.StatusInternalServerError, "Failed to logout: "+err.Error())
			return
		}

		if userUUID, exists := c.Get("userUUID"); exists {
			h.redisClient.DeleteRefreshToken(c.Request.Context(), userUUID.(string))
		}
	}

	h.successResponse(c, gin.H{"message": "Logged out successfully"})
}

func (h *AuthHandler) errorResponse(ctx *gin.Context, statusCode int, message string) {
	ctx.JSON(statusCode, gin.H{
		"message": message,
	})
}

func (h *AuthHandler) successResponse(ctx *gin.Context, data interface{}) {
	ctx.JSON(http.StatusOK, gin.H{
		"data": data,
	})
}
