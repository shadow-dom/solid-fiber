// Package routes wires HTTP endpoints to their handlers.
package routes

import (
	"github.com/gofiber/fiber/v3"

	"github.com/shadow-dom/solid-fiber/api/handlers"
	"github.com/shadow-dom/solid-fiber/pkg/work_item"
)

// WorkItemRouter registers the work item CRUD endpoints under the given router.
func WorkItemRouter(router fiber.Router, service work_item.Service) {
	group := router.Group("/work-items")
	group.Post("", handlers.AddWorkItem(service))
	group.Get("", handlers.ListWorkItems(service))
	group.Get("/:id", handlers.GetWorkItem(service))
	group.Put("/:id", handlers.UpdateWorkItem(service))
	group.Delete("/:id", handlers.DeleteWorkItem(service))
}
