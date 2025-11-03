package handler

import (
	"fmt"
	"lab_1/internal/app/auth"
	"lab_1/internal/app/ds"
	"lab_1/internal/app/middleware"
	"lab_1/internal/app/redis"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

/* func getUserIDFromContext(c *gin.Context) (uint, error) {
	if id, ok := c.Get("userID"); ok {
		return id.(uint), nil
	}
	return 1, nil
} */

/* func isModeratorFromContext(c *gin.Context) bool {
	if isMod, ok := c.Get("isModerator"); ok {
		return isMod.(bool)
	}
	// Заглушка: ID 2 - модератор
	userID, _ := getUserIDFromContext(c)
	return userID == 2
}

func mockAuthMiddleware(c *gin.Context) {
	token := c.GetHeader("Authorization")

	// Установка значений по умолчанию
	userID := uint(1)
	isModerator := false

	// Проверка токенов
	if strings.HasPrefix(token, "Bearer dummy-jwt-token-for-moderator") {
		userID = 2
		isModerator = true
		logrus.Printf("Moderator access: userID=%d", userID)
	} else if strings.HasPrefix(token, "Bearer dummy-jwt-token-for-creator") {
		userID = 1
		isModerator = false
		logrus.Printf("Creator access: userID=%d", userID)
	} else if token != "" {
		logrus.Printf("Unknown token, using default: userID=%d", userID)
	} else {
		logrus.Printf("No Authorization header, using default: userID=%d", userID)
	}

	// ВСЕГДА устанавливаем контекст и продолжаем
	c.Set("userID", userID)
	c.Set("isModerator", isModerator)
	c.Next() // Важно: всегда вызываем c.Next()
} */

func (h *Handler) RegisterRoutes(router *gin.Engine, jwtService *auth.JWTService, redisClient *redis.Client) {
	api := router.Group("/api")
	api.POST("/register", h.Register)
	api.POST("/login", h.Login)

	authApi := api.Group("/", middleware.AuthMiddleware(jwtService, redisClient))
	{
		authApi.POST("/logout", h.Logout)
		authApi.GET("/users/profile", h.GetUserProfile)
		authApi.PUT("/users/profile", h.UpdateUserProfile)

		authApi.GET("/accounts", h.GetAccounts)
		authApi.GET("/accounts/:id", h.GetAccountByID)
		authApi.POST("/accounts", h.CreateAccount)
		authApi.PUT("/accounts/:id", h.UpdateAccount)
		authApi.DELETE("/accounts/:id", h.DeleteAccount)
		authApi.POST("/accounts/:id/image", h.UploadAccountImage)

		authApi.GET("/cash-forecasts", h.GetCashForecasts)
		authApi.GET("/cash-forecasts/cart", h.GetCartIcon)
		authApi.GET("/cash-forecasts/:id", h.GetCashForecast)
		authApi.PUT("/cash-forecasts/:id", h.UpdateCashForecast)
		authApi.PUT("/cash-forecasts/:id/form", h.FormCashForecast)
		authApi.PUT("/cash-forecasts/:id/complete", h.CompleteCashForecast)
		authApi.PUT("/cash-forecasts/:id/reject", h.RejectCashForecast)
		authApi.DELETE("/cash-forecasts/:id", h.DeleteCashForecast)

		authApi.POST("/cash-forecast-items", h.AddAccountToCashForecast)
		authApi.DELETE("/cash-forecast-items/:itemID", h.RemoveItemFromCashForecast)
		authApi.PUT("/cash-forecast-items", h.UpdateCashForecastItem)
	}
}

func (h *Handler) errorResponse(ctx *gin.Context, statusCode int, message string) {
	logrus.Error(message)
	ctx.JSON(statusCode, gin.H{
		"message": message,
	})
}

func (h *Handler) successResponse(ctx *gin.Context, data interface{}) {
	ctx.JSON(http.StatusOK, gin.H{
		"data": data,
	})
}

// Register godoc
// @Summary Регистрация пользователя
// @Description Создание нового пользователя в системе
// @Tags users
// @Accept json
// @Produce json
// @Param request body ds.RegisterRequest true "Данные для регистрации"
// @Success 200 {object} ds.Users
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/register [post]
func (h *Handler) Register(c *gin.Context) {
	var req ds.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.errorResponse(c, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	user := &ds.Users{
		Login:       req.Login,
		Password:    req.Password,
		IsModerator: req.IsModerator,
	}

	if err := h.Repository.CreateUser(user); err != nil {
		h.errorResponse(c, http.StatusInternalServerError, "Failed to create user: "+err.Error())
		return
	}
	user.Password = ""
	h.successResponse(c, user)
}

// Login godoc
// @Summary Аутентификация пользователя
// @Description Вход пользователя в систему
// @Tags users
// @Accept json
// @Produce json
// @Param request body ds.LoginRequest true "Данные для входа"
// @Success 200 {object} map[string]interface{} "token и user"
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/login [post]
func (h *Handler) Login(c *gin.Context) {
	var req ds.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.errorResponse(c, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	user, err := h.Repository.GetUserByLogin(req.Login)
	if err != nil {
		h.errorResponse(c, http.StatusUnauthorized, "Invalid credentials or user not found")
		return
	}

	if !h.Repository.CheckPasswordHash(req.Password, user.Password) {
		h.errorResponse(c, http.StatusUnauthorized, "Invalid credentials")
		return
	}
	token := "dummy-jwt-token-for-" + user.Login
	user.Password = ""
	h.successResponse(c, gin.H{"token": token, "user": user})
}

// Logout godoc
// @Summary Выход из системы
// @Description Завершение сессии пользователя
// @Tags users
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]string
// @Router /api/logout [post]
func (h *Handler) Logout(c *gin.Context) {
	h.successResponse(c, gin.H{"message": "Logged out successfully"})
}

// GetUserProfile godoc
// @Summary Получить профиль пользователя
// @Description Получение данных текущего пользователя для личного кабинета
// @Tags users
// @Security BearerAuth
// @Produce json
// @Success 200 {object} ds.Users
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/users/profile [get]
func (h *Handler) GetUserProfile(c *gin.Context) {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		h.errorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	user, err := h.Repository.GetUserByID(userID)
	if err != nil {
		h.errorResponse(c, http.StatusNotFound, "User profile not found")
		return
	}
	user.Password = ""
	h.successResponse(c, user)
}

// UpdateUserProfile godoc
// @Summary Обновить профиль пользователя
// @Description Изменение данных пользователя в личном кабинете
// @Tags users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body ds.UpdateUserRequest true "Данные для обновления"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/users/profile [put]
func (h *Handler) UpdateUserProfile(c *gin.Context) {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		h.errorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var req ds.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.errorResponse(c, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	updates := make(map[string]interface{})
	if req.Login != "" {
		updates["login"] = req.Login
	}
	if len(updates) == 0 {
		h.errorResponse(c, http.StatusBadRequest, "No valid fields to update")
		return
	}

	if err := h.Repository.UpdateUser(userID, updates); err != nil {
		h.errorResponse(c, http.StatusInternalServerError, "Failed to update profile: "+err.Error())
		return
	}
	h.successResponse(c, gin.H{"message": "Profile updated successfully"})
}

// CreateAccount godoc
// @Summary Создать новый счет
// @Description Создание нового счета в системе
// @Tags accounts
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body ds.CreateAccountRequest true "Данные для создания счета"
// @Success 200 {object} ds.Account
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/accounts [post]
func (h *Handler) CreateAccount(c *gin.Context) {
	var req ds.CreateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.errorResponse(c, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	account := &ds.Account{
		Code:        req.Code,
		Title:       req.Title,
		Description: req.Description,
		Type:        req.Type,
		Category:    req.Category,
		IsActive:    true,
	}

	if err := h.Repository.CreateAccount(account); err != nil {
		h.errorResponse(c, http.StatusInternalServerError, "Failed to create account: "+err.Error())
		return
	}
	h.successResponse(c, account)
}

// UpdateAccount godoc
// @Summary Обновить счет
// @Description Изменение данных счета
// @Tags accounts
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "ID счета"
// @Param request body ds.UpdateAccountRequest true "Данные для обновления"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/accounts/{id} [put]
func (h *Handler) UpdateAccount(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		h.errorResponse(c, http.StatusBadRequest, "Invalid account ID")
		return
	}

	var req ds.UpdateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.errorResponse(c, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	updates := make(map[string]interface{})
	if req.Code != "" {
		updates["code"] = req.Code
	}
	if req.Title != "" {
		updates["title"] = req.Title
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.Type != "" {
		updates["type"] = req.Type
	}
	if req.Category != "" {
		updates["category"] = req.Category
	}

	if len(updates) == 0 {
		h.errorResponse(c, http.StatusBadRequest, "No valid fields to update")
		return
	}

	if err := h.Repository.UpdateAccount(uint(id), updates); err != nil {
		h.errorResponse(c, http.StatusInternalServerError, "Failed to update account: "+err.Error())
		return
	}
	h.successResponse(c, gin.H{"message": "Account updated successfully"})
}

// DeleteAccount godoc
// @Summary Удалить счет
// @Description Удаление счета (включая изображение)
// @Tags accounts
// @Security BearerAuth
// @Produce json
// @Param id path int true "ID счета"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/accounts/{id} [delete]
func (h *Handler) DeleteAccount(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		h.errorResponse(c, http.StatusBadRequest, "Invalid account ID")
		return
	}

	if err := h.Repository.DeleteAccount(uint(id)); err != nil {
		h.errorResponse(c, http.StatusInternalServerError, "Failed to delete account: "+err.Error())
		return
	}
	h.successResponse(c, gin.H{"message": "Account deleted successfully"})
}

// UploadAccountImage godoc
// @Summary Загрузить изображение счета
// @Description Добавление/замена изображения для счета. Название генерируется на латинице
// @Tags accounts
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param id path int true "ID счета"
// @Param image formData file true "Изображение счета"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/accounts/{id}/image [post]
func (h *Handler) UploadAccountImage(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		h.errorResponse(c, http.StatusBadRequest, "Invalid account ID")
		return
	}

	file, err := c.FormFile("image")
	if err != nil {
		h.errorResponse(c, http.StatusBadRequest, "Failed to get image file: "+err.Error())
		return
	}
	src, err := file.Open()
	if err != nil {
		h.errorResponse(c, http.StatusInternalServerError, "Failed to open file: "+err.Error())
		return
	}
	defer src.Close()
	fileExtension := ""
	if parts := strings.Split(file.Filename, "."); len(parts) > 1 {
		fileExtension = "." + parts[len(parts)-1]
	}
	imageName := fmt.Sprintf("%s%s", uuid.New().String(), fileExtension)

	if err := h.Repository.SetAccountImage(uint(id), imageName); err != nil {
		h.errorResponse(c, http.StatusInternalServerError, "Failed to update account image field: "+err.Error())
		return
	}

	h.successResponse(c, gin.H{"message": "Image uploaded successfully", "image_name": imageName})
}

// GetAccounts godoc
// @Summary Получить список счетов
// @Description Возвращает список счетов с возможностью фильтрации
// @Tags accounts
// @Security BearerAuth
// @Produce json
// @Param search query string false "Поиск по названию"
// @Param type query string false "Фильтр по типу"
// @Param category query string false "Фильтр по категории"
// @Success 200 {array} ds.Account
// @Failure 500 {object} map[string]string
// @Router /api/accounts [get]
func (h *Handler) GetAccounts(c *gin.Context) {
	var filter ds.AccountsFilter
	filter.Search = c.Query("search")
	filter.Type = c.Query("type")
	filter.Category = c.Query("category")

	accounts, err := h.Repository.GetAccountsByTitleByFilter(filter)
	if err != nil {
		h.errorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	h.successResponse(c, accounts)
}

// GetAccountByID godoc
// @Summary Получить счет по ID
// @Description Получение информации о конкретном счете
// @Tags accounts
// @Security BearerAuth
// @Produce json
// @Param id path int true "ID счета"
// @Success 200 {object} ds.Account
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/accounts/{id} [get]
func (h *Handler) GetAccountByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		h.errorResponse(c, http.StatusBadRequest, "Invalid account ID")
		return
	}
	account, err := h.Repository.GetAccountByID(uint(id))
	if err != nil {
		h.errorResponse(c, http.StatusNotFound, "Account not found")
		return
	}
	h.successResponse(c, account)
}

// GetCashForecasts godoc
// @Summary Get cash forecasts list
// @Description Get filtered list of cash forecasts with date range
// @Tags cash-forecasts
// @Security BearerAuth
// @Produce json
// @Param date_from query string false "Start date (YYYY-MM-DD)"
// @Param date_to query string false "End date (YYYY-MM-DD)"
// @Success 200 {array} ds.FundsApplication
// @Failure 500 {object} map[string]string
// @Router /api/cash-forecasts [get]
func (h *Handler) GetCashForecasts(c *gin.Context) {
	var filter ds.ApplicationFilter
	dateFromStr := c.Query("date_from")
	dateToStr := c.Query("date_to")
	if dateFromStr != "" {
		t, err := time.Parse("2006-01-02", dateFromStr)
		if err != nil {
			h.errorResponse(c, http.StatusBadRequest, "Invalid date_from format. Use YYYY-MM-DD")
			return
		}
		filter.DateFrom = &t
	}
	if dateToStr != "" {
		t, err := time.Parse("2006-01-02", dateToStr)
		if err != nil {
			h.errorResponse(c, http.StatusBadRequest, "Invalid date_to format. Use YYYY-MM-DD")
			return
		}
		filter.DateTo = &t
	}

	applications, err := h.Repository.GetCashForecastsList(filter)
	if err != nil {
		h.errorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	h.successResponse(c, applications)
}

// GetCartIcon godoc
// @Summary Получить иконку корзины
// @Description Получение ID заявки-черновика пользователя и количества услуг в ней
// @Tags cash-forecasts
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{} "application_id и item_count"
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/cash-forecasts/cart [get]
func (h *Handler) GetCartIcon(c *gin.Context) {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		h.errorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	appID, count, err := h.Repository.GetCashForecastCountForUser(userID)
	if err != nil {
		h.errorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	h.successResponse(c, gin.H{
		"application_id": appID,
		"item_count":     count,
	})
}

// GetCashForecasts godoc
// @Summary Получить список заявок
// @Description Получение списка заявок с фильтрацией по диапазону дат и статусу (кроме удаленных и черновиков)
// @Tags cash-forecasts
// @Security BearerAuth
// @Produce json
// @Param date_from query string false "Начальная дата (YYYY-MM-DD)"
// @Param date_to query string false "Конечная дата (YYYY-MM-DD)"
// @Param status query string false "Статус заявки"
// @Success 200 {array} ds.FundsApplication
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/cash-forecasts [get]
func (h *Handler) GetCashForecast(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		h.errorResponse(c, http.StatusBadRequest, "Invalid application ID")
		return
	}
	app, err := h.Repository.GetCashForecastByID(uint(id))
	if err != nil {
		h.errorResponse(c, http.StatusNotFound, "Application not found")
		return
	}
	h.successResponse(c, app)
}

// UpdateCashForecast godoc
// @Summary Обновить заявку
// @Description Изменение полей заявки по теме
// @Tags cash-forecasts
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "ID заявки"
// @Param request body ds.UpdateFundsApplicationRequest true "Данные для обновления"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /api/cash-forecasts/{id} [put]
func (h *Handler) UpdateCashForecast(c *gin.Context) {
	appID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		h.errorResponse(c, http.StatusBadRequest, "Invalid application ID")
		return
	}

	var req ds.UpdateFundsApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.errorResponse(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}
	if err := h.Repository.UpdateCashForecast(uint(appID), req); err != nil {
		h.errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	h.successResponse(c, gin.H{"message": "Application updated"})
}

// FormCashForecast godoc
// @Summary Сформировать заявку
// @Description Отправка заявки на модерацию с проверкой обязательных полей
// @Tags cash-forecasts
// @Security BearerAuth
// @Produce json
// @Param id path int true "ID заявки"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Router /api/cash-forecasts/{id}/form [put]
func (h *Handler) FormCashForecast(c *gin.Context) {
	appID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		h.errorResponse(c, http.StatusBadRequest, "Invalid application ID")
		return
	}
	userID, _ := middleware.GetUserIDFromContext(c)
	app, err := h.Repository.GetCashForecastByID(uint(appID))
	if err != nil || app.CreatorID != userID {
		h.errorResponse(c, http.StatusForbidden, "You are not the creator of this application or it does not exist")
		return
	}
	if err := h.Repository.FormCashForecast(uint(appID)); err != nil {
		h.errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	h.successResponse(c, gin.H{"message": "Application has been formed and submitted for moderation"})
}

// CompleteCashForecast godoc
// @Summary Завершить заявку
// @Description Завершение заявки модератором с расчетом результата
// @Tags cash-forecasts
// @Security BearerAuth
// @Produce json
// @Param id path int true "ID заявки"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Router /api/cash-forecasts/{id}/complete [put]
func (h *Handler) CompleteCashForecast(c *gin.Context) {
	appID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		h.errorResponse(c, http.StatusBadRequest, "Invalid application ID")
		return
	}

	if !middleware.IsModeratorFromContext(c) {
		h.errorResponse(c, http.StatusForbidden, "Only moderators can complete applications")
		return
	}
	moderatorID, _ := middleware.GetUserIDFromContext(c)

	if err := h.Repository.CompleteCashForecast(uint(appID), moderatorID); err != nil {
		h.errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	h.successResponse(c, gin.H{"message": "Application completed successfully and result calculated"})
}

// RejectCashForecast godoc
// @Summary Отклонить заявку
// @Description Отклонение заявки модератором
// @Tags cash-forecasts
// @Security BearerAuth
// @Produce json
// @Param id path int true "ID заявки"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Router /api/cash-forecasts/{id}/reject [put]
func (h *Handler) RejectCashForecast(c *gin.Context) {
	appID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		h.errorResponse(c, http.StatusBadRequest, "Invalid application ID")
		return
	}

	if !middleware.IsModeratorFromContext(c) {
		h.errorResponse(c, http.StatusForbidden, "Only moderators can reject applications")
		return
	}
	moderatorID, _ := middleware.GetUserIDFromContext(c)

	if err := h.Repository.RejectCashForecast(uint(appID), moderatorID); err != nil {
		h.errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	h.successResponse(c, gin.H{"message": "Application rejected successfully"})
}

// DeleteCashForecast godoc
// @Summary Удалить заявку
// @Description Удаление заявки-черновика (мягкое удаление)
// @Tags cash-forecasts
// @Security BearerAuth
// @Produce json
// @Param id path int true "ID заявки"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Router /api/cash-forecasts/{id} [delete]
func (h *Handler) DeleteCashForecast(c *gin.Context) {
	appID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		h.errorResponse(c, http.StatusBadRequest, "Invalid application ID")
		return
	}
	userID, _ := middleware.GetUserIDFromContext(c)
	app, err := h.Repository.GetCashForecastByID(uint(appID))
	if err != nil || app.CreatorID != userID {
		h.errorResponse(c, http.StatusForbidden, "You are not the creator of this application or it does not exist")
		return
	}

	if err := h.Repository.DeleteCashForecast(uint(appID)); err != nil {
		h.errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	h.successResponse(c, gin.H{"message": "Application (Draft) soft deleted successfully"})
}

// AddAccountToCashForecast godoc
// @Summary Добавить счет в заявку
// @Description Добавление счета в заявку-черновик (создается автоматически если не существует)
// @Tags cash-forecast-items
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body ds.AddToFundsApplicationRequest true "Данные для добавления"
// @Success 200 {object} map[string]interface{} "message и application_id"
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/cash-forecast-items [post]
func (h *Handler) AddAccountToCashForecast(c *gin.Context) {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		h.errorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var req ds.AddToFundsApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.errorResponse(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}
	draftApp, err := h.Repository.GetUserDraftCashForecast(userID)
	if err != nil {
		h.errorResponse(c, http.StatusInternalServerError, "Failed to get/create draft application: "+err.Error())
		return
	}

	if err := h.Repository.AddAccountToCashForecast(draftApp.ID, req.AccountID); err != nil {
		h.errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	h.successResponse(c, gin.H{"message": "Account added to draft application", "application_id": draftApp.ID})
}

// RemoveItemFromCashForecast godoc
// @Summary Удалить элемент из заявки
// @Description Удаление элемента из заявки-черновика
// @Tags cash-forecast-items
// @Security BearerAuth
// @Produce json
// @Param itemID path int true "ID элемента"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /api/cash-forecast-items/{itemID} [delete]
func (h *Handler) RemoveItemFromCashForecast(c *gin.Context) {
	itemID, err := strconv.ParseUint(c.Param("itemID"), 10, 32)
	if err != nil {
		h.errorResponse(c, http.StatusBadRequest, "Invalid item ID")
		return
	}
	if err := h.Repository.RemoveItemFromCashForecast(uint(itemID)); err != nil {
		h.errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	h.successResponse(c, gin.H{"message": "Item removed from application"})
}

// UpdateCashForecastItem godoc
// @Summary Обновить элемент заявки
// @Description Изменение количества/порядка/значения в элементе заявки
// @Tags cash-forecast-items
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body object{item_id=uint} true "Данные для обновления элемента"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /api/cash-forecast-items [put]
func (h *Handler) UpdateCashForecastItem(c *gin.Context) {
	var req struct {
		ItemID uint `json:"item_id" binding:"required"`
		ds.UpdateFundsApplicationItemRequest
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.errorResponse(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}
	if err := h.Repository.UpdateCashForecastItem(req.ItemID, req.UpdateFundsApplicationItemRequest); err != nil {
		h.errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	h.successResponse(c, gin.H{"message": "Item updated successfully"})
}
