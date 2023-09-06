package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/briandowns/spinner"
)

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

		ch, err := containerzClient.PullImage(command.Context(), image, tag, nil)
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
}
