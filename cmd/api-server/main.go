package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"gpu-platform/internal/builder"
	"gpu-platform/internal/config"
	"gpu-platform/internal/handlers"
	"gpu-platform/internal/middleware"
	"gpu-platform/internal/repository"
	"gpu-platform/internal/services"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.Init("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := repository.InitDatabase(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer repository.CloseDatabase()

	if !cfg.IsDevelopment() {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()
	router.Use(gin.Recovery())

	userRepo := repository.NewUserRepository(db)
	gpuPoolRepo := repository.NewGPUPoolRepository(db)
	gpuServerRepo := repository.NewGPUServerRepository(db)
	containerRepo := repository.NewContainerRepository(db)
	templateRepo := repository.NewTemplateRepository(db)

	jwtSvc := middleware.NewJWTService(cfg.JWT.Secret, cfg.JWT.AccessTokenExpire, cfg.JWT.RefreshTokenExpire)
	authMiddleware := middleware.NewAuthMiddleware(jwtSvc, userRepo)

	userSvc := services.NewUserService(userRepo, cfg)
	resourceSvc := services.NewResourceService(cfg, gpuPoolRepo, gpuServerRepo, containerRepo)
	containerSvc := services.NewContainerService(cfg, containerRepo, resourceSvc, nil)
	templateBuilder := builder.NewTemplateBuilder()
	templateSvc := services.NewTemplateService(templateRepo, templateBuilder)

	handler := handlers.NewHandler(
		cfg,
		userSvc,
		resourceSvc,
		containerSvc,
		templateSvc,
		authMiddleware,
	)

	handler.SetupRoutes(router)

	addr := fmt.Sprintf("%s:%d", cfg.App.Host, cfg.App.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	go func() {
		log.Printf("Starting API server on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	log.Println("Server exited properly")
}
