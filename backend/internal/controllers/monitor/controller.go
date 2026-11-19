package monitorcontroller

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"unicode/utf8"

	"gorm.io/gorm"
	"uptime-app/backend/internal/auth"
	"uptime-app/backend/internal/httpapi"
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
	user, ok := httpapi.CurrentUser(c.auth, w, r)
	if !ok {
		return
	}
	var monitors []models.Monitor
	if err := c.db.Where("user_id = ?", user.ID).Order("created_at DESC").Find(&monitors).Error; err != nil {
		httpapi.WriteError(w, http.StatusInternalServerError, "could not load monitors")
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, map[string]any{"monitors": monitors})
}

func (c *Controller) create(w http.ResponseWriter, r *http.Request) {
	user, ok := httpapi.CurrentUser(c.auth, w, r)
	if !ok {
		return
	}
	var value input
	decodeErr := decodeInput(w, r, &value)
	var tooLarge *http.MaxBytesError
	if errors.As(decodeErr, &tooLarge) {
		httpapi.WriteError(w, http.StatusRequestEntityTooLarge, "request body too large")
		return
	}
	seconds, intervalErr := intervalSeconds(value.IntervalValue, value.IntervalUnit)
	if decodeErr != nil || !validURL(value.URL) || intervalErr != nil {
		httpapi.WriteError(w, http.StatusBadRequest, "valid URL and interval are required")
		return
	}
	monitor := models.Monitor{UserID: user.ID, URL: value.URL, IntervalSeconds: seconds}
	if err := c.db.Create(&monitor).Error; err != nil {
		httpapi.WriteError(w, http.StatusInternalServerError, "could not create monitor")
		return
	}
	httpapi.WriteJSON(w, http.StatusCreated, map[string]any{"monitor": monitor})
}

func decodeInput(w http.ResponseWriter, r *http.Request, value *input) error {
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
	decoder := json.NewDecoder(r.Body)
	decodeErr := decoder.Decode(value)
	if decodeErr == nil {
		var extra any
		if err := decoder.Decode(&extra); err != io.EOF {
			if err == nil {
				decodeErr = errors.New("multiple JSON values")
			} else {
				decodeErr = err
			}
		}
	}
	return decodeErr
}

func validURL(raw string) bool {
	parsed, err := url.ParseRequestURI(raw)
	return utf8.RuneCountInString(raw) <= 2048 && err == nil && (parsed.Scheme == "http" || parsed.Scheme == "https") && parsed.Host != ""
}
func intervalSeconds(value int, unit string) (int, error) {
	if value < 1 {
		return 0, fmt.Errorf("interval must be positive")
	}
	multiplier := map[string]int{"seconds": 1, "minutes": 60, "hours": 3600}[unit]
	if multiplier == 0 || value > 9007199254740991/multiplier {
		return 0, fmt.Errorf("invalid interval")
	}
	return value * multiplier, nil
}
