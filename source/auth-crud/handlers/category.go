package handlers

import (
	"auth-crud/config"
	"auth-crud/models"
	"auth-crud/utils"
	"encoding/json"
	"net/http"
)

func GetCategories(w http.ResponseWriter, r *http.Request) {
	var categories []models.Category
	if err := config.DB.Find(&categories).Error; err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to get categories")
		return
	}
	utils.JSON(w, http.StatusOK, categories)
}

func GetCategory(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var category models.Category
	if err := config.DB.Where("id = ?", id).First(&category).Error; err != nil {
		utils.Error(w, http.StatusNotFound, "Category not found")
		return
	}
	utils.JSON(w, http.StatusOK, category)
}

func CreateCategory(w http.ResponseWriter, r *http.Request) {
	var input models.Category
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if input.Name == "" {
		utils.Error(w, http.StatusBadRequest, "Name is required")
		return
	}

	// ensure unique name
	var exists models.Category
	if err := config.DB.Where("name = ?", input.Name).First(&exists).Error; err == nil {
		utils.Error(w, http.StatusBadRequest, "Category already exists")
		return
	}

	category := models.Category{Name: input.Name}
	if err := config.DB.Create(&category).Error; err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to create category")
		return
	}
	utils.JSON(w, http.StatusCreated, category)
}
