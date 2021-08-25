package cmd

import (
	"github.com/spf13/cobra"
	"github.com/tenyo/brbr/lib/brbrserver"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop a brbr server running in the background",
	Long:  `Gracefully shuts down a daemonized brbr server`,
	Run: func(cmd *cobra.Command, args []string) {
		brbrserver.Stop()
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
