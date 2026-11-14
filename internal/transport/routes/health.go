package routes

import (
	"os"

	"github.com/gofiber/fiber/v2"
)

// healthHandler reports process liveness and public release metadata only.
// It does not check database or external service availability.
func healthHandler() fiber.Handler {
	value := func(key, fallback string) string {
		if v := os.Getenv(key); v != "" {
			return v
		}
		return fallback
	}
	metadata := fiber.Map{
		"status":     "UP",
		"version":    value("APP_VERSION", "dev"),
		"commit":     value("APP_COMMIT", "unknown"),
		"buildDate":  value("APP_BUILD_DATE", "unknown"),
		"deployedAt": value("APP_DEPLOYED_AT", "unknown"),
		"deployedBy": value("APP_DEPLOYED_BY", "unknown"),
	}
	return func(c *fiber.Ctx) error {
		c.Set("Cache-Control", "no-store")
		return c.JSON(metadata)
	}
}
