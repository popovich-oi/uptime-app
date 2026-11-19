package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"

	"uptime-app/backend/internal/auth"
	"uptime-app/backend/internal/models"
)

// CORS разрешает браузерные запросы только с указанного origin (схема, хост и порт).
// Это не авторизация: серверные клиенты по-прежнему должны проходить проверку доступа.
func CORS(next http.Handler, allowedOrigin string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Ответ зависит от Origin, поэтому общие кэши должны учитывать этот заголовок.
		w.Header().Add("Vary", "Origin")
		origin := r.Header.Get("Origin")
		if origin != "" && origin == allowedOrigin && allowedOrigin != "*" {
			// Для запросов с cookie браузер требует точный origin вместо «*».
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		// Браузер предварительно запрашивает разрешения для нестандартных методов и заголовков.
		if r.Method == http.MethodOptions && origin != "" && r.Header.Get("Access-Control-Request-Method") != "" {
			w.Header().Add("Vary", "Access-Control-Request-Method")
			w.Header().Add("Vary", "Access-Control-Request-Headers")
			if origin != allowedOrigin || allowedOrigin == "*" {
				http.Error(w, "origin is not allowed", http.StatusForbidden)
				return
			}
			w.Header().Set("Access-Control-Allow-Methods", "GET, HEAD, POST, PATCH")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

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
