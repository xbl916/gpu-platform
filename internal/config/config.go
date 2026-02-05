package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

var cfg *Config

type Config struct {
	App          AppConfig          `mapstructure:"app"`
	JWT          JWTConfig          `mapstructure:"jwt"`
	Database     DatabaseConfig     `mapstructure:"database"`
	Redis        RedisConfig        `mapstructure:"redis"`
	Kubernetes   KubernetesConfig   `mapstructure:"kubernetes"`
	GPU          GPUConfig          `mapstructure:"gpu"`
	Container    ContainerConfig    `mapstructure:"container"`
	Storage      StorageConfig      `mapstructure:"storage"`
	Monitoring   MonitoringConfig   `mapstructure:"monitoring"`
	Notification NotificationConfig `mapstructure:"notification"`
	Registry     RegistryConfig     `mapstructure:"registry"`
	RateLimit    RateLimitConfig    `mapstructure:"rate_limit"`
	Session      SessionConfig      `mapstructure:"session"`
}

type AppConfig struct {
	Name     string `mapstructure:"name"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Env      string `mapstructure:"env"`
	Debug    bool   `mapstructure:"debug"`
	LogLevel string `mapstructure:"log_level"`
}

type JWTConfig struct {
	Secret             string `mapstructure:"secret"`
	AccessTokenExpire  int    `mapstructure:"access_token_expire"`
	RefreshTokenExpire int    `mapstructure:"refresh_token_expire"`
}

type DatabaseConfig struct {
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	Username        string `mapstructure:"username"`
	Password        string `mapstructure:"password"`
	Name            string `mapstructure:"name"`
	SSLMode         string `mapstructure:"sslmode"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"`
}

func (d *DatabaseConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.Username, d.Password, d.Name, d.SSLMode)
}

type RedisConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	Password     string `mapstructure:"password"`
	DB           int    `mapstructure:"db"`
	PoolSize     int    `mapstructure:"pool_size"`
	MinIdleConns int    `mapstructure:"min_idle_conns"`
}

func (r *RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

type KubernetesConfig struct {
	Kubeconfig    string `mapstructure:"kubeconfig"`
	Namespace     string `mapstructure:"namespace"`
	GPUPoolLabel  string `mapstructure:"gpu_pool_label"`
	InstanceLabel string `mapstructure:"instance_label"`
}

type GPUConfig struct {
	DefaultGPUModel   string `mapstructure:"default_gpu_model"`
	DefaultGPUCount   int    `mapstructure:"default_gpu_count"`
	MaxGPUPerInstance int    `mapstructure:"max_gpu_per_instance"`
	DefaultCPUCores   int    `mapstructure:"default_cpu_cores"`
	MaxCPUCores       int    `mapstructure:"max_cpu_cores"`
	DefaultMemoryMB   int    `mapstructure:"default_memory_mb"`
	MaxMemoryMB       int    `mapstructure:"max_memory_mb"`
	DefaultStorageGB  int    `mapstructure:"default_storage_gb"`
	MaxStorageGB      int    `mapstructure:"max_storage_gb"`
}

type ContainerConfig struct {
	DefaultImage        string `mapstructure:"default_image"`
	DefaultCommand      string `mapstructure:"default_command"`
	WorkspaceSizeGB     int    `mapstructure:"workspace_size_gb"`
	DataSizeGB          int    `mapstructure:"data_size_gb"`
	MaxInstancesPerUser int    `mapstructure:"max_instances_per_user"`
	MaxRunningHours     int    `mapstructure:"max_running_hours"`
}

type StorageConfig struct {
	Endpoint     string `mapstructure:"endpoint"`
	AccessKey    string `mapstructure:"access_key"`
	SecretKey    string `mapstructure:"secret_key"`
	UseSSL       bool   `mapstructure:"use_ssl"`
	BucketPrefix string `mapstructure:"bucket_prefix"`
	Region       string `mapstructure:"region"`
}

type MonitoringConfig struct {
	Enabled                 bool `mapstructure:"enabled"`
	MetricsPort             int  `mapstructure:"metrics_port"`
	ScrapeInterval          int  `mapstructure:"scrape_interval"`
	AlertThresholdGPUTemp   int  `mapstructure:"alert_threshold_gpu_temp"`
	AlertThresholdGPUMemory int  `mapstructure:"alert_threshold_gpu_memory"`
	AlertThresholdCPUUsage  int  `mapstructure:"alert_threshold_cpu_usage"`
}

type NotificationConfig struct {
	Enabled bool          `mapstructure:"enabled"`
	Email   EmailConfig   `mapstructure:"email"`
	Webhook WebhookConfig `mapstructure:"webhook"`
}

type EmailConfig struct {
	SMTPHost     string `mapstructure:"smtp_host"`
	SMTPPort     int    `mapstructure:"smtp_port"`
	SMTPUser     string `mapstructure:"smtp_user"`
	SMTPPassword string `mapstructure:"smtp_password"`
	FromAddress  string `mapstructure:"from_address"`
}

type WebhookConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	URL     string `mapstructure:"url"`
}

type RegistryConfig struct {
	Address  string `mapstructure:"address"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Insecure bool   `mapstructure:"insecure"`
}

type RateLimitConfig struct {
	Enabled           bool `mapstructure:"enabled"`
	RequestsPerMinute int  `mapstructure:"requests_per_minute"`
	BurstSize         int  `mapstructure:"burst_size"`
}

type SessionConfig struct {
	CookieName     string `mapstructure:"cookie_name"`
	CookieSecure   bool   `mapstructure:"cookie_secure"`
	CookieHTTPOnly bool   `mapstructure:"cookie_http_only"`
	MaxSessionAge  int    `mapstructure:"max_session_age"`
}

func (s *SessionConfig) MaxSessionDuration() time.Duration {
	return time.Duration(s.MaxSessionAge) * time.Second
}

func Init(configPath string) (*Config, error) {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	cfg.processEnvVars()

	return cfg, nil
}

func (c *Config) processEnvVars() {
	if c.JWT.Secret == "" || strings.HasPrefix(c.JWT.Secret, "${") {
		c.JWT.Secret = os.Getenv("JWT_SECRET_KEY")
	}
	if c.Database.Host == "" || strings.HasPrefix(c.Database.Host, "${") {
		c.Database.Host = os.Getenv("DB_HOST")
	}
	if c.Database.Password == "" || strings.HasPrefix(c.Database.Password, "${") {
		c.Database.Password = os.Getenv("DB_PASSWORD")
	}
	if c.Redis.Host == "" || strings.HasPrefix(c.Redis.Host, "${") {
		c.Redis.Host = os.Getenv("REDIS_HOST")
	}
	if c.Redis.Password == "" || strings.HasPrefix(c.Redis.Password, "${") {
		c.Redis.Password = os.Getenv("REDIS_PASSWORD")
	}
}

func Get() *Config {
	return cfg
}

func (c *Config) IsProduction() bool {
	return c.App.Env == "production"
}

func (c *Config) IsDevelopment() bool {
	return c.App.Env == "development"
}
