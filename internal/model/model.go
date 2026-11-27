package model

import "context"

// Model defines the minimal interface for an LLM provider.
type Model interface {
	Generate(ctx context.Context, prompt string) (string, error)
}
