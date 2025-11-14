// @kthulu:module:notifier
package notifier

import (
	"go.uber.org/fx"

	"backend/internal/infrastructure/notifier"
)

// Module provides fx.Options for notifier (communications) module.
// Includes email, SMS, and push notifications.
var Module = fx.Options(
	fx.Provide(
		fx.Annotated{
			Name:   "smtpProvider",
			Target: notifier.NewSMTPProvider,
		},
	),
)
