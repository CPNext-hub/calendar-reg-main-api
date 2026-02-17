package handler

import (
	"github.com/CPNext-hub/calendar-reg-main-api/internal/delivery/http/adapter"
	"github.com/CPNext-hub/calendar-reg-main-api/pkg/queue"
	"github.com/CPNext-hub/calendar-reg-main-api/pkg/response"
	"github.com/gofiber/fiber/v2"
)

// QueueHandler handles HTTP requests for queue status.
type QueueHandler struct {
	queue *queue.RefreshQueue
}

// NewQueueHandler creates a new QueueHandler instance.
func NewQueueHandler(q *queue.RefreshQueue) *QueueHandler {
	return &QueueHandler{queue: q}
}

// GetStatus returns the current refresh queue status.
// @Summary Get queue status
// @Description Returns pending, processed, dropped counts and capacity
// @Tags queue
// @Produce json
// @Success 200 {object} queue.QueueStatus
// @Router /queue/status [get]
func (h *QueueHandler) GetStatus(c *fiber.Ctx) error {
	return response.OK(adapter.NewFiberResponder(c), h.queue.Status())
}
