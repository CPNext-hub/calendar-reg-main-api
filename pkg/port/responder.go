package port

// Responder abstracts HTTP response writing.
// This allows handlers and response helpers to be framework-agnostic.
type Responder interface {
	// Status sets the HTTP status code and returns itself for chaining.
	Status(code int) Responder
	// JSON sends a JSON response body.
	JSON(data interface{}) error
	// SendStatus sends a response with only a status code (no body).
	SendStatus(code int) error
}
