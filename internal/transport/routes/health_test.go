package routes

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestHealthMetadata(t *testing.T) {
	for _, populated := range []bool{false, true} {
		t.Run(map[bool]string{false: "defaults", true: "release"}[populated], func(t *testing.T) {
			want := map[string]string{"status": "UP", "version": "dev", "commit": "unknown", "buildDate": "unknown", "deployedAt": "unknown", "deployedBy": "unknown"}
			for field, key := range map[string]string{"version": "APP_VERSION", "commit": "APP_COMMIT", "buildDate": "APP_BUILD_DATE", "deployedAt": "APP_DEPLOYED_AT", "deployedBy": "APP_DEPLOYED_BY"} {
				v := ""
				if populated {
					v = "release-" + field
					want[field] = v
				}
				t.Setenv(key, v)
			}
			t.Setenv("JWT_SECRET", "must-not-be-returned")
			app := fiber.New()
			app.Get("/health", healthHandler())
			resp, err := app.Test(httptest.NewRequest("GET", "/health", nil))
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != 200 || resp.Header.Get("Cache-Control") != "no-store" {
				t.Fatalf("unexpected health response: %v", resp)
			}
			var got map[string]string
			if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
				t.Fatal(err)
			}
			if len(got) != len(want) {
				t.Fatalf("unexpected fields: %v", got)
			}
			for key, value := range want {
				if got[key] != value {
					t.Errorf("%s=%q, want %q", key, got[key], value)
				}
			}
		})
	}
}
