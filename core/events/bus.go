package events

import (
	"errors"
	"sync"
)

var ErrSubscriberQueueFull = errors.New("subscriber queue is full")

// Bus provides in-process pub/sub for internal service events.
type Bus struct {
	mu          sync.RWMutex
	subscribers map[string]map[uint64]chan Event
	nextID      uint64
	closed      bool
}

func NewBus() *Bus {
	return &Bus{
		subscribers: make(map[string]map[uint64]chan Event),
	}
}

// Subscribe registers a subscriber for an event name.
// It returns a receive-only channel and an unsubscribe function.
func (b *Bus) Subscribe(eventName string, buffer int) (<-chan Event, func()) {
	if buffer < 0 {
		buffer = 0
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	ch := make(chan Event, buffer)
	if b.closed {
		close(ch)
		return ch, func() {}
	}

	if _, ok := b.subscribers[eventName]; !ok {
		b.subscribers[eventName] = make(map[uint64]chan Event)
	}

	id := b.nextID
	b.nextID++
	b.subscribers[eventName][id] = ch

	unsubscribed := false
	unsubscribe := func() {
		b.mu.Lock()
		defer b.mu.Unlock()
		if unsubscribed {
			return
		}
		unsubscribed = true

		subsForEvent, ok := b.subscribers[eventName]
		if !ok {
			return
		}
		out, ok := subsForEvent[id]
		if !ok {
			return
		}

		delete(subsForEvent, id)
		if len(subsForEvent) == 0 {
			delete(b.subscribers, eventName)
		}
		close(out)
	}

	return ch, unsubscribe
}

// Publish delivers an event to all subscribers for the event name.
// It is non-blocking and returns ErrSubscriberQueueFull if any subscriber queue is full.
func (b *Bus) Publish(event Event) error {
	b.mu.RLock()
	if b.closed {
		b.mu.RUnlock()
		return nil
	}

	subsForEvent := b.subscribers[event.Name]
	channels := make([]chan Event, 0, len(subsForEvent))
	for _, ch := range subsForEvent {
		channels = append(channels, ch)
	}
	b.mu.RUnlock()

	queueFull := false
	for _, ch := range channels {
		select {
		case ch <- event:
		default:
			queueFull = true
		}
	}

	if queueFull {
		return ErrSubscriberQueueFull
	}
	return nil
}

// Close closes the bus and all active subscriber channels.
func (b *Bus) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return
	}
	b.closed = true

	for eventName, subsForEvent := range b.subscribers {
		for id, ch := range subsForEvent {
			close(ch)
			delete(subsForEvent, id)
		}
		delete(b.subscribers, eventName)
	}
}
