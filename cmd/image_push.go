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
	"time"

	"github.com/spf13/cobra"
	"github.com/briandowns/spinner"
)

var (
	file     string
	isPlugin bool
)

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push a local image tarball to the containerz server.",
	Long:  "Push the result of 'docker save' to the containerz server.",
	RunE: func(command *cobra.Command, args []string) error {
		if file == "" {
			return fmt.Errorf("--file cannot be empty")
		}

		output := image + "/" + tag
		if image == "" {
			output = "image"
		}

		ch, err := containerzClient.PushImage(command.Context(), image, tag, file, isPlugin)
		if err != nil {
			return err
		}

		s := spinner.New(spinner.CharSets[69], 100*time.Millisecond)
		s.Start()
		defer s.Stop()
		s.Prefix = fmt.Sprintf("Pushing %s ", output)
		s.Suffix = " 0"

		for prog := range ch {
			if prog.Error != nil {
				return prog.Error
			}

			if prog.Finished {
				s.FinalMSG = fmt.Sprintf("Pushed %s/%s\n", prog.Image, prog.Tag)
			} else {
				s.Suffix = fmt.Sprintf(" %d", prog.BytesReceived)
			}
		}

		return nil
	},
}

func init() {
	imageCmd.AddCommand(pushCmd)
	pushCmd.PersistentFlags().StringVar(&file, "file", "", "Image tar to upload.")
	pushCmd.PersistentFlags().BoolVar(&isPlugin, "is_plugin", false, "If set to true, a plugin will be uploaded rather than loading the image into the container runtime")
}
