package worker

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
)

var ErrQueueFull = errors.New("worker queue is full")

type Job struct {
	Type        string
	Payload     interface{}
	Attempt     int
	MaxAttempts int
}

type Processor func(ctx context.Context, job Job) error

type Runner struct {
	processor Processor
	queue     chan Job

	workerCount int
	cancel      context.CancelFunc
	wg          sync.WaitGroup
}

func NewRunner(bufferSize int, workerCount int, processor Processor) *Runner {
	if bufferSize <= 0 {
		bufferSize = 64
	}
	if workerCount <= 0 {
		workerCount = 1
	}

	return &Runner{
		processor:   processor,
		queue:       make(chan Job, bufferSize),
		workerCount: workerCount,
	}
}

func (r *Runner) Start(ctx context.Context) {
	workerCtx, cancel := context.WithCancel(ctx)
	r.cancel = cancel

	for i := 0; i < r.workerCount; i++ {
		r.wg.Add(1)
		go func(workerID int) {
			defer r.wg.Done()
			r.runLoop(workerCtx, workerID)
		}(i + 1)
	}
}

func (r *Runner) runLoop(ctx context.Context, workerID int) {
	for {
		select {
		case <-ctx.Done():
			return
		case job, ok := <-r.queue:
			if !ok {
				return
			}

			if job.MaxAttempts <= 0 {
				job.MaxAttempts = 1
			}
			if job.Attempt <= 0 {
				job.Attempt = 1
			}

			err := r.processor(ctx, job)
			if err == nil {
				continue
			}

			if job.Attempt < job.MaxAttempts {
				job.Attempt++
				if enqueueErr := r.Enqueue(job); enqueueErr != nil {
					log.Printf("worker[%d] failed to retry job type=%s: %v", workerID, job.Type, enqueueErr)
				}
				continue
			}

			log.Printf("worker[%d] job failed type=%s attempts=%d err=%v", workerID, job.Type, job.Attempt, err)
		}
	}
}

func (r *Runner) Enqueue(job Job) error {
	if job.Type == "" {
		return fmt.Errorf("job type is required")
	}

	if job.MaxAttempts <= 0 {
		job.MaxAttempts = 1
	}
	if job.Attempt <= 0 {
		job.Attempt = 1
	}

	select {
	case r.queue <- job:
		return nil
	default:
		return ErrQueueFull
	}
}

func (r *Runner) Shutdown() {
	if r.cancel != nil {
		r.cancel()
	}
	close(r.queue)
	r.wg.Wait()
}
