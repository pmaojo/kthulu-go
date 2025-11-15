// @kthulu:module:notifier
package notifier

import (
	"context"
	"fmt"

	"github.com/pmaojo/kthulu-go/backend/core"
	"github.com/pmaojo/kthulu-go/backend/internal/repository"
)

// ConsoleProvider implements NotificationProvider by logging to console
type ConsoleProvider struct {
	logger core.Logger
}

// NewConsoleProvider creates a new console notification provider
func NewConsoleProvider(logger core.Logger) repository.NotificationProvider {
	return &ConsoleProvider{
		logger: logger,
	}
}

// SendNotification sends a notification by logging it to console
func (c *ConsoleProvider) SendNotification(ctx context.Context, req repository.NotificationRequest) error {
	c.logger.Info("Notification sent",
		"type", string(req.Type),
		"to", req.To,
		"subject", req.Subject,
		"body", req.Body,
		"data", req.Data,
	)

	// Also print to stdout for development visibility
	fmt.Printf("\n=== NOTIFICATION ===\n")
	fmt.Printf("Type: %s\n", req.Type)
	fmt.Printf("To: %s\n", req.To)
	fmt.Printf("Subject: %s\n", req.Subject)
	fmt.Printf("Body: %s\n", req.Body)
	if len(req.Data) > 0 {
		fmt.Printf("Data: %+v\n", req.Data)
	}
	fmt.Printf("==================\n\n")

	return nil
}

// SendEmailConfirmation sends an email confirmation notification
func (c *ConsoleProvider) SendEmailConfirmation(ctx context.Context, email, confirmationCode string) error {
	req := repository.NotificationRequest{
		To:      email,
		Subject: "Confirm Your Email Address",
		Body:    fmt.Sprintf("Please confirm your email address by using this code: %s", confirmationCode),
		Type:    repository.NotificationTypeEmailConfirmation,
		Data: map[string]interface{}{
			"confirmationCode": confirmationCode,
		},
	}

	return c.SendNotification(ctx, req)
}

// SendPasswordReset sends a password reset notification
func (c *ConsoleProvider) SendPasswordReset(ctx context.Context, email, resetCode string) error {
	req := repository.NotificationRequest{
		To:      email,
		Subject: "Reset Your Password",
		Body:    fmt.Sprintf("Use this code to reset your password: %s", resetCode),
		Type:    repository.NotificationTypePasswordReset,
		Data: map[string]interface{}{
			"resetCode": resetCode,
		},
	}

	return c.SendNotification(ctx, req)
}

// SendWelcomeEmail sends a welcome email notification
func (c *ConsoleProvider) SendWelcomeEmail(ctx context.Context, email, name string) error {
	req := repository.NotificationRequest{
		To:      email,
		Subject: "Welcome!",
		Body:    fmt.Sprintf("Welcome %s! Thank you for joining us.", name),
		Type:    repository.NotificationTypeWelcome,
		Data: map[string]interface{}{
			"name": name,
		},
	}

	return c.SendNotification(ctx, req)
}

// Ensure ConsoleProvider implements NotificationProvider
var _ repository.NotificationProvider = (*ConsoleProvider)(nil)
