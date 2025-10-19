package handler

import (
	"fmt"
	"lab_1/internal/app/ds"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

func getUserIDFromContext(c *gin.Context) (uint, error) {
	if id, ok := c.Get("userID"); ok {
		return id.(uint), nil
	}
	return 1, nil
}

func isModeratorFromContext(c *gin.Context) bool {
	if isMod, ok := c.Get("isModerator"); ok {
		return isMod.(bool)
	}
	// Заглушка: ID 2 - модератор
	userID, _ := getUserIDFromContext(c)
	return userID == 2
}

func mockAuthMiddleware(c *gin.Context) {
	token := c.GetHeader("Authorization")
	userID := uint(1) // По умолчанию Creator
	isModerator := false

	if strings.HasPrefix(token, "Bearer dummy-jwt-token-for-moderator") {
		userID = 2
		isModerator = true
	} else if strings.HasPrefix(token, "Bearer dummy-jwt-token-for-creator") {
		userID = 1
		isModerator = false
	}
	c.Set("userID", userID)
	c.Set("isModerator", isModerator)
	c.Next()
}

func (h *Handler) RegisterRoutes(router *gin.Engine) {
	api := router.Group("/api")
	api.POST("/register", h.Register)
	api.POST("/login", h.Login)
	authApi := api.Group("/", mockAuthMiddleware)
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

func (h *Handler) Logout(c *gin.Context) {
	h.successResponse(c, gin.H{"message": "Logged out successfully"})
}

func (h *Handler) GetUserProfile(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
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

func (h *Handler) UpdateUserProfile(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
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

func (h *Handler) GetCartIcon(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
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

func (h *Handler) FormCashForecast(c *gin.Context) {
	appID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		h.errorResponse(c, http.StatusBadRequest, "Invalid application ID")
		return
	}
	userID, _ := getUserIDFromContext(c)
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

func (h *Handler) CompleteCashForecast(c *gin.Context) {
	appID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		h.errorResponse(c, http.StatusBadRequest, "Invalid application ID")
		return
	}

	if !isModeratorFromContext(c) {
		h.errorResponse(c, http.StatusForbidden, "Only moderators can complete applications")
		return
	}
	moderatorID, _ := getUserIDFromContext(c)

	if err := h.Repository.CompleteCashForecast(uint(appID), moderatorID); err != nil {
		h.errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	h.successResponse(c, gin.H{"message": "Application completed successfully and result calculated"})
}

func (h *Handler) RejectCashForecast(c *gin.Context) {
	appID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		h.errorResponse(c, http.StatusBadRequest, "Invalid application ID")
		return
	}

	if !isModeratorFromContext(c) {
		h.errorResponse(c, http.StatusForbidden, "Only moderators can reject applications")
		return
	}
	moderatorID, _ := getUserIDFromContext(c)

	if err := h.Repository.RejectCashForecast(uint(appID), moderatorID); err != nil {
		h.errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	h.successResponse(c, gin.H{"message": "Application rejected successfully"})
}

func (h *Handler) DeleteCashForecast(c *gin.Context) {
	appID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		h.errorResponse(c, http.StatusBadRequest, "Invalid application ID")
		return
	}
	userID, _ := getUserIDFromContext(c)
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

func (h *Handler) AddAccountToCashForecast(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
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
