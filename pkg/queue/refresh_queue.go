package queue

import (
	"log"
	"sync"
	"sync/atomic"
	"time"
)

// RefreshJob represents a background course refresh job.
type RefreshJob struct {
	Code      string
	IsNew     bool // true = first fetch (Create), false = stale refresh (Update)
	EnqueueAt time.Time
	Result    chan<- JobResult // optional: caller can wait for the result
}

// JobResult holds the outcome of a processed refresh job.
type JobResult struct {
	Data interface{} // the refreshed entity (caller must type-assert)
	Err  error
}

// QueueStatus holds snapshot stats about the queue.
type QueueStatus struct {
	Workers    int      `json:"workers"`
	Pending    int      `json:"pending"`
	Processing int      `json:"processing"`
	Processed  int64    `json:"processed"`
	Codes      []string `json:"codes"`
}

// RefreshQueue is a bounded worker pool that processes background refresh jobs.
type RefreshQueue struct {
	jobs      chan RefreshJob
	mu        sync.Mutex
	inflight  map[string]bool
	wg        sync.WaitGroup
	processed atomic.Int64
	workers   int
}

// New creates a new RefreshQueue with the given buffer size and worker count.
func New(bufferSize, workers int) *RefreshQueue {
	if workers < 1 {
		workers = 1
	}
	return &RefreshQueue{
		jobs:     make(chan RefreshJob, bufferSize),
		inflight: make(map[string]bool),
		workers:  workers,
	}
}

// Enqueue tries to register a refresh for the given code.
// Returns false if the code is already being refreshed (dedup) or the queue is full.
func (q *RefreshQueue) Enqueue(job RefreshJob) bool {
	q.mu.Lock()
	if q.inflight[job.Code] {
		q.mu.Unlock()
		log.Printf("[queue] refresh already in progress for %s, skipping", job.Code)
		return false
	}
	q.inflight[job.Code] = true
	q.mu.Unlock()

	job.EnqueueAt = time.Now()

	select {
	case q.jobs <- job:
		log.Printf("[queue] enqueued refresh for %s", job.Code)
		return true
	default:
		// Queue full — remove from inflight and reject.
		q.mu.Lock()
		delete(q.inflight, job.Code)
		q.mu.Unlock()
		log.Printf("[queue] queue full, dropped refresh for %s", job.Code)
		return false
	}
}

// MarkDone removes a code from the inflight set and increments processed counter.
func (q *RefreshQueue) MarkDone(code string) {
	q.mu.Lock()
	delete(q.inflight, code)
	q.mu.Unlock()
	q.processed.Add(1)
}

// Start spawns worker goroutines that consume jobs from the queue.
func (q *RefreshQueue) Start(handler func(RefreshJob)) {
	for i := 0; i < q.workers; i++ {
		q.wg.Add(1)
		go func(id int) {
			defer q.wg.Done()
			log.Printf("[queue] worker %d started", id)
			for job := range q.jobs {
				handler(job)
			}
			log.Printf("[queue] worker %d stopped", id)
		}(i)
	}
	log.Printf("[queue] started %d workers (buffer=%d)", q.workers, cap(q.jobs))
}

// Stop closes the jobs channel and waits for all workers to finish.
func (q *RefreshQueue) Stop() {
	log.Println("[queue] stopping — waiting for workers to drain...")
	close(q.jobs)
	q.wg.Wait()
	log.Println("[queue] all workers stopped")
}

// Status returns a snapshot of the current queue state.
func (q *RefreshQueue) Status() QueueStatus {
	q.mu.Lock()
	codes := make([]string, 0, len(q.inflight))
	for code := range q.inflight {
		codes = append(codes, code)
	}
	processing := len(q.inflight)
	q.mu.Unlock()

	return QueueStatus{
		Workers:    q.workers,
		Pending:    len(q.jobs),
		Processing: processing,
		Processed:  q.processed.Load(),
		Codes:      codes,
	}
}
