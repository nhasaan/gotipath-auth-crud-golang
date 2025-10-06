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
	utils.Success(w, http.StatusOK, "Categories fetched successfully")
	utils.JSON(w, http.StatusOK, categories)
	return
}

func GetCategory(w http.ResponseWriter, r *http.Request) {
	var category models.Category
	if err := config.DB.Where("id = ?", r.URL.Query().Get("id")).First(&category).Error; err != nil {
		utils.Error(w, http.StatusNotFound, "Category not found")
		return
	}
	utils.Success(w, http.StatusOK, "Category fetched successfully")
	utils.JSON(w, http.StatusOK, category)
	return
}

func CreateCategory(w http.ResponseWriter, r *http.Request) {
	var category models.Category
	if err := json.NewDecoder(r.Body).Decode(&category); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	utils.Success(w, http.StatusCreated, "Category created successfully")
	utils.JSON(w, http.StatusCreated, category)
	return
}
