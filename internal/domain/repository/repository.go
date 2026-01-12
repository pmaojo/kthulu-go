package repository

import (
	"context"
)

// ConnectionRepository defines the interface for managing websocket connections
type ConnectionRepository interface {
	Add(ctx context.Context, conn any) (string, error)
	Remove(ctx context.Context, id string) error
}

// NotificationType defines the type of notification
type NotificationType string

const (
	NotificationTypeEmailConfirmation NotificationType = "EMAIL_CONFIRMATION"
	NotificationTypePasswordReset     NotificationType = "PASSWORD_RESET"
	NotificationTypeWelcome           NotificationType = "WELCOME"
)

// NotificationRequest represents a notification request
type NotificationRequest struct {
	To      string
	Subject string
	Body    string
	Type    NotificationType
	Data    map[string]interface{}
}

// NotificationProvider defines the interface for sending notifications
type NotificationProvider interface {
	SendNotification(ctx context.Context, req NotificationRequest) error
	SendEmailConfirmation(ctx context.Context, email, confirmationCode string) error
	SendPasswordReset(ctx context.Context, email, resetCode string) error
	SendWelcomeEmail(ctx context.Context, email, name string) error
}
