// @kthulu:core
package queues

import "context"

// Queue defines the interface for message queue implementations.
type Queue interface {
	// Publish sends a payload to the specified queue name.
	Publish(ctx context.Context, queue string, payload []byte) error
	// Consume subscribes to messages from the specified queue and returns a channel.
	Consume(ctx context.Context, queue string) (<-chan []byte, error)
	// Close releases resources used by the queue driver.
	Close() error
}
