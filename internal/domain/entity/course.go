package entity

// Course represents a university course.
type Course struct {
	BaseEntity
	Code         string    // e.g., "CP353004"
	NameEN       string    // e.g., "Software Engineering"
	NameTH       string    // e.g., "วิศวกรรมซอฟต์แวร์"
	Faculty      string    // e.g., "วิทยาลัยการคอมพิวเตอร์, วิทยาการคอมพิวเตอร์"
	Credits      string    // e.g., "3 (2-2-5)"
	Prerequisite string    // e.g., "CP353002 หรือ SC313002"
	Semester     int       // e.g., 2
	Year         int       // e.g., 2568
	Program      string    // e.g., "ปริญญาตรี ภาคปกติ"
	Sections     []Section // multiple sections per course
}

// Section represents a course section with schedule and instructor info.
type Section struct {
	Number     string     // e.g., "01"
	Schedules  []Schedule // multiple schedule slots per section
	Seats      int        // e.g., 40
	Instructor string     // e.g., "ผศ.ดร.ชิตสุธา สุ่มเล็ก"
	ExamDate   string     // e.g., "31 มี.ค. 2569 เวลา 13:00 - 16:00 ..."
}

// Schedule represents a single class meeting (day + time + room).
type Schedule struct {
	Day  string // e.g., "จันทร์"
	Time string // e.g., "13:00-15:00"
	Room string // e.g., "CP9 CP9127"
	Type string // e.g., "C" (lecture)
}
