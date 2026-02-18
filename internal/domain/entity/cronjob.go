package entity

// CronJob represents a scheduled job that refreshes course data.
type CronJob struct {
	BaseEntity
	Name        string   // human-readable label, e.g. "Refresh CP courses"
	CourseCodes []string // subject codes to refresh, e.g. ["CP353004", "SC313002"]
	Acadyear    int      // academic year, e.g. 2568
	Semester    int      // semester, e.g. 2
	CronExpr    string   // cron expression, e.g. "0 */6 * * *" (every 6 hours)
	Enabled     bool     // toggle on/off
}
