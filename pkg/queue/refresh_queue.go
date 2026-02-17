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
	EnqueueAt time.Time
}

// QueueStatus holds snapshot stats about the queue.
type QueueStatus struct {
	Processing int      `json:"processing"`
	Processed  int64    `json:"processed"`
	Codes      []string `json:"codes"`
}

// RefreshQueue tracks in-flight background refresh goroutines.
type RefreshQueue struct {
	mu        sync.Mutex
	inflight  map[string]bool
	processed atomic.Int64
}

// New creates a new RefreshQueue.
func New(bufferSize int) *RefreshQueue {
	return &RefreshQueue{
		inflight: make(map[string]bool),
	}
}

// Enqueue tries to register a refresh for the given code.
// Returns false if the code is already being refreshed (dedup).
func (q *RefreshQueue) Enqueue(job RefreshJob) bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.inflight[job.Code] {
		log.Printf("[queue] refresh already in progress for %s, skipping", job.Code)
		return false
	}

	q.inflight[job.Code] = true
	log.Printf("[queue] background refresh in progress for %s", job.Code)
	return true
}

// MarkDone removes a code from the inflight set and increments processed.
func (q *RefreshQueue) MarkDone(code string) {
	q.mu.Lock()
	delete(q.inflight, code)
	q.mu.Unlock()
	q.processed.Add(1)
}

// Status returns a snapshot of the current queue state.
// Processing count is derived from the inflight map — single source of truth.
func (q *RefreshQueue) Status() QueueStatus {
	q.mu.Lock()
	codes := make([]string, 0, len(q.inflight))
	for code := range q.inflight {
		codes = append(codes, code)
	}
	processing := len(q.inflight)
	q.mu.Unlock()

	return QueueStatus{
		Processing: processing,
		Processed:  q.processed.Load(),
		Codes:      codes,
	}
}

// Start is a no-op — goroutines handle the work directly.
func (q *RefreshQueue) Start(handler func(RefreshJob)) {
	log.Println("[queue] refresh tracker started")
}

// Stop is a no-op — goroutines finish on their own.
func (q *RefreshQueue) Stop() {
	log.Println("[queue] refresh tracker stopped")
}
