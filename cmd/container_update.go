package cmd

import (
	"fmt"
	"context"
	"google.golang.org/grpc/metadata"

	"github.com/spf13/cobra"
	"github.com/openconfig/containerz/client"
)

var (
	updateAsync bool
)

var cntUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a running container with the specified image, tag, and other parameters",
	RunE: func(command *cobra.Command, args []string) error {
		if instance == "" {
			return fmt.Errorf("--instance must be specified")
		}
		if image == "" {
			return fmt.Errorf("--image must be specified")
		}
		ctx, cancel := context.WithCancel(command.Context())
		defer cancel()
		ctx = metadata.AppendToOutgoingContext(ctx, "username", "cisco", "password", "cisco123")

		id, err := containerzClient.UpdateContainer(ctx, instance, image, tag, cntCommand, updateAsync, client.WithEnv(envs), client.WithPorts(ports), client.WithVolumes(volumes))
		if err != nil {
			return err
		}

		fmt.Printf("Container updated with id - %s\n", id)
		return nil
	},
}

func init() {
	containerCmd.AddCommand(cntUpdateCmd)

	cntUpdateCmd.PersistentFlags().StringVar(&cntCommand, "command", "/bin/bash", "command to run.")
	cntUpdateCmd.PersistentFlags().StringVar(&instance, "instance", "", "Name of the container to update.")
	cntUpdateCmd.PersistentFlags().StringArrayVar(&ports, "port", []string{}, "Ports to expose (format: <internal_port>:<external_port>")
	cntUpdateCmd.PersistentFlags().StringArrayVar(&envs, "env", []string{}, "Environment vars to set (format: <VAR_NAME>=<VAR_VALUE>")
	cntUpdateCmd.PersistentFlags().StringArrayVarP(&volumes, "volume", "v", []string{}, "Volumes to attach to the container (format: <volume-name>:<mountpoint>[:ro])")
	cntUpdateCmd.PersistentFlags().BoolVar(&updateAsync, "async", false, "Run the update operation asynchronously")
}
