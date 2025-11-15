// @kthulu:module:notifier
package modules

import (
	"go.uber.org/fx"

	"github.com/kthulu/kthulu-go/backend/internal/infrastructure/notifier"
)

// NotifierModule provides notification functionality
var NotifierModule = fx.Options(
	// Notification providers
	notifier.NotifierModule,
)
