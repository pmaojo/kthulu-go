package ai

import "context"

// Client is the interface used by AI features. Implementations may be real
// Gemini clients or local mocks used for tests.
type Client interface {
	GenerateText(ctx context.Context, prompt string) (string, error)
	Close() error
}
