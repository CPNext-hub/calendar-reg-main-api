package entity

import (
	"fmt"
	"time"
)

// Course represents a university course.
type Course struct {
	BaseEntity
	Code         string    // e.g., "CP353004"
	NameEN       string    // e.g., "Software Engineering"
	NameTH       string    // e.g., "วิศวกรรมซอฟต์แวร์"
	Faculty      string    // e.g., "วิทยาลัยการคอมพิวเตอร์"
	Department   string    // e.g., "วิทยาการคอมพิวเตอร์"
	Credits      string    // e.g., "3 (2-2-5)"
	Prerequisite string    // e.g., "CP353002 หรือ SC313002"
	Semester     int       // e.g., 2
	Year         int       // e.g., 2568
	Sections     []Section // multiple sections per course
}

// Key returns the composite lookup key: "code:year:semester".
func (c *Course) Key() string {
	return fmt.Sprintf("%s:%d:%d", c.Code, c.Year, c.Semester)
}

// Section represents a course section with schedule and instructor info.
type Section struct {
	Number       string     // e.g., "01"
	Schedules    []Schedule // multiple schedule slots per section
	Seats        int        // e.g., 40
	Instructor   string     // e.g., "ผศ.ดร.ชิตสุธา สุ่มเล็ก"
	ExamStart    time.Time  // e.g., 2026-03-31 13:00:00
	ExamEnd      time.Time  // e.g., 2026-03-31 16:00:00
	MidtermStart time.Time  // สอบกลางภาค start
	MidtermEnd   time.Time  // สอบกลางภาค end
	Note         string     // หมายเหตุ e.g., "ผู้สอบไม่ผ่าน", "Closed"
	ReservedFor  []string   // สำรองสำหรับ e.g., ["ผู้ที่สอบไม่ผ่าน 50-49-1"]
	Campus       string     // e.g., "ขอนแก่น", "หนองคาย"
	Program      string     // e.g., "ปริญญาตรี ภาคปกติ"
}

// Schedule represents a single class meeting (day + time + room).
type Schedule struct {
	Day       string    // e.g., "จันทร์"
	StartTime time.Time // e.g., 13:00 parsed from "13:00-15:00"
	EndTime   time.Time // e.g., 15:00 parsed from "13:00-15:00"
	Room      string    // e.g., "CP9 CP9127"
	Type      string    // e.g., "C" (lecture)
}
