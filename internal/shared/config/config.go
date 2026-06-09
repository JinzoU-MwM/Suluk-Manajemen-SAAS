package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	App      AppConfig
	Database DatabaseConfig
	Redis    RedisConfig
	NATS     NATSConfig
	MinIO    MinIOConfig
	JWT      JWTConfig
	Server   ServerConfig
	Gemini   GeminiConfig
	Pakasir  PakasirConfig
	Internal InternalConfig
	Email    EmailConfig
}

// EmailConfig holds transactional-email settings (SMTP preferred, Resend fallback).
type EmailConfig struct {
	ResendAPIKey string
	From         string
	SMTPHost     string
	SMTPPort     int
	SMTPUser     string
	SMTPPass     string
}

type AppConfig struct {
	Env string
	// PublicURL is the public base URL of the app (e.g. https://app.suluk.id),
	// used to build payment redirect-back links. No trailing slash.
	PublicURL string
}

// PakasirConfig holds the Pakasir payment-gateway credentials/settings.
type PakasirConfig struct {
	APIKey      string
	ProjectSlug string
	BaseURL     string
	Sandbox     bool
}

// InternalConfig holds the shared secret used to authenticate
// service-to-service (internal) endpoints.
type InternalConfig struct {
	APIKey string
}

type ServerConfig struct {
	Port int
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		d.User, d.Password, d.Host, d.Port, d.DBName, d.SSLMode)
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type NATSConfig struct {
	Addr string
}

type MinIOConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	UseSSL    bool
}

type JWTConfig struct {
	PrivateKeyPath string
	PublicKeyPath  string
	AccessTTL      string
	RefreshTTL     string
}

type GeminiConfig struct {
	APIKey string
}

func Load() *Config {
	return &Config{
		App: AppConfig{
			Env:       envOr("APP_ENV", "development"),
			PublicURL: strings.TrimRight(envOr("APP_PUBLIC_URL", "http://localhost:5173"), "/"),
		},
		Server: ServerConfig{
			Port: intEnvOr("SERVER_PORT", 8080),
		},
		Database: DatabaseConfig{
			Host:     envOr("POSTGRES_HOST", "localhost"),
			Port:     intEnvOr("POSTGRES_PORT", 5433),
			User:     envOr("POSTGRES_USER", "jamaah"),
			Password: envOr("POSTGRES_PASSWORD", "Jamaah123!"),
			DBName:   envOr("POSTGRES_DB", "jamaah_auth"),
			SSLMode:  envOr("POSTGRES_SSLMODE", "disable"),
		},
		Redis: RedisConfig{
			Addr:     envOr("REDIS_ADDR", "localhost:6379"),
			Password: envOr("REDIS_PASSWORD", ""),
			DB:       intEnvOr("REDIS_DB", 0),
		},
		NATS: NATSConfig{
			Addr: envOr("NATS_ADDR", "nats://localhost:4222"),
		},
		MinIO: MinIOConfig{
			Endpoint:  envOr("MINIO_ENDPOINT", "localhost:9000"),
			AccessKey: envOr("MINIO_ACCESS_KEY", "minioadmin"),
			SecretKey: envOr("MINIO_SECRET_KEY", "minioadmin"),
			Bucket:    envOr("MINIO_BUCKET", "jamaah-docs"),
			UseSSL:    envOr("MINIO_USE_SSL", "false") == "true",
		},
		JWT: JWTConfig{
			PrivateKeyPath: envOr("JWT_PRIVATE_KEY_PATH", "./certs/private.pem"),
			PublicKeyPath:  envOr("JWT_PUBLIC_KEY_PATH", "./certs/public.pem"),
			AccessTTL:      envOr("JWT_ACCESS_TTL", "15m"),
			RefreshTTL:     envOr("JWT_REFRESH_TTL", "168h"),
		},
		Gemini: GeminiConfig{
			APIKey: envOr("GEMINI_API_KEY", ""),
		},
		Pakasir: PakasirConfig{
			APIKey:      envOr("PAKASIR_API_KEY", ""),
			ProjectSlug: envOr("PAKASIR_PROJECT_SLUG", ""),
			BaseURL:     strings.TrimRight(envOr("PAKASIR_BASE_URL", "https://app.pakasir.com"), "/"),
			Sandbox:     envOr("PAKASIR_SANDBOX", "false") == "true",
		},
		Internal: InternalConfig{
			APIKey: envOr("INTERNAL_API_KEY", ""),
		},
		Email: EmailConfig{
			ResendAPIKey: envOr("RESEND_API_KEY", ""),
			From:         envOr("EMAIL_FROM", "Suluk <onboarding@resend.dev>"),
			SMTPHost:     envOr("SMTP_HOST", ""),
			SMTPPort:     intEnvOr("SMTP_PORT", 465),
			SMTPUser:     envOr("SMTP_USER", ""),
			SMTPPass:     envOr("SMTP_PASS", ""),
		},
	}
}

// Validate fails fast (in production) when critical secrets are not explicitly
// set, so a service never silently falls back to a built-in default credential.
// POSTGRES_PASSWORD is required for every service (they all connect to the DB).
// Service-specific secrets are passed via extraRequired by the services that use
// them — e.g. auth/invoice require INTERNAL_API_KEY, invoice also the Pakasir
// keys. (INTERNAL_API_KEY must NOT be globally required: most services neither
// set nor use it, and requiring it would fatal them on boot.)
// Dev keeps the convenient defaults. Call once after Load() in each service main.
func (c *Config) Validate(extraRequired ...string) {
	if c.App.Env != "production" {
		return
	}
	var missing []string
	required := append([]string{"POSTGRES_PASSWORD"}, extraRequired...)
	for _, key := range required {
		if os.Getenv(key) == "" {
			missing = append(missing, key)
		}
	}
	if len(missing) > 0 {
		log.Fatalf("config: required env vars not set in production: %s", strings.Join(missing, ", "))
	}
	if c.Database.SSLMode == "disable" {
		log.Printf("config: WARNING POSTGRES_SSLMODE=disable in production — database traffic is unencrypted")
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func intEnvOr(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}
