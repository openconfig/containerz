package cmd

import (
	"github.com/spf13/cobra"
	"google.golang.org/grpc/metadata"
	"google3/third_party/openconfig/containerz/client/client"
)

var containerCmd = &cobra.Command{
	Use:   "container",
	Short: "General container operations",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// copybara:strip_begin(google-context)
		ctx := metadata.AppendToOutgoingContext(cmd.Context(), "deviceFqdn", addr)
		cmd.SetContext(ctx)
		// copybara:end_begin
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