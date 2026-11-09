// Copyright 2025 Google LLC
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

	"github.com/spf13/cobra"
)

var (
	configFile string
)

var pluginStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a plugin",
	RunE: func(command *cobra.Command, args []string) error {
		if instance == "" {
			return fmt.Errorf("--instance must be provided")
		}

		if name == "" {
			return fmt.Errorf("--name must be provided")
		}

		if configFile == "" {
			return fmt.Errorf("--config must be provided")
		}

		if err := containerzClient.StartPlugin(command.Context(), name, instance, configFile); err != nil {
			return err
		}

		fmt.Printf("Successfully started %s\n", instance)
		return nil
	},
}

var pluginRestartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart a plugin",
	RunE: func(command *cobra.Command, args []string) error {
		if instance == "" {
			return fmt.Errorf("--instance must be provided")
		}

		if err := containerzClient.StartPlugin(command.Context(), name, instance, configFile); err != nil {
			return err
		}

		fmt.Printf("Successfully restarted %s\n", instance)
		return nil
	},
}

func init() {
	pluginCmd.AddCommand(pluginStartCmd)
	pluginStartCmd.PersistentFlags().StringVar(&instance, "instance", "", "plugin instance name")
	pluginStartCmd.PersistentFlags().StringVar(&name, "name", "", "plugin name")
	pluginStartCmd.PersistentFlags().StringVar(&configFile, "config", "", "plugin config file")

	pluginCmd.AddCommand(pluginRestartCmd)
	pluginRestartCmd.PersistentFlags().StringVar(&instance, "instance", "", "plugin instance name")
}
