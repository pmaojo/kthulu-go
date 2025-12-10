// @kthulu:module:ai
package modules

import (
	"context"
	"os"
	"time"

	"go.uber.org/fx"

	adapterhttp "github.com/pmaojo/kthulu-go/backend/internal/adapters/http"
	"github.com/pmaojo/kthulu-go/backend/internal/infrastructure/ai"
	"github.com/pmaojo/kthulu-go/backend/internal/infrastructure/config"
	"github.com/pmaojo/kthulu-go/backend/internal/usecase"
)

var AIModule = fx.Options(
	// AI client provider with config-driven choice
	fx.Provide(
		func(cfg config.AIConfig) (ai.Client, error) {
			multi := ai.NewMultiProviderClient()

			// Always register mock for fallback/testing
			mockClient := ai.NewMockClientWithCache(cfg.CacheSize, time.Duration(cfg.CacheTTL)*time.Second)
			multi.RegisterProvider("mock", mockClient)

			// Register Gemini if configured or valid
			geminiClient, err := ai.NewGeminiClient(cfg.Model, time.Duration(cfg.CacheTTL)*time.Second)
			if err == nil {
				multi.RegisterProvider("gemini", geminiClient)
			}

			// Register OpenAI if configured
			if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
				openaiClient := ai.NewOpenAIProvider(apiKey, "gpt-4", cfg.CacheSize, time.Duration(cfg.CacheTTL)*time.Second)
				multi.RegisterProvider("openai", openaiClient)
			}

			// Register LiteLLM if configured
			if baseURL := os.Getenv("LITELLM_BASE_URL"); baseURL != "" {
				litellmClient := ai.NewLiteLLMClient(ai.LiteLLMConfig{
					BaseURL: baseURL,
					Timeout: 60 * time.Second,
				}, time.Duration(cfg.CacheTTL)*time.Second)
				multi.RegisterProvider("litellm", litellmClient)
			}

			// Set initial provider
			if cfg.UseMock {
				multi.SetProvider("mock")
			} else if os.Getenv("OPENAI_API_KEY") != "" {
				multi.SetProvider("openai")
			} else {
				// Default to gemini (which falls back to mock internally if no key) or whatever was registered first
				if err := multi.SetProvider("gemini"); err != nil {
					multi.SetProvider("mock")
				}
			}

			return multi, nil
		},
	),
	// Use case
	fx.Provide(
		usecase.NewAIUseCase,
	),
	// HTTP handler
	fx.Provide(
		adapterhttp.NewAIHandler,
	),
	fx.Invoke(func(handler *adapterhttp.AIHandler, registry *RouteRegistry) {
		registry.Register(handler)
	}),
	fx.Invoke(func(lc fx.Lifecycle, client ai.Client) {
		lc.Append(fx.Hook{OnStop: func(ctx context.Context) error {
			return client.Close()
		}})
	}),
)
