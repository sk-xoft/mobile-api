package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestAPIContextRejectsMissingLegacyHeaders(t *testing.T) {
	app := fiber.New()
	app.Get("/v1/test", APIContext(), func(c *fiber.Ctx) error { return c.SendStatus(fiber.StatusNoContent) })

	response, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/v1/test", nil))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if response.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.StatusCode, fiber.StatusBadRequest)
	}
}

func TestAPIContextAcceptsRequiredLegacyHeaders(t *testing.T) {
	app := fiber.New()
	app.Get("/v1/test", APIContext(), func(c *fiber.Ctx) error { return c.SendStatus(fiber.StatusNoContent) })
	request := httptest.NewRequest(fiber.MethodGet, "/v1/test", nil)
	request.Header.Set(fiber.HeaderAcceptLanguage, "TH")
	request.Header.Set("devicemodel", "iPhone")
	request.Header.Set("correlationid", "correlation-1")

	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if response.StatusCode != fiber.StatusNoContent {
		t.Fatalf("status = %d, want %d", response.StatusCode, fiber.StatusNoContent)
	}
}
