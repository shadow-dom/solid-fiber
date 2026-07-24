package main

import (
	"context"
	"io/fs"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/static"

	"github.com/shadow-dom/solid-fiber/api/routes"
	"github.com/shadow-dom/solid-fiber/pkg/storage"
	"github.com/shadow-dom/solid-fiber/pkg/work_item"
	"github.com/shadow-dom/solid-fiber/web"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Open the datastore. DB_PATH overrides the default location; the file is
	// created on first run.
	dbPath := "work_items.db"
	if v := os.Getenv("DB_PATH"); v != "" {
		dbPath = v
	}
	db, err := storage.OpenSQLite(dbPath)
	if err != nil {
		slog.Error("open database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Dependency wiring: repository -> service -> handlers.
	repo, err := work_item.NewSQLiteRepository(db)
	if err != nil {
		slog.Error("init work item repository", "error", err)
		os.Exit(1)
	}
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
		slog.Error("load embedded SPA", "error", err)
		os.Exit(1)
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

	slog.Info("starting server", "addr", addr)
	if err := app.Listen(addr, fiber.ListenConfig{GracefulContext: ctx}); err != nil {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
}
