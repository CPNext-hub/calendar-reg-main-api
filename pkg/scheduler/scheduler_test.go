package scheduler

import (
	"testing"
	"time"

	"github.com/CPNext-hub/calendar-reg-main-api/internal/domain/entity"
	"github.com/CPNext-hub/calendar-reg-main-api/pkg/queue"
)

func newTestScheduler() (*Scheduler, *queue.RefreshQueue) {
	q := queue.New(100, 1)
	s := New(q)
	return s, q
}

// ---- New ----

func TestNew(t *testing.T) {
	q := queue.New(10, 1)
	s := New(q)
	if s == nil {
		t.Fatal("expected non-nil scheduler")
	}
	if s.c == nil {
		t.Error("expected cron instance")
	}
	if s.entries == nil {
		t.Error("expected entries map")
	}
	if s.refreshQueue != q {
		t.Error("expected refresh queue to be set")
	}
}

// ---- Start / Stop ----

func TestStartStop(t *testing.T) {
	s, _ := newTestScheduler()
	s.Start()

	done := make(chan struct{})
	go func() {
		s.Stop()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for Stop")
	}
}

// ---- AddJob ----

func TestAddJob_Enabled(t *testing.T) {
	s, _ := newTestScheduler()
	s.Start()
	defer s.Stop()

	job := &entity.CronJob{
		BaseEntity:  entity.BaseEntity{ID: "job1"},
		Name:        "Test Job",
		CourseCodes: []string{"CS101"},
		Acadyear:    2568,
		Semester:    1,
		CronExpr:    "* * * * *",
		Enabled:     true,
	}

	err := s.AddJob(job)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	s.mu.Lock()
	_, ok := s.entries["job1"]
	s.mu.Unlock()
	if !ok {
		t.Error("expected job to be registered in entries")
	}
}

func TestAddJob_Disabled(t *testing.T) {
	s, _ := newTestScheduler()

	job := &entity.CronJob{
		BaseEntity: entity.BaseEntity{ID: "job1"},
		Name:       "Disabled Job",
		CronExpr:   "* * * * *",
		Enabled:    false,
	}

	err := s.AddJob(job)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	s.mu.Lock()
	_, ok := s.entries["job1"]
	s.mu.Unlock()
	if ok {
		t.Error("disabled job should not be in entries")
	}
}

func TestAddJob_ReplaceExisting(t *testing.T) {
	s, _ := newTestScheduler()
	s.Start()
	defer s.Stop()

	job := &entity.CronJob{
		BaseEntity:  entity.BaseEntity{ID: "job1"},
		Name:        "Original",
		CourseCodes: []string{"CS101"},
		CronExpr:    "* * * * *",
		Enabled:     true,
	}
	s.AddJob(job)

	s.mu.Lock()
	oldEntryID := s.entries["job1"]
	s.mu.Unlock()

	// Replace with new cron expr
	job2 := &entity.CronJob{
		BaseEntity:  entity.BaseEntity{ID: "job1"},
		Name:        "Replaced",
		CourseCodes: []string{"CS102"},
		CronExpr:    "*/5 * * * *",
		Enabled:     true,
	}
	err := s.AddJob(job2)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	s.mu.Lock()
	newEntryID := s.entries["job1"]
	s.mu.Unlock()

	if newEntryID == oldEntryID {
		t.Error("expected entry ID to change after replace")
	}
}

func TestAddJob_DisableExisting(t *testing.T) {
	s, _ := newTestScheduler()
	s.Start()
	defer s.Stop()

	job := &entity.CronJob{
		BaseEntity: entity.BaseEntity{ID: "job1"},
		Name:       "Active",
		CronExpr:   "* * * * *",
		Enabled:    true,
	}
	s.AddJob(job)

	// Disable it
	job.Enabled = false
	s.AddJob(job)

	s.mu.Lock()
	_, ok := s.entries["job1"]
	s.mu.Unlock()
	if ok {
		t.Error("expected disabled job to be removed from entries")
	}
}

func TestAddJob_InvalidCronExpr(t *testing.T) {
	s, _ := newTestScheduler()

	job := &entity.CronJob{
		BaseEntity: entity.BaseEntity{ID: "job1"},
		Name:       "Bad Cron",
		CronExpr:   "invalid cron",
		Enabled:    true,
	}

	err := s.AddJob(job)
	if err == nil {
		t.Fatal("expected error for invalid cron expression")
	}
}

