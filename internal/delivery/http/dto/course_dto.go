package dto

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/CPNext-hub/calendar-reg-main-api/internal/domain/entity"
)

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
			var startTime, endTime time.Time
			parts := strings.Split(sc.Time, "-")
			if len(parts) == 2 {
				// Parse "13:00" -> time.Time
				// Using a dummy date or just parsing time
				// time.Parse("15:04", ...) returns 0000-01-01 13:00:00 +0000 UTC
				st, err1 := time.Parse("15:04", strings.TrimSpace(parts[0]))
				et, err2 := time.Parse("15:04", strings.TrimSpace(parts[1]))
				if err1 == nil && err2 == nil {
					startTime = st
					endTime = et
				} else {
					log.Printf("Error parsing time for schedule: %v, %v", err1, err2)
				}
			}

			schedules[j] = entity.Schedule{
				Day:       sc.Day,
				StartTime: startTime,
				EndTime:   endTime,
				Room:      sc.Room,
				Type:      sc.Type,
			}
		}

		var examStart, examEnd time.Time
		if s.ExamDate != "" {
			es, ee, err := parseThaiExamDate(s.ExamDate)
			if err == nil {
				examStart = es
				examEnd = ee
			} else {
				log.Printf("Error parsing exam date: %v", err)
			}
		}

		sections[i] = entity.Section{
			Number:     s.Number,
			Schedules:  schedules,
			Seats:      s.Seats,
			Instructor: s.Instructor,
			ExamStart:  examStart,
			ExamEnd:    examEnd,
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
	UpdatedAt    string            `json:"updated_at"`
	Sections     []SectionResponse `json:"sections"`
}

// SectionResponse represents a section in the response.
type SectionResponse struct {
	Number     string             `json:"number"`
	Schedules  []ScheduleResponse `json:"schedules"`
	Seats      int                `json:"seats"`
	Instructor string             `json:"instructor"`
	ExamStart  string             `json:"exam_start,omitempty"`
	ExamEnd    string             `json:"exam_end,omitempty"`
}

// ScheduleResponse represents a schedule slot in the response.
type ScheduleResponse struct {
	Day       string `json:"day"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
	Room      string `json:"room"`
	Type      string `json:"type"`
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
				Day:       sc.Day,
				StartTime: sc.StartTime.Format("15:04"),
				EndTime:   sc.EndTime.Format("15:04"),
				Room:      sc.Room,
				Type:      sc.Type,
			}
		}

		var examStartStr, examEndStr string
		if !s.ExamStart.IsZero() {
			examStartStr = s.ExamStart.Format("2006-01-02 15:04:05")
		}
		if !s.ExamEnd.IsZero() {
			examEndStr = s.ExamEnd.Format("2006-01-02 15:04:05")
		}

		sections[i] = SectionResponse{
			Number:     s.Number,
			Schedules:  schedules,
			Seats:      s.Seats,
			Instructor: s.Instructor,
			ExamStart:  examStartStr,
			ExamEnd:    examEndStr,
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
		UpdatedAt:    c.UpdatedAt.Format(time.RFC3339),
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

// parseThaiExamDate parses strings like "31 มี.ค. 2569 เวลา 13:00 - 16:00"
func parseThaiExamDate(dateStr string) (time.Time, time.Time, error) {
	// Expected format: "dd MMM yyyy เวลา HH:mm - HH:mm"
	// Example: "31 มี.ค. 2569 เวลา 13:00 - 16:00"

	// Remove known suffixes or noise if any (though example is clean)
	dateStr = strings.TrimSpace(dateStr)

	parts := strings.Split(dateStr, " เวลา ")
	if len(parts) != 2 {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid format: missing ' เวลา ' separator")
	}

	datePart := strings.TrimSpace(parts[0]) // "31 มี.ค. 2569"
	timePart := strings.TrimSpace(parts[1]) // "13:00 - 16:00"

	// Parse date part
	dateFields := strings.Fields(datePart)
	if len(dateFields) != 3 {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid date format")
	}

	day, err := strconv.Atoi(dateFields[0])
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid day: %v", err)
	}

	monthStr := dateFields[1]
	yearBE, err := strconv.Atoi(dateFields[2])
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid year: %v", err)
	}
	// Convert Buddhist Era to Common Era
	yearCE := yearBE - 543

	month := parseThaiMonth(monthStr)
	if month == 0 {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid month: %s", monthStr)
	}

	// Parse time part
	times := strings.Split(timePart, "-")
	if len(times) != 2 {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid time range format")
	}

	startStr := strings.TrimSpace(times[0])
	endStr := strings.TrimSpace(times[1])

	startT, err := time.Parse("15:04", startStr)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid start time: %v", err)
	}
	endT, err := time.Parse("15:04", endStr)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid end time: %v", err)
	}

	// Combine
	examStart := time.Date(yearCE, month, day, startT.Hour(), startT.Minute(), 0, 0, time.Local)
	examEnd := time.Date(yearCE, month, day, endT.Hour(), endT.Minute(), 0, 0, time.Local)

	return examStart, examEnd, nil
}

func parseThaiMonth(abbr string) time.Month {
	switch abbr {
	case "ม.ค.":
		return time.January
	case "ก.พ.":
		return time.February
	case "มี.ค.":
		return time.March
	case "เม.ย.":
		return time.April
	case "พ.ค.":
		return time.May
	case "มิ.ย.":
		return time.June
	case "ก.ค.":
		return time.July
	case "ส.ค.":
		return time.August
	case "ก.ย.":
		return time.September
	case "ต.ค.":
		return time.October
	case "พ.ย.":
		return time.November
	case "ธ.ค.":
		return time.December
	default:
		return 0
	}
}
