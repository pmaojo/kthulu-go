package queues

import (
	"context"
	"os"

	"github.com/hibiken/asynq"
)

// AsynqQueue implements the Queue interface using github.com/hibiken/asynq.
type AsynqQueue struct {
	client   *asynq.Client
	redisOpt asynq.RedisClientOpt
}

// NewAsynqQueue creates a new Asynq-backed queue driver.
// It uses the REDIS_ADDR environment variable or defaults to localhost:6379.
func NewAsynqQueue() (Queue, error) {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}
	opt := asynq.RedisClientOpt{Addr: addr}
	client := asynq.NewClient(opt)
	return &AsynqQueue{client: client, redisOpt: opt}, nil
}

// Publish enqueues a payload with the given queue name as task type.
func (a *AsynqQueue) Publish(ctx context.Context, queue string, payload []byte) error {
	task := asynq.NewTask(queue, payload)
	_, err := a.client.EnqueueContext(ctx, task)
	return err
}

// Consume subscribes to tasks of the given queue name and delivers their payloads on a channel.
func (a *AsynqQueue) Consume(ctx context.Context, queue string) (<-chan []byte, error) {
	ch := make(chan []byte)
	server := asynq.NewServer(a.redisOpt, asynq.Config{
		Concurrency: 1,
		Queues:      map[string]int{queue: 1},
	})
	go func() {
		defer close(ch)
		handler := asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
			ch <- t.Payload()
			return nil
		})
		if err := server.Run(handler); err != nil {
			// server exited with error; no further action needed for this simple driver
		}
	}()
	go func() {
		<-ctx.Done()
		server.Shutdown()
	}()
	return ch, nil
}

// Close closes the underlying Asynq client.
func (a *AsynqQueue) Close() error {
	return a.client.Close()
}
