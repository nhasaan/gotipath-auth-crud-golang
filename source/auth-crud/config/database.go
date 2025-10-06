package config

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"auth-crud/loggers"
	"auth-crud/models"
	"auth-crud/utils"

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

	loggers.Info("Connected to database successfully")
	loggers.Info("Running DB migrations...")
	DB.AutoMigrate(&models.User{}, &models.Video{}, &models.Category{})

	if os.Getenv("SEED_DATA") == "true" {
		seedDatabase()
	}

	return nil
}

func seedDatabase() {
	rand.Seed(time.Now().UnixNano())
	// Categories
	var count int64
	DB.Model(&models.Category{}).Count(&count)
	if count == 0 {
		for i := 1; i <= 30; i++ {
			DB.Create(&models.Category{Name: fmt.Sprintf("Category %d", i)})
		}
	}
	// Users
	DB.Model(&models.User{}).Count(&count)
	if count == 0 {
		for i := 1; i <= 30; i++ {
			email := fmt.Sprintf("user%02d@example.com", i)
			pwd, _ := utils.HashPassword("Passw0rd!")
			DB.Create(&models.User{Email: email, Password: pwd, IsAdmin: i == 1})
		}
	}
	// Videos
	DB.Model(&models.Video{}).Count(&count)
	if count == 0 {
		var cats []models.Category
		DB.Find(&cats)
		for i := 1; i <= 30; i++ {
			c := cats[rand.Intn(len(cats))]
			DB.Create(&models.Video{
				Title:         fmt.Sprintf("Video %02d", i),
				Duration:      fmt.Sprintf("%dm", 5+(i%15)),
				URL:           fmt.Sprintf("https://example.com/video%02d.mp4", i),
				ThumbnailPath: "/uploads/sample.png",
				CategoryID:    c.ID,
			})
		}
	}
	loggers.Info("Seeding complete")
}

func CloseDB() error {
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("Failed to get database instance")
	}
	sqlDB.Close()
	return nil
}
