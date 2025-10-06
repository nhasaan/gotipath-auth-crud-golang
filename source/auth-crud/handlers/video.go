package handlers

import (
	"auth-crud/config"
	"auth-crud/models"
	"auth-crud/utils"
	"encoding/json"
	"net/http"
	"strconv"
)

func GetVideos(w http.ResponseWriter, r *http.Request) {
	var videos []models.Video
	if err := config.DB.Find(&videos).Error; err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to get videos")
		return
	}
	utils.JSON(w, http.StatusOK, videos)
}

func GetVideo(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if idStr == "" {
		utils.Error(w, http.StatusBadRequest, "Missing video id")
		return
	}
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid video id")
		return
	}

	var video models.Video
	if err := config.DB.First(&video, id).Error; err != nil {
		utils.Error(w, http.StatusNotFound, "Video not found")
		return
	}
	utils.JSON(w, http.StatusOK, video)
}

// VideoInput represents the payload for creating/updating a video.
// Only fields that clients are allowed to set are included here.
type VideoInput struct {
	Title         string `json:"title"`
	Duration      string `json:"duration"`
	URL           string `json:"url"`
	ThumbnailPath string `json:"thumbnailPath"`
	CategoryID    uint   `json:"categoryId"`
}

func CreateVideo(w http.ResponseWriter, r *http.Request) {
	var input VideoInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if input.Title == "" || input.Duration == "" || input.URL == "" || input.ThumbnailPath == "" || input.CategoryID == 0 {
		utils.Error(w, http.StatusBadRequest, "Missing required fields")
		return
	}

	// Ensure category exists
	var category models.Category
	if err := config.DB.First(&category, input.CategoryID).Error; err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid categoryId")
		return
	}

	video := models.Video{
		Title:         input.Title,
		Duration:      input.Duration,
		URL:           input.URL,
		ThumbnailPath: input.ThumbnailPath,
		CategoryID:    input.CategoryID,
	}

	if err := config.DB.Create(&video).Error; err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to create video")
		return
	}
	utils.JSON(w, http.StatusCreated, video)
}

func UpdateVideo(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if idStr == "" {
		utils.Error(w, http.StatusBadRequest, "Missing video id")
		return
	}
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid video id")
		return
	}

	var existing models.Video
	if err := config.DB.First(&existing, id).Error; err != nil {
		utils.Error(w, http.StatusNotFound, "Video not found")
		return
	}

	var input VideoInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// If categoryId is provided, validate it
	if input.CategoryID != 0 {
		var category models.Category
		if err := config.DB.First(&category, input.CategoryID).Error; err != nil {
			utils.Error(w, http.StatusBadRequest, "Invalid categoryId")
			return
		}
		existing.CategoryID = input.CategoryID
	}

	if input.Title != "" {
		existing.Title = input.Title
	}
	if input.Duration != "" {
		existing.Duration = input.Duration
	}
	if input.URL != "" {
		existing.URL = input.URL
	}
	if input.ThumbnailPath != "" {
		existing.ThumbnailPath = input.ThumbnailPath
	}

	if err := config.DB.Save(&existing).Error; err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to update video")
		return
	}
	utils.JSON(w, http.StatusOK, existing)
}
