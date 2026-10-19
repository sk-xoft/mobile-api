package config

import (
	"encoding/base64"
	"strings"
	"testing"
)

func TestLoadDecodesSpringCompatibleJWTSecret(t *testing.T) {
	t.Setenv("JWT_SECRET", base64.StdEncoding.EncodeToString([]byte(strings.Repeat("k", 32))))

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Security.JWTSecret != strings.Repeat("k", 32) {
		t.Fatalf("JWT secret was not decoded")
	}
}

func TestLoadRejectsInvalidJWTSecret(t *testing.T) {
	t.Setenv("JWT_SECRET", "not base64")

	if _, err := Load(); err == nil {
		t.Fatal("Load() expected an error")
	}
}

func TestPaymentModeMustMatchKeyAndFlagsAreValidated(t *testing.T) {
	t.Setenv("JWT_SECRET", "MDEyMzQ1Njc4OTAxMjM0NTY3ODkwMTIzNDU2Nzg5MDE=")
	t.Setenv("PAYMENT_WORKER_ENABLED", "true")
	t.Setenv("OMISE_SECRET_KEY", "skey_live_placeholder")
	t.Setenv("OMISE_LIVE_MODE", "false")
	if _, err := Load(); err == nil {
		t.Fatal("live key accepted in test mode")
	}
	t.Setenv("OMISE_LIVE_MODE", "true")
	if _, err := Load(); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PAYMENT_WORKER_ENABLED", "invalid")
	if _, err := Load(); err == nil {
		t.Fatal("invalid worker flag accepted")
	}
}
