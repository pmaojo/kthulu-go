// @kthulu:module:queue
package queue

import (
	"go.uber.org/fx"

	"backend/internal/infrastructure/queues"
)

// Module provides fx.Options for queue (async processing) module.
// Includes background job processing and message queues.
var Module = fx.Options(
	queues.Module,
)
