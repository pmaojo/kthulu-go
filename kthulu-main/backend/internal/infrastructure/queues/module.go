// @kthulu:module:queues
package queues

import (
	"context"

	"backend/core"
	"go.uber.org/fx"
)

// Module provides the queue driver for Fx dependency injection.
var Module = fx.Options(
	fx.Provide(NewQueue),
)

// NewQueue initializes an Asynq-backed queue driver.
func NewQueue(lc fx.Lifecycle, logger core.Logger) (Queue, error) {
	q, err := NewAsynqQueue()
	if err != nil {
		return nil, err
	}
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return q.Close()
		},
	})
	logger.Info("Queue driver initialized", "driver", "asynq")
	return q, nil
}
