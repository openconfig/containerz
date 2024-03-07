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
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
)

var volListCmd = &cobra.Command{
	Use:   "list",
	Short: "List volumes",
	RunE: func(command *cobra.Command, args []string) error {
		ch, err := containerzClient.ListVolume(command.Context(), nil)
		if err != nil {
			return err
		}

		writer := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', tabwriter.AlignRight)
		fmt.Fprint(writer, "Name\tDriver\tOptions\tLabels\tCreation Time\n")
		defer writer.Flush()
		for info := range ch {
			if info.Error != nil {
				return info.Error
			}
			fmt.Fprintf(writer, "%s\t%s\t%v\t%v\t%s\n", info.Name, info.Driver, info.Options, info.Labels, info.CreationTime.Format(time.RFC822))
		}

		return nil
	},
}

func init() {
	volumesCmd.AddCommand(volListCmd)
}
