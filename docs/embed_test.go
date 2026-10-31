package docs

import (
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"net/http/httptest"
	"testing"
)

func TestDocumentation(t *testing.T) {
	for _, enabled := range []bool{false, true} {
		app := fiber.New()
		Register(app, enabled)
		for _, path := range []string{"/docs", "/docs/openapi.yaml", "/docs/assets/init.js", "/docs/assets/swagger-ui-bundle.js", "/docs/assets/swagger-ui.css"} {
			r, err := app.Test(httptest.NewRequest("GET", path, nil))
			if err != nil {
				t.Fatal(err)
			}
			r.Body.Close()
			want := 404
			if enabled {
				want = 200
			}
			if r.StatusCode != want {
				t.Fatalf("%s: got %d want %d", path, r.StatusCode, want)
			}
		}
	}
	b, err := files.ReadFile("openapi.yaml")
	if err != nil {
		t.Fatal(err)
	}
	var spec map[string]any
	if err = json.Unmarshal(b, &spec); err != nil {
		t.Fatal(err)
	}
}
