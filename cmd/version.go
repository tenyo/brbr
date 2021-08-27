package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

type CmdVersion struct {
	AppVersion string
	GitCommit  string
	BuildTime  string
}

var Ver *CmdVersion

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show the current version",
	Long:  `Displays version and build information`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("brbr version: %s\ngit commit: %s\nbuildtime: %s\n", Ver.AppVersion, Ver.GitCommit, Ver.BuildTime)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
