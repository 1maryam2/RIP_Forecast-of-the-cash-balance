package api

import (
	"lab_1/internal/app/handler"
	"lab_1/internal/app/repository"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func StartServer() {
	log.Println("Starting server")

	repo, err := repository.NewRepository()
	if err != nil {
		logrus.Error("ошибка инициализации репозитория")
	}

	handler := handler.NewHandler(repo)

	r := gin.Default()
	r.LoadHTMLGlob("D:/5 семестр/РИП/lab_1/templates/*")
	r.Static("/static", "D:/5 семестр/РИП/lab_1/resources")
	r.GET("/mainPage", handler.GetAllAccounts)
	r.GET("/account/:id", handler.GetAccount)
	r.GET("/cart/:id", handler.GetCart)
	r.Run()
	log.Println("Server down")
}
