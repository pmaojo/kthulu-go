package config

// AIConfig holds AI-related configuration
type AIConfig struct {
	// UsesMock indicates whether to use a mock AI client for testing/dev
	UseMock bool
	// Model is the AI model to use (e.g., "gemini-1.5-pro")
	Model string
	// CacheSize is the max number of entries in the AI response cache
	CacheSize int
	// CacheTTL is the time-to-live for cache entries (in seconds)
	CacheTTL int
}

// DefaultAIConfig returns sensible defaults
func DefaultAIConfig() AIConfig {
	return AIConfig{
		UseMock:   false,
		Model:     "gemini-1.5-pro",
		CacheSize: 256,
		CacheTTL:  300, // 5 minutes
	}
}
