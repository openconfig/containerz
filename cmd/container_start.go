// Copyright 2023 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
        "context"
        "google.golang.org/grpc/metadata"

	"github.com/spf13/cobra"
	"github.com/openconfig/containerz/client"
)

var (
	cntCommand, instance string
	ports                []string
	envs                 []string
	volumes              []string
)

var cntStartCmd = &cobra.Command{
	Use:   "start",
	Short: "start a container wirth the specified image and tag",
	RunE: func(command *cobra.Command, args []string) error {
		if image == "" {
			return fmt.Errorf("--image must be specified")
		}
                ctx, cancel := context.WithCancel(command.Context())
                defer cancel()
                ctx = metadata.AppendToOutgoingContext(ctx, "username","cisco", "password", "cisco123")

		id, err := containerzClient.StartContainer(ctx, image, tag, cntCommand, instance, client.WithEnv(envs), client.WithPorts(ports), client.WithVolumes(volumes))
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
	cntStartCmd.PersistentFlags().StringArrayVarP(&volumes, "volume", "v", []string{}, "Volumes to attach to the container (format: <volume-name>:<mountpoint>[:ro])")
}
