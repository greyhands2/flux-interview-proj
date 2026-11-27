package model

import (
	"context"
	"errors"
	"sync"
)

var ErrNoMoreResponses = errors.New("scripted model: no more scripted responses available")

// ScriptedModel implements Model returning pre-scripted responses sequentially or by step.
type ScriptedModel struct {
	mu        sync.Mutex
	responses []string
	index     int
}

// NewScriptedModel creates a ScriptedModel with a predetermined sequence of responses.
func NewScriptedModel(responses []string) *ScriptedModel {
	return &ScriptedModel{
		responses: responses,
	}
}

// Generate returns the next scripted response in the sequence.
func (s *ScriptedModel) Generate(ctx context.Context, prompt string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.index >= len(s.responses) {
		return "", ErrNoMoreResponses
	}
	res := s.responses[s.index]
	s.index++
	return res, nil
}
