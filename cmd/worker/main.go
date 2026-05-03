package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MattInnovates/gitforge/core/config"
	"github.com/MattInnovates/gitforge/core/events"
	"github.com/MattInnovates/gitforge/core/storage/database"
	"github.com/MattInnovates/gitforge/core/worker"
)

func main() {
	cfg, err := config.LoadFromEnv()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	db, err := database.OpenAndMigrate(context.Background(), cfg)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer db.Close()

	runner := worker.NewRunner(128, 4, func(ctx context.Context, job worker.Job) error {
		log.Printf("processing job type=%s attempt=%d payload=%v", job.Type, job.Attempt, job.Payload)
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			return nil
		}
	})
	runner.Start(context.Background())
	defer runner.Shutdown()

	bus := events.NewBus()
	defer bus.Close()

	tickCh, unsubscribe := bus.Subscribe("worker.tick", 64)
	defer unsubscribe()

	go func() {
		for event := range tickCh {
			if err := runner.Enqueue(worker.Job{Type: "maintenance.tick", Payload: event.Payload, MaxAttempts: 3}); err != nil {
				log.Printf("enqueue maintenance job: %v", err)
			}
		}
	}()

	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for ts := range ticker.C {
			_ = bus.Publish(events.Event{
				Name:      "worker.tick",
				Timestamp: ts.UTC(),
				Payload: map[string]string{
					"source": "worker",
					"job":    "maintenance.tick",
				},
			})
		}
	}()

	log.Printf("worker running (repos=%s db_driver=%s)", cfg.ReposPath, cfg.DatabaseDriver)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Printf("worker shutting down")
}
