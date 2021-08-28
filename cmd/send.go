package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tenyo/brbr/lib/brbrclient"
)

var sendCmd = &cobra.Command{
	Use:   "send [address]",
	Short: "Send a metagram to a brbr server",
	Long: `Send a metagram to a listening brbr onion server at the specified address.
Will prompt the user to enter a message at the command line. Alternatively,
messages can be piped into this command.`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("unexpected number of arguments: %d", len(args))
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		brbrclient.Send(dataDir, args[0])
	},
}

func init() {
	rootCmd.AddCommand(sendCmd)
}
