package repository

import (
	"lab_1/internal/app/ds"

	"github.com/minio/minio-go/v7"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const AccountBucket = "account-images"

type Repository struct {
	db          *gorm.DB
	MinioClient *minio.Client
}

func New(dsn string, minioClient *minio.Client) (*Repository, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	err = db.AutoMigrate(&ds.Users{}, &ds.Account{}, &ds.FundsApplication{}, &ds.FundsApplicationItem{})
	if err != nil {
		return nil, err
	}
	return &Repository{
		db:          db,
		MinioClient: minioClient,
	}, nil
}
