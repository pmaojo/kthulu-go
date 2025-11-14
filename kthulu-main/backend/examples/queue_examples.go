// @kthulu:core
package examples

import (
	"context"
	"fmt"

	"backend/internal/infrastructure/queues"
)

// QueueProducerExample muestra cómo enviar un mensaje a una cola.
func QueueProducerExample() error {
	q, err := queues.NewAsynqQueue()
	if err != nil {
		return err
	}
	defer q.Close()
	return q.Publish(context.Background(), "demo", []byte("hola mundo"))
}

// QueueConsumerExample muestra cómo consumir un mensaje de una cola.
func QueueConsumerExample() error {
	q, err := queues.NewAsynqQueue()
	if err != nil {
		return err
	}
	defer q.Close()
	ch, err := q.Consume(context.Background(), "demo")
	if err != nil {
		return err
	}
	msg := <-ch
	fmt.Println(string(msg))
	return nil
}
