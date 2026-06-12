package main

import (
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "get resources",
}

func init() {
	getCmd.AddCommand(getServerCmd)
}
