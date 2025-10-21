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
	"strings"

	"github.com/spf13/cobra"
)

var (
	name    string
	driver  string
	options []string
	labels  []string
)

var volCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create volumes",
	RunE: func(command *cobra.Command, args []string) error {
		opts := map[string]string{}
		for _, o := range options {
			parts := strings.SplitN(o, "=", 2)
			opts[parts[0]] = parts[1]
		}

		lbls := map[string]string{}
		for _, l := range labels {
			parts := strings.SplitN(l, "=", 2)
			lbls[parts[0]] = parts[1]
		}

		resp, err := containerzClient.CreateVolume(command.Context(), name, driver, lbls, opts)
		if err != nil {
			return err
		}

		fmt.Printf("Volume %q created!\n", resp)
		return nil
	},
}

func init() {
	volumesCmd.AddCommand(volCreateCmd)

	volCreateCmd.PersistentFlags().StringVar(&name, "name", "", "Name of the volume to create.")
	volCreateCmd.PersistentFlags().StringVar(&driver, "driver", "", "Type of driver to use to create the volume.")
	volCreateCmd.PersistentFlags().StringSliceVarP(&options, "options", "o", []string{}, "Options to pass to the driver in the form k1=v1,k2=v2,...")
	volCreateCmd.PersistentFlags().StringSliceVarP(&labels, "labels", "l", []string{}, "Labels to tag. the volume with, in the form k1=v1")
}
