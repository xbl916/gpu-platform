package middleware

import (
	"net/http"
	"strings"

	"gpu-platform/internal/auth"
	"gpu-platform/internal/config"
	"gpu-platform/internal/models"
	"gpu-platform/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AuthMiddleware struct {
	jwtSvc   *auth.JWTService
	userRepo *repository.UserRepository
	cfg      *config.Config
}

func NewAuthMiddleware(jwtSvc *auth.JWTService, userRepo *repository.UserRepository, cfg *config.Config) *AuthMiddleware {
	return &AuthMiddleware{
		jwtSvc:   jwtSvc,
		userRepo: userRepo,
		cfg:      cfg,
	}
}

func (m *AuthMiddleware) AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "authorization header required",
			})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid authorization header format",
			})
			return
		}

		claims, err := m.jwtSvc.ValidateToken(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(),
			})
			return
		}

		userID, err := uuid.Parse(claims.UserID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid user ID in token",
			})
			return
		}

		user, err := m.userRepo.FindByID(c.Request.Context(), userID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "user not found",
			})
			return
		}

		if user.Status == string(models.UserStatusLocked) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "user account is locked",
			})
			return
		}

		if user.Status == string(models.UserStatusInactive) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "user account is inactive",
			})
			return
		}

		c.Set("userID", userID)
		c.Set("user", user)
		c.Set("userRole", user.Role)
		c.Set("claims", claims)

		c.Next()
	}
}

func (m *AuthMiddleware) AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("userRole")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "authentication required",
			})
			return
		}

		if role != string(models.RoleAdmin) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "admin access required",
			})
			return
		}

		c.Next()
	}
}

func (m *AuthMiddleware) RoleRequired(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("userRole")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "authentication required",
			})
			return
		}

		for _, allowed := range allowedRoles {
			if role == allowed {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error": "insufficient permissions",
		})
	}
}

func GetUserID(c *gin.Context) string {
	if userID, exists := c.Get("userID"); exists {
		return userID.(string)
	}
	return ""
}

func GetUserRole(c *gin.Context) string {
	if role, exists := c.Get("userRole"); exists {
		return role.(string)
	}
	return ""
}

func ParseUUID(s string) (string, error) {
	if len(s) != 36 {
		return "", &ValidationError{Message: "invalid UUID format"}
	}
	return s, nil
}

type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

type RateLimitMiddleware struct {
	cfg *config.RateLimitConfig
}

func NewRateLimitMiddleware(cfg *config.RateLimitConfig) *RateLimitMiddleware {
	return &RateLimitMiddleware{cfg: cfg}
}

func (m *RateLimitMiddleware) RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !m.cfg.Enabled {
			c.Next()
			return
		}

		c.Next()
	}
}

type CORSMiddleware struct{}

func NewCORSMiddleware() *CORSMiddleware {
	return &CORSMiddleware{}
}

func (m *CORSMiddleware) CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

type LoggingMiddleware struct{}

func NewLoggingMiddleware() *LoggingMiddleware {
	return &LoggingMiddleware{}
}

func (m *LoggingMiddleware) Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}

type RecoveryMiddleware struct {
	recovery gin.RecoveryFunc
}

func NewRecoveryMiddleware() *RecoveryMiddleware {
	return &RecoveryMiddleware{}
}

func (m *RecoveryMiddleware) Recovery() gin.HandlerFunc {
	return gin.Recovery()
}
