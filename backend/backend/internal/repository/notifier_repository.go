// @kthulu:module:notifier
package repository

import (
	"context"
)

// NotificationRequest represents a notification to be sent
type NotificationRequest struct {
	To      string                 `json:"to"`
	Subject string                 `json:"subject"`
	Body    string                 `json:"body"`
	Type    NotificationType       `json:"type"`
	Data    map[string]interface{} `json:"data,omitempty"`
}

// NotificationType represents the type of notification
type NotificationType string

const (
	NotificationTypeEmail             NotificationType = "email"
	NotificationTypeEmailConfirmation NotificationType = "email_confirmation"
	NotificationTypePasswordReset     NotificationType = "password_reset"
	NotificationTypeWelcome           NotificationType = "welcome"
	NotificationTypeInvitation        NotificationType = "invitation"
)

// NotificationProvider defines the interface for sending notifications
type NotificationProvider interface {
	SendNotification(ctx context.Context, req NotificationRequest) error
	SendEmailConfirmation(ctx context.Context, email, confirmationCode string) error
	SendPasswordReset(ctx context.Context, email, resetCode string) error
	SendWelcomeEmail(ctx context.Context, email, name string) error
}
