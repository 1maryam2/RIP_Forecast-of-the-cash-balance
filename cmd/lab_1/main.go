package main

import (
	"fmt"
	"os"

	"lab_1/internal/app/config"
	"lab_1/internal/app/dsn"
	"lab_1/internal/app/handler"
	"lab_1/internal/app/repository"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/sirupsen/logrus"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file, trying parent dirs...")
		godotenv.Load("../../.env")
	}

	router := gin.Default()
	conf, err := config.NewConfig()
	if err != nil {
		logrus.Fatalf("error loading config: %v", err)
	}

	var minioClient *minio.Client
	endpoint := os.Getenv("MINIO_ENDPOINT")
	accessKeyID := os.Getenv("MINIO_ACCESS_KEY_ID")
	secretAccessKey := os.Getenv("MINIO_SECRET_ACCESS_KEY")
	useSSL := os.Getenv("MINIO_USE_SSL") == "true"

	if endpoint != "" && accessKeyID != "" && secretAccessKey != "" {
		minioClient, err = minio.New(endpoint, &minio.Options{
			Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
			Secure: useSSL,
		})
		if err != nil {
			logrus.Errorf("Error initializing MinIO client, continuing with mock: %v", err)
			minioClient = nil
		} else {
			logrus.Info("MinIO client initialized successfully.")
		}
	} else {
		logrus.Warn("MinIO environment variables not found. Using MockMinioClient in repository.")
	}

	postgresString := dsn.FromEnv()
	fmt.Println("DSN string:", postgresString)
	rep, errRep := repository.New(postgresString, minioClient)
	if errRep != nil {
		logrus.Fatalf("error initializing repository: %v", errRep)
	}
	hand := handler.NewHandler(rep)
	hand.RegisterRoutes(router)
	serverAddress := fmt.Sprintf("%s:%d", conf.ServiceHost, conf.ServicePort)
	if serverAddress == ":0" {
		serverAddress = ":8080"
	}

	fmt.Printf("Starting server on %s\n", serverAddress)
	if err := router.Run(serverAddress); err != nil {
		logrus.Fatalf("Failed to run server: %v", err)
	}
}
