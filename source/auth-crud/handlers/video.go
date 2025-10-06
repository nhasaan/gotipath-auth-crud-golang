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
	limit, cursor, sortBy, order := utils.ParsePagination(r)
	var videos []models.Video

	q := config.DB.Model(&models.Video{})
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
	if err := q.Limit(limit).Find(&videos).Error; err != nil {
		utils.JSONError(w, r, http.StatusInternalServerError, "Failed to get videos", "db_query_failed", err.Error())
		return
	}

	nextCursor := ""
	if len(videos) > 0 {
		last := videos[len(videos)-1]
		nextCursor = utils.BuildNextCursor(len(videos), limit, last.ID)
	}

	utils.JSONSuccess(w, r, "Successfully retrieved the videos", map[string]interface{}{
		"items":       videos,
		"next_cursor": nextCursor,
	})
}

func GetVideo(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if idStr == "" {
		utils.JSONError(w, r, http.StatusBadRequest, "Missing video id", "validation_error", "")
		return
	}
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		utils.JSONError(w, r, http.StatusBadRequest, "Invalid video id", "validation_error", "")
		return
	}

	var video models.Video
	if err := config.DB.First(&video, id).Error; err != nil {
		utils.JSONError(w, r, http.StatusNotFound, "Video not found", "not_found", "")
		return
	}
	utils.JSONSuccess(w, r, "Successfully retrieved the video", video)
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
		utils.JSONError(w, r, http.StatusBadRequest, "Invalid request body", "invalid_request", err.Error())
		return
	}

	if input.Title == "" || input.Duration == "" || input.URL == "" || input.ThumbnailPath == "" || input.CategoryID == 0 {
		utils.JSONError(w, r, http.StatusBadRequest, "Missing required fields", "validation_error", "")
		return
	}

	// Ensure category exists
	var category models.Category
	if err := config.DB.First(&category, input.CategoryID).Error; err != nil {
		utils.JSONError(w, r, http.StatusBadRequest, "Invalid categoryId", "validation_error", "")
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
		utils.JSONError(w, r, http.StatusInternalServerError, "Failed to create video", "db_create_failed", err.Error())
		return
	}
	utils.JSONCreated(w, r, "Video created successfully", video)
}

func UpdateVideo(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if idStr == "" {
		utils.JSONError(w, r, http.StatusBadRequest, "Missing video id", "validation_error", "")
		return
	}
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		utils.JSONError(w, r, http.StatusBadRequest, "Invalid video id", "validation_error", "")
		return
	}

	var existing models.Video
	if err := config.DB.First(&existing, id).Error; err != nil {
		utils.JSONError(w, r, http.StatusNotFound, "Video not found", "not_found", "")
		return
	}

	var input VideoInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.JSONError(w, r, http.StatusBadRequest, "Invalid request body", "invalid_request", err.Error())
		return
	}

	// If categoryId is provided, validate it
	if input.CategoryID != 0 {
		var category models.Category
		if err := config.DB.First(&category, input.CategoryID).Error; err != nil {
			utils.JSONError(w, r, http.StatusBadRequest, "Invalid categoryId", "validation_error", "")
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
		utils.JSONError(w, r, http.StatusInternalServerError, "Failed to update video", "db_update_failed", err.Error())
		return
	}
	utils.JSONSuccess(w, r, "Video updated successfully", existing)
}
