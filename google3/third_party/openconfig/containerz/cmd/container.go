package cmd

import (
	"github.com/spf13/cobra"
	"google.golang.org/grpc/metadata"
	"/client/client"
)

var containerCmd = &cobra.Command{
	Use:   "container",
	Short: "General container operations",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		containerzClient, err = client.NewClient(cmd.Context(), addr)
		return err
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func init() {
	RootCmd.AddCommand(containerCmd)
	containerCmd.PersistentFlags().StringVar(&image, "image", "", "container image name")
	containerCmd.PersistentFlags().StringVar(&tag, "tag", "latest", "container image tag")
}
