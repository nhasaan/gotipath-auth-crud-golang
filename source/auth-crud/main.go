package main

import (
	"net/http"

	"auth-crud/config"
	"auth-crud/handlers"
	"auth-crud/loggers"
	"auth-crud/middlewares"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	// Load environment variables
	if err := godotenv.Load(); err != nil {
		loggers.Info(".env not found, relying on environment variables")
	}

	if err := config.ConnectDB(); err != nil {
		loggers.Error("Failed to connect to database:", err)
		return
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/auth/register", handlers.Register)
	mux.HandleFunc("/api/v1/auth/login", handlers.Login)
	mux.HandleFunc("/api/v1/videos", handlers.GetVideos)
	mux.HandleFunc("/api/v1/videos/{id}", handlers.GetVideo)
	mux.HandleFunc("/api/admin/v1/videos", middlewares.RequireAdmin(handlers.CreateVideo))
	mux.HandleFunc("/api/admin/v1/videos/{id}", middlewares.RequireAdmin(handlers.UpdateVideo))

	mux.HandleFunc("/api/v1/categories", handlers.GetCategories)
	mux.HandleFunc("/api/v1/categories/{id}", handlers.GetCategory)
	mux.HandleFunc("/api/admin/v1/categories", middlewares.RequireAdmin(handlers.CreateCategory))

	// Uploads
	mux.HandleFunc("/api/admin/v1/uploads", middlewares.RequireAdmin(handlers.UploadFile))
	mux.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("/uploads"))))

	loggers.Info("HTTP server listening on :8080")
	_ = http.ListenAndServe(":8080", mux)
}
