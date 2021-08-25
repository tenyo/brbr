package cmd

import (
	"github.com/spf13/cobra"
	"github.com/tenyo/brbr/lib/brbrserver"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run a brbr server in the foreground",
	Long: `Start an onion server in the foreground to listen for incoming connections from 
brbr senders, receive and save metagrams to local disk`,
	Run: func(cmd *cobra.Command, args []string) {
		brbrserver.Start()
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
