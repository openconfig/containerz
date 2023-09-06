package cmd

import (
	"context"
	"os"
	"os/signal"

	"github.com/spf13/cobra"
	"google3/third_party/golang/github_com/moby/moby/v/v24/client/client"
	"/containers/docker/docker"
	"/server/server"
)

var (
	dockerHost string
	chunkSize  int
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Launch the containerz server",
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
		mgr := docker.New(cli)
		s := server.New(docker.New(cli), opts...)

		// listen for ctrl-c
		interrupt := make(chan os.Signal, 1)
		signal.Notify(interrupt, os.Interrupt)
		go func() {
			<-interrupt // wait for signal
			cancel()
			s.Halt(ctx)
			mgr.Stop()
		}()

		return s.Serve(ctx)
	},
}

func init() {
	RootCmd.AddCommand(startCmd)
	startCmd.PersistentFlags().StringVar(&dockerHost, "docker_host", "unix:///var/run/docker.sock", "Docker host to connect to.")
	startCmd.PersistentFlags().IntVar(&chunkSize, "chunk_size", 64000, "the size of the chunks supported by this server")
}
