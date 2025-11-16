package domain

import "errors"

// Session links a user to a client and its authentication methods.
type Session struct {
	UserID      string
	ClientID    string
	AuthMethods []string
}

// NewSession creates a new Session with minimal validation.
func NewSession(userID, clientID string, authMethods []string) (*Session, error) {
	if userID == "" {
		return nil, errors.New("userID is required")
	}
	if clientID == "" {
		return nil, errors.New("clientID is required")
	}
	return &Session{UserID: userID, ClientID: clientID, AuthMethods: authMethods}, nil
}
