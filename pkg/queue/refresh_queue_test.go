package queue

import (
	"sync"
	"testing"
	"time"
)

// ---- RefreshJob tests ----

func TestRefreshJob_Key(t *testing.T) {
	job := RefreshJob{Code: "CS101", Acadyear: 2568, Semester: 1}
	expected := "CS101:2568:1"
	if got := job.Key(); got != expected {
		t.Errorf("expected key %q, got %q", expected, got)
	}
}

// ---- New tests ----

func TestNew_DefaultWorkers(t *testing.T) {
	q := New(10, 0)
	if q.workers != 1 {
		t.Errorf("expected workers=1 when given 0, got %d", q.workers)
	}
}

func TestNew_NegativeWorkers(t *testing.T) {
	q := New(10, -5)
	if q.workers != 1 {
		t.Errorf("expected workers=1 when given -5, got %d", q.workers)
	}
}

func TestNew_ValidWorkers(t *testing.T) {
	q := New(10, 3)
	if q.workers != 3 {
		t.Errorf("expected workers=3, got %d", q.workers)
	}
}

// ---- Enqueue tests ----

func TestEnqueue_Success(t *testing.T) {
	q := New(10, 1)
	job := RefreshJob{Code: "CS101", Acadyear: 2568, Semester: 1}

	ok := q.Enqueue(job)
	if !ok {
		t.Error("expected enqueue to succeed")
	}

	// Key should be inflight
	q.mu.Lock()
	if !q.inflight[job.Key()] {
		t.Error("expected key to be inflight")
	}
	q.mu.Unlock()
}

func TestEnqueue_Duplicate(t *testing.T) {
	q := New(10, 1)
	job := RefreshJob{Code: "CS101", Acadyear: 2568, Semester: 1}

	q.Enqueue(job)
	ok := q.Enqueue(job) // same key
	if ok {
		t.Error("expected duplicate enqueue to be rejected")
	}
}

func TestEnqueue_QueueFull(t *testing.T) {
	q := New(0, 1) // buffer size 0 = no room
	job := RefreshJob{Code: "CS101", Acadyear: 2568, Semester: 1}

	ok := q.Enqueue(job)
	if ok {
		t.Error("expected enqueue to fail when queue is full")
	}

	// Key should NOT be inflight after rejection
	q.mu.Lock()
	if q.inflight[job.Key()] {
		t.Error("expected key to be removed from inflight after full queue")
	}
	q.mu.Unlock()
}

func TestEnqueue_SetsEnqueueAt(t *testing.T) {
	q := New(10, 1)
	before := time.Now()

	job := RefreshJob{Code: "CS101", Acadyear: 2568, Semester: 1}
	q.Enqueue(job)

	// Read the job back from channel to check EnqueueAt
	received := <-q.jobs
	if received.EnqueueAt.Before(before) {
		t.Error("expected EnqueueAt to be set to current time")
	}
}

// ---- MarkDone tests ----

func TestMarkDone(t *testing.T) {
	q := New(10, 1)
	job := RefreshJob{Code: "CS101", Acadyear: 2568, Semester: 1}
	q.Enqueue(job)

	q.MarkDone(job.Key())

	// Should not be inflight anymore
	q.mu.Lock()
	if q.inflight[job.Key()] {
		t.Error("expected key to be removed from inflight after MarkDone")
	}
	q.mu.Unlock()

	// Processed counter should be 1
	if q.processed.Load() != 1 {
		t.Errorf("expected processed=1, got %d", q.processed.Load())
	}
}

// ---- Start / Stop tests ----

func TestStartStop(t *testing.T) {
	q := New(10, 2)

	processed := make(chan string, 10)
	q.Start(func(job RefreshJob) {
		processed <- job.Key()
	})

	q.Enqueue(RefreshJob{Code: "A", Acadyear: 2568, Semester: 1})
	q.Enqueue(RefreshJob{Code: "B", Acadyear: 2568, Semester: 1})

	// Wait for processing
	for i := 0; i < 2; i++ {
		select {
		case <-processed:
		case <-time.After(2 * time.Second):
			t.Fatal("timed out waiting for job processing")
		}
	}

	q.Stop()
}

