package handler

import (
	"lab_1/internal/app/repository"
	"net/http"
	"strconv"

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

func (h *Handler) GetAllAccounts(ctx *gin.Context) {
	var accounts []repository.Account
	var err error

	searchQuery := ctx.Query("query") // получаем значение из поля поиска
	if searchQuery == "" {            // если поле поиска пусто, то просто получаем из репозитория все записи
		accounts, err = h.Repository.GetAllAccounts()
		if err != nil {
			logrus.Error(err)
		}
	} else {
		accounts, err = h.Repository.GetAccountsByTitle(searchQuery) // в ином случае ищем заказ по заголовку
		if err != nil {
			logrus.Error(err)
		}
	}

	ctx.HTML(http.StatusOK, "mainPage.html", gin.H{
		"accounts": accounts,
		"query":    searchQuery, // передаем введенный запрос обратно на страницу
		// в ином случае оно будет очищаться при нажатии на кнопку
	})
}
func (h *Handler) GetAccount(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logrus.Error(err)
	}
	account, err := h.Repository.GetAccount(id)
	if err != nil {
		logrus.Error(err)
	}

	ctx.HTML(http.StatusOK, "accountPage.html", gin.H{
		"account": account,
	})
}
func (h *Handler) GetCart(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logrus.Error(err)
	}

	cartItems, err := h.Repository.GetCart(id)
	if err != nil {
		logrus.Error(err)
	}

	ctx.HTML(http.StatusOK, "cartPage.html", gin.H{
		"cartItems": cartItems,
	})
}
