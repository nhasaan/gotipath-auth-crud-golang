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

	user := models.User{
		Email:    input.Email,
		Password: input.Password,
	}

	if err := config.DB.Create(&user).Error; err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	utils.Success(w, http.StatusCreated, "User created successfully")
	return
}

func Login(w http.ResponseWriter, r *http.Request) {
	var input LoginInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	user := models.User{
		Email:    input.Email,
		Password: input.Password,
	}

	if err := config.DB.Where("email = ?", strings.ToLower(input.Email)).First(&user).Error; err != nil {
		utils.Error(w, http.StatusNotFound, "User not found")
		return
	}

	if err := utils.VerifyPassword(input.Password, user.Password); err != nil {
		utils.Error(w, http.StatusUnauthorized, "Invalid password")
		return
	}

	token, err := utils.GenerateToken(user.ID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	utils.Success(w, http.StatusOK, "User logged in successfully")
	utils.JSON(w, http.StatusOK, map[string]string{"token": token})
	return
}