// ---- RemoveJob ----

func TestRemoveJob_Existing(t *testing.T) {
	s, _ := newTestScheduler()
	s.Start()
	defer s.Stop()

	job := &entity.CronJob{
		BaseEntity: entity.BaseEntity{ID: "job1"},
		Name:       "To Remove",
		CronExpr:   "* * * * *",
		Enabled:    true,
	}
	s.AddJob(job)

	s.RemoveJob("job1")

	s.mu.Lock()
	_, ok := s.entries["job1"]
	s.mu.Unlock()
	if ok {
		t.Error("expected job to be removed")
	}
}

func TestRemoveJob_NonExistent(t *testing.T) {
	s, _ := newTestScheduler()
	// Should not panic
	s.RemoveJob("nonexistent")
}

// ---- LoadJobs ----

func TestLoadJobs(t *testing.T) {
	s, _ := newTestScheduler()
	s.Start()
	defer s.Stop()

	jobs := []*entity.CronJob{
		{BaseEntity: entity.BaseEntity{ID: "j1"}, Name: "Job 1", CronExpr: "* * * * *", Enabled: true},
		{BaseEntity: entity.BaseEntity{ID: "j2"}, Name: "Job 2", CronExpr: "*/5 * * * *", Enabled: true},
		{BaseEntity: entity.BaseEntity{ID: "j3"}, Name: "Job 3", CronExpr: "* * * * *", Enabled: false},
	}

	s.LoadJobs(jobs)

	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.entries) != 2 {
		t.Errorf("expected 2 entries (disabled excluded), got %d", len(s.entries))
	}
}

func TestLoadJobs_InvalidCron(t *testing.T) {
	s, _ := newTestScheduler()

	jobs := []*entity.CronJob{
		{BaseEntity: entity.BaseEntity{ID: "j1"}, Name: "Good", CronExpr: "* * * * *", Enabled: true},
		{BaseEntity: entity.BaseEntity{ID: "j2"}, Name: "Bad", CronExpr: "bad", Enabled: true},
	}

	// Should not panic, should log error and continue
	s.LoadJobs(jobs)

	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.entries) != 1 {
		t.Errorf("expected 1 valid entry, got %d", len(s.entries))
	}
}

// ---- TriggerJob ----

func TestTriggerJob(t *testing.T) {
	s, q := newTestScheduler()

	job := &entity.CronJob{
		BaseEntity:  entity.BaseEntity{ID: "job1"},
		Name:        "Trigger Me",
		CourseCodes: []string{"CS101", "CS102"},
		Acadyear:    2568,
		Semester:    1,
	}

	s.TriggerJob(job)

	st := q.Status()
	if st.Processing != 2 {
		t.Errorf("expected 2 jobs enqueued, got processing=%d", st.Processing)
	}
}

// ---- makeHandler ----

func TestMakeHandler(t *testing.T) {
	s, q := newTestScheduler()

	job := &entity.CronJob{
		BaseEntity:  entity.BaseEntity{ID: "job1"},
		Name:        "Handler Job",
		CourseCodes: []string{"A", "B", "C"},
		Acadyear:    2568,
		Semester:    2,
	}

	handler := s.makeHandler(job)
	handler() // invoke manually

	st := q.Status()
	if st.Processing != 3 {
		t.Errorf("expected 3 jobs enqueued by handler, got processing=%d", st.Processing)
	}
}

// ---- ValidateCronExpr ----

func TestValidateCronExpr_Valid(t *testing.T) {
	tests := []string{
		"* * * * *",
		"0 */6 * * *",
		"30 2 * * 1-5",
		"0 0 1 * *",
	}

	for _, expr := range tests {
		if err := ValidateCronExpr(expr); err != nil {
			t.Errorf("expected %q to be valid, got: %v", expr, err)
		}
	}
}

func TestValidateCronExpr_Invalid(t *testing.T) {
	tests := []string{
		"invalid",
		"* * *",
		"60 * * * *",
		"",
	}

	for _, expr := range tests {
		if err := ValidateCronExpr(expr); err == nil {
			t.Errorf("expected %q to be invalid", expr)
		}
	}
}
