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
	"context"
	"os"
	"os/signal"

	"github.com/spf13/cobra"
	"google3/third_party/golang/github_com/moby/moby/v/v24/client/client"
	"github.com/openconfig/containerz/containers/docker"
	"github.com/openconfig/containerz/server"
)

var (
	dockerHost string
	chunkSize  int
	useALTS    bool
)

var startCmd = &cobra.Command{
	Use:              "start",
	Short:            "Launch the containerz server",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {},
	RunE: func(command *cobra.Command, args []string) error {
		ctx, cancel := context.WithCancel(command.Context())
		defer cancel()

		cli, err := client.NewClientWithOpts(client.WithHost(dockerHost), client.WithAPIVersionNegotiation())
		if err != nil {
			return err
		}

		opts := []server.Option{
			server.WithAddr(addr),
			server.WithChunkSize(chunkSize),
		}

		if useALTS {
			opts = append(opts, server.UseALTS())
		}

		mgr := docker.New(cli)
		s := server.New(docker.New(cli), opts...)
		mgr.Start(ctx)

		// listen for ctrl-c
		interrupt := make(chan os.Signal, 1)
		signal.Notify(interrupt, os.Interrupt)
		go func() {
			<-interrupt // wait for signal
			cancel()
			s.Halt(ctx)
			mgr.Stop(ctx)
		}()

		return s.Serve(ctx)
	},
}

func init() {
	RootCmd.AddCommand(startCmd)
	startCmd.PersistentFlags().StringVar(&dockerHost, "docker_host", "unix:///var/run/docker.sock", "Docker host to connect to.")
	startCmd.PersistentFlags().IntVar(&chunkSize, "chunk_size", 3000000, "the size of the chunks supported by this server")
	startCmd.PersistentFlags().BoolVar(&useALTS, "use_alts", false, "Use ALTS authentication.")
}
