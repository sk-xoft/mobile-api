package docs

import (
	"embed"
	"github.com/gofiber/fiber/v2"
)

//go:embed openapi.yaml index.html assets/*
var files embed.FS

func Register(app *fiber.App, enabled bool) {
	if !enabled {
		return
	}
	app.Get("/docs", func(c *fiber.Ctx) error { b, _ := files.ReadFile("index.html"); c.Type("html"); return c.Send(b) })
	app.Get("/docs/*", func(c *fiber.Ctx) error {
		name := c.Params("*")
		if name != "openapi.yaml" && name != "assets/init.js" && name != "assets/swagger-ui-bundle.js" && name != "assets/swagger-ui.css" {
			return c.SendStatus(404)
		}
		b, err := files.ReadFile(name)
		if err != nil {
			return c.SendStatus(404)
		}
		if name == "openapi.yaml" {
			c.Set("Content-Type", "application/yaml")
		} else if name == "assets/swagger-ui.css" {
			c.Type("css")
		} else {
			c.Type("js")
		}
		return c.Send(b)
	})
}
