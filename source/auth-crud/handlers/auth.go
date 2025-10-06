package handlers

import (
	"auth-crud/config"
	"auth-crud/models"
	"auth-crud/utils"
	"encoding/json"
	"net/http"
	"strings"
)

type RegisterInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func Register(w http.ResponseWriter, r *http.Request) {
	var input RegisterInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.JSONError(w, r, http.StatusBadRequest, "Invalid request body", "invalid_request", err.Error())
		return
	}

	input.Email = strings.TrimSpace(strings.ToLower(input.Email))
	if input.Email == "" || input.Password == "" {
		utils.JSONError(w, r, http.StatusBadRequest, "Email and password are required", "validation_error", "missing email or password")
		return
	}

	// check uniqueness
	var existing models.User
	if err := config.DB.Where("email = ?", input.Email).First(&existing).Error; err == nil {
		utils.JSONError(w, r, http.StatusBadRequest, "Email already in use", "duplicate_email", "email exists")
		return
	}

	hashed, err := utils.HashPassword(input.Password)
	if err != nil {
		utils.JSONError(w, r, http.StatusInternalServerError, "Failed to process password", "hash_failed", err.Error())
		return
	}

	user := models.User{
		Email:    input.Email,
		Password: hashed,
	}

	if err := config.DB.Create(&user).Error; err != nil {
		utils.JSONError(w, r, http.StatusInternalServerError, "Failed to create user", "db_create_failed", err.Error())
		return
	}

	utils.JSONCreated(w, r, "User created successfully", map[string]interface{}{
		"id":    user.ID,
		"email": user.Email,
	})
}

func Login(w http.ResponseWriter, r *http.Request) {
	var input LoginInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.JSONError(w, r, http.StatusBadRequest, "Invalid request body", "invalid_request", err.Error())
		return
	}

	email := strings.TrimSpace(strings.ToLower(input.Email))
	if email == "" || input.Password == "" {
		utils.JSONError(w, r, http.StatusBadRequest, "Email and password are required", "validation_error", "missing email or password")
		return
	}

	var user models.User
	if err := config.DB.Where("email = ?", email).First(&user).Error; err != nil {
		utils.JSONError(w, r, http.StatusNotFound, "User not found", "user_not_found", "no user for email")
		return
	}

	if err := utils.VerifyPassword(input.Password, user.Password); err != nil {
		utils.JSONError(w, r, http.StatusUnauthorized, "Invalid credentials", "invalid_credentials", "password mismatch")
		return
	}

	token, err := utils.GenerateToken(user.ID)
	if err != nil {
		utils.JSONError(w, r, http.StatusInternalServerError, "Failed to generate token", "token_failed", err.Error())
		return
	}

	utils.JSONSuccess(w, r, "User logged in successfully", map[string]string{"token": token})
}
