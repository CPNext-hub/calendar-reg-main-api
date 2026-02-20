package dto

import (
	"log"
	"testing"
	"time"

	"github.com/CPNext-hub/calendar-reg-main-api/internal/domain/entity"
	"github.com/stretchr/testify/assert"
)

func TestToEntity(t *testing.T) {
	req := CreateCourseRequest{
		Code:       "CP353004",
		Faculty:    "วิทยาลัยการคอมพิวเตอร์",
		Department: "วิทยาการคอมพิวเตอร์",
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
				MidtermDate: "15 ก.พ. 2569 เวลา 09:00 - 12:00",
				ReservedFor: []string{"Students who failed"},
			},
		},
	}

	entityCourse := req.ToEntity()

	assert.NotNil(t, entityCourse)
	assert.Equal(t, "วิทยาลัยการคอมพิวเตอร์", entityCourse.Faculty)
	assert.Equal(t, "วิทยาการคอมพิวเตอร์", entityCourse.Department)
	assert.Equal(t, 1, len(entityCourse.Sections))

	section := entityCourse.Sections[0]
	assert.Equal(t, "01", section.Number)

	// Verify Schedule Time Split
	schedule := section.Schedules[0]
	assert.Equal(t, "13:00", schedule.StartTime)
	assert.Equal(t, "15:00", schedule.EndTime)

	// Verify ExamDate Parsing
	// 2569 - 543 = 2026
	// 31 March
	// 13:00 - 16:00
	expectedExamStart := "2026-03-31 13:00:00"
	expectedExamEnd := "2026-03-31 16:00:00"

	assert.Equal(t, expectedExamStart, section.ExamStart)
	assert.Equal(t, expectedExamEnd, section.ExamEnd)
	assert.Equal(t, []string{"Students who failed"}, section.ReservedFor)

	// Verify MidtermDate Parsing
	expectedMidtermStart := "2026-02-15 09:00:00"
	expectedMidtermEnd := "2026-02-15 12:00:00"
	assert.Equal(t, expectedMidtermStart, section.MidtermStart)
	assert.Equal(t, expectedMidtermEnd, section.MidtermEnd)
}

func TestToEntity_InvalidTimes(t *testing.T) {
	// Capture stderr to avoid polluting test output with logs
	// But in this environment we can just let it log.
	// Ideally we mock the logger or just accept the output.
	log.SetOutput(log.Writer())

	req := CreateCourseRequest{
		Sections: []SectionRequest{
			{
				Schedules: []ScheduleRequest{
					{Time: "Invalid"},       // Invalid split
					{Time: "13:00-Invalid"}, // Invalid parse
				},
				ExamDate:    "Invalid Exam Date",
				MidtermDate: "Invalid Midterm Date",
			},
		},
	}
	entityCourse := req.ToEntity()
	assert.NotNil(t, entityCourse)
	assert.Equal(t, "", entityCourse.Sections[0].Schedules[0].StartTime)
	assert.Equal(t, "13:00", entityCourse.Sections[0].Schedules[1].StartTime)
	assert.Equal(t, "Invalid", entityCourse.Sections[0].Schedules[1].EndTime)
	assert.Equal(t, "", entityCourse.Sections[0].ExamStart)
	assert.Equal(t, "", entityCourse.Sections[0].MidtermStart)
}

func TestToCourseResponse(t *testing.T) {
	start := "09:00"
	end := "12:00"

	examStart := "2026-03-31 13:00:00"
	examEnd := "2026-03-31 16:00:00"

	// Mock UpdatedAt
	updatedAt := time.Date(2026, time.February, 17, 14, 0, 0, 0, time.Local)

	entityCourse := &entity.Course{
		Code:       "CP353004",
		Faculty:    "วิทยาลัยการคอมพิวเตอร์",
		Department: "วิทยาการคอมพิวเตอร์",
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
				// Midterm is zero
			},
		},
	}

	response := ToCourseResponse(entityCourse)

	assert.NotNil(t, response)
	assert.Equal(t, "วิทยาลัยการคอมพิวเตอร์", response.Faculty)
	assert.Equal(t, "วิทยาการคอมพิวเตอร์", response.Department)
	assert.Equal(t, 1, len(response.Sections))
	assert.Equal(t, updatedAt.Format(time.RFC3339), response.UpdatedAt)

	secResp := response.Sections[0]
	assert.Equal(t, "2026-03-31 13:00:00", secResp.ExamStart)
	assert.Equal(t, "2026-03-31 16:00:00", secResp.ExamEnd)
	assert.Empty(t, secResp.MidtermStart)
	assert.Empty(t, secResp.MidtermEnd)

	schedule := secResp.Schedules[0]
	assert.Equal(t, "Wednesday", schedule.Day)
	assert.Equal(t, "09:00", schedule.StartTime)
	assert.Equal(t, "12:00", schedule.EndTime)
}

