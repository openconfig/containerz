package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	force bool
)

var cntStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "stop a container by intance name",
	RunE: func(command *cobra.Command, args []string) error {
		if instance == "" {
			fmt.Println("--instance must be provided")
		}

		if err := containerzClient.Stop(command.Context(), instance, force); err != nil {
			return err
		}

		fmt.Printf("Successfully stopped %s\n", instance)
		return nil
	},
}

func init() {
	containerCmd.AddCommand(cntStopCmd)

	cntStopCmd.PersistentFlags().StringVar(&instance, "instance", "", "Container instance to stop.")
	cntStopCmd.PersistentFlags().BoolVar(&force, "force", false, "Forcefully stop the container.")
}
