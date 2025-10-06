package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"auth-crud/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() error {
	// dsn1 := "host=host.docker.internal user=postgres password=1qazZAQ! dbname=auth_crud_db port=5432 sslmode=disable"

	dsn := os.Getenv("DB_URL")
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("Failed to connect to database")
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("Failed to get database instance")
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Println("Connected to database successfully")
	log.Println("Running DB migrations...")
	DB.AutoMigrate(&models.User{}, &models.Video{}, &models.Category{})

	return nil
}

func CloseDB() error {
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("Failed to get database instance")
	}
	sqlDB.Close()
	return nil
}
