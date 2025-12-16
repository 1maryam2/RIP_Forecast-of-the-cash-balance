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

	"github.com/gin-contrib/cors"
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
	eventManager := handler.NewEventManager()
	// Передаем его в Handler (нужно обновить структуру NewHandler)
	mainHandler := handler.NewHandler(repo, eventManager)
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	config := cors.DefaultConfig()
	config.AllowOriginFunc = func(origin string) bool {
		return origin == "http://localhost:5173" || origin == "https://tauri.localhost"
	}
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	router.Use(cors.New(config))
	api := router.Group("/api")
	{
		api.POST("/register", authHandler.Register)
		api.POST("/login", authHandler.Login)
		api.POST("/refresh", authHandler.Refresh)
		api.GET("/accounts/:id", mainHandler.GetAccountByID)
		api.GET("/accounts", mainHandler.GetAccounts)
		api.POST("/accounts", mainHandler.CreateAccount)
		api.GET("/cash-forecasts/cart", mainHandler.GetCartIcon)
		api.POST("/internal/result", mainHandler.ReceiveCalculationResult)
		authorized := api.Group("/")
		authorized.Use(middleware.AuthMiddleware(jwtService, redisClient))
		{
			authorized.POST("/logout", authHandler.Logout)
			authorized.GET("/users/profile", mainHandler.GetUserProfile)
			authorized.PUT("/users/profile", mainHandler.UpdateUserProfile)
			authorized.PUT("/accounts/:id", mainHandler.UpdateAccount)
			authorized.DELETE("/accounts/:id", mainHandler.DeleteAccount)
			authorized.POST("/accounts/:id/image", mainHandler.UploadAccountImage)

			authorized.GET("/cash-forecasts/:id", mainHandler.GetCashForecast)
			authorized.PUT("/cash-forecasts/:id", mainHandler.UpdateCashForecast)
			authorized.PUT("/cash-forecasts/:id/form", mainHandler.FormCashForecast)
			authorized.PUT("/cash-forecasts/:id/complete", mainHandler.CompleteCashForecast)
			authorized.PUT("/cash-forecasts/:id/reject", mainHandler.RejectCashForecast)
			authorized.DELETE("/cash-forecasts/:id", mainHandler.DeleteCashForecast)

			authorized.POST("/cash-forecast-items", mainHandler.AddAccountToCashForecast)
			authorized.DELETE("/cash-forecast-items/:itemID", mainHandler.RemoveItemFromCashForecast)
			authorized.PUT("/cash-forecast-items", mainHandler.UpdateCashForecastItem)
		}
	}

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	serverAddress := fmt.Sprintf("%s:%d", conf.ServiceHost, conf.ServicePort)
	if serverAddress == ":0" {
		serverAddress = ":8080"
	}

	logrus.Infof("Starting server on %s", serverAddress)
	if err := router.Run(serverAddress); err != nil {
		logrus.Fatalf("Failed to run server: %v", err)
	}
}
