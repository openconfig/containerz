package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	follow bool
)

var cntLogCmd = &cobra.Command{
	Use:   "logs",
	Short: "fetch the logs for a container",
	RunE: func(command *cobra.Command, args []string) error {
		if instance == "" {
			fmt.Println("--instance must be provided")
		}

		ch, err := containerzClient.Logs(command.Context(), instance, follow)
		if err != nil {
			return err
		}

		for msg := range ch {
			if msg.Error != nil {
				return msg.Error
			}
			fmt.Print(msg.Msg)
		}
		return nil
	},
}

func init() {
	containerCmd.AddCommand(cntLogCmd)

	cntLogCmd.PersistentFlags().StringVar(&instance, "instance", "", "Container instance to stop.")
	cntLogCmd.PersistentFlags().BoolVar(&follow, "follow", false, "Follow logs.")
}
