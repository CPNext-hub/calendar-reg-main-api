package dto

import "github.com/CPNext-hub/calendar-reg-main-api/internal/domain/entity"

// --- Request DTOs ---

// CreateCourseRequest represents the request body for creating a course.
type CreateCourseRequest struct {
	Code         string           `json:"code"`
	NameEN       string           `json:"name_en"`
	NameTH       string           `json:"name_th"`
	Faculty      string           `json:"faculty"`
	Credits      string           `json:"credits"`
	Prerequisite string           `json:"prerequisite,omitempty"`
	Semester     int              `json:"semester"`
	Year         int              `json:"year"`
	Program      string           `json:"program"`
	Sections     []SectionRequest `json:"sections"`
}

// SectionRequest represents a section in a create/update request.
type SectionRequest struct {
	Number     string            `json:"number"`
	Schedules  []ScheduleRequest `json:"schedules"`
	Seats      int               `json:"seats"`
	Instructor string            `json:"instructor"`
	ExamDate   string            `json:"exam_date,omitempty"`
}

// ScheduleRequest represents a schedule slot in a request.
type ScheduleRequest struct {
	Day  string `json:"day"`
	Time string `json:"time"`
	Room string `json:"room"`
	Type string `json:"type"`
}

// ToEntity converts a CreateCourseRequest to a domain entity.
func (r *CreateCourseRequest) ToEntity() *entity.Course {
	sections := make([]entity.Section, len(r.Sections))
	for i, s := range r.Sections {
		schedules := make([]entity.Schedule, len(s.Schedules))
		for j, sc := range s.Schedules {
			schedules[j] = entity.Schedule{
				Day:  sc.Day,
				Time: sc.Time,
				Room: sc.Room,
				Type: sc.Type,
			}
		}
		sections[i] = entity.Section{
			Number:     s.Number,
			Schedules:  schedules,
			Seats:      s.Seats,
			Instructor: s.Instructor,
			ExamDate:   s.ExamDate,
		}
	}

	return &entity.Course{
		Code:         r.Code,
		NameEN:       r.NameEN,
		NameTH:       r.NameTH,
		Faculty:      r.Faculty,
		Credits:      r.Credits,
		Prerequisite: r.Prerequisite,
		Semester:     r.Semester,
		Year:         r.Year,
		Program:      r.Program,
		Sections:     sections,
	}
}

// --- Response DTOs ---

// CourseResponse represents the response body for a course.
type CourseResponse struct {
	ID           string            `json:"id"`
	Code         string            `json:"code"`
	NameEN       string            `json:"name_en"`
	NameTH       string            `json:"name_th"`
	Faculty      string            `json:"faculty"`
	Credits      string            `json:"credits"`
	Prerequisite string            `json:"prerequisite,omitempty"`
	Semester     int               `json:"semester"`
	Year         int               `json:"year"`
	Program      string            `json:"program"`
	Sections     []SectionResponse `json:"sections"`
}

// SectionResponse represents a section in the response.
type SectionResponse struct {
	Number     string             `json:"number"`
	Schedules  []ScheduleResponse `json:"schedules"`
	Seats      int                `json:"seats"`
	Instructor string             `json:"instructor"`
	ExamDate   string             `json:"exam_date,omitempty"`
}

// ScheduleResponse represents a schedule slot in the response.
type ScheduleResponse struct {
	Day  string `json:"day"`
	Time string `json:"time"`
	Room string `json:"room"`
	Type string `json:"type"`
}

// ToCourseResponse converts a Course entity to a CourseResponse DTO.
func ToCourseResponse(c *entity.Course) *CourseResponse {
	if c == nil {
		return nil
	}

	sections := make([]SectionResponse, len(c.Sections))
	for i, s := range c.Sections {
		schedules := make([]ScheduleResponse, len(s.Schedules))
		for j, sc := range s.Schedules {
			schedules[j] = ScheduleResponse{
				Day:  sc.Day,
				Time: sc.Time,
				Room: sc.Room,
				Type: sc.Type,
			}
		}
		sections[i] = SectionResponse{
			Number:     s.Number,
			Schedules:  schedules,
			Seats:      s.Seats,
			Instructor: s.Instructor,
			ExamDate:   s.ExamDate,
		}
	}

	return &CourseResponse{
		ID:           c.ID,
		Code:         c.Code,
		NameEN:       c.NameEN,
		NameTH:       c.NameTH,
		Faculty:      c.Faculty,
		Credits:      c.Credits,
		Prerequisite: c.Prerequisite,
		Semester:     c.Semester,
		Year:         c.Year,
		Program:      c.Program,
		Sections:     sections,
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
