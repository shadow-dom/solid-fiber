package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v3"
)

// Pinger is the subset of *sql.DB the health check needs. Keeping it an
// interface avoids coupling the handler package to a concrete driver.
type Pinger interface {
	PingContext(ctx context.Context) error
}

// Health reports service liveness, verifying datastore connectivity.
func Health(p Pinger) fiber.Handler {
	return func(c fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := p.PingContext(ctx); err != nil {
			c.Status(http.StatusServiceUnavailable)
			return c.JSON(fiber.Map{"status": false, "data": fiber.Map{"status": "unavailable"}, "error": err.Error()})
		}
		return c.JSON(fiber.Map{"status": true, "data": fiber.Map{"status": "ok"}, "error": nil})
	}
}
