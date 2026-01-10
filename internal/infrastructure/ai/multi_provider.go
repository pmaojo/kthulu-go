package ai

import (
	"context"
	"fmt"
	"sync"
)

// MultiProviderClient allows switching between multiple AI providers
type MultiProviderClient struct {
	providers       map[string]Client
	currentProvider string
	mu              sync.RWMutex
}

// NewMultiProviderClient creates a new MultiProviderClient
func NewMultiProviderClient() *MultiProviderClient {
	return &MultiProviderClient{
		providers: make(map[string]Client),
	}
}

// RegisterProvider registers a provider with a name
func (c *MultiProviderClient) RegisterProvider(name string, client Client) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.providers[name] = client
	if c.currentProvider == "" {
		c.currentProvider = name
	}
}

// SetProvider sets the active provider
func (c *MultiProviderClient) SetProvider(name string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.providers[name]; !ok {
		return fmt.Errorf("provider not found: %s", name)
	}
	c.currentProvider = name
	return nil
}

// GenerateText delegates to the current provider
func (c *MultiProviderClient) GenerateText(ctx context.Context, prompt string) (string, error) {
	c.mu.RLock()
	providerName := c.currentProvider
	provider, ok := c.providers[providerName]
	c.mu.RUnlock()

	if !ok {
		return "", fmt.Errorf("no provider selected")
	}
	return provider.GenerateText(ctx, prompt)
}

// Close closes all providers
func (c *MultiProviderClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var errs []error
	for _, p := range c.providers {
		if err := p.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("errors closing providers: %v", errs)
	}
	return nil
}
