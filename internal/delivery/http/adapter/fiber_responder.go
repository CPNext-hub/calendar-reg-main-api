package adapter

import (
	"github.com/CPNext-hub/calendar-reg-main-api/pkg/port"
	"github.com/gofiber/fiber/v2"
)

// FiberResponder adapts *fiber.Ctx to the port.Responder interface.
type FiberResponder struct {
	ctx *fiber.Ctx
}

// NewFiberResponder wraps a Fiber context as a Responder.
func NewFiberResponder(c *fiber.Ctx) port.Responder {
	return &FiberResponder{ctx: c}
}

func (r *FiberResponder) Status(code int) port.Responder {
	r.ctx.Status(code)
	return r
}

func (r *FiberResponder) JSON(data interface{}) error {
	return r.ctx.JSON(data)
}

func (r *FiberResponder) SendStatus(code int) error {
	return r.ctx.SendStatus(code)
}
