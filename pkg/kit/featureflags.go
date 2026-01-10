package core

import (
	"context"
	"fmt"
	"net/http"

	flagsmith "github.com/Flagsmith/flagsmith-go-client"
	unleash "github.com/Unleash/unleash-client-go/v4"
)

// FeatureFlagClient defines the interface for evaluating feature flags
// It allows the application to toggle functionality for dark launches.
type FeatureFlagClient interface {
	IsEnabled(ctx context.Context, flag string) bool
}

// NewFeatureFlagClient constructs a feature flag client based on configuration.
// Supported providers: "unleash", "flagsmith".
func NewFeatureFlagClient(cfg *Config) (FeatureFlagClient, error) {
	if cfg.FeatureFlags.Provider == "" {
		return nil, fmt.Errorf("feature flag provider not configured")
	}
	switch cfg.FeatureFlags.Provider {
	case "unleash":
		options := []unleash.ConfigOption{unleash.WithAppName(cfg.FeatureFlags.AppName)}
		if cfg.FeatureFlags.URL != "" {
			options = append(options, unleash.WithUrl(cfg.FeatureFlags.URL))
		}
		if cfg.FeatureFlags.APIKey != "" {
			headers := http.Header{}
			headers.Set("Authorization", cfg.FeatureFlags.APIKey)
			options = append(options, unleash.WithCustomHeaders(headers))
		}
		if cfg.FeatureFlags.Environment != "" {
			options = append(options, unleash.WithEnvironment(cfg.FeatureFlags.Environment))
		}
		client, err := unleash.NewClient(options...)
		if err != nil {
			return nil, err
		}
		return &unleashWrapper{client: client}, nil
	case "flagsmith":
		cfgOpts := flagsmith.DefaultConfig()
		if cfg.FeatureFlags.URL != "" {
			cfgOpts.BaseURI = cfg.FeatureFlags.URL
		}
		client := flagsmith.NewClient(cfg.FeatureFlags.APIKey, cfgOpts)
		return &flagsmithWrapper{client: client}, nil
	default:
		return nil, fmt.Errorf("unknown feature flag provider: %s", cfg.FeatureFlags.Provider)
	}
}

type unleashWrapper struct{ client *unleash.Client }

func (u *unleashWrapper) IsEnabled(ctx context.Context, flag string) bool {
	// Unleash client does not use context for simple checks
	return u.client.IsEnabled(flag)
}

type flagsmithWrapper struct{ client *flagsmith.Client }

func (f *flagsmithWrapper) IsEnabled(ctx context.Context, flag string) bool {
	enabled, err := f.client.FeatureEnabledWithContext(ctx, flag)
	if err != nil {
		return false
	}
	return enabled
}
