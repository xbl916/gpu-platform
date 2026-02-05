package auth

import (
	"errors"
	"fmt"
	"time"

	"gpu-platform/internal/config"
	"gpu-platform/internal/models"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidToken  = errors.New("invalid token")
	ErrExpiredToken  = errors.New("token has expired")
	ErrInvalidClaims = errors.New("invalid token claims")
	ErrWrongPassword = errors.New("wrong password")
	ErrUserLocked    = errors.New("user account is locked")
	ErrUserInactive  = errors.New("user account is inactive")
)

type JWTService struct {
	config *config.JWTConfig
}

type TokenClaims struct {
	UserID    string `json:"userId"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	ProjectID string `json:"projectId,omitempty"`
	jwt.RegisteredClaims
}

func NewJWTService(cfg *config.JWTConfig) *JWTService {
	return &JWTService{config: cfg}
}

func (s *JWTService) GenerateAccessToken(user *models.User) (string, error) {
	claims := TokenClaims{
		UserID: user.ID.String(),
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(s.config.AccessTokenExpire) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "gpu-platform",
			Subject:   user.ID.String(),
		},
	}

	if user.ProjectID != nil {
		claims.ProjectID = user.ProjectID.String()
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.config.Secret))
}

func (s *JWTService) GenerateRefreshToken(user *models.User) (string, error) {
	claims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(s.config.RefreshTokenExpire) * time.Second)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		NotBefore: jwt.NewNumericDate(time.Now()),
		Issuer:    "gpu-platform",
		Subject:   user.ID.String(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.config.Secret))
}

func (s *JWTService) ValidateToken(tokenString string) (*TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.config.Secret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*TokenClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidClaims
	}

	return claims, nil
}

func (s *JWTService) RefreshTokenPair(refreshToken string) (accessToken, newRefreshToken string, err error) {
	claims, err := s.ValidateToken(refreshToken)
	if err != nil {
		return "", "", err
	}

	user := &models.User{
		ID:    uuid.MustParse(claims.UserID),
		Email: claims.Email,
		Role:  claims.Role,
	}

	if claims.ProjectID != "" {
		projectID := uuid.MustParse(claims.ProjectID)
		user.ProjectID = &projectID
	}

	accessToken, err = s.GenerateAccessToken(user)
	if err != nil {
		return "", "", err
	}

	newRefreshToken, err = s.GenerateRefreshToken(user)
	if err != nil {
		return "", "", err
	}

	return accessToken, newRefreshToken, nil
}

type PasswordService struct {
	cost int
}

func NewPasswordService() *PasswordService {
	return &PasswordService{cost: bcrypt.DefaultCost}
}

func (s *PasswordService) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), s.cost)
	return string(bytes), err
}

func (s *PasswordService) CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func (s *PasswordService) ValidatePassword(password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters")
	}
	return nil
}

type SessionService struct {
	config *config.SessionConfig
}

func NewSessionService(cfg *config.SessionConfig) *SessionService {
	return &SessionService{config: cfg}
}

func (s *SessionService) CreateSession(userID string) *Session {
	return &Session{
		UserID:   userID,
		ExpireAt: time.Now().Add(s.config.MaxSessionDuration()),
	}
}

type Session struct {
	UserID   string
	ExpireAt time.Time
}

func (s *Session) IsValid() bool {
	return time.Now().Before(s.ExpireAt)
}

func (s *Session) TTL() time.Duration {
	return time.Until(s.ExpireAt)
}
