package utils

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

type ResponseStatus string

const (
	StatusSuccess ResponseStatus = "SUCCESS"
	StatusFail    ResponseStatus = "FAIL"
)

type ErrorInfo struct {
	Code        string `json:"code,omitempty"`
	Description string `json:"description,omitempty"`
}

type MetaInfo struct {
	Timestamp string `json:"timestamp"`
	RequestID string `json:"request_id"`
	TraceID   string `json:"trace_id,omitempty"`
}

type StandardResponse struct {
	Status  ResponseStatus `json:"status"`
	Message string         `json:"message"`
	Data    interface{}    `json:"data"`
	Error   *ErrorInfo     `json:"error,omitempty"`
	Meta    MetaInfo       `json:"meta"`
}

func generateRequestID() string {
	b := make([]byte, 12)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func GetOrSetRequestID(w http.ResponseWriter, r *http.Request) string {
	reqID := r.Header.Get("X-Request-Id")
	if reqID == "" {
		reqID = generateRequestID()
	}
	w.Header().Set("X-Request-Id", reqID)
	return reqID
}

func writeJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}

func JSONSuccess(w http.ResponseWriter, r *http.Request, message string, data interface{}) {
	resp := StandardResponse{
		Status:  StatusSuccess,
		Message: message,
		Data:    data,
		Meta: MetaInfo{
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			RequestID: GetOrSetRequestID(w, r),
		},
	}
	writeJSON(w, http.StatusOK, resp)
}

func JSONCreated(w http.ResponseWriter, r *http.Request, message string, data interface{}) {
	resp := StandardResponse{
		Status:  StatusSuccess,
		Message: message,
		Data:    data,
		Meta: MetaInfo{
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			RequestID: GetOrSetRequestID(w, r),
		},
	}
	writeJSON(w, http.StatusCreated, resp)
}

func JSONError(w http.ResponseWriter, r *http.Request, httpStatus int, message string, code string, description string) {
	resp := StandardResponse{
		Status:  StatusFail,
		Message: message,
		Data:    nil,
		Error: &ErrorInfo{
			Code:        code,
			Description: description,
		},
		Meta: MetaInfo{
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			RequestID: GetOrSetRequestID(w, r),
		},
	}
	writeJSON(w, httpStatus, resp)
}

// Password helpers
func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hashedPassword), err
}

func VerifyPassword(password, hashedPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// JWT helpers
func GenerateToken(userID uint) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(time.Hour * 24).Unix(),
	})
	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}

// Pagination helpers (cursor + sort)
// cursor is the last seen numeric id (string). sortBy: id|created_at (default id). order: asc|desc (default asc)
func ParsePagination(r *http.Request) (limit int, cursor string, sortBy string, order string) {
	limit = 20
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 100 {
			limit = n
		}
	}
	cursor = r.URL.Query().Get("cursor")
	sortBy = r.URL.Query().Get("sort_by")
	if sortBy == "" {
		sortBy = "id"
	}
	order = strings.ToLower(r.URL.Query().Get("order"))
	if order != "desc" {
		order = "asc"
	}
	return
}

func BuildNextCursor(itemsLen int, limit int, lastID uint) string {
	if itemsLen < limit {
		return ""
	}
	return strconv.FormatUint(uint64(lastID), 10)
}
