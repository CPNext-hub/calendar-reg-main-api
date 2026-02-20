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

	expectedStart := "2026-03-31 13:00:00"
	expectedEnd := "2026-03-31 16:00:00"

	if start != expectedStart {
		t.Errorf("expected start %s, got %s", expectedStart, start)
	}
	if end != expectedEnd {
		t.Errorf("expected end %s, got %s", expectedEnd, end)
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
	expectedStart := "2027-02-15 09:00:00"
	expectedEnd := "2027-02-15 12:00:00"

	if start != expectedStart {
		t.Errorf("expected start %s, got %s", expectedStart, start)
	}
	if end != expectedEnd {
		t.Errorf("expected end %s, got %s", expectedEnd, end)
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
