package repository

import (
	"fmt"
	"time"

	"gpu-platform/internal/config"
	"gpu-platform/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB

func InitDatabase(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	var err error

	gormLogger := logger.Default.LogMode(logger.Silent)
	if cfg.Name == "gpu_platform" {
		gormLogger = logger.Default.LogMode(logger.Info)
	}

	db, err = gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	if err := migrate(db); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return db, nil
}

func migrate(database *gorm.DB) error {
	return database.AutoMigrate(
		&models.User{},
		&models.Project{},
		&models.ProjectMember{},
		&models.GPUPool{},
		&models.GPUServer{},
		&models.GPUDevice{},
		&models.ContainerTemplate{},
		&models.ContainerInstance{},
		&models.ContainerMetrics{},
	)
}

func GetDB() *gorm.DB {
	return db
}

func CloseDatabase() error {
	if db != nil {
		sqlDB, err := db.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}
