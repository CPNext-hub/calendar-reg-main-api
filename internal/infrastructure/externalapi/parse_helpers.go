package externalapi

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// parseThaiExamDate parses strings like "31 มี.ค. 2569 เวลา 13:00 - 16:00"
func parseThaiExamDate(dateStr string) (time.Time, time.Time, error) {
	dateStr = strings.TrimSpace(dateStr)

	parts := strings.Split(dateStr, " เวลา ")
	if len(parts) != 2 {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid format: missing ' เวลา ' separator")
	}

	datePart := strings.TrimSpace(parts[0])
	timePart := strings.TrimSpace(parts[1])

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
	yearCE := yearBE - 543

	month := parseThaiMonth(monthStr)
	if month == 0 {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid month: %s", monthStr)
	}

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
