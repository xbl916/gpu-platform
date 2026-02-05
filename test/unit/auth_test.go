package unit

import (
	"testing"

	"gpu-platform/internal/auth"
	"gpu-platform/internal/config"
	"gpu-platform/internal/models"

	"github.com/google/uuid"
)

func TestPasswordService(t *testing.T) {
	svc := auth.NewPasswordService()

	password := "TestPassword123!"
	hash, err := svc.HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	if hash == password {
		t.Error("Hash should not equal original password")
	}

	if !svc.CheckPassword(password, hash) {
		t.Error("CheckPassword should return true for correct password")
	}

	if svc.CheckPassword("WrongPassword123!", hash) {
		t.Error("CheckPassword should return false for wrong password")
	}
}

func TestPasswordService_Validation(t *testing.T) {
	svc := auth.NewPasswordService()

	err := svc.ValidatePassword("short")
	if err == nil {
		t.Error("Expected error for short password")
	}

	err = svc.ValidatePassword("validpassword123")
	if err != nil {
		t.Errorf("Expected no error for valid password, got: %v", err)
	}
}

func TestJWTService(t *testing.T) {
	cfg := &config.JWTConfig{
		Secret:             "test-secret-key",
		AccessTokenExpire:  3600,
		RefreshTokenExpire: 604800,
	}

	jwtSvc := auth.NewJWTService(cfg)

	user := &models.User{
		ID:    uuid.New(),
		Email: "test@example.com",
		Name:  "Test User",
		Role:  string(models.RoleDeveloper),
	}

	accessToken, err := jwtSvc.GenerateAccessToken(user)
	if err != nil {
		t.Fatalf("Failed to generate access token: %v", err)
	}

	if accessToken == "" {
		t.Error("Access token should not be empty")
	}

	claims, err := jwtSvc.ValidateToken(accessToken)
	if err != nil {
		t.Fatalf("Failed to validate token: %v", err)
	}

	if claims.Email != user.Email {
		t.Errorf("Expected email %s, got %s", user.Email, claims.Email)
	}

	if claims.UserID != user.ID.String() {
		t.Errorf("Expected user ID %s, got %s", user.ID, claims.UserID)
	}
}

func TestJWTService_InvalidToken(t *testing.T) {
	cfg := &config.JWTConfig{
		Secret: "test-secret-key",
	}

	jwtSvc := auth.NewJWTService(cfg)

	_, err := jwtSvc.ValidateToken("invalid-token")
	if err == nil {
		t.Error("Expected error for invalid token")
	}
}

func TestSessionService(t *testing.T) {
	cfg := &config.SessionConfig{
		MaxSessionAge: 86400,
	}

	svc := auth.NewSessionService(cfg)

	userID := uuid.New().String()
	session := svc.CreateSession(userID)

	if session.UserID != userID {
		t.Errorf("Expected user ID %s, got %s", userID, session.UserID)
	}

	if !session.IsValid() {
		t.Error("Session should be valid immediately after creation")
	}

	if session.TTL() <= 0 {
		t.Error("Session TTL should be positive")
	}
}
