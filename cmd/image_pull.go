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

        "context"
        "google.golang.org/grpc/metadata"
	"github.com/spf13/cobra"
	"github.com/briandowns/spinner"
)
var path string
var vrf string
var proto string

var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pull the specified container image",
	RunE: func(command *cobra.Command, args []string) error {
		if image == "" {
			return fmt.Errorf("--image must be specified")
		}

		s := spinner.New(spinner.CharSets[69], 100*time.Millisecond)
		s.Start()
		defer s.Stop()
		s.Prefix = fmt.Sprintf("Pulling %s/%s ", image, tag)
		s.Suffix = " 0"
		s.FinalMSG = fmt.Sprintf("Pulled %s/%s\n", image, tag)

                ctx, cancel := context.WithCancel(command.Context())
                defer cancel()
                ctx = metadata.AppendToOutgoingContext(ctx, "username","cisco", "password", "cisco123")
		ch, err := containerzClient.PullImage(ctx, image, tag, path, vrf, proto, nil)
		if err != nil {
			return err
		}

		for progress := range ch {
			s.Suffix = fmt.Sprintf(" %d", progress.BytesReceived)
		}

		return nil
	},
}

func init() {
	imageCmd.AddCommand(pullCmd)
	pullCmd.PersistentFlags().StringVar(&path, "path", "192.168.122.1:8080/vrf-relay.tar.gz", "Image tar path to download.")
        pullCmd.PersistentFlags().StringVar(&vrf, "vrf", "", "vrf to used")
        pullCmd.PersistentFlags().StringVar(&proto, "protocol", "http", "protocol to use : default is http")
}
