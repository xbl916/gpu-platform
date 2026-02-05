package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gpu-platform/internal/auth"
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
		log.Fatalf("Failed to initialize config: %v", err)
	}

	db, err := repository.InitDatabase(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer repository.CloseDatabase()

	userRepo := repository.NewUserRepository(db)
	gpuServerRepo := repository.NewGPUServerRepository(db)
	gpuPoolRepo := repository.NewGPUPoolRepository(db)
	containerRepo := repository.NewContainerRepository(db)
	templateRepo := repository.NewTemplateRepository(db)

	userSvc := services.NewUserService(userRepo, cfg)
	resourceSvc := services.NewResourceService(cfg, gpuPoolRepo, gpuServerRepo, containerRepo)
	containerSvc := services.NewContainerService(cfg, containerRepo, resourceSvc, nil)
	templateSvc := services.NewTemplateService(templateRepo)

	jwtSvc := auth.NewJWTService(&cfg.JWT)
	authMiddleware := middleware.NewAuthMiddleware(jwtSvc, userRepo, cfg)

	handler := handlers.NewHandler(cfg, userSvc, resourceSvc, containerSvc, templateSvc, authMiddleware)

	if cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.Logger())
	r.Use(middleware.NewCORSMiddleware().CORS())

	handler.SetupRoutes(r)

	addr := fmt.Sprintf("%s:%d", cfg.App.Host, cfg.App.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: r,
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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
