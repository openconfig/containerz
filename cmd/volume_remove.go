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

	"github.com/spf13/cobra"
)

var volRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove volumes",
	RunE: func(command *cobra.Command, args []string) error {
		if err := containerzClient.RemoveVolume(command.Context(), name, force); err != nil {
			return err
		}

		fmt.Printf("Successfully removed volume %q\n", name)
		return nil
	},
}

func init() {
	volumesCmd.AddCommand(volRemoveCmd)

	volRemoveCmd.PersistentFlags().StringVar(&name, "name", "", "Name of the volume to remove.")
	volRemoveCmd.PersistentFlags().BoolVar(&force, "force", false, "Force removal of the volume.")
}
