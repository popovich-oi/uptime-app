package authcontroller

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"uptime-app/backend/internal/auth"
	"uptime-app/backend/internal/config"
	"uptime-app/backend/internal/models"
	"uptime-app/backend/internal/storage"
)

type Controller struct {
	auth *auth.Service
	cfg  config.Config
}

type credentials struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	User        any    `json:"user"`
}

type profileInput struct {
	Username string `json:"username"`
}

var emailPattern = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)

func NewController(service *auth.Service, cfg config.Config) *Controller {
	return &Controller{auth: service, cfg: cfg}
}

func (c *Controller) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /auth/register", c.register)
	mux.HandleFunc("POST /auth/login", c.login)
	mux.HandleFunc("POST /auth/refresh", c.refresh)
	mux.HandleFunc("POST /auth/logout", c.logout)
	mux.HandleFunc("GET /auth/profile", c.profile)
	mux.HandleFunc("PATCH /auth/profile", c.updateProfile)
	mux.HandleFunc("POST /auth/avatar", c.uploadAvatar)
	mux.HandleFunc("POST /auth/files", c.uploadFile)
	mux.HandleFunc("GET /files/{name}", c.downloadFile)
	mux.Handle("GET /avatars/", http.StripPrefix("/avatars/", http.FileServer(http.Dir(c.cfg.AvatarDir))))
}

func (c *Controller) register(w http.ResponseWriter, r *http.Request) {
	var input credentials
	if !decode(w, r, &input) {
		return
	}
	if !valid(input, true) {
		writeError(w, http.StatusBadRequest, "username, valid email and password of at least 8 characters are required")
		return
	}
	user, refresh, err := c.auth.Register(input.Username, input.Email, input.Password)
	if err != nil {
		writeError(w, http.StatusConflict, "username or email already exists")
		return
	}
	c.respond(w, user, refresh)
}

func (c *Controller) login(w http.ResponseWriter, r *http.Request) {
	var input credentials
	if !decode(w, r, &input) || input.Email == "" || input.Password == "" {
		writeError(w, http.StatusBadRequest, "email and password are required")
		return
	}
	user, refresh, err := c.auth.Login(input.Email, input.Password)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	c.respond(w, user, refresh)
}

func (c *Controller) refresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(c.cfg.RefreshCookieName)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid refresh token")
		return
	}
	user, refresh, err := c.auth.RotateRefresh(cookie.Value)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid refresh token")
		return
	}
	c.respond(w, user, refresh)
}

func (c *Controller) logout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(c.cfg.RefreshCookieName); err == nil {
		if err := c.auth.RevokeRefresh(cookie.Value); err != nil && !errors.Is(err, auth.ErrInvalidRefreshToken) {
			writeError(w, http.StatusInternalServerError, "could not revoke session")
			return
		}
	}
	http.SetCookie(w, &http.Cookie{
		Name: c.cfg.RefreshCookieName, Value: "", Path: "/", HttpOnly: true,
		Secure: c.cfg.CookieSecure, SameSite: parseSameSite(c.cfg.CookieSameSite), MaxAge: -1,
	})
	w.WriteHeader(http.StatusNoContent)
}

func (c *Controller) profile(w http.ResponseWriter, r *http.Request) {
	user, ok := c.currentUser(w, r)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"user": user})
}

func (c *Controller) updateProfile(w http.ResponseWriter, r *http.Request) {
	current, ok := c.currentUser(w, r)
	if !ok {
		return
	}
	var input profileInput
	if !decode(w, r, &input) {
		return
	}
	if len([]rune(input.Username)) < 3 || len([]rune(input.Username)) > 32 {
		writeError(w, http.StatusBadRequest, "username must be between 3 and 32 characters")
		return
	}
	user, err := c.auth.UpdateUsername(current.ID, input.Username)
	if err != nil {
		writeError(w, http.StatusConflict, "username already exists")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"user": user})
}

func (c *Controller) uploadAvatar(w http.ResponseWriter, r *http.Request) {
	current, ok := c.currentUser(w, r)
	if !ok {
		return
	}
	name, err := saveAvatar(c.cfg.AvatarDir, r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Выберите корректное изображение JPEG или PNG размером до 5 МБ")
		return
	}
	username := r.FormValue("username")
	user, err := c.auth.UpdateProfile(current.ID, username, "/avatars/"+name)
	if err != nil {
		_ = os.Remove(filepath.Join(c.cfg.AvatarDir, name))
		writeError(w, http.StatusBadRequest, "Не удалось сохранить профиль. Проверьте имя пользователя")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"user": user})
}

// uploadFile accepts any file without trusting its name or declared content type.
func (c *Controller) uploadFile(w http.ResponseWriter, r *http.Request) {
	if _, ok := c.currentUser(w, r); !ok {
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 26<<20)
	if err := r.ParseMultipartForm(1 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "invalid file or file exceeds 25 MB")
		return
	}
	defer r.MultipartForm.RemoveAll()
	file, _, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "file is required")
		return
	}
	defer file.Close()
	name, err := storage.Save(c.cfg.FileDir, file, 25<<20)
	if errors.Is(err, storage.ErrTooLarge) {
		writeError(w, http.StatusRequestEntityTooLarge, "file exceeds 25 MB")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not save file")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"url": "/files/" + name})
}

func (c *Controller) downloadFile(w http.ResponseWriter, r *http.Request) {
	path, ok := storage.Path(c.cfg.FileDir, r.PathValue("name"))
	if !ok {
		http.NotFound(w, r)
		return
	}
	file, err := os.Open(path)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	http.ServeContent(w, r, "download", info.ModTime(), file)
}

func (c *Controller) currentUser(w http.ResponseWriter, r *http.Request) (*models.User, bool) {
	value := r.Header.Get("Authorization")
	if len(value) < len("Bearer ") || value[:len("Bearer ")] != "Bearer " {
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

func (c *Controller) respond(w http.ResponseWriter, user *models.User, refresh string) {
	access, err := c.auth.AccessToken(user)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not issue token")
		return
	}
	http.SetCookie(w, &http.Cookie{Name: c.cfg.RefreshCookieName, Value: refresh, Path: "/", HttpOnly: true, Secure: c.cfg.CookieSecure, SameSite: parseSameSite(c.cfg.CookieSameSite), Expires: time.Now().Add(c.cfg.RefreshTTL)})
	writeJSON(w, http.StatusOK, tokenResponse{AccessToken: access, User: user})
}

func decode(w http.ResponseWriter, r *http.Request, value any) bool {
	if err := json.NewDecoder(r.Body).Decode(value); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return false
	}
	return true
}

func valid(value credentials, registration bool) bool {
	if registration && (len([]rune(value.Username)) < 3 || len([]rune(value.Username)) > 32) {
		return false
	}
	return value.Email != "" && emailPattern.MatchString(value.Email) && value.Password != "" && (!registration || len([]rune(value.Password)) >= 8)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func parseSameSite(value string) http.SameSite {
	switch value {
	case "Strict":
		return http.SameSiteStrictMode
	case "None":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteLaxMode
	}
}
