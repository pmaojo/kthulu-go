package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/pmaojo/kthulu-go/internal/infrastructure/queues"
)

var queueCmd = &cobra.Command{
	Use:   "queue",
	Short: "Commands to interact with message queues",
}

var publishCmd = &cobra.Command{
	Use:   "publish [queue] [message]",
	Short: "Publish a message to a queue",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		q, err := queues.NewAsynqQueue()
		if err != nil {
			return err
		}
		defer q.Close()
		return q.Publish(cmd.Context(), args[0], []byte(args[1]))
	},
}

var consumeCmd = &cobra.Command{
	Use:   "consume [queue]",
	Short: "Consume messages from a queue",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		q, err := queues.NewAsynqQueue()
		if err != nil {
			return err
		}
		defer q.Close()
		ch, err := q.Consume(cmd.Context(), args[0])
		if err != nil {
			return err
		}
		for msg := range ch {
			fmt.Println(string(msg))
		}
		return nil
	},
}

func init() {
	queueCmd.AddCommand(publishCmd)
	queueCmd.AddCommand(consumeCmd)
	rootCmd.AddCommand(queueCmd)
}
