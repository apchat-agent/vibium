package main

import (
	"github.com/spf13/cobra"
)

func newQuitCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "close",
		Aliases: []string{"quit"},
		Short:   "Close the browser session",
		Example: `  vibium close
  # Close the browser (daemon keeps running)`,
		Args: cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			result, err := daemonCall("browser_quit", map[string]interface{}{})
			if err != nil {
				printError(err)
				return
			}
			printResult(result)
		},
	}
}
