package routes

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"github.com/smartseaman/mobile-api/internal/domain/auth"
	"github.com/smartseaman/mobile-api/internal/middleware"
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

func TestHardDeleteStillRequiresAuthentication(t *testing.T) {
	app := fiber.New()
	registerProtected(app, middleware.Authenticate("test-secret", nil), nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	response, err := app.Test(httptest.NewRequest("DELETE", "/document-renewals/requests/261000004/hard", nil))
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("hard delete without JWT status=%d, want 401", response.StatusCode)
	}
}

func TestTimelineStatusRequiresAuthentication(t *testing.T) {
	app := fiber.New()
	registerProtected(app, middleware.Authenticate("test-secret", nil), nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	response, err := app.Test(httptest.NewRequest("GET", "/document-renewals/261000023/timeline/DOCUMENT_REVIEW/status", nil))
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("timeline status without JWT status=%d, want 401", response.StatusCode)
	}
}

func (securityMessages) Message(context.Context, string, string) string { return "invalid credential" }
