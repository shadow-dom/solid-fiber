package main

import (
	"context"
	"io/fs"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/static"

	"github.com/shadow-dom/solid-fiber/api/routes"
	"github.com/shadow-dom/solid-fiber/pkg/work_item"
	"github.com/shadow-dom/solid-fiber/web"
)

func main() {
	// Dependency wiring: repository -> service -> handlers.
	repo := work_item.NewInMemoryRepository()
	workItemService := work_item.NewService(repo)

	app := fiber.New()

	api := app.Group("/api")
	api.Get("/hello", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "hello from fiber"})
	})
	routes.WorkItemRouter(api, workItemService)

	// Any unmatched /api/* request is a real 404 (JSON), not the SPA shell.
	api.Use(func(c fiber.Ctx) error {
		return fiber.ErrNotFound
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

	// Graceful shutdown: cancel the context on SIGINT/SIGTERM so in-flight
	// requests can drain before the process exits.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := app.Listen(addr, fiber.ListenConfig{GracefulContext: ctx}); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
