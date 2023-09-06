// Binary containerz is the entry point for all container based operations.
package main

import (
	"context"
	"os"

	"/cmd/cmd"
)

func main() {
	if err := cmd.RootCmd.ExecuteContext(context.Background()); err != nil {
		// no need to report error; cobra already did
		os.Exit(1)
	}
}
