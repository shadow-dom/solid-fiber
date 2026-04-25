package main

import (
	"io/fs"
	"log"
	"os"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/static"

	"github.com/shadow-dom/solid-fiber/web"
)

func main() {
	app := fiber.New()

	api := app.Group("/api")
	api.Get("/hello", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "hello from fiber"})
	})

	dist, err := web.Dist()
	if err != nil {
		log.Fatalf("load embedded SPA: %v", err)
	}

	app.Use("/", static.New("", static.Config{
		FS: dist,
	}))

	// SPA fallback: any non-API GET that didn't match a static file returns index.html
	// so the client-side router can handle it.
	app.Get("/*", func(c fiber.Ctx) error {
		index, err := fs.ReadFile(dist, "index.html")
		if err != nil {
			return fiber.ErrNotFound
		}
		c.Set(fiber.HeaderContentType, fiber.MIMETextHTMLCharsetUTF8)
		return c.Send(index)
	})

	addr := ":3000"
	if v := os.Getenv("ADDR"); v != "" {
		addr = v
	}
	log.Fatal(app.Listen(addr))
}
