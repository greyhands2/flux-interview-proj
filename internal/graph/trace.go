package graph

import (
	"time"
)

// LogTrace appends a timestamped audit event to the state's trace log.
func (s *State) LogTrace(step string, details interface{}) {
	entry := TraceEntry{
		Step:      step,
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		Details:   details,
	}
	s.Trace = append(s.Trace, entry)
}

// RecordRoute records a visited node or final outcome in the state machine's path.
func (s *State) RecordRoute(node string) {
	s.RouteHistory = append(s.RouteHistory, node)
}
