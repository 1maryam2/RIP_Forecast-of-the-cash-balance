package handler

import (
	"sync"
)

// EventManager управляет подписками на обновления
type EventManager struct {
	subscribers map[chan bool]bool // Карта каналов, которые ждут обновления
	mu          sync.Mutex
}

func NewEventManager() *EventManager {
	return &EventManager{
		subscribers: make(map[chan bool]bool),
	}
}

// Subscribe создает новый канал и добавляет его в список слушателей
func (em *EventManager) Subscribe() chan bool {
	em.mu.Lock()
	defer em.mu.Unlock()
	ch := make(chan bool, 1) // Буферизированный канал, чтобы не блокировать
	em.subscribers[ch] = true
	return ch
}

// Unsubscribe удаляет канал из списка
func (em *EventManager) Unsubscribe(ch chan bool) {
	em.mu.Lock()
	defer em.mu.Unlock()
	delete(em.subscribers, ch)
	close(ch)
}

// Broadcast отправляет сигнал всем слушателям, что данные изменились
func (em *EventManager) Broadcast() {
	em.mu.Lock()
	defer em.mu.Unlock()
	for ch := range em.subscribers {
		// Неблокирующая отправка
		select {
		case ch <- true:
		default:
		}
	}
}
