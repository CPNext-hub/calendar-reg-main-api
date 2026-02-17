package response

import (
	"encoding/json"
	"testing"

	"github.com/CPNext-hub/calendar-reg-main-api/pkg/port"
)

// mockResponder captures the status code and JSON body for testing.
type mockResponder struct {
	statusCode int
	body       interface{}
	sentStatus int // for SendStatus
}

func (m *mockResponder) Status(code int) port.Responder {
	m.statusCode = code
	return m
}

func (m *mockResponder) JSON(data interface{}) error {
	m.body = data
	return nil
}

func (m *mockResponder) SendStatus(code int) error {
	m.sentStatus = code
	return nil
}

func newMock() *mockResponder { return &mockResponder{} }

// bodyAs unmarshals the captured body into a Body struct for assertions.
func bodyAs(t *testing.T, m *mockResponder) Body {
	t.Helper()
	raw, err := json.Marshal(m.body)
	if err != nil {
		t.Fatalf("failed to marshal body: %v", err)
	}
	var b Body
	if err := json.Unmarshal(raw, &b); err != nil {
		t.Fatalf("failed to unmarshal body: %v", err)
	}
	return b
}

// ---------- Success helpers ----------

func TestOK(t *testing.T) {
	m := newMock()
	err := OK(m, map[string]string{"key": "value"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.statusCode != StatusOK {
		t.Errorf("expected status %d, got %d", StatusOK, m.statusCode)
	}
	b := bodyAs(t, m)
	if !b.Success {
		t.Error("expected success=true")
	}
	if b.Error != nil {
		t.Error("expected no error body")
	}
}

func TestCreated(t *testing.T) {
	m := newMock()
	err := Created(m, "new-item")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.statusCode != StatusCreated {
		t.Errorf("expected status %d, got %d", StatusCreated, m.statusCode)
	}
	b := bodyAs(t, m)
	if !b.Success {
		t.Error("expected success=true")
	}
}

func TestNoContent(t *testing.T) {
	m := newMock()
	err := NoContent(m)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.sentStatus != StatusNoContent {
		t.Errorf("expected sentStatus %d, got %d", StatusNoContent, m.sentStatus)
	}
}

// ---------- Error helpers ----------

func TestBadRequest(t *testing.T) {
	m := newMock()
	_ = BadRequest(m, "bad input")
	if m.statusCode != StatusBadRequest {
		t.Errorf("expected status %d, got %d", StatusBadRequest, m.statusCode)
	}
	b := bodyAs(t, m)
	if b.Success {
		t.Error("expected success=false")
	}
	if b.Error == nil || b.Error.Code != StatusBadRequest {
		t.Errorf("expected error code %d", StatusBadRequest)
	}
	if b.Error.Message != "bad input" {
		t.Errorf("expected message 'bad input', got %q", b.Error.Message)
	}
}

func TestUnauthorized(t *testing.T) {
	m := newMock()
	_ = Unauthorized(m, "no token")
	if m.statusCode != StatusUnauthorized {
		t.Errorf("expected status %d, got %d", StatusUnauthorized, m.statusCode)
	}
	b := bodyAs(t, m)
	if b.Success {
		t.Error("expected success=false")
	}
	if b.Error == nil || b.Error.Code != StatusUnauthorized {
		t.Errorf("expected error code %d", StatusUnauthorized)
	}
}

func TestForbidden(t *testing.T) {
	m := newMock()
	_ = Forbidden(m, "access denied")
	if m.statusCode != StatusForbidden {
		t.Errorf("expected status %d, got %d", StatusForbidden, m.statusCode)
	}
	b := bodyAs(t, m)
	if b.Error == nil || b.Error.Code != StatusForbidden {
		t.Errorf("expected error code %d", StatusForbidden)
	}
}

func TestNotFound(t *testing.T) {
	m := newMock()
	_ = NotFound(m, "not here")
	if m.statusCode != StatusNotFound {
		t.Errorf("expected status %d, got %d", StatusNotFound, m.statusCode)
	}
	b := bodyAs(t, m)
	if b.Error == nil || b.Error.Code != StatusNotFound {
		t.Errorf("expected error code %d", StatusNotFound)
	}
}

func TestConflict(t *testing.T) {
	m := newMock()
	_ = Conflict(m, "duplicate")
	if m.statusCode != StatusConflict {
		t.Errorf("expected status %d, got %d", StatusConflict, m.statusCode)
	}
	b := bodyAs(t, m)
	if b.Error == nil || b.Error.Code != StatusConflict {
		t.Errorf("expected error code %d", StatusConflict)
	}
}

func TestUnprocessableEntity(t *testing.T) {
	m := newMock()
	_ = UnprocessableEntity(m, "bad entity")
	if m.statusCode != StatusUnprocessableEntity {
		t.Errorf("expected status %d, got %d", StatusUnprocessableEntity, m.statusCode)
	}
	b := bodyAs(t, m)
	if b.Error == nil || b.Error.Code != StatusUnprocessableEntity {
		t.Errorf("expected error code %d", StatusUnprocessableEntity)
	}
}

func TestInternalError(t *testing.T) {
	m := newMock()
	_ = InternalError(m, "oops")
	if m.statusCode != StatusInternalServerError {
		t.Errorf("expected status %d, got %d", StatusInternalServerError, m.statusCode)
	}
	b := bodyAs(t, m)
	if b.Success {
		t.Error("expected success=false")
	}
	if b.Error == nil || b.Error.Code != StatusInternalServerError {
		t.Errorf("expected error code %d", StatusInternalServerError)
	}
}

func TestValidationError(t *testing.T) {
	m := newMock()
	fieldErrors := map[string]string{"email": "required"}
	_ = ValidationError(m, fieldErrors)
	if m.statusCode != StatusUnprocessableEntity {
		t.Errorf("expected status %d, got %d", StatusUnprocessableEntity, m.statusCode)
	}
	b := bodyAs(t, m)
	if b.Success {
		t.Error("expected success=false")
	}
	if b.Error == nil || b.Error.Message != "validation failed" {
		t.Errorf("expected message 'validation failed', got %q", b.Error.Message)
	}
	if b.Data == nil {
		t.Error("expected data to contain field errors")
	}
}

// ---------- Success body structure ----------

func TestOK_DataIsPresent(t *testing.T) {
	m := newMock()
	_ = OK(m, []string{"a", "b"})
	b := bodyAs(t, m)
	if b.Data == nil {
		t.Error("expected data to be present")
	}
}

func TestErrorBody_NoData(t *testing.T) {
	m := newMock()
	_ = InternalError(m, "err")
	b := bodyAs(t, m)
	if b.Data != nil {
		t.Error("expected no data for error response")
	}
}
