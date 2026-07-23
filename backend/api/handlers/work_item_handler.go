package handlers

import (
	"errors"
	"net/http"

	"github.com/gofiber/fiber/v3"

	"github.com/shadow-dom/solid-fiber/api/presenter"
	"github.com/shadow-dom/solid-fiber/pkg/work_item"
)

// statusForError maps domain errors to HTTP status codes.
func statusForError(err error) int {
	switch {
	case errors.Is(err, work_item.ErrTitleRequired):
		return http.StatusBadRequest
	case errors.Is(err, work_item.ErrNotFound):
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}

// AddWorkItem handles POST /api/work-items.
func AddWorkItem(service work_item.Service) fiber.Handler {
	return func(c fiber.Ctx) error {
		var requestBody work_item.WorkItem
		if err := c.Bind().Body(&requestBody); err != nil {
			c.Status(http.StatusBadRequest)
			return c.JSON(presenter.WorkItemErrorResponse(err))
		}
		result, err := service.CreateWorkItem(&requestBody)
		if err != nil {
			c.Status(statusForError(err))
			return c.JSON(presenter.WorkItemErrorResponse(err))
		}
		c.Status(http.StatusCreated)
		return c.JSON(presenter.WorkItemSuccessResponse(result))
	}
}

// GetWorkItem handles GET /api/work-items/:id.
func GetWorkItem(service work_item.Service) fiber.Handler {
	return func(c fiber.Ctx) error {
		result, err := service.GetWorkItemByID(c.Params("id"))
		if err != nil {
			c.Status(statusForError(err))
			return c.JSON(presenter.WorkItemErrorResponse(err))
		}
		return c.JSON(presenter.WorkItemSuccessResponse(result))
	}
}

// UpdateWorkItem handles PUT /api/work-items/:id.
func UpdateWorkItem(service work_item.Service) fiber.Handler {
	return func(c fiber.Ctx) error {
		var requestBody work_item.WorkItem
		if err := c.Bind().Body(&requestBody); err != nil {
			c.Status(http.StatusBadRequest)
			return c.JSON(presenter.WorkItemErrorResponse(err))
		}
		// The path is the source of truth for identity.
		requestBody.ID = c.Params("id")
		result, err := service.UpdateWorkItem(&requestBody)
		if err != nil {
			c.Status(statusForError(err))
			return c.JSON(presenter.WorkItemErrorResponse(err))
		}
		return c.JSON(presenter.WorkItemSuccessResponse(result))
	}
}

// DeleteWorkItem handles DELETE /api/work-items/:id.
func DeleteWorkItem(service work_item.Service) fiber.Handler {
	return func(c fiber.Ctx) error {
		if err := service.DeleteWorkItem(c.Params("id")); err != nil {
			c.Status(statusForError(err))
			return c.JSON(presenter.WorkItemErrorResponse(err))
		}
		return c.SendStatus(http.StatusNoContent)
	}
}

// ListWorkItems handles GET /api/work-items?project_id=...
func ListWorkItems(service work_item.Service) fiber.Handler {
	return func(c fiber.Ctx) error {
		projectID := c.Query("project_id")
		if projectID == "" {
			c.Status(http.StatusBadRequest)
			return c.JSON(presenter.WorkItemErrorResponse(errors.New("project_id query parameter is required")))
		}
		results, err := service.ListWorkItemsByProjectID(projectID)
		if err != nil {
			c.Status(statusForError(err))
			return c.JSON(presenter.WorkItemErrorResponse(err))
		}
		return c.JSON(presenter.WorkItemsSuccessResponse(results))
	}
}
