package stream

import (
	"context"
	"log"
	"sync"
	"time"
)

type StreamHub struct {
	mu      sync.RWMutex
	streams map[string]*Stream
	timers  map[string]*time.Timer
}

type Stream struct {
	subscribers map[chan []byte]struct{}
}

func NewStreamHub() *StreamHub {
	return &StreamHub{
		streams: make(map[string]*Stream),
		timers:  make(map[string]*time.Timer),
	}
}

// Subscribe to a stream
func (b *StreamHub) Subscribe(ctx context.Context, streamID string) (chan []byte, error) {
	ch := make(chan []byte, 10)

	b.mu.Lock()
	stream, exists := b.streams[streamID]
	if !exists {
		stream = &Stream{
			subscribers: make(map[chan []byte]struct{}),
		}
		b.streams[streamID] = stream
	}
	stream.subscribers[ch] = struct{}{}

	// Cancel pending deletion if stream is re-used
	if timer, exists := b.timers[streamID]; exists {
		timer.Stop()
		delete(b.timers, streamID)
	}
	b.mu.Unlock()

	log.Printf("[📥] Subscribed to stream: %s", streamID)
	return ch, nil
}

// Publish a message to all subscribers of a stream
func (b *StreamHub) Publish(streamID string, data []byte) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	stream, exists := b.streams[streamID]
	if !exists {
		return
	}

	for ch := range stream.subscribers {
		select {
		case ch <- data:
		default:
			// Drop if buffer is full
		}
	}
}

// Unsubscribe a channel from a stream
func (b *StreamHub) Unsubscribe(streamID string, target chan []byte) {
	b.mu.Lock()
	defer b.mu.Unlock()

	stream, exists := b.streams[streamID]
	if !exists {
		return
	}

	if _, ok := stream.subscribers[target]; ok {
		close(target)
		delete(stream.subscribers, target)
	}

	// Clean up stream if no subscribers remain
	if len(stream.subscribers) == 0 {
		b.timers[streamID] = time.AfterFunc(2*time.Minute, func() {
			b.mu.Lock()
			defer b.mu.Unlock()
			delete(b.streams, streamID)
			delete(b.timers, streamID)
			log.Printf("[🗑️] Deleted inactive stream: %s", streamID)
		})
	}
}

// ListStreams returns all currently active stream IDs
func (b *StreamHub) ListStreams() []string {
	b.mu.RLock()
	defer b.mu.RUnlock()

	ids := make([]string, 0, len(b.streams))
	for streamID := range b.streams {
		ids = append(ids, streamID)
	}
	return ids
}

// Exists checks if a stream exists
func (b *StreamHub) Exists(streamID string) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	_, ok := b.streams[streamID]
	return ok
}

// CreateTemporaryStream creates a stream with automatic expiration
func (b *StreamHub) CreateTemporaryStream(streamID string, ttl time.Duration) {
	b.mu.Lock()
	if _, exists := b.streams[streamID]; exists {
		b.mu.Unlock()
		log.Printf("[ℹ️] Stream %s already exists, skipping creation", streamID)
		return
	}

	b.streams[streamID] = &Stream{
		subscribers: make(map[chan []byte]struct{}),
	}
	b.mu.Unlock()

	log.Printf("[🆕] Created temporary stream: %s (expires in %s)", streamID, ttl)

	b.timers[streamID] = time.AfterFunc(ttl, func() {
		b.mu.Lock()
		defer b.mu.Unlock()
		if _, exists := b.streams[streamID]; exists {
			delete(b.streams, streamID)
			delete(b.timers, streamID)
			log.Printf("[🗑️] Automatically deleted expired stream: %s", streamID)
		}
	})
}
