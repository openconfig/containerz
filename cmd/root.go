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

// Package cmd contains all the commands to run and operate with containerz.
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	addr         string
	grpcMetadata map[string]string
)

// RootCmd is the cmd entrypoint for all containerz commands.
var RootCmd = &cobra.Command{
	Use:   "containerz",
	Short: "Containerz suite of CLI tools",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
	},
	RunE: func(command *cobra.Command, args []string) error {
		command.HelpFunc()(command, args)
		return fmt.Errorf("no command specified")
	},
}

func init() {
	RootCmd.PersistentFlags().StringVar(&addr, "addr", ":19999", "Containerz listen port.")
	RootCmd.PersistentFlags().StringToStringVar(&grpcMetadata, "grpc_metadata", nil, "gRPC metadata to attach to all outgoing requests.")
}
