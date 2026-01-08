package domain

import "errors"

// Client represents an OAuth client.
type Client struct {
	ID           string
	Secret       string
	RedirectURIs []string
	Scopes       []string
}

// NewClient creates a new Client with minimal validation.
func NewClient(id, secret string, redirectURIs, scopes []string) (*Client, error) {
	if id == "" {
		return nil, errors.New("id is required")
	}
	if secret == "" {
		return nil, errors.New("secret is required")
	}
	return &Client{ID: id, Secret: secret, RedirectURIs: redirectURIs, Scopes: scopes}, nil
}
