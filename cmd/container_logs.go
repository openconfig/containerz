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
	follow bool
)

var cntLogCmd = &cobra.Command{
	Use:   "logs",
	Short: "fetch the logs for a container",
	RunE: func(command *cobra.Command, args []string) error {
		if instance == "" {
			fmt.Println("--instance must be provided")
		}

                ctx, cancel := context.WithCancel(command.Context())
                defer cancel()
                ctx = metadata.AppendToOutgoingContext(ctx, "username","cisco", "password", "cisco123")
		ch, err := containerzClient.Logs(ctx, instance, follow)
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
