package handlers

import (
	"auth-crud/config"
	"auth-crud/models"
	"auth-crud/utils"
	"encoding/json"
	"net/http"
	"strconv"
)

func GetCategories(w http.ResponseWriter, r *http.Request) {
	limit, cursor, sortBy, order := utils.ParsePagination(r)
	var categories []models.Category

	q := config.DB.Model(&models.Category{})
	if cursor != "" {
		if id, err := strconv.Atoi(cursor); err == nil {
			if order == "asc" {
				q = q.Where("id > ?", id)
			} else {
				q = q.Where("id < ?", id)
			}
		}
	}
	if sortBy == "created_at" {
		q = q.Order("created_at " + order)
	} else {
		q = q.Order("id " + order)
	}
	if err := q.Limit(limit).Find(&categories).Error; err != nil {
		utils.JSONError(w, r, http.StatusInternalServerError, "Failed to get categories", "db_query_failed", err.Error())
		return
	}

	nextCursor := ""
	if len(categories) > 0 {
		last := categories[len(categories)-1]
		nextCursor = utils.BuildNextCursor(len(categories), limit, last.ID)
	}

	utils.JSONSuccess(w, r, "Successfully retrieved the categories", map[string]interface{}{
		"items":       categories,
		"next_cursor": nextCursor,
	})
}

func GetCategory(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var category models.Category
	if err := config.DB.Where("id = ?", id).First(&category).Error; err != nil {
		utils.JSONError(w, r, http.StatusNotFound, "Category not found", "not_found", "")
		return
	}
	utils.JSONSuccess(w, r, "Successfully retrieved the category", category)
}

func CreateCategory(w http.ResponseWriter, r *http.Request) {
	var input models.Category
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.JSONError(w, r, http.StatusBadRequest, "Invalid request body", "invalid_request", err.Error())
		return
	}
	if input.Name == "" {
		utils.JSONError(w, r, http.StatusBadRequest, "Name is required", "validation_error", "")
		return
	}

	// ensure unique name
	var exists models.Category
	if err := config.DB.Where("name = ?", input.Name).First(&exists).Error; err == nil {
		utils.JSONError(w, r, http.StatusBadRequest, "Category already exists", "duplicate_name", "")
		return
	}

	category := models.Category{Name: input.Name}
	if err := config.DB.Create(&category).Error; err != nil {
		utils.JSONError(w, r, http.StatusInternalServerError, "Failed to create category", "db_create_failed", err.Error())
		return
	}
	utils.JSONCreated(w, r, "Category created successfully", category)
}
