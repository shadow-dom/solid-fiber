package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v3"

	"github.com/shadow-dom/solid-fiber/api"
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

	dist, err := web.Dist()
	if err != nil {
		slog.Error("load embedded SPA", "error", err)
		os.Exit(1)
	}

	app := api.New(api.Config{
		WorkItems: workItemService,
		Pinger:    db,
		SPA:       dist,
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
