package main

import (
	"log"
	"net/http"

	"auth-crud/config"
	"auth-crud/handlers"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("Error loading .env file")
	}

	if err := config.ConnectDB(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/auth/register", handlers.Register)
	mux.HandleFunc("/api/v1/auth/login", handlers.Login)
	mux.HandleFunc("/api/v1/videos", handlers.GetVideos)
	mux.HandleFunc("/api/v1/videos/{id}", handlers.GetVideo)
	mux.HandleFunc("/api/admin/v1/videos", handlers.CreateVideo)
	mux.HandleFunc("/api/admin/v1/videos/{id}", handlers.UpdateVideo)

	mux.HandleFunc("/api/v1/categories", handlers.GetCategories)
	mux.HandleFunc("/api/v1/categories/{id}", handlers.GetCategory)
	mux.HandleFunc("/api/admin/v1/categories", handlers.CreateCategory)

	log.Fatal(http.ListenAndServe(":8080", mux))
}
