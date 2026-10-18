package middleware

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/smartseaman/mobile-api/internal/transport/httpx"
)

const (
	TraceID   = "trace.id"
	StartedAt = "request.started_at"
)

func APIContext() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Locals(TraceID, strings.ToLower(uuid.NewString()))
		c.Locals(StartedAt, time.Now())

		for _, header := range []struct {
			name  string
			label string
		}{
			{fiber.HeaderAcceptLanguage, "Language"},
			{"devicemodel", "device model"},
			{"correlationid", "correlation id"},
		} {
			if strings.TrimSpace(c.Get(header.name)) == "" {
				return httpx.Error(c, fiber.StatusBadRequest, "MA00004", "[Header] Missing parameter "+header.label+".")
			}
		}
		return c.Next()
	}
}
