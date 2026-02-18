package scheduler

import (
	"log"
	"sync"

	"github.com/CPNext-hub/calendar-reg-main-api/internal/domain/entity"
	"github.com/CPNext-hub/calendar-reg-main-api/pkg/queue"
	"github.com/robfig/cron/v3"
)

// Scheduler wraps robfig/cron and manages dynamic cron job registration.
type Scheduler struct {
	c            *cron.Cron
	mu           sync.Mutex
	entries      map[string]cron.EntryID // jobID → cron entryID
	refreshQueue *queue.RefreshQueue
}

// New creates a new Scheduler.
func New(refreshQueue *queue.RefreshQueue) *Scheduler {
	return &Scheduler{
		c:            cron.New(),
		entries:      make(map[string]cron.EntryID),
		refreshQueue: refreshQueue,
	}
}

// Start begins the cron scheduler.
func (s *Scheduler) Start() {
	s.c.Start()
	log.Println("[scheduler] started")
}

// Stop gracefully stops the cron scheduler and waits for running jobs.
func (s *Scheduler) Stop() {
	ctx := s.c.Stop()
	<-ctx.Done()
	log.Println("[scheduler] stopped")
}

// AddJob registers or replaces a cron job in the scheduler.
// If the job is disabled, it will be removed from the scheduler instead.
func (s *Scheduler) AddJob(job *entity.CronJob) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Remove existing entry if present.
	if entryID, ok := s.entries[job.ID]; ok {
		s.c.Remove(entryID)
		delete(s.entries, job.ID)
	}

	// If disabled, just remove (already done above).
	if !job.Enabled {
		log.Printf("[scheduler] job %s (%s) disabled, removed from scheduler", job.ID, job.Name)
		return nil
	}

	// Register the cron entry.
	entryID, err := s.c.AddFunc(job.CronExpr, s.makeHandler(job))
	if err != nil {
		return err
	}

	s.entries[job.ID] = entryID
	log.Printf("[scheduler] job %s (%s) registered with cron expr: %s", job.ID, job.Name, job.CronExpr)
	return nil
}

// RemoveJob removes a cron job from the scheduler.
func (s *Scheduler) RemoveJob(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if entryID, ok := s.entries[id]; ok {
		s.c.Remove(entryID)
		delete(s.entries, id)
		log.Printf("[scheduler] job %s removed from scheduler", id)
	}
}

// LoadJobs registers a batch of cron jobs (used at startup).
func (s *Scheduler) LoadJobs(jobs []*entity.CronJob) {
	for _, job := range jobs {
		if err := s.AddJob(job); err != nil {
			log.Printf("[scheduler] failed to load job %s (%s): %v", job.ID, job.Name, err)
		}
	}
	log.Printf("[scheduler] loaded %d jobs", len(jobs))
}

// TriggerJob immediately enqueues all course codes for the given job.
func (s *Scheduler) TriggerJob(job *entity.CronJob) {
	log.Printf("[scheduler] manually triggering job %s (%s)", job.ID, job.Name)
	s.enqueueCourseCodes(job)
}

// makeHandler creates the function called by cron for a specific job.
func (s *Scheduler) makeHandler(job *entity.CronJob) func() {
	// Capture job data at registration time.
	jobID := job.ID
	jobName := job.Name
	codes := make([]string, len(job.CourseCodes))
	copy(codes, job.CourseCodes)
	acadyear := job.Acadyear
	semester := job.Semester

	return func() {
		log.Printf("[scheduler] executing job %s (%s): refreshing %d courses", jobID, jobName, len(codes))
		for _, code := range codes {
			s.refreshQueue.Enqueue(queue.RefreshJob{
				Code:     code,
				Acadyear: acadyear,
				Semester: semester,
				IsNew:    false,
			})
		}
	}
}

// enqueueCourseCodes enqueues all course codes for a job.
func (s *Scheduler) enqueueCourseCodes(job *entity.CronJob) {
	for _, code := range job.CourseCodes {
		s.refreshQueue.Enqueue(queue.RefreshJob{
			Code:     code,
			Acadyear: job.Acadyear,
			Semester: job.Semester,
			IsNew:    false,
		})
	}
}

// ValidateCronExpr validates a cron expression.
func ValidateCronExpr(expr string) error {
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	_, err := parser.Parse(expr)
	return err
}
