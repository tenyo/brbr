package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"github.com/tenyo/brbr/lib/brbrserver"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a brbr server in the background",
	Long: `Start and daemonize an onion server to constantly listen for incoming connections from 
brbr senders, receive and save metagrams to local disk`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Starting in background ...")

		path := os.Args[0]
		proc := exec.Command(path, "run")
		proc.Stderr = os.Stderr

		err := proc.Start()
		if err != nil {
			fmt.Printf("error starting brbr server process: %v", err)
		}

		err = brbrserver.SavePid(proc.Process.Pid)
		if err != nil {
			fmt.Printf("error saving brbr server pid file: %v", err)
		}

		fmt.Printf("brbr started and listening in the background (pid %d)\n", proc.Process.Pid)
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}
