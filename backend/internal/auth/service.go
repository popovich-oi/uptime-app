package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"uptime-app/backend/internal/models"
)

var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrInvalidRefreshToken = errors.New("invalid refresh token")
var ErrInvalidAccessToken = errors.New("invalid access token")

type Service struct {
	db                    *gorm.DB
	jwtSecret             []byte
	accessTTL, refreshTTL time.Duration
}

func NewService(db *gorm.DB, secret string, accessTTL, refreshTTL time.Duration) *Service {
	return &Service{db: db, jwtSecret: []byte(secret), accessTTL: accessTTL, refreshTTL: refreshTTL}
}

func NormalizeUsername(value string) string { return strings.TrimSpace(value) }
func NormalizeEmail(value string) string    { return strings.ToLower(strings.TrimSpace(value)) }
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}
func CheckPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func (s *Service) Register(username, email, password string) (*models.User, string, error) {
	hash, err := HashPassword(password)
	if err != nil {
		return nil, "", err
	}
	user := &models.User{Username: strings.ToLower(NormalizeUsername(username)), Email: NormalizeEmail(email), PasswordHash: hash}
	if err = s.db.Create(user).Error; err != nil {
		return nil, "", err
	}
	refresh, err := s.issueRefresh(user.ID)
	if err != nil {
		return nil, "", err
	}
	return user, refresh, nil
}

func (s *Service) Login(email, password string) (*models.User, string, error) {
	var user models.User
	if err := s.db.Where("email = ?", NormalizeEmail(email)).First(&user).Error; err != nil || !CheckPassword(user.PasswordHash, password) {
		return nil, "", ErrInvalidCredentials
	}
	refresh, err := s.issueRefresh(user.ID)
	if err != nil {
		return nil, "", err
	}
	return &user, refresh, nil
}

func (s *Service) AccessToken(user *models.User) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{"sub": user.ID, "username": user.Username, "email": user.Email, "iat": now.Unix(), "exp": now.Add(s.accessTTL).Unix(), "jti": randomID()}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.jwtSecret)
}

func (s *Service) UserFromAccessToken(raw string) (*models.User, error) {
	token, err := jwt.Parse(raw, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, ErrInvalidAccessToken
		}
		return s.jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return nil, ErrInvalidAccessToken
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrInvalidAccessToken
	}
	sub, ok := claims["sub"].(float64)
	if !ok || sub <= 0 {
		return nil, ErrInvalidAccessToken
	}
	var user models.User
	if err := s.db.First(&user, uint(sub)).Error; err != nil {
		return nil, ErrInvalidAccessToken
	}
	return &user, nil
}

func (s *Service) UpdateUsername(userID uint, username string) (*models.User, error) {
	username = strings.ToLower(NormalizeUsername(username))
	if len([]rune(username)) < 3 || len([]rune(username)) > 32 {
		return nil, errors.New("invalid username")
	}
	if err := s.db.Model(&models.User{}).Where("id = ?", userID).Update("username", username).Error; err != nil {
		return nil, err
	}
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *Service) UpdateProfile(userID uint, username, avatarURL string) (*models.User, error) {
	username = strings.ToLower(NormalizeUsername(username))
	if len([]rune(username)) < 3 || len([]rune(username)) > 32 {
		return nil, errors.New("invalid username")
	}
	if err := s.db.Model(&models.User{}).Where("id = ?", userID).Updates(map[string]any{"username": username, "avatar_url": avatarURL}).Error; err != nil {
		return nil, err
	}
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *Service) RotateRefresh(raw string) (*models.User, string, error) {
	hash := tokenHash(raw)
	var token models.RefreshToken
	if err := s.db.Preload("User").Where("token_hash = ?", hash).First(&token).Error; err != nil || token.RevokedAt != nil || time.Now().After(token.ExpiresAt) {
		return nil, "", ErrInvalidRefreshToken
	}
	var next string
	err := s.db.Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		if err := tx.Model(&token).Update("revoked_at", now).Error; err != nil {
			return err
		}
		next = randomToken()
		return tx.Create(&models.RefreshToken{UserID: token.UserID, TokenHash: tokenHash(next), ExpiresAt: now.Add(s.refreshTTL)}).Error
	})
	return &token.User, next, err
}

func (s *Service) RevokeRefresh(raw string) error {
	hash := tokenHash(raw)
	result := s.db.Model(&models.RefreshToken{}).
		Where("token_hash = ? AND revoked_at IS NULL", hash).
		Update("revoked_at", time.Now())
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrInvalidRefreshToken
	}
	return nil
}

func (s *Service) issueRefresh(userID uint) (string, error) {
	raw := randomToken()
	return raw, s.db.Create(&models.RefreshToken{UserID: userID, TokenHash: tokenHash(raw), ExpiresAt: time.Now().Add(s.refreshTTL)}).Error
}
func randomToken() string {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(bytes)
}
func randomID() string              { return randomToken() }
func tokenHash(value string) string { return fmtHash(value) }
func fmtHash(value string) string {
	sum := sha256.Sum256([]byte(value))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}
