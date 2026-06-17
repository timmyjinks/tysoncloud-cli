package main

import "github.com/spf13/cobra"

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "update resources",
}

func init() {
	updateCmd.AddCommand(updateServerCmd)
}
