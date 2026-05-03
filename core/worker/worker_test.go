package worker

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestRunnerProcessesJobs(t *testing.T) {
	ctx := context.Background()
	processed := make(chan string, 1)

	r := NewRunner(8, 1, func(_ context.Context, job Job) error {
		processed <- job.Type
		return nil
	})
	r.Start(ctx)
	t.Cleanup(r.Shutdown)

	if err := r.Enqueue(Job{Type: "cleanup"}); err != nil {
		t.Fatalf("enqueue: %v", err)
	}

	select {
	case got := <-processed:
		if got != "cleanup" {
			t.Fatalf("expected job type cleanup, got %q", got)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for job processing")
	}
}

func TestRunnerRetriesUntilSuccess(t *testing.T) {
	ctx := context.Background()
	var attempts int32
	completed := make(chan struct{}, 1)

	r := NewRunner(8, 1, func(_ context.Context, job Job) error {
		current := atomic.AddInt32(&attempts, 1)
		if current < 2 {
			return errors.New("transient error")
		}
		completed <- struct{}{}
		return nil
	})
	r.Start(ctx)
	t.Cleanup(r.Shutdown)

	if err := r.Enqueue(Job{Type: "webhook.dispatch", MaxAttempts: 3}); err != nil {
		t.Fatalf("enqueue: %v", err)
	}

	select {
	case <-completed:
		if atomic.LoadInt32(&attempts) != 2 {
			t.Fatalf("expected 2 attempts, got %d", atomic.LoadInt32(&attempts))
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for retried job success")
	}
}

func TestRunnerQueueFull(t *testing.T) {
	r := NewRunner(1, 1, func(_ context.Context, _ Job) error { return nil })

	if err := r.Enqueue(Job{Type: "job-1"}); err != nil {
		t.Fatalf("enqueue first job: %v", err)
	}
	if err := r.Enqueue(Job{Type: "job-2"}); !errors.Is(err, ErrQueueFull) {
		t.Fatalf("expected ErrQueueFull, got %v", err)
	}
}
