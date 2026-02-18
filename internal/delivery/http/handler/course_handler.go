package handler

import (
	"context"
	"strconv"
	"time"

	"github.com/CPNext-hub/calendar-reg-main-api/internal/delivery/http/adapter"
	"github.com/CPNext-hub/calendar-reg-main-api/internal/delivery/http/dto"
	"github.com/CPNext-hub/calendar-reg-main-api/internal/domain/usecase"
	"github.com/CPNext-hub/calendar-reg-main-api/pkg/pagination"
	"github.com/CPNext-hub/calendar-reg-main-api/pkg/response"
	"github.com/gofiber/fiber/v2"
)

// CourseHandler handles HTTP requests for courses.
type CourseHandler struct {
	usecase usecase.CourseUsecase
}

// NewCourseHandler creates a new CourseHandler instance.
func NewCourseHandler(uc usecase.CourseUsecase) *CourseHandler {
	return &CourseHandler{usecase: uc}
}

// CreateCourse creates a new course.
// @Summary Create a new course
// @Description Create a new course with sections and schedules
// @Tags courses
// @Accept json
// @Produce json
// @Param request body dto.CreateCourseRequest true "Course Request"
// @Success 201 {object} dto.CourseResponse
// @Failure 400 {object} interface{}
// @Failure 500 {object} interface{}
// @Router /courses [post]
func (h *CourseHandler) CreateCourse(c *fiber.Ctx) error {
	var req dto.CreateCourseRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(adapter.NewFiberResponder(c), "Invalid request body")
	}

	if req.Code == "" || req.NameEN == "" || req.Credits == "" {
		return response.BadRequest(adapter.NewFiberResponder(c), "Missing required fields")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	course := req.ToEntity()
	if err := h.usecase.CreateCourse(ctx, course); err != nil {
		return response.InternalError(adapter.NewFiberResponder(c), err.Error())
	}

	return response.Created(adapter.NewFiberResponder(c), dto.ToCourseResponse(course))
}

// GetCourses retrieves courses with pagination.
// @Summary Get courses (paginated)
// @Description Retrieve a paginated list of courses. Use limit=0 to fetch all.
// @Tags courses
// @Accept json
// @Produce json
// @Param page query int false "Page number (default 1)"
// @Param limit query int false "Items per page (default 10, 0=all)"
// @Success 200 {object} interface{}
// @Failure 500 {object} interface{}
// @Router /courses [get]
func (h *CourseHandler) GetCourses(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page"))
	limit, _ := strconv.Atoi(c.Query("limit"))
	pq := pagination.FromQuery(page, limit)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := h.usecase.GetCoursesPaginated(ctx, pq)
	if err != nil {
		return response.InternalError(adapter.NewFiberResponder(c), err.Error())
	}

	return response.OK(adapter.NewFiberResponder(c),
		dto.ToCourseSummaryResponses(result.Items),
		result.GetMeta(),
	)
}

// GetCourse retrieves a course by code.
// @Summary Get course by code
// @Description Retrieve a specific course details by its code
// @Tags courses
// @Accept json
// @Produce json
// @Param code path string true "Course Code"
// @Success 200 {object} dto.CourseResponse
// @Failure 404 {object} interface{}
// @Failure 500 {object} interface{}
// @Router /courses/{code} [get]
func (h *CourseHandler) GetCourse(c *fiber.Ctx) error {
	code := c.Params("code")
	acadyear, _ := strconv.Atoi(c.Query("acadyear"))
	semester, _ := strconv.Atoi(c.Query("semester"))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	course, err := h.usecase.GetCourseByCode(ctx, code, acadyear, semester)
	if err != nil {
		return response.InternalError(adapter.NewFiberResponder(c), err.Error())
	}
	if course == nil {
		return response.NotFound(adapter.NewFiberResponder(c), "Course not found")
	}

	return response.OK(adapter.NewFiberResponder(c), dto.ToCourseResponse(course))
}

// DeleteCourse deletes a course by code.
// @Summary Soft delete course by code
// @Description Soft delete a course (set deleted_at timestamp)
// @Tags courses
// @Accept json
// @Produce json
// @Param code path string true "Course Code"
// @Success 200 {object} interface{}
// @Failure 500 {object} interface{}
// @Router /courses/{code} [delete]
func (h *CourseHandler) DeleteCourse(c *fiber.Ctx) error {
	code := c.Params("code")
	acadyear, _ := strconv.Atoi(c.Query("acadyear"))
	semester, _ := strconv.Atoi(c.Query("semester"))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := h.usecase.DeleteCourse(ctx, code, acadyear, semester); err != nil {
		return response.InternalError(adapter.NewFiberResponder(c), err.Error())
	}

	return response.OK(adapter.NewFiberResponder(c), map[string]string{"message": "Course deleted"})
}
