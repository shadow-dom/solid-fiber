// Package api assembles the HTTP application: middleware, routes, health, and
// the embedded SPA. It is kept separate from main so it can be exercised in
// tests via app.Test.
package api

import (
	"errors"
	"io/fs"
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/helmet"
	recoverer "github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/requestid"
	"github.com/gofiber/fiber/v3/middleware/static"

	"github.com/shadow-dom/solid-fiber/api/handlers"
	"github.com/shadow-dom/solid-fiber/api/routes"
	"github.com/shadow-dom/solid-fiber/pkg/work_item"
)

// Config holds the dependencies needed to build the application.
type Config struct {
	WorkItems work_item.Service
	Pinger    handlers.Pinger
	SPA       fs.FS
}

// New builds the Fiber application with all middleware and routes wired up.
func New(cfg Config) *fiber.App {
	app := fiber.New(fiber.Config{
		ErrorHandler: jsonErrorHandler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
		BodyLimit:    1 * 1024 * 1024, // 1 MiB
	})

	// Global middleware. requestid runs first so it's available to the logger
	// and error handler; recover sits above the routes to catch panics.
	app.Use(requestid.New())
	app.Use(requestLogger())
	app.Use(recoverer.New())
	app.Use(helmet.New())

	api := app.Group("/api")
	api.Get("/hello", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "hello from fiber"})
	})
	api.Get("/health", handlers.Health(cfg.Pinger))
	routes.WorkItemRouter(api, cfg.WorkItems)

	// Any unmatched /api/* request is a real 404 (JSON), not the SPA shell.
	api.Use(func(c fiber.Ctx) error {
		return fiber.ErrNotFound
	})

	app.Use("/", static.New("", static.Config{FS: cfg.SPA}))

	// SPA fallback: any non-API GET that didn't match a static file returns
	// index.html so the client-side router can handle it.
	app.Get("/*", func(c fiber.Ctx) error {
		index, err := fs.ReadFile(cfg.SPA, "index.html")
		if err != nil {
			return fiber.ErrNotFound
		}
		c.Set(fiber.HeaderContentType, fiber.MIMETextHTMLCharsetUTF8)
		return c.Send(index)
	})

	return app
}

// jsonErrorHandler renders every error as the standard JSON envelope so API
// clients get a consistent shape (including 404s and recovered panics).
func jsonErrorHandler(c fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	var fe *fiber.Error
	if errors.As(err, &fe) {
		code = fe.Code
	}
	if code >= fiber.StatusInternalServerError {
		slog.Error("request error",
			"error", err,
			"method", c.Method(),
			"path", c.Path(),
			"request_id", requestid.FromContext(c),
		)
	}
	return c.Status(code).JSON(fiber.Map{"status": false, "data": nil, "error": err.Error()})
}

// requestLogger emits one structured line per request with the final status,
// latency, and request id.
func requestLogger() fiber.Handler {
	return func(c fiber.Ctx) error {
		start := time.Now()
		err := c.Next()

		status := c.Response().StatusCode()
		if err != nil {
			// Mirror jsonErrorHandler so the logged status matches the response.
			status = fiber.StatusInternalServerError
			var fe *fiber.Error
			if errors.As(err, &fe) {
				status = fe.Code
			}
		}
		slog.Info("http request",
			"method", c.Method(),
			"path", c.Path(),
			"status", status,
			"duration_ms", time.Since(start).Milliseconds(),
			"request_id", requestid.FromContext(c),
		)
		return err
	}
}
