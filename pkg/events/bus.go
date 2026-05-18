package events

func New() *Bus {
	return &Bus{
		subs: make(map[string][]EventSubscriberFunc),
	}
}

// Subscribe registers a handler for a given topic.
// The handler is called asynchronously (in a new goroutine) each time
// an event is published on that topic.
func (b *Bus) Subscribe(topic EventType, handler EventSubscriberFunc) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.subs[string(topic)] = append(b.subs[string(topic)], handler)
}

// Publish sends an event to all subscribers of the given topic.
// Each handler runs in its own goroutine so publishers never block.
func (b *Bus) Publish(topic EventType, payload any) {
	b.mu.RLock()
	handlers := make([]EventSubscriberFunc, len(b.subs[string(topic)]))
	copy(handlers, b.subs[string(topic)])
	b.mu.RUnlock()

	event := Event{Type: topic, Payload: payload}
	for _, h := range handlers {
		go h(event, payload)
	}
}

// PublishSync is like Publish but calls handlers in the caller's goroutine.
// Useful when you need to guarantee ordering or want backpressure.
func (b *Bus) PublishSync(topic EventType, payload any) {
	b.mu.RLock()
	handlers := make([]EventSubscriberFunc, len(b.subs[string(topic)]))
	copy(handlers, b.subs[string(topic)])
	b.mu.RUnlock()

	event := Event{Type: topic, Payload: payload}
	for _, h := range handlers {
		h(event, payload)
	}
}
