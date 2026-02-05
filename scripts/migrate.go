package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"gpu-platform/internal/config"
	"gpu-platform/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	configFile = flag.String("config", "config.yaml", "Path to config file")
	action     = flag.String("action", "up", "Migration action: up, down, status, seed")
)

func main() {
	flag.Parse()

	cfg, err := config.Init(*configFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := initDB(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	switch *action {
	case "up":
		runMigrations(db, "up")
	case "down":
		runMigrations(db, "down")
	case "status":
		printMigrationStatus(db)
	case "seed":
		runSeed(db)
	default:
		fmt.Printf("Unknown action: %s\n", *action)
		os.Exit(1)
	}
}

func initDB(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Username,
		cfg.Database.Password,
		cfg.Database.Name,
		cfg.Database.SSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	return db, nil
}

type Migration struct {
	ID        uint   `gorm:"primaryKey"`
	Version   int64  `gorm:"uniqueIndex"`
	Name      string `gorm:"size:255"`
	AppliedAt time.Time
	Script    string `gorm:"type:text"`
}

var migrations = []struct {
	Version int64
	Name    string
	Script  string
}{
	{
		Version: 1,
		Name:    "initial_schema",
		Script: `
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TYPE user_role AS ENUM ('admin', 'developer', 'guest');
CREATE TYPE user_status AS ENUM ('active', 'inactive', 'locked');
CREATE TYPE server_status AS ENUM ('online', 'offline', 'maintenance');
CREATE TYPE gpu_status AS ENUM ('available', 'used', 'reserved', 'faulty');
CREATE TYPE instance_status AS ENUM ('pending', 'creating', 'running', 'stopped', 'error', 'deleted');
CREATE TYPE template_status AS ENUM ('draft', 'building', 'ready', 'failed', 'deprecated');

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(100) NOT NULL,
    role user_role DEFAULT 'developer',
    status user_status DEFAULT 'active',
    project_id UUID,
    last_login_at TIMESTAMP,
    login_count INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_project ON users(project_id);

CREATE TABLE projects (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    description VARCHAR(1000),
    quota_config JSONB DEFAULT '{}',
    owner_id UUID,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_projects_owner ON projects(owner_id);

CREATE TABLE gpu_pools (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    description VARCHAR(500),
    scheduling_policy VARCHAR(50) DEFAULT 'binpack',
    quota_config JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE gpu_servers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    hostname VARCHAR(255) UNIQUE NOT NULL,
    ip_address VARCHAR(45) NOT NULL,
    status server_status DEFAULT 'offline',
    specs JSONB DEFAULT '{}',
    gpu_pool_id UUID,
    last_heartbeat TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_servers_pool ON gpu_servers(gpu_pool_id);
CREATE INDEX idx_servers_status ON gpu_servers(status);

CREATE TABLE gpu_devices (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    server_id UUID REFERENCES gpu_servers(id) ON DELETE CASCADE,
    gpu_id VARCHAR(50) NOT NULL,
    name VARCHAR(100),
    model VARCHAR(50),
    memory_mb INTEGER,
    compute_cap INTEGER,
    cuda_version VARCHAR(20),
    temperature INTEGER DEFAULT 0,
    power_usage INTEGER DEFAULT 0,
    utilization INTEGER DEFAULT 0,
    status gpu_status DEFAULT 'available',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_devices_server ON gpu_devices(server_id);
CREATE INDEX idx_devices_status ON gpu_devices(status);

CREATE TABLE container_templates (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    description VARCHAR(1000),
    user_id UUID REFERENCES users(id),
    project_id UUID REFERENCES projects(id),
    config JSONB DEFAULT '{}',
    docker_image VARCHAR(500),
    dockerfile TEXT,
    version INTEGER DEFAULT 1,
    status template_status DEFAULT 'draft',
    is_public BOOLEAN DEFAULT FALSE,
    usage_count INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_templates_user ON container_templates(user_id);
CREATE INDEX idx_templates_project ON container_templates(project_id);
CREATE INDEX idx_templates_name ON container_templates(name);

CREATE TABLE container_instances (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    user_id UUID REFERENCES users(id),
    project_id UUID REFERENCES projects(id),
    template_id UUID REFERENCES container_templates(id),
    gpu_server_id UUID REFERENCES gpu_servers(id),
    status instance_status DEFAULT 'pending',
    resources JSONB DEFAULT '{}',
    access_token VARCHAR(255),
    ssh_port INTEGER DEFAULT 22,
    web_port INTEGER DEFAULT 8888,
    container_id VARCHAR(100),
    pod_name VARCHAR(255),
    ip_address VARCHAR(45),
    started_at TIMESTAMP,
    stopped_at TIMESTAMP,
    expires_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_instances_user ON container_instances(user_id);
CREATE INDEX idx_instances_project ON container_instances(project_id);
CREATE INDEX idx_instances_server ON container_instances(gpu_server_id);
CREATE INDEX idx_instances_status ON container_instances(status);

CREATE TABLE project_members (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(50) DEFAULT 'member',
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX idx_project_members_unique ON project_members(project_id, user_id);

CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id),
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(50),
    resource_id VARCHAR(100),
    details JSONB DEFAULT '{}',
    ip_address VARCHAR(45),
    user_agent TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_audit_user ON audit_logs(user_id);
CREATE INDEX idx_audit_action ON audit_logs(action);
CREATE INDEX idx_audit_created ON audit_logs(created_at);

CREATE TABLE notifications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL,
    title VARCHAR(255) NOT NULL,
    message TEXT,
    is_read BOOLEAN DEFAULT FALSE,
    read_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_notifications_user ON notifications(user_id);
CREATE INDEX idx_notifications_unread ON notifications(user_id, is_read) WHERE is_read = FALSE;
`,
	},
}

func runMigrations(db *gorm.DB, direction string) {
	log.Printf("Running migrations %s...\n", direction)

	if direction == "up" {
		for _, m := range migrations {
			var count int64
			db.Table("migrations").Where("version = ?", m.Version).Count(&count)
			if count > 0 {
				log.Printf("Migration %d already applied, skipping\n", m.Version)
				continue
			}

			log.Printf("Applying migration %d: %s\n", m.Version, m.Name)

			if err := db.Exec(m.Script).Error; err != nil {
				log.Fatalf("Failed to apply migration %d: %v\n", m.Version, err)
			}

			migration := Migration{
				Version:   m.Version,
				Name:      m.Name,
				AppliedAt: time.Now(),
				Script:    m.Script,
			}
			if err := db.Table("migrations").Create(&migration).Error; err != nil {
				log.Printf("Warning: failed to record migration: %v\n", err)
			}
		}
	}

	log.Println("Migrations completed successfully")
}

func printMigrationStatus(db *gorm.DB) {
	log.Println("Migration Status:")
	log.Println("===============")

	for _, m := range migrations {
		log.Printf("[PENDING] %d: %s\n", m.Version, m.Name)
	}
}

func runSeed(db *gorm.DB) {
	log.Println("Running database seed...")

	adminUser := models.User{
		Email:        "admin@gpu-platform.local",
		PasswordHash: "admin123",
		Name:         "Administrator",
		Role:         "admin",
		Status:       "active",
	}
	db.Create(&adminUser)
	log.Println("Created admin user")

	project := models.Project{
		Name:        "Demo Project",
		Description: "Demo project for testing GPU containers",
	}
	db.Create(&project)
	log.Println("Created demo project")

	log.Println("Database seed completed")
}
