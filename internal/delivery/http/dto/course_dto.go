package dto

import "github.com/CPNext-hub/calendar-reg-main-api/internal/domain/entity"

// CreateCourseRequest represents the request body for creating a course.
type CreateCourseRequest struct {
	Code    string `json:"code"`
	Name    string `json:"name"`
	Credits string `json:"credits"`
}

// CourseResponse represents the response body for a course.
type CourseResponse struct {
	Code    string `json:"code"`
	Name    string `json:"name"`
	Credits string `json:"credits"`
}

// ToCourseResponse converts a Course entity to a CourseResponse DTO.
func ToCourseResponse(c *entity.Course) *CourseResponse {
	if c == nil {
		return nil
	}
	return &CourseResponse{
		Code:    c.Code,
		Name:    c.Name,
		Credits: c.Credits,
	}
}

// ToCourseResponses converts a slice of Course entities to CourseResponse DTOs.
func ToCourseResponses(courses []*entity.Course) []*CourseResponse {
	responses := make([]*CourseResponse, len(courses))
	for i, c := range courses {
		responses[i] = ToCourseResponse(c)
	}
	return responses
}
