package routes

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"github.com/smartseaman/mobile-api/internal/domain/auth"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPublicActivationRejectsEmailOnlyWithoutCallingService(t *testing.T) {
	app := fiber.New()
	registerPublic(app, auth.NewHandler(nil, securityMessages{}), nil, nil)
	req := httptest.NewRequest("POST", "/profile/active", strings.NewReader(`{"email":"victim@example.com"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 400 {
		t.Fatalf("email-only activation status=%d", resp.StatusCode)
	}
}

type securityMessages struct{}

func (securityMessages) Message(context.Context, string, string) string { return "invalid credential" }
