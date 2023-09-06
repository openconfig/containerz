package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var (
	all   bool
	limit int32
)

var cntListCmd = &cobra.Command{
	Use:   "list",
	Short: "List containers",
	RunE: func(command *cobra.Command, args []string) error {
		ch, err := containerzClient.List(command.Context(), all, limit, nil)
		if err != nil {
			return err
		}

		writer := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', tabwriter.AlignRight)
		fmt.Fprint(writer, "ID\tName\tImage\tState\n")
		defer writer.Flush()
		for info := range ch {
			if info.Error != nil {
				return info.Error
			}
			fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n", info.ID[:5], info.Name, info.ImageName, info.State)
		}

		return nil
	},
}

func init() {
	containerCmd.AddCommand(cntListCmd)

	cntListCmd.PersistentFlags().BoolVar(&all, "all", false, "Return all containers.")
	cntListCmd.PersistentFlags().Int32Var(&limit, "limit", -1, "number of containers to return")
}
