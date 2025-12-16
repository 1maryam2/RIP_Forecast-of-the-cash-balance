package main

import (
	"lab_1/internal/app/ds"
	"lab_1/internal/app/dsn"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	_ = godotenv.Load("../../.env")
	db, err := gorm.Open(postgres.Open(dsn.FromEnv()), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	err = db.AutoMigrate(
		&ds.Account{},
		&ds.FundsApplication{},
		&ds.FundsApplicationItem{},
		&ds.Users{},
	)
	if err != nil {
		panic("cant migrate db")
	}
}
