package handlers

import (
	"errors"
	"net/http"

	"github.com/gofiber/fiber/v3"
	"github.com/shadow-dom/solid-fiber/pkg/work_item"
)

func AddWorkItem(service work_item.Service) fiber.Handler {
	return func(c fiber.Ctx) error {
		var requestBody work_item.WorkItem
		err := c.BodyParser(&requestBody)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return c.JSON(presenter.WorkItemErrorResponse(err))
		}
		if requestBody.Title == "" {
			c.Status(http.StatusInternalServerError)
			return c.JSON(presenter.WorkItemErrorResponse(errors.New(
				"Please specify title")))
		}
		result, err := service.CreateWorkItem(&requestBody)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return c.JSON(presenter.WorkItemErrorResponse(err))
		}
		return c.JSON(presenter.WorkItemSuccessResponse(result))
	}
}
