package config

import (
	"encoding/base64"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	App      App
	Database Database
	Security Security
	Payment  Payment
	Email    Email
	Storage  Storage
}

type App struct {
	DocsEnabled bool
	Address     string
	Profile     string
}

type Database struct {
	DSN             string
	MaxOpen         int
	MaxIdle         int
	ConnMaxLifetime time.Duration
}

type Security struct {
	JWTSecret      string
	AllowedOrigins []string
}

type Payment struct {
	QRExpire                         time.Duration
	WorkerEnabled                    bool
	LiveMode                         bool
	APIURL, SecretKey, WebhookSecret string
}

type Email struct {
	Host             string
	Port             int
	Username         string
	Password         string
	RegisterURL      string
	ResetPasswordURL string
}

type Storage struct {
	Endpoint        string
	AccessKey       string
	SecretKey       string
	Region          string
	Bucket          string
	ProfilePath     string
	DocumentPath    string
	FormPath        string
	NewsPath        string
	BannerPath      string
	VoucherPath     string
	VoucherQRPath   string
	RequestItemPath string
}

func Load() (Config, error) {
	cfg := Config{
		App: App{
			DocsEnabled: env("API_DOCS_ENABLED", "false") == "true",
			Address:     env("APP_ADDRESS", ":8080"),
			Profile:     env("APP_PROFILE", "prod"),
		},
		Database: Database{
			DSN:             os.Getenv("DATABASE_DSN"),
			MaxOpen:         envInt("DATABASE_MAX_OPEN_CONNS", 5),
			MaxIdle:         envInt("DATABASE_MAX_IDLE_CONNS", 3),
			ConnMaxLifetime: envDuration("DATABASE_CONN_MAX_LIFETIME", 30*time.Minute),
		},
		Security: Security{
			JWTSecret:      os.Getenv("JWT_SECRET"),
			AllowedOrigins: envList("CORS_ALLOWED_ORIGINS", []string{"https://mobile.smartseaman.com", "https://admin.smartseaman.com"}),
		},
		Payment: Payment{
			WorkerEnabled: env("PAYMENT_WORKER_ENABLED", "true") == "true",
			LiveMode:      env("OMISE_LIVE_MODE", "false") == "true",
			APIURL:        env("OMISE_API_URL", "https://api.omise.co"),
			SecretKey:     os.Getenv("OMISE_SECRET_KEY"),
			WebhookSecret: os.Getenv("OMISE_WEBHOOK_SECRET"),
			QRExpire:      time.Duration(envInt("PAYMENT_QR_EXPIRE_MINUTES", 15)) * time.Minute,
		},
		Email: Email{
			Host:             env("SMTP_HOST", "smtp.gmail.com"),
			Port:             envInt("SMTP_PORT", 587),
			Username:         os.Getenv("SMTP_USERNAME"),
			Password:         os.Getenv("SMTP_PASSWORD"),
			RegisterURL:      os.Getenv("REGISTER_CONFIRM_URL"),
			ResetPasswordURL: os.Getenv("FORGOT_CONFIRM_URL"),
		},
		Storage: Storage{
			Endpoint:        os.Getenv("OBJECT_STORE_ENDPOINT"),
			AccessKey:       os.Getenv("OBJECT_STORE_ACCESS_KEY"),
			SecretKey:       os.Getenv("OBJECT_STORE_SECRET_KEY"),
			Region:          env("OBJECT_STORE_REGION", "Singapore"),
			Bucket:          os.Getenv("OBJECT_STORE_BUCKET"),
			ProfilePath:     env("OBJECT_STORE_PROFILE_PATH", "profiles/images"),
			DocumentPath:    env("OBJECT_STORE_DOCUMENT_PATH", "documents/%s/%s"),
			FormPath:        env("OBJECT_STORE_FORM_PATH", "document_download"),
			NewsPath:        env("OBJECT_STORE_NEWS_PATH", "news"),
			BannerPath:      env("OBJECT_STORE_BANNER_PATH", "banner"),
			VoucherPath:     env("OBJECT_STORE_VOUCHER_PATH", "voucher"),
			VoucherQRPath:   env("OBJECT_STORE_VOUCHER_QR_PATH", "voucher/qr"),
			RequestItemPath: env("OBJECT_STORE_REQUEST_ITEM_PATH", "documents/%s/request-items/%s"),
		},
	}

	for _, flag := range []string{"API_DOCS_ENABLED", "PAYMENT_WORKER_ENABLED", "OMISE_LIVE_MODE"} {
		if value := os.Getenv(flag); value != "" {
			if _, err := strconv.ParseBool(value); err != nil {
				return Config{}, fmt.Errorf("%s must be a boolean", flag)
			}
		}
	}
	cfg.App.DocsEnabled, _ = strconv.ParseBool(env("API_DOCS_ENABLED", "false"))
	cfg.Payment.WorkerEnabled, _ = strconv.ParseBool(env("PAYMENT_WORKER_ENABLED", "true"))
	cfg.Payment.LiveMode, _ = strconv.ParseBool(env("OMISE_LIVE_MODE", "false"))
	if strings.HasPrefix(cfg.Payment.SecretKey, "skey_live_") && !cfg.Payment.LiveMode {
		return Config{}, fmt.Errorf("OMISE_LIVE_MODE must be true for a live key")
	}
	if strings.HasPrefix(cfg.Payment.SecretKey, "skey_test_") && cfg.Payment.LiveMode {
		return Config{}, fmt.Errorf("OMISE_LIVE_MODE must be false for a test key")
	}
	if cfg.Security.JWTSecret == "" {
		return Config{}, fmt.Errorf("JWT_SECRET is required")
	}
	if cfg.Payment.QRExpire > 24*time.Hour {
		return Config{}, fmt.Errorf("PAYMENT_QR_EXPIRE_MINUTES must not exceed 1440")
	}
	decodedSecret, err := base64.StdEncoding.DecodeString(cfg.Security.JWTSecret)
	if err != nil {
		return Config{}, fmt.Errorf("JWT_SECRET must be base64 encoded: %w", err)
	}
	cfg.Security.JWTSecret = string(decodedSecret)
	return cfg, nil
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func envInt(key string, fallback int) int {
	value, err := strconv.Atoi(os.Getenv(key))
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

func envDuration(key string, fallback time.Duration) time.Duration {
	value, err := time.ParseDuration(os.Getenv(key))
	if err != nil {
		return fallback
	}
	return value
}

func envList(key string, fallback []string) []string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if item := strings.TrimSpace(part); item != "" {
			result = append(result, item)
		}
	}
	return result
}
