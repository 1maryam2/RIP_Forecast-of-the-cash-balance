package main

import (
	"context"
	"fmt"
	_ "lab_1/docs"
	"lab_1/internal/app/auth"
	"lab_1/internal/app/config"
	"lab_1/internal/app/dsn"
	"lab_1/internal/app/handler"
	"lab_1/internal/app/middleware"
	"lab_1/internal/app/redis"
	"lab_1/internal/app/repository"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/minio/minio-go/v7"
	"github.com/sirupsen/logrus"
)

// @title Cash Forecast Service
// @version 1.0
// @description Service for cash flow forecasting and fund applications

// @contact.name API Support
// @contact.url http://example.com
// @contact.email support@example.com

// @license.name MIT

// @host localhost:8081
// @BasePath /
// @schemes http

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description JWT Authorization header using the Bearer scheme. Example: "Bearer {token}"

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file, trying parent dirs...")
		godotenv.Load("../.env")
	}
	ctx := context.Background()
	conf, err := config.NewConfig(ctx)
	if err != nil {
		logrus.Fatalf("error loading config: %v", err)
	}
	redisClient, err := redis.New(ctx, conf.Redis)
	if err != nil {
		logrus.Fatalf("error initializing redis: %v", err)
	}
	defer redisClient.Close()
	var minioClient *minio.Client
	postgresString := dsn.FromEnv()
	repo, err := repository.New(postgresString, minioClient)
	if err != nil {
		logrus.Fatalf("error initializing repository: %v", err)
	}
	jwtService := auth.NewJWTService(conf)
	authHandler := handler.NewAuthHandler(repo, jwtService, redisClient)
	mainHandler := handler.NewHandler(repo)
	router := gin.Default()
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	router.POST("/api/register", authHandler.Register)
	router.POST("/api/login", authHandler.Login)
	router.POST("/api/refresh", authHandler.Refresh)
	authRoutes := router.Group("/api")
	authRoutes.Use(middleware.AuthMiddleware(jwtService, redisClient))
	{
		authRoutes.POST("/logout", authHandler.Logout)

		// User routes
		authRoutes.GET("/users/profile", mainHandler.GetUserProfile)
		authRoutes.PUT("/users/profile", mainHandler.UpdateUserProfile)

		// Account routes
		authRoutes.GET("/accounts", mainHandler.GetAccounts)
		authRoutes.GET("/accounts/:id", mainHandler.GetAccountByID)
		authRoutes.POST("/accounts", mainHandler.CreateAccount)
		authRoutes.PUT("/accounts/:id", mainHandler.UpdateAccount)
		authRoutes.DELETE("/accounts/:id", mainHandler.DeleteAccount)
		authRoutes.POST("/accounts/:id/image", mainHandler.UploadAccountImage)

		// Funds application routes
		authRoutes.GET("/cash-forecasts", mainHandler.GetCashForecasts)
		authRoutes.GET("/cash-forecasts/cart", mainHandler.GetCartIcon)
		authRoutes.GET("/cash-forecasts/:id", mainHandler.GetCashForecast)
		authRoutes.PUT("/cash-forecasts/:id", mainHandler.UpdateCashForecast)
		authRoutes.PUT("/cash-forecasts/:id/form", mainHandler.FormCashForecast)
		authRoutes.PUT("/cash-forecasts/:id/complete", mainHandler.CompleteCashForecast)
		authRoutes.PUT("/cash-forecasts/:id/reject", mainHandler.RejectCashForecast)
		authRoutes.DELETE("/cash-forecasts/:id", mainHandler.DeleteCashForecast)

		// Funds application items routes
		authRoutes.POST("/cash-forecast-items", mainHandler.AddAccountToCashForecast)
		authRoutes.DELETE("/cash-forecast-items/:itemID", mainHandler.RemoveItemFromCashForecast)
		authRoutes.PUT("/cash-forecast-items", mainHandler.UpdateCashForecastItem)
	}
	serverAddress := fmt.Sprintf("%s:%d", conf.ServiceHost, conf.ServicePort)
	if serverAddress == ":0" {
		serverAddress = ":8080"
	}

	logrus.Infof("Starting server on %s", serverAddress)
	if err := router.Run(serverAddress); err != nil {
		logrus.Fatalf("Failed to run server: %v", err)
	}
}
