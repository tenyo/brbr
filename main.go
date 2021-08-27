package main

import "github.com/tenyo/brbr/cmd"

var (
	version = "v0.0.0"
	commit  = "unset"
	date    = "unset"
)

func main() {
	cmd.Ver = &cmd.CmdVersion{
		AppVersion: version,
		GitCommit:  commit,
		BuildTime:  date,
	}

	cmd.Execute()
}
