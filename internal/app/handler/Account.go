package handler

import (
	"fmt"
	"lab_1/internal/app/ds"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func (h *Handler) AddToFundsApplication(ctx *gin.Context) {
	accountIDStr := ctx.PostForm("account_id")

	accountID, err := strconv.Atoi(accountIDStr)
	if err != nil {
		logrus.Errorf("Invalid account_id: %s", accountIDStr)
		ctx.Redirect(http.StatusFound, "/mainPage")
		return
	}
	userID := uint(1)
	fundsApplication, err := h.Repository.GetOrCreateFundsApplicationForUser(userID)
	if err != nil {
		logrus.Errorf("Error getting or creating cart: %v", err)
		ctx.Redirect(http.StatusFound, "/mainPage")
		return
	}

	err = h.Repository.AddToFundsApplication(uint(accountID), fundsApplication.ID)
	if err != nil {
		logrus.Errorf("Error adding to cart: %v", err)
	}

	ctx.Redirect(http.StatusFound, "/mainPage")
}

func (h *Handler) GetAllAccounts(ctx *gin.Context) {
	var accounts []ds.Account
	var err error
	userID := uint(1)
	fundsApplicationID, err := h.Repository.GetOrCreateFundsApplicationForUser(userID)
	if err != nil {
		logrus.Error(err)
		return
	}
	searchQuery := ctx.Query("accountSearch")
	if searchQuery == "" {
		accounts, err = h.Repository.GetAllAccounts()
	} else {
		accounts, err = h.Repository.GetAccountsByTitle(searchQuery)
	}
	if err != nil {
		logrus.Error(err)
	}
	fundsApplicationItemIDs, err := h.Repository.GetAccountIDsInCart(fundsApplicationID.ID)
	if err != nil {
		logrus.Error(err)
	}
	fundsIn := make(map[uint]bool)
	for _, id := range fundsApplicationItemIDs {
		fundsIn[id] = true
	}
	for i := range accounts {
		if _, found := fundsIn[accounts[i].ID]; found {
			accounts[i].InCart = true
		}
	}

	ctx.HTML(http.StatusOK, "mainPage.html", gin.H{
		"accounts":           accounts,
		"accounts_count":     h.Repository.GetFundsApplicationCount(fundsApplicationID.ID),
		"accountSearch":      searchQuery,
		"FundsApplicationID": fundsApplicationID.ID,
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

func (h *Handler) GetFundsApplication(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logrus.Error(err)
		ctx.HTML(http.StatusBadRequest, "FundsApplicationPage.html", gin.H{
			"FundsApplicationItems": nil,
		})
		return
	}

	fundsApplicationData, err := h.Repository.GetFundsApplication(id)
	if err != nil {
		logrus.Error(err)
		ctx.HTML(http.StatusNotFound, "FundsApplicationPage.html", gin.H{
			"FundsApplicationItems": nil,
		})
		return
	}
	fundsApplicationData["FundsApplicationID"] = id
	ctx.HTML(http.StatusOK, "FundsApplicationPage.html", fundsApplicationData)
}

func (h *Handler) DeleteAccount(ctx *gin.Context) {
	strId := ctx.PostForm("funds_application_item_id")
	if strId == "" {
		logrus.Error("Empty funds_application_item_id received")
		ctx.Redirect(http.StatusFound, "/mainPage")
		return
	}
	strFundsApplicationId := ctx.PostForm("funds_application_id")
	if strFundsApplicationId == "" {
		logrus.Error("Empty funds_application_id received")
		ctx.Redirect(http.StatusFound, "/mainPage")
		return
	}

	FundsApplicationItemId, err := strconv.Atoi(strId)
	if err != nil {
		logrus.Errorf("Invalid funds_application_item_id: %s", strId)
		ctx.Redirect(http.StatusFound, "/mainPage")
		return
	}
	FundsApplicationId, err := strconv.Atoi(strFundsApplicationId)
	if err != nil {
		logrus.Errorf("Invalid funds_application_id: %s", strFundsApplicationId)
		ctx.Redirect(http.StatusFound, "/mainPage")
		return
	}
	err = h.Repository.DeleteFundsApplicationItem(uint(FundsApplicationItemId))
	if err != nil {
		logrus.Errorf("Error deleting item: %v", err)
	}

	result, err := h.Repository.CalculateFundsApplicationResult(FundsApplicationId)
	if err != nil {
		logrus.Errorf("Error calculating result: %v", err)
	}
	err = h.Repository.UpdateFundsApplicationResult(FundsApplicationId, result)
	if err != nil {
		logrus.Errorf("Error updating result: %v", err)
	}
	ctx.Redirect(http.StatusFound, fmt.Sprintf("/fundsApplication/%d", FundsApplicationId))
}
