package monitorcontroller

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/url"

	"gorm.io/gorm"
	"uptime-app/backend/internal/auth"
	"uptime-app/backend/internal/models"
)

type Controller struct {
	auth *auth.Service
	db   *gorm.DB
}
type input struct {
	URL           string `json:"url"`
	IntervalValue int    `json:"interval_value"`
	IntervalUnit  string `json:"interval_unit"`
}

func NewController(service *auth.Service, db *gorm.DB) *Controller {
	return &Controller{auth: service, db: db}
}

func (c *Controller) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /monitors", c.list)
	mux.HandleFunc("POST /monitors", c.create)
}

func (c *Controller) list(w http.ResponseWriter, r *http.Request) {
	user, ok := c.currentUser(w, r)
	if !ok {
		return
	}
	var monitors []models.Monitor
	if err := c.db.Where("user_id = ?", user.ID).Order("created_at DESC").Find(&monitors).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "could not load monitors")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"monitors": monitors})
}

func (c *Controller) create(w http.ResponseWriter, r *http.Request) {
	user, ok := c.currentUser(w, r)
	if !ok {
		return
	}
	var value input
	decodeErr := json.NewDecoder(r.Body).Decode(&value)
	seconds, intervalErr := intervalSeconds(value.IntervalValue, value.IntervalUnit)
	if decodeErr != nil || !validURL(value.URL) || intervalErr != nil {
		writeError(w, http.StatusBadRequest, "valid URL and interval are required")
		return
	}
	monitor := models.Monitor{UserID: user.ID, URL: value.URL, IntervalSeconds: seconds}
	if err := c.db.Create(&monitor).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "could not create monitor")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"monitor": monitor})
}

func (c *Controller) currentUser(w http.ResponseWriter, r *http.Request) (*models.User, bool) {
	value := r.Header.Get("Authorization")
	if len(value) <= len("Bearer ") || value[:len("Bearer ")] != "Bearer " {
		writeError(w, http.StatusUnauthorized, "authorization is required")
		return nil, false
	}
	user, err := c.auth.UserFromAccessToken(value[len("Bearer "):])
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid access token")
		return nil, false
	}
	return user, true
}

func validURL(raw string) bool {
	parsed, err := url.ParseRequestURI(raw)
	return err == nil && (parsed.Scheme == "http" || parsed.Scheme == "https") && parsed.Host != ""
}
func intervalSeconds(value int, unit string) (int, error) {
	if value < 1 {
		return 0, fmt.Errorf("interval must be positive")
	}
	multiplier := map[string]int{"seconds": 1, "minutes": 60, "hours": 3600}[unit]
	if multiplier == 0 || value > math.MaxInt/multiplier {
		return 0, fmt.Errorf("invalid interval")
	}
	return value * multiplier, nil
}
func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
