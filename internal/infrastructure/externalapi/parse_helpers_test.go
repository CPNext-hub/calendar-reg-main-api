package externalapi

import (
	"testing"
	"time"
)

// ---- parseThaiExamDate tests ----

func TestParseThaiExamDate_ValidDate(t *testing.T) {
	start, end, err := parseThaiExamDate("31 มี.ค. 2569 เวลา 13:00 - 16:00")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 2569 BE = 2026 CE
	if start.Year() != 2026 {
		t.Errorf("expected year 2026, got %d", start.Year())
	}
	if start.Month() != time.March {
		t.Errorf("expected March, got %v", start.Month())
	}
	if start.Day() != 31 {
		t.Errorf("expected day 31, got %d", start.Day())
	}
	if start.Hour() != 13 || start.Minute() != 0 {
		t.Errorf("expected start 13:00, got %02d:%02d", start.Hour(), start.Minute())
	}
	if end.Hour() != 16 || end.Minute() != 0 {
		t.Errorf("expected end 16:00, got %02d:%02d", end.Hour(), end.Minute())
	}
	if end.Year() != start.Year() || end.Month() != start.Month() || end.Day() != start.Day() {
		t.Error("expected start and end to be on the same date")
	}
}

func TestParseThaiExamDate_MissingSeparator(t *testing.T) {
	_, _, err := parseThaiExamDate("31 มี.ค. 2569 13:00 - 16:00")
	if err == nil {
		t.Fatal("expected error for missing ' เวลา ' separator")
	}
}

func TestParseThaiExamDate_InvalidDay(t *testing.T) {
	_, _, err := parseThaiExamDate("XX มี.ค. 2569 เวลา 13:00 - 16:00")
	if err == nil {
		t.Fatal("expected error for invalid day")
	}
}

func TestParseThaiExamDate_InvalidYear(t *testing.T) {
	_, _, err := parseThaiExamDate("31 มี.ค. YYYY เวลา 13:00 - 16:00")
	if err == nil {
		t.Fatal("expected error for invalid year")
	}
}

func TestParseThaiExamDate_InvalidMonth(t *testing.T) {
	_, _, err := parseThaiExamDate("31 zzz. 2569 เวลา 13:00 - 16:00")
	if err == nil {
		t.Fatal("expected error for invalid month")
	}
}

func TestParseThaiExamDate_InvalidTimeRange(t *testing.T) {
	_, _, err := parseThaiExamDate("31 มี.ค. 2569 เวลา 13:00")
	if err == nil {
		t.Fatal("expected error for missing time range separator")
	}
}

func TestParseThaiExamDate_InvalidStartTime(t *testing.T) {
	_, _, err := parseThaiExamDate("31 มี.ค. 2569 เวลา XX:XX - 16:00")
	if err == nil {
		t.Fatal("expected error for invalid start time")
	}
}

func TestParseThaiExamDate_InvalidEndTime(t *testing.T) {
	_, _, err := parseThaiExamDate("31 มี.ค. 2569 เวลา 13:00 - XX:XX")
	if err == nil {
		t.Fatal("expected error for invalid end time")
	}
}

func TestParseThaiExamDate_InvalidDateFormat(t *testing.T) {
	// Only 2 fields instead of 3 (day month year)
	_, _, err := parseThaiExamDate("31 2569 เวลา 13:00 - 16:00")
	if err == nil {
		t.Fatal("expected error for invalid date format (less than 3 fields)")
	}
}

func TestParseThaiExamDate_WithWhitespace(t *testing.T) {
	start, end, err := parseThaiExamDate("  15 ก.พ. 2570 เวลา 09:00 - 12:00  ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 2570 BE = 2027 CE
	if start.Year() != 2027 || start.Month() != time.February || start.Day() != 15 {
		t.Errorf("expected 2027-02-15, got %v", start)
	}
	if end.Hour() != 12 {
		t.Errorf("expected end hour 12, got %d", end.Hour())
	}
}

// ---- parseThaiMonth tests ----

func TestParseThaiMonth_AllMonths(t *testing.T) {
	tests := []struct {
		abbr     string
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
	}

	for _, tc := range tests {
		got := parseThaiMonth(tc.abbr)
		if got != tc.expected {
			t.Errorf("parseThaiMonth(%q) = %v, want %v", tc.abbr, got, tc.expected)
		}
	}
}

func TestParseThaiMonth_Unknown(t *testing.T) {
	got := parseThaiMonth("unknown")
	if got != 0 {
		t.Errorf("expected 0 for unknown month, got %v", got)
	}
}
