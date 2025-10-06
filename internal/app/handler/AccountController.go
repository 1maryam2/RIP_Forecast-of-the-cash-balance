package handler

import (
	"lab_1/internal/app/repository"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	Repository *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{
		Repository: r,
	}
}
func (h *Handler) RegisterHandler(router *gin.Engine) {
	router.GET("/mainPage", h.GetAllAccounts)
	router.GET("/account/:id", h.GetAccount)
	router.GET("/", h.GetAllAccounts)
	router.GET("/fundsApplication/:id", h.GetFundsApplication)
	router.POST("/add-to-fundsApplication", h.AddToFundsApplication)
	router.POST("/delete-account", h.DeleteAccount)
}

func (h *Handler) RegisterStatic(router *gin.Engine) {
	router.LoadHTMLGlob("../../templates/*")
	router.Static("/static", "D:/5 семестр/РИП/lab_1/resources")
}
func (h *Handler) errorHandler(ctx *gin.Context, errorStatusCode int, err error) {
	logrus.Error(err.Error())
	ctx.JSON(errorStatusCode, gin.H{
		"status":      "error",
		"description": err.Error(),
	})
}
