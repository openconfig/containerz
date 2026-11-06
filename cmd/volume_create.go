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
		opts, err := parseKeyValues("option", options)
		if err != nil {
			return err
		}

		lbls, err := parseKeyValues("label", labels)
		if err != nil {
			return err
		}

		resp, err := containerzClient.CreateVolume(command.Context(), name, driver, lbls, opts)
		if err != nil {
			return err
		}

		fmt.Printf("Volume %q created!\n", resp)
		return nil
	},
}

// parseKeyValues parses a list of key=value entries into a map. Only the first
// "=" separates the key from the value, so values may contain "=" and ",".
// kind names the entry type in error messages.
func parseKeyValues(kind string, entries []string) (map[string]string, error) {
	kvs := map[string]string{}
	for _, e := range entries {
		key, value, ok := strings.Cut(e, "=")
		if !ok || key == "" {
			return nil, fmt.Errorf("invalid %s %q: expected key=value", kind, e)
		}
		kvs[key] = value
	}
	return kvs, nil
}

func init() {
	volumesCmd.AddCommand(volCreateCmd)

	volCreateCmd.PersistentFlags().StringVar(&name, "name", "", "Name of the volume to create.")
	volCreateCmd.PersistentFlags().StringVar(&driver, "driver", "", "Type of driver to use to create the volume.")
	volCreateCmd.PersistentFlags().StringArrayVarP(&options, "options", "o", []string{}, "Option to pass to the driver in the form key=value. Repeat for multiple options, e.g. -o options=rbind,ro -o mountpoint=/data")
	volCreateCmd.PersistentFlags().StringArrayVarP(&labels, "labels", "l", []string{}, "Label to tag the volume with in the form key=value. Repeat for multiple labels.")
}
