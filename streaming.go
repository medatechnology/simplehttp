package simplehttp

import (
	"fmt"
	"net/http"
)

// SSEEvent represents a Server-Sent Event
type SSEEvent struct {
	ID    string // Optional event ID
	Event string // Event type (defaults to "message")
	Data  string // Event data
	Retry int    // Optional retry interval in milliseconds
}

// String formats the SSE event for transmission
func (e SSEEvent) String() string {
	var result string

	if e.ID != "" {
		result += fmt.Sprintf("id: %s\n", e.ID)
	}
	if e.Event != "" && e.Event != "message" {
		result += fmt.Sprintf("event: %s\n", e.Event)
	}
	if e.Retry > 0 {
		result += fmt.Sprintf("retry: %d\n", e.Retry)
	}
	result += fmt.Sprintf("data: %s\n\n", e.Data)

	return result
}

// SSEWriter provides methods for writing SSE events
type SSEWriter interface {
	// Send sends a simple data message
	Send(data string) error
	// SendEvent sends a full SSE event
	SendEvent(event SSEEvent) error
	// Flush flushes the response writer
	Flush()
}

// sseWriter implements SSEWriter
type sseWriter struct {
	w       http.ResponseWriter
	flusher http.Flusher
}

// NewSSEWriter creates a new SSE writer from an http.ResponseWriter
func NewSSEWriter(w http.ResponseWriter) SSEWriter {
	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)

	flusher, _ := w.(http.Flusher)
	return &sseWriter{w: w, flusher: flusher}
}

// Send sends a simple data message
func (s *sseWriter) Send(data string) error {
	return s.SendEvent(SSEEvent{Data: data})
}

// SendEvent sends a full SSE event
func (s *sseWriter) SendEvent(event SSEEvent) error {
	_, err := s.w.Write([]byte(event.String()))
	if err != nil {
		return err
	}
	s.Flush()
	return nil
}

// Flush flushes the response writer
func (s *sseWriter) Flush() {
	if s.flusher != nil {
		s.flusher.Flush()
	}
}

// SSEHandler is a function that receives an SSEWriter for streaming events
type SSEHandler func(SSEWriter) error
