package services

import (
	"context"
	"errors"
	"time"

	"gpu-platform/internal/auth"
	"gpu-platform/internal/config"
	"gpu-platform/internal/models"
	"gpu-platform/internal/repository"

	"github.com/google/uuid"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrInvalidEmail      = errors.New("invalid email format")
	ErrWeakPassword      = errors.New("password does not meet requirements")
	ErrUserLocked        = errors.New("user account is locked")
	ErrUserInactive      = errors.New("user account is inactive")
	ErrWrongPassword     = errors.New("wrong password")
)

type UserService struct {
	userRepo    *repository.UserRepository
	passwordSvc *auth.PasswordService
	jwtSvc      *auth.JWTService
	sessionSvc  *auth.SessionService
	cfg         *config.Config
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=12"`
	Name     string `json:"name" binding:"required,min=2,max=100"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	User         *UserResponse `json:"user"`
	AccessToken  string        `json:"accessToken"`
	RefreshToken string        `json:"refreshToken"`
	ExpiresIn    int           `json:"expiresIn"`
}

type UserResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	Role      string `json:"role"`
	Status    string `json:"status"`
	CreatedAt string `json:"createdAt"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"oldPassword" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required,min=12"`
}

func NewUserService(userRepo *repository.UserRepository, cfg *config.Config) *UserService {
	return &UserService{
		userRepo:    userRepo,
		passwordSvc: auth.NewPasswordService(),
		jwtSvc:      auth.NewJWTService(&cfg.JWT),
		sessionSvc:  auth.NewSessionService(&cfg.Session),
		cfg:         cfg,
	}
}

func (s *UserService) Register(ctx context.Context, req *RegisterRequest) (*LoginResponse, error) {
	existing, _ := s.userRepo.FindByEmail(ctx, req.Email)
	if existing != nil {
		return nil, ErrUserAlreadyExists
	}

	if err := s.validateEmail(req.Email); err != nil {
		return nil, err
	}

	if err := s.passwordSvc.ValidatePassword(req.Password); err != nil {
		return nil, ErrWeakPassword
	}

	passwordHash, err := s.passwordSvc.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Email:        req.Email,
		PasswordHash: passwordHash,
		Name:         req.Name,
		Role:         string(models.RoleDeveloper),
		Status:       string(models.UserStatusActive),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return s.generateLoginResponse(user)
}

func (s *UserService) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	if user.Status == string(models.UserStatusLocked) {
		return nil, ErrUserLocked
	}

	if user.Status == string(models.UserStatusInactive) {
		return nil, ErrUserInactive
	}

	if !s.passwordSvc.CheckPassword(req.Password, user.PasswordHash) {
		return nil, ErrWrongPassword
	}

	s.userRepo.IncrementLoginCount(ctx, user.ID)

	return s.generateLoginResponse(user)
}

func (s *UserService) GetUserByID(ctx context.Context, id uuid.UUID) (*UserResponse, error) {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.toUserResponse(user), nil
}

func (s *UserService) GetCurrentUser(ctx context.Context, userID uuid.UUID) (*UserResponse, error) {
	return s.GetUserByID(ctx, userID)
}

func (s *UserService) UpdateProfile(ctx context.Context, userID uuid.UUID, name string) (*UserResponse, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	user.Name = name
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return s.toUserResponse(user), nil
}

func (s *UserService) ChangePassword(ctx context.Context, userID uuid.UUID, req *ChangePasswordRequest) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}

	if !s.passwordSvc.CheckPassword(req.OldPassword, user.PasswordHash) {
		return ErrWrongPassword
	}

	if err := s.passwordSvc.ValidatePassword(req.NewPassword); err != nil {
		return ErrWeakPassword
	}

	newHash, err := s.passwordSvc.HashPassword(req.NewPassword)
	if err != nil {
		return err
	}

	return s.userRepo.UpdatePassword(ctx, userID, newHash)
}

func (s *UserService) ListUsers(ctx context.Context, limit, offset int) ([]*UserResponse, int64, error) {
	users, total, err := s.userRepo.List(ctx, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]*UserResponse, len(users))
	for i, user := range users {
		responses[i] = s.toUserResponse(&user)
	}

	return responses, total, nil
}

func (s *UserService) LockUser(ctx context.Context, id uuid.UUID) error {
	return s.userRepo.UpdateStatus(ctx, id, string(models.UserStatusLocked))
}

func (s *UserService) UnlockUser(ctx context.Context, id uuid.UUID) error {
	return s.userRepo.UpdateStatus(ctx, id, string(models.UserStatusActive))
}

func (s *UserService) DeleteUser(ctx context.Context, id uuid.UUID) error {
	return s.userRepo.Delete(ctx, id)
}

func (s *UserService) generateLoginResponse(user *models.User) (*LoginResponse, error) {
	accessToken, err := s.jwtSvc.GenerateAccessToken(user)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.jwtSvc.GenerateRefreshToken(user)
	if err != nil {
		return nil, err
	}

	return &LoginResponse{
		User:         s.toUserResponse(user),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    s.cfg.JWT.AccessTokenExpire,
	}, nil
}

func (s *UserService) toUserResponse(user *models.User) *UserResponse {
	return &UserResponse{
		ID:        user.ID.String(),
		Email:     user.Email,
		Name:      user.Name,
		Role:      user.Role,
		Status:    user.Status,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
	}
}

func (s *UserService) validateEmail(email string) error {
	if len(email) < 5 || len(email) > 255 {
		return ErrInvalidEmail
	}
	return nil
}

func (s *UserService) RefreshToken(ctx context.Context, refreshToken string) (string, string, error) {
	accessToken, newRefreshToken, err := s.jwtSvc.RefreshTokenPair(refreshToken)
	if err != nil {
		return "", "", err
	}
	return accessToken, newRefreshToken, nil
}
