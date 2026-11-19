package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"

	"uptime-app/backend/internal/auth"
	"uptime-app/backend/internal/models"
)

func CurrentUser(service *auth.Service, w http.ResponseWriter, r *http.Request) (*models.User, bool) {
	value := r.Header.Get("Authorization")
	if !strings.HasPrefix(value, "Bearer ") || len(value) == len("Bearer ") {
		WriteError(w, http.StatusUnauthorized, "authorization is required")
		return nil, false
	}
	user, err := service.UserFromAccessToken(strings.TrimPrefix(value, "Bearer "))
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "invalid access token")
		return nil, false
	}
	return user, true
}

func WriteJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func WriteError(w http.ResponseWriter, status int, message string) {
	WriteJSON(w, status, map[string]string{"error": message})
}