func TestToCourseResponse_Nil(t *testing.T) {
	assert.Nil(t, ToCourseResponse(nil))
}

func TestToCourseResponse_Midterm(t *testing.T) {
	midtermStart := "2026-02-15 09:00:00"
	midtermEnd := "2026-02-15 12:00:00"

	entityCourse := &entity.Course{
		Sections: []entity.Section{
			{
				MidtermStart: midtermStart,
				MidtermEnd:   midtermEnd,
			},
		},
	}

	response := ToCourseResponse(entityCourse)
	assert.Equal(t, "2026-02-15 09:00:00", response.Sections[0].MidtermStart)
	assert.Equal(t, "2026-02-15 12:00:00", response.Sections[0].MidtermEnd)
}

func TestToCourseResponses(t *testing.T) {
	courses := []*entity.Course{
		{Code: "C1"},
		{Code: "C2"},
	}
	responses := ToCourseResponses(courses)
	assert.Len(t, responses, 2)
	assert.Equal(t, "C1", responses[0].Code)
	assert.Equal(t, "C2", responses[1].Code)
}

func TestToCourseSummaryResponse(t *testing.T) {
	updatedAt := time.Date(2026, time.February, 17, 14, 0, 0, 0, time.Local)
	entityCourse := &entity.Course{
		Code:       "CP353004",
		Faculty:    "วิทยาลัยการคอมพิวเตอร์",
		Department: "วิทยาการคอมพิวเตอร์",
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
	assert.Equal(t, "วิทยาลัยการคอมพิวเตอร์", response.Faculty)
	assert.Equal(t, "วิทยาการคอมพิวเตอร์", response.Department)
	assert.Equal(t, updatedAt.Format(time.RFC3339), response.UpdatedAt)
}

func TestToCourseSummaryResponse_Nil(t *testing.T) {
	assert.Nil(t, ToCourseSummaryResponse(nil))
}

func TestToCourseSummaryResponses(t *testing.T) {
	courses := []*entity.Course{
		{Code: "C1"},
		{Code: "C2"},
	}
	responses := ToCourseSummaryResponses(courses)
	assert.Len(t, responses, 2)
	assert.Equal(t, "C1", responses[0].Code)
	assert.Equal(t, "C2", responses[1].Code)
}

func TestParseThaiExamDate_InvalidFormat(t *testing.T) {
	// Missing " เวลา "
	_, _, err := parseThaiExamDate("31 มี.ค. 2569 13:00 - 16:00")
	assert.Error(t, err)

	// Invalid Date Part
	_, _, err = parseThaiExamDate("31 Invalid 2569 เวลา 13:00 - 16:00")
	assert.Error(t, err)

	// Invalid Date Fields Count
	_, _, err = parseThaiExamDate("31 มี.ค. 2569 Extra เวลา 13:00 - 16:00")
	assert.Error(t, err)

	// Invalid Day
	_, _, err = parseThaiExamDate("Invalid มี.ค. 2569 เวลา 13:00 - 16:00")
	assert.Error(t, err)

	// Invalid Year
	_, _, err = parseThaiExamDate("31 มี.ค. Invalid เวลา 13:00 - 16:00")
	assert.Error(t, err)

	// Invalid Month
	_, _, err = parseThaiExamDate("31 Inval 2569 เวลา 13:00 - 16:00")
	assert.Error(t, err)

	// Invalid Time Range
	_, _, err = parseThaiExamDate("31 มี.ค. 2569 เวลา 13:00")
	assert.Error(t, err)

	// Invalid Start Time
	_, _, err = parseThaiExamDate("31 มี.ค. 2569 เวลา In:va - 16:00")
	assert.Error(t, err)

	// Invalid End Time
	_, _, err = parseThaiExamDate("31 มี.ค. 2569 เวลา 13:00 - In:va")
	assert.Error(t, err)
}

func TestParseThaiMonth(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Month
	}{
		{"ม.ค.", time.January},
		{"ก.พ.", time.February},
		{"มี.ค.", time.March},
		{"เม.ย.", time.April},
		{"พ.ค.", time.May},
		{"มิ.ย.", time.June},
		{"ก.ค.", time.July},
		{"ส.ค.", time.August},
		{"ก.ย.", time.September},
		{"ต.ค.", time.October},
		{"พ.ย.", time.November},
		{"ธ.ค.", time.December},
		{"Invalid", 0},
	}

	for _, test := range tests {
		result := parseThaiMonth(test.input)
		assert.Equal(t, test.expected, result)
	}
}
