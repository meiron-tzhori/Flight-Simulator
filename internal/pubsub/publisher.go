package pubsub

import (
	"sync"

	"github.com/meiron-tzhori/Flight-Simulator/internal/models"
)

// StatePublisher manages state update subscriptions using a fan-out pattern.
type StatePublisher struct {
	mu          sync.RWMutex
	subscribers map[string]chan models.AircraftState
	bufferSize  int
}

// NewStatePublisher creates a new state publisher.
func NewStatePublisher(bufferSize int) *StatePublisher {
	return &StatePublisher{
		subscribers: make(map[string]chan models.AircraftState),
		bufferSize:  bufferSize,
	}
}

// Subscribe creates a new subscription and returns a channel for state updates.
func (p *StatePublisher) Subscribe(id string) <-chan models.AircraftState {
	p.mu.Lock()
	defer p.mu.Unlock()

	ch := make(chan models.AircraftState, p.bufferSize)
	p.subscribers[id] = ch
	return ch
}

// Unsubscribe removes a subscription and closes its channel.
func (p *StatePublisher) Unsubscribe(id string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if ch, exists := p.subscribers[id]; exists {
		close(ch)
		delete(p.subscribers, id)
	}
}

// Publish sends a state update to all subscribers (non-blocking).
func (p *StatePublisher) Publish(state models.AircraftState) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	for id, ch := range p.subscribers {
		select {
		case ch <- state:
			// Successfully sent
		default:
			// Channel full, skip (subscriber is lagging)
			// TODO: Add logging or metrics
			_ = id // avoid unused variable warning
		}
	}
}

// SubscriberCount returns the current number of subscribers.
func (p *StatePublisher) SubscriberCount() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.subscribers)
}
