package response

import "github.com/CPNext-hub/calendar-reg-main-api/pkg/port"

// HTTP status codes (framework-agnostic).
const (
	StatusOK                  = 200
	StatusCreated             = 201
	StatusNoContent           = 204
	StatusBadRequest          = 400
	StatusUnauthorized        = 401
	StatusForbidden           = 403
	StatusNotFound            = 404
	StatusConflict            = 409
	StatusUnprocessableEntity = 422
	StatusInternalServerError = 500
)

// Body is the standard API response envelope.
type Body struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorBody  `json:"error,omitempty"`
}

// ErrorBody represents an error payload inside the response.
type ErrorBody struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// ---------- success helpers ----------

// OK sends a 200 response with data.
func OK(r port.Responder, data interface{}) error {
	return r.Status(StatusOK).JSON(Body{
		Success: true,
		Data:    data,
	})
}

// Created sends a 201 response with data.
func Created(r port.Responder, data interface{}) error {
	return r.Status(StatusCreated).JSON(Body{
		Success: true,
		Data:    data,
	})
}

// NoContent sends a 204 response with no body.
func NoContent(r port.Responder) error {
	return r.SendStatus(StatusNoContent)
}

// ---------- error helpers ----------

// BadRequest sends a 400 error response.
func BadRequest(r port.Responder, message string) error {
	return errResponse(r, StatusBadRequest, message)
}

// Unauthorized sends a 401 error response.
func Unauthorized(r port.Responder, message string) error {
	return errResponse(r, StatusUnauthorized, message)
}

// Forbidden sends a 403 error response.
func Forbidden(r port.Responder, message string) error {
	return errResponse(r, StatusForbidden, message)
}

// NotFound sends a 404 error response.
func NotFound(r port.Responder, message string) error {
	return errResponse(r, StatusNotFound, message)
}

// Conflict sends a 409 error response.
func Conflict(r port.Responder, message string) error {
	return errResponse(r, StatusConflict, message)
}

// UnprocessableEntity sends a 422 error response.
func UnprocessableEntity(r port.Responder, message string) error {
	return errResponse(r, StatusUnprocessableEntity, message)
}

// InternalError sends a 500 error response.
func InternalError(r port.Responder, message string) error {
	return errResponse(r, StatusInternalServerError, message)
}

// ---------- validation helper ----------

// ValidationError sends a 422 response with field-level validation errors.
func ValidationError(r port.Responder, errors interface{}) error {
	return r.Status(StatusUnprocessableEntity).JSON(Body{
		Success: false,
		Data:    errors,
		Error: &ErrorBody{
			Code:    StatusUnprocessableEntity,
			Message: "validation failed",
		},
	})
}

// ---------- internal ----------

func errResponse(r port.Responder, code int, message string) error {
	return r.Status(code).JSON(Body{
		Success: false,
		Error: &ErrorBody{
			Code:    code,
			Message: message,
		},
	})
}
