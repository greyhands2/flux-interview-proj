package graph

import (
	"context"
	"errors"

	"adstudio/internal/model"
)

// ErrNotImplemented is returned when the pipeline logic is stubbed out.
var ErrNotImplemented = errors.New("graph runner: pipeline logic not implemented yet")

// Runner executes the Ad Studio agent pipeline graph.
type Runner interface {
	Run(ctx context.Context, idea string, m model.Model) (*State, error)
}

// Graph implements the state machine pipeline.
type Graph struct{}

// New creates a new Graph runner.
func New() *Graph {
	return &Graph{}
}

// Run executes the graph state machine (stubbed).
func (g *Graph) Run(ctx context.Context, idea string, m model.Model) (*State, error) {
	// Project logic intentionally not implemented yet.
	return nil, ErrNotImplemented
}
