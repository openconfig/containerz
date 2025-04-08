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
	"github.com/openconfig/containerz/client"
)

var (
	forceImgRemoval bool
)
var removeCmd = &cobra.Command{
	Use:   "remove",
	Short: "Removes the image from the containerz server",
	RunE: func(command *cobra.Command, args []string) error {
		err := containerzClient.RemoveImage(command.Context(), image, tag, forceImgRemoval)
		switch err {
		case nil:
			fmt.Printf("Image %s:%s has been removed.\n", image, tag)
		case client.ErrNotFound:
			return fmt.Errorf("Image %s:%s does not exist in containerz.\n", image, tag)
		case client.ErrRunning:
			return fmt.Errorf("Image %s:%s has a container running; use force option or stop the container.\n", image, tag)
		default:
			return err
		}
		return nil
	},
}

func init() {
	imageCmd.AddCommand(removeCmd)
	removeCmd.PersistentFlags().BoolVarP(&forceImgRemoval, "force", "f", false, "Force Image removal.")
}
