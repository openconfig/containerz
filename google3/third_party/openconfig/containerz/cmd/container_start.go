package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"/client/client"
)

var (
	cntCommand, instance string
	ports                []string
	envs                 []string
)

var cntStartCmd = &cobra.Command{
	Use:   "start",
	Short: "start a container wirth the specified image and tag",
	RunE: func(command *cobra.Command, args []string) error {
		if image == "" {
			return fmt.Errorf("--image must be specified")
		}

		id, err := containerzClient.Start(command.Context(), image, tag, cntCommand, instance, client.WithEnv(envs), client.WithPorts(ports))
		if err != nil {
			return err
		}

		fmt.Printf("Container started with id - %s\n", id)
		return nil
	},
}

func init() {
	containerCmd.AddCommand(cntStartCmd)

	cntStartCmd.PersistentFlags().StringVar(&cntCommand, "command", "/bin/bash", "command to run.")
	cntStartCmd.PersistentFlags().StringVar(&instance, "instance", "", "Name to give to the container.")
	cntStartCmd.PersistentFlags().StringArrayVar(&ports, "port", []string{}, "Ports to expose (format: <internal_port>:<external_port>")
	cntStartCmd.PersistentFlags().StringArrayVar(&envs, "env", []string{}, "Environment vars to set (format: <VAR_NAMEt>=<VAR_VALUE>")
}
