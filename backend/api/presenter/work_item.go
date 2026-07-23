// Package presenter shapes domain values into the JSON envelope returned by the API.
package presenter

import (
	"github.com/gofiber/fiber/v3"
	"github.com/shadow-dom/solid-fiber/pkg/work_item"
)

// WorkItemSuccessResponse wraps a single work item in the standard success envelope.
func WorkItemSuccessResponse(data *work_item.WorkItem) fiber.Map {
	return fiber.Map{
		"status": true,
		"data":   data,
		"error":  nil,
	}
}

// WorkItemsSuccessResponse wraps a list of work items in the standard success envelope.
func WorkItemsSuccessResponse(data []*work_item.WorkItem) fiber.Map {
	return fiber.Map{
		"status": true,
		"data":   data,
		"error":  nil,
	}
}

// WorkItemErrorResponse wraps an error in the standard error envelope.
func WorkItemErrorResponse(err error) fiber.Map {
	return fiber.Map{
		"status": false,
		"data":   nil,
		"error":  err.Error(),
	}
}
