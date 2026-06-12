package main

import "github.com/spf13/cobra"

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "add resource",
	Annotations: map[string]string{
		"database": "true",
	},
}

func init() {
	addCmd.AddCommand(addServerCmd)
}
