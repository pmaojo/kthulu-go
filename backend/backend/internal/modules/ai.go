// @kthulu:module:ai
package modules

import (
	"context"
	"time"

	"go.uber.org/fx"

	adapterhttp "github.com/pmaojo/kthulu-go/backend/internal/adapters/http"
	"github.com/pmaojo/kthulu-go/backend/internal/ai"
	"github.com/pmaojo/kthulu-go/backend/internal/config"
	"github.com/pmaojo/kthulu-go/backend/internal/usecase"
)

var AIModule = fx.Options(
	// AI client provider with config-driven choice
	fx.Provide(
		func(cfg config.AIConfig) (ai.Client, error) {
			if cfg.UseMock {
				// Return mock client for testing/dev
				return ai.NewMockClientWithCache(cfg.CacheSize, time.Duration(cfg.CacheTTL)*time.Second), nil
			}
			// Use real Gemini client (or fallback to mock if GEMINI_API_KEY not set)
			return ai.NewGeminiClient(cfg.Model, time.Duration(cfg.CacheTTL)*time.Second)
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
