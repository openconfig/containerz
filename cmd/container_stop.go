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
)

var (
	force bool
	restart bool
)

var cntStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "stop a container by intance name",
	RunE: func(command *cobra.Command, args []string) error {
		if instance == "" {
			fmt.Println("--instance must be provided")
		}
                ctx, cancel := context.WithCancel(command.Context())
                defer cancel()
                ctx = metadata.AppendToOutgoingContext(ctx, "username","cisco", "password", "cisco123")

		if err := containerzClient.StopContainer(ctx, instance, force,restart); err != nil {
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
	cntStopCmd.PersistentFlags().BoolVar(&restart, "restart", false, "restart the container.")
}
