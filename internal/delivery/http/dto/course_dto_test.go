package dto

import (
	"testing"
	"time"

	"github.com/CPNext-hub/calendar-reg-main-api/internal/domain/entity"
	"github.com/stretchr/testify/assert"
)

func TestToEntity(t *testing.T) {
	req := CreateCourseRequest{
		Code: "CP353004",
		Sections: []SectionRequest{
			{
				Number: "01",
				Schedules: []ScheduleRequest{
					{
						Day:  "Monday",
						Time: "13:00-15:00",
						Room: "Room 1",
						Type: "Lecture",
					},
				},
				ExamDate:    "31 มี.ค. 2569 เวลา 13:00 - 16:00",
				ReservedFor: []string{"Students who failed"},
			},
		},
	}

	entityCourse := req.ToEntity()

	assert.NotNil(t, entityCourse)
	assert.Equal(t, 1, len(entityCourse.Sections))

	section := entityCourse.Sections[0]
	assert.Equal(t, "01", section.Number)

	// Verify Schedule Time Split
	schedule := section.Schedules[0]
	expectedStart, _ := time.Parse("15:04", "13:00")
	expectedEnd, _ := time.Parse("15:04", "15:00")
	assert.Equal(t, expectedStart.Hour(), schedule.StartTime.Hour())
	assert.Equal(t, expectedStart.Minute(), schedule.StartTime.Minute())
	assert.Equal(t, expectedEnd.Hour(), schedule.EndTime.Hour())
	assert.Equal(t, expectedEnd.Minute(), schedule.EndTime.Minute())

	// Verify ExamDate Parsing
	// 2569 - 543 = 2026
	// 31 March
	// 13:00 - 16:00
	expectedExamStart := time.Date(2026, time.March, 31, 13, 0, 0, 0, time.Local)
	expectedExamEnd := time.Date(2026, time.March, 31, 16, 0, 0, 0, time.Local)

	assert.Equal(t, expectedExamStart.Year(), section.ExamStart.Year())
	assert.Equal(t, expectedExamStart.Month(), section.ExamStart.Month())
	assert.Equal(t, expectedExamStart.Day(), section.ExamStart.Day())
	assert.Equal(t, expectedExamStart.Hour(), section.ExamStart.Hour())

	assert.Equal(t, expectedExamEnd.Hour(), section.ExamEnd.Hour())
	assert.Equal(t, []string{"Students who failed"}, section.ReservedFor)
}

func TestToCourseResponse(t *testing.T) {
	start, _ := time.Parse("15:04", "09:00")
	end, _ := time.Parse("15:04", "12:00")

	examStart := time.Date(2026, time.March, 31, 13, 0, 0, 0, time.Local)
	examEnd := time.Date(2026, time.March, 31, 16, 0, 0, 0, time.Local)

	// Mock UpdatedAt
	updatedAt := time.Date(2026, time.February, 17, 14, 0, 0, 0, time.Local)

	entityCourse := &entity.Course{
		Code: "CP353004",
		BaseEntity: entity.BaseEntity{
			UpdatedAt: updatedAt,
		},
		Sections: []entity.Section{
			{
				Number: "01",
				Schedules: []entity.Schedule{
					{
						Day:       "Wednesday",
						StartTime: start,
						EndTime:   end,
						Room:      "Room 2",
						Type:      "Lab",
					},
				},
				ExamStart: examStart,
				ExamEnd:   examEnd,
			},
		},
	}

	response := ToCourseResponse(entityCourse)

	assert.NotNil(t, response)
	assert.Equal(t, 1, len(response.Sections))
	assert.Equal(t, updatedAt.Format(time.RFC3339), response.UpdatedAt)

	secResp := response.Sections[0]
	assert.Equal(t, "2026-03-31 13:00:00", secResp.ExamStart)
	assert.Equal(t, "2026-03-31 16:00:00", secResp.ExamEnd)

	schedule := secResp.Schedules[0]
	assert.Equal(t, "Wednesday", schedule.Day)
	assert.Equal(t, "09:00", schedule.StartTime)
	assert.Equal(t, "12:00", schedule.EndTime)
}

func TestToCourseSummaryResponse(t *testing.T) {
	updatedAt := time.Date(2026, time.February, 17, 14, 0, 0, 0, time.Local)
	entityCourse := &entity.Course{
		Code: "CP353004",
		BaseEntity: entity.BaseEntity{
			UpdatedAt: updatedAt,
		},
		Sections: []entity.Section{
			{Number: "01"}, // Should be ignored
		},
	}

	response := ToCourseSummaryResponse(entityCourse)

	assert.NotNil(t, response)
	assert.Equal(t, "CP353004", response.Code)
	assert.Equal(t, updatedAt.Format(time.RFC3339), response.UpdatedAt)
	// response.Sections does not exist, so we can't check it, which is the point.
}

func TestParseThaiExamDate_InvalidFormat(t *testing.T) {
	// Missing " เวลา "
	_, _, err := parseThaiExamDate("31 มี.ค. 2569 13:00 - 16:00")
	assert.Error(t, err)

	// Invalid Date Part
	_, _, err = parseThaiExamDate("31 Invalid 2569 เวลา 13:00 - 16:00")
	assert.Error(t, err)

	// Invalid Time Part
	_, _, err = parseThaiExamDate("31 มี.ค. 2569 เวลา 13:00")
	assert.Error(t, err)
}
