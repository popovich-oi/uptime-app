package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	AvatarDir         string
	FileDir           string
	DatabaseURL       string
	JWTSecret         string
	HTTPAddr          string
	CORSAllowedOrigin string
	AccessTTL         time.Duration
	RefreshTTL        time.Duration
	RefreshCookieName string
	CookieSecure      bool
	CookieSameSite    string
}

func Load() Config {
	return Config{
		FileDir:           getenv("UPLOAD_DIR", "uploads/files"),
		AvatarDir:         getenv("AVATAR_DIR", "uploads/avatars"),
		DatabaseURL:       getenv("DATABASE_URL", "host=localhost user=uptime password=uptime dbname=uptime port=5432 sslmode=disable"),
		JWTSecret:         getenv("JWT_SECRET", "local-development-secret-change-me"),
		HTTPAddr:          getenv("HTTP_ADDR", ":8080"),
		CORSAllowedOrigin: getenv("CORS_ALLOWED_ORIGIN", "http://localhost:3000"),
		AccessTTL:         getDuration("ACCESS_TTL", 15*time.Minute),
		RefreshTTL:        getDuration("REFRESH_TTL", 30*24*time.Hour),
		RefreshCookieName: getenv("REFRESH_COOKIE_NAME", "refresh_token"),
		CookieSecure:      getBool("COOKIE_SECURE", false),
		CookieSameSite:    getenv("COOKIE_SAMESITE", "Lax"),
	}
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
func getDuration(key string, fallback time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if parsed, err := time.ParseDuration(value); err == nil {
			return parsed
		}
	}
	return fallback
}
func getBool(key string, fallback bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return fallback
}
