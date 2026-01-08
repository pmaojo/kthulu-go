package ai

import (
	"context"
	"fmt"
	"time"
)

// AIProvider es el tipo de proveedor de IA
type AIProvider string

const (
	ProviderGemini    AIProvider = "gemini"
	ProviderLiteLLM   AIProvider = "litellm"
	ProviderOpenAI    AIProvider = "openai"
	ProviderAnthropic AIProvider = "anthropic"
)

// AIService proporciona una abstracción sobre los clientes de IA
type AIService struct {
	provider AIProvider
	client   Client
	model    string
}

// NewAIService crea un nuevo servicio de IA
func NewAIService(provider AIProvider, config map[string]interface{}, cacheTTL time.Duration) (*AIService, error) {
	var client Client
	var err error

	switch provider {
	case ProviderLiteLLM:
		baseURL := "http://localhost:4000"
		if url, ok := config["baseUrl"].(string); ok {
			baseURL = url
		}
		timeout := 30 * time.Second
		if t, ok := config["timeout"].(int); ok {
			timeout = time.Duration(t) * time.Second
		}
		client = NewLiteLLMClient(LiteLLMConfig{
			BaseURL: baseURL,
			Timeout: timeout,
		}, cacheTTL)

	case ProviderGemini:
		apiKey, ok := config["apiKey"].(string)
		if !ok {
			return nil, fmt.Errorf("gemini provider requires apiKey")
		}
		client, err = NewGeminiClient(apiKey, cacheTTL)
		if err != nil {
			return nil, fmt.Errorf("failed to create gemini client: %w", err)
		}

	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}

	model := "gpt-4"
	if m, ok := config["model"].(string); ok {
		model = m
	}

	return &AIService{
		provider: provider,
		client:   client,
		model:    model,
	}, nil
}

// GenerateText genera texto usando el cliente configurado
func (s *AIService) GenerateText(ctx context.Context, prompt string) (string, error) {
	return s.client.GenerateText(ctx, prompt)
}

// SetModel cambia el modelo actual
func (s *AIService) SetModel(model string) {
	s.model = model
}

// GetModel retorna el modelo actual
func (s *AIService) GetModel() string {
	return s.model
}

// GetProvider retorna el proveedor actual
func (s *AIService) GetProvider() AIProvider {
	return s.provider
}

// Close cierra el cliente
func (s *AIService) Close() error {
	return s.client.Close()
}
