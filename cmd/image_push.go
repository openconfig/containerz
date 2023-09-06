package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/briandowns/spinner"
)

var file string

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

		ch, err := containerzClient.PushImage(command.Context(), image, tag, file)
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
}
