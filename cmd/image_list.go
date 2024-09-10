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
        "context"
        "google.golang.org/grpc/metadata"
	"github.com/spf13/cobra"
)

var (
	imgLimit int32
	imgFilter              []string
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List Images",
	RunE: func(command *cobra.Command, args []string) error {
                ctx, cancel := context.WithCancel(command.Context())
                defer cancel()
                ctx = metadata.AppendToOutgoingContext(ctx, "username","cisco", "password", "cisco123")
                ch, err := containerzClient.ListImage(ctx, imgLimit, imgFilter)
		if err != nil {
			return err
		}
                //fmt.Printf("Filters: %v\n", filter)
		writer := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', tabwriter.AlignRight)
		fmt.Fprint(writer, "ID\tImage\tTag\n")
		defer writer.Flush()
		for info := range ch {
			if info.Error != nil {
				return info.Error
			}
			fmt.Fprintf(writer, "%s\t%s\t%s\n", info.ID[:5], info.ImageName, info.Tag)
		}

		return nil
	},
}

func init() {
	imageCmd.AddCommand(listCmd)

	listCmd.PersistentFlags().Int32Var(&imgLimit, "limit", -1, "number of containers to return")
	listCmd.PersistentFlags().StringArrayVarP(&imgFilter, "filter", "f", []string{}, "filter to apply" )
}
