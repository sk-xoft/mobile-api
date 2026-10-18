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
