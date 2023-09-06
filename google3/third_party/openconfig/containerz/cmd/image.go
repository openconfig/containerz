package cmd

import (
	"github.com/spf13/cobra"
	"google.golang.org/grpc/metadata"
	"/client/client"
	cpb "github.com/openconfig/gnoi/containerz"
)

var (
	containerzClient *client.Client
	cli              cpb.ContainerzClient
	image, tag       string
)

var imageCmd = &cobra.Command{
	Use:   "image",
	Short: "General image operations",
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
	RootCmd.AddCommand(imageCmd)
	imageCmd.PersistentFlags().StringVar(&image, "image", "", "container image name")
	imageCmd.PersistentFlags().StringVar(&tag, "tag", "latest", "container image tag")
}