func TestStart_WorkerProcessesAllJobs(t *testing.T) {
	q := New(10, 1)

	var mu sync.Mutex
	results := map[string]bool{}

	q.Start(func(job RefreshJob) {
		mu.Lock()
		results[job.Key()] = true
		mu.Unlock()
		q.MarkDone(job.Key())
	})

	q.Enqueue(RefreshJob{Code: "X", Acadyear: 2568, Semester: 1})
	q.Enqueue(RefreshJob{Code: "Y", Acadyear: 2568, Semester: 2})
	q.Enqueue(RefreshJob{Code: "Z", Acadyear: 2568, Semester: 3})

	// Allow time for workers
	time.Sleep(200 * time.Millisecond)
	q.Stop()

	mu.Lock()
	defer mu.Unlock()

	if len(results) != 3 {
		t.Errorf("expected 3 results, got %d", len(results))
	}
	if q.processed.Load() != 3 {
		t.Errorf("expected processed=3, got %d", q.processed.Load())
	}
}

// ---- Status tests ----

func TestStatus_Empty(t *testing.T) {
	q := New(10, 2)

	st := q.Status()
	if st.Workers != 2 {
		t.Errorf("expected workers=2, got %d", st.Workers)
	}
	if st.Pending != 0 {
		t.Errorf("expected pending=0, got %d", st.Pending)
	}
	if st.Processing != 0 {
		t.Errorf("expected processing=0, got %d", st.Processing)
	}
	if st.Processed != 0 {
		t.Errorf("expected processed=0, got %d", st.Processed)
	}
	if len(st.Codes) != 0 {
		t.Errorf("expected no codes, got %v", st.Codes)
	}
}

func TestStatus_WithInflight(t *testing.T) {
	q := New(10, 1)
	q.Enqueue(RefreshJob{Code: "CS101", Acadyear: 2568, Semester: 1})
	q.Enqueue(RefreshJob{Code: "CS102", Acadyear: 2568, Semester: 2})

	st := q.Status()
	if st.Pending != 2 {
		t.Errorf("expected pending=2, got %d", st.Pending)
	}
	if st.Processing != 2 {
		t.Errorf("expected processing=2 (inflight), got %d", st.Processing)
	}
	if len(st.Codes) != 2 {
		t.Errorf("expected 2 codes, got %d", len(st.Codes))
	}
}

func TestStatus_AfterProcessing(t *testing.T) {
	q := New(10, 1)
	q.Enqueue(RefreshJob{Code: "CS101", Acadyear: 2568, Semester: 1})
	q.MarkDone("CS101:2568:1")

	st := q.Status()
	if st.Processing != 0 {
		t.Errorf("expected processing=0 after MarkDone, got %d", st.Processing)
	}
	if st.Processed != 1 {
		t.Errorf("expected processed=1, got %d", st.Processed)
	}
}

// ---- Enqueue after MarkDone (re-enqueue) ----

func TestEnqueue_AfterMarkDone(t *testing.T) {
	q := New(10, 1)
	job := RefreshJob{Code: "CS101", Acadyear: 2568, Semester: 1}

	q.Enqueue(job)
	q.MarkDone(job.Key())

	// Should be able to enqueue same key again
	ok := q.Enqueue(job)
	if !ok {
		t.Error("expected re-enqueue after MarkDone to succeed")
	}
}

// ---- JobResult channel ----

func TestEnqueue_WithResultChannel(t *testing.T) {
	q := New(10, 1)

	resultCh := make(chan JobResult, 1)
	job := RefreshJob{Code: "CS101", Acadyear: 2568, Semester: 1, Result: resultCh}

	q.Start(func(j RefreshJob) {
		j.Result <- JobResult{Data: "done", Err: nil}
		q.MarkDone(j.Key())
	})

	q.Enqueue(job)

	select {
	case res := <-resultCh:
		if res.Err != nil {
			t.Errorf("expected no error, got %v", res.Err)
		}
		if res.Data != "done" {
			t.Errorf("expected data 'done', got %v", res.Data)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for result")
	}

	q.Stop()
}
