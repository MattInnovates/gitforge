package events

import (
	"errors"
	"testing"
	"time"
)

func TestBusPublishSubscribe(t *testing.T) {
	bus := NewBus()
	t.Cleanup(bus.Close)

	ch, unsubscribe := bus.Subscribe("repo.created", 1)
	t.Cleanup(unsubscribe)

	evt := Event{Name: "repo.created", Timestamp: time.Now().UTC(), Payload: "demo"}
	if err := bus.Publish(evt); err != nil {
		t.Fatalf("publish returned error: %v", err)
	}

	select {
	case got := <-ch:
		if got.Name != evt.Name {
			t.Fatalf("expected name %q, got %q", evt.Name, got.Name)
		}
		if got.Payload != evt.Payload {
			t.Fatalf("expected payload %v, got %v", evt.Payload, got.Payload)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for event")
	}
}

func TestBusUnsubscribeClosesChannel(t *testing.T) {
	bus := NewBus()

	ch, unsubscribe := bus.Subscribe("repo.deleted", 0)
	unsubscribe()

	_, ok := <-ch
	if ok {
		t.Fatal("expected channel to be closed after unsubscribe")
	}
}

func TestBusPublishQueueFull(t *testing.T) {
	bus := NewBus()
	t.Cleanup(bus.Close)

	_, unsubscribe := bus.Subscribe("repo.pushed", 1)
	t.Cleanup(unsubscribe)

	evt := Event{Name: "repo.pushed", Timestamp: time.Now().UTC()}
	if err := bus.Publish(evt); err != nil {
		t.Fatalf("first publish returned error: %v", err)
	}

	err := bus.Publish(evt)
	if !errors.Is(err, ErrSubscriberQueueFull) {
		t.Fatalf("expected ErrSubscriberQueueFull, got %v", err)
	}
}

func TestBusCloseClosesSubscriberChannels(t *testing.T) {
	bus := NewBus()
	ch, _ := bus.Subscribe("repo.created", 0)

	bus.Close()

	_, ok := <-ch
	if ok {
		t.Fatal("expected channel closed after bus close")
	}
}
