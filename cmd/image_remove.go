package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"google3/third_party/openconfig/containerz/client/client"
)

var removeCmd = &cobra.Command{
	Use:   "remove",
	Short: "Removes the image from the containerz server",
	RunE: func(command *cobra.Command, args []string) error {
		err := containerzClient.Remove(command.Context(), image, tag)
		switch err {
		case nil:
			fmt.Printf("Image %s:%s has been removed.\n", image, tag)
		case client.ErrNotFound:
			fmt.Printf("Image %s:%s does not exist in containerz.\n", image, tag)
		case client.ErrRunning:
			fmt.Printf("Image %s:%s has a container running; use force option or stop the container.\n", image, tag)
		default:
			return err
		}
		return nil
	},
}

func init() {
	imageCmd.AddCommand(removeCmd)
}
