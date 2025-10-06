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
		utils.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	input.Email = strings.TrimSpace(strings.ToLower(input.Email))
	if input.Email == "" || input.Password == "" {
		utils.Error(w, http.StatusBadRequest, "Email and password are required")
		return
	}

	// check uniqueness
	var existing models.User
	if err := config.DB.Where("email = ?", input.Email).First(&existing).Error; err == nil {
		utils.Error(w, http.StatusBadRequest, "Email already in use")
		return
	}

	hashed, err := utils.HashPassword(input.Password)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to process password")
		return
	}

	user := models.User{
		Email:    input.Email,
		Password: hashed,
	}

	if err := config.DB.Create(&user).Error; err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	utils.JSON(w, http.StatusCreated, map[string]interface{}{
		"id":    user.ID,
		"email": user.Email,
	})
}

func Login(w http.ResponseWriter, r *http.Request) {
	var input LoginInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	email := strings.TrimSpace(strings.ToLower(input.Email))
	if email == "" || input.Password == "" {
		utils.Error(w, http.StatusBadRequest, "Email and password are required")
		return
	}

	var user models.User
	if err := config.DB.Where("email = ?", email).First(&user).Error; err != nil {
		utils.Error(w, http.StatusNotFound, "User not found")
		return
	}

	if err := utils.VerifyPassword(input.Password, user.Password); err != nil {
		utils.Error(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	token, err := utils.GenerateToken(user.ID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]string{"token": token})
}
