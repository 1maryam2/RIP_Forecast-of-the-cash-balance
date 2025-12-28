package handler

import (
	"sync"
)

type EventManager struct {
	subscribers map[chan bool]bool
	mu          sync.Mutex
}

func NewEventManager() *EventManager {
	return &EventManager{
		subscribers: make(map[chan bool]bool),
	}
}

func (em *EventManager) Subscribe() chan bool {
	em.mu.Lock()
	defer em.mu.Unlock()
	ch := make(chan bool, 1)
	em.subscribers[ch] = true
	return ch
}

func (em *EventManager) Unsubscribe(ch chan bool) {
	em.mu.Lock()
	defer em.mu.Unlock()
	delete(em.subscribers, ch)
	close(ch)
}

func (em *EventManager) Broadcast() {
	em.mu.Lock()
	defer em.mu.Unlock()
	for ch := range em.subscribers {
		select {
		case ch <- true:
		default:
		}
	}
}
