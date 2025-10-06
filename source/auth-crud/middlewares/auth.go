package middlewares

import (
	"auth-crud/config"
	"auth-crud/models"
	"net/http"
	"os"
	"strconv"
	"strings"

	"context"

	"github.com/golang-jwt/jwt/v4"
)

// contextKey is an unexported type to avoid key collisions in context
// when storing authenticated user information.
type contextKey string

const (
	contextUserIDKey contextKey = "auth.userId"
	contextUserKey   contextKey = "auth.user"
)

// GetAuthenticatedUser returns the authenticated user from the request context, if present.
func GetAuthenticatedUser(r *http.Request) (*models.User, bool) {
	val := r.Context().Value(contextUserKey)
	user, ok := val.(*models.User)
	return user, ok
}

// RequireAuth validates the Bearer token, loads the user, and injects it into the request context.
func RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authz := r.Header.Get("Authorization")
		if authz == "" || !strings.HasPrefix(strings.ToLower(authz), "bearer ") {
			http.Error(w, "missing or invalid authorization header", http.StatusUnauthorized)
			return
		}
		tokenString := strings.TrimSpace(authz[len("Bearer "):])
		if tokenString == "" {
			http.Error(w, "missing token", http.StatusUnauthorized)
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(os.Getenv("JWT_SECRET")), nil
		})
		if err != nil || !token.Valid {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "invalid token claims", http.StatusUnauthorized)
			return
		}

		// we stored user id in "sub"
		sub := claims["sub"]
		var userID uint64
		switch v := sub.(type) {
		case float64:
			userID = uint64(v)
		case string:
			if parsed, err := strconv.ParseUint(v, 10, 64); err == nil {
				userID = parsed
			}
		}
		if userID == 0 {
			http.Error(w, "invalid subject in token", http.StatusUnauthorized)
			return
		}

		var user models.User
		if err := config.DB.First(&user, userID).Error; err != nil {
			http.Error(w, "user not found", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), contextUserIDKey, user.ID)
		ctx = context.WithValue(ctx, contextUserKey, &user)
		next(w, r.WithContext(ctx))
	}
}

// RequireAdmin ensures the requester is authenticated and has admin privileges.
func RequireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		user, ok := GetAuthenticatedUser(r)
		if !ok || user == nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		if !user.IsAdmin {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		next(w, r)
	})
}
