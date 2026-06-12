package main

import "github.com/spf13/cobra"

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "add resource",
}

func init() {
	addCmd.AddCommand(addServerCmd)
}
