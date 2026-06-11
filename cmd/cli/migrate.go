package main

import (
	"github.com/spf13/cobra"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "migrate infra to other machines",
	Args:  cobra.ArbitraryArgs,
	Run:   migrateRun,
}

func migrateRun(cmd *cobra.Command, args []string) {
	defer app.db.Close()
	if len(args) != 1 {
		cmd.Usage()
		return
	}

	name := args[0]
	if name == "" {
		return
	}

}
