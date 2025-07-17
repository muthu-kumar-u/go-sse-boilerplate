package stream

import (
	"context"
	"sync"
)

type StreamHub struct {
	mu      sync.RWMutex
	streams map[string][]chan []byte
}

func NewStreamHub() *StreamHub {
	return &StreamHub{
		streams: make(map[string][]chan []byte),
	}
}

// Publish sends data to all subscribers of a stream.
// Returns true if at least one subscriber received the message.
func (s *StreamHub) Publish(stream string, data []byte) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	subs, exists := s.streams[stream]
	if !exists {
		return false
	}

	// Create a copy of the data to prevent race conditions
	msg := make([]byte, len(data))
	copy(msg, data)

	sent := false
	for _, ch := range subs {
		select {
		case ch <- msg:
			sent = true
		default:
			// Skip blocked channels to prevent publisher blocking
		}
	}
	return sent
}

// Subscribe creates a new subscription to the specified stream.
// The subscription will be automatically closed when the context is cancelled.
func (s *StreamHub) Subscribe(ctx context.Context, stream string) (<-chan []byte, error) {
	ch := make(chan []byte, 10)

	s.mu.Lock()
	s.streams[stream] = append(s.streams[stream], ch)
	s.mu.Unlock()

	// Setup automatic unsubscription when context is done
	go func() {
		<-ctx.Done()
		s.Unsubscribe(stream, ch)
	}()

	return ch, nil
}

// Unsubscribe removes a specific channel from a stream's subscribers
// and cleans up the channel. Safe to call multiple times.
func (s *StreamHub) Unsubscribe(stream string, ch chan []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()

	subs, exists := s.streams[stream]
	if !exists {
		return
	}

	for i, sub := range subs {
		if sub == ch {
			// Remove the channel from the slice
			s.streams[stream] = append(subs[:i], subs[i+1:]...)
			
			// Close the channel safely
			select {
			case _, ok := <-ch:
				if ok {
					close(ch)
				}
			default:
				close(ch)
			}
			
			// Clean up empty streams
			if len(s.streams[stream]) == 0 {
				delete(s.streams, stream)
			}
			return
		}
	}
}

// CloseStream closes all channels for a stream and removes it
func (s *StreamHub) CloseStream(stream string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	subs, exists := s.streams[stream]
	if !exists {
		return
	}

	for _, ch := range subs {
		close(ch)
	}
	delete(s.streams, stream)
}

// SubscriberCount returns the number of active subscribers for a stream
func (s *StreamHub) SubscriberCount(stream string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.streams[stream])
}