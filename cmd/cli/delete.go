package main

import "github.com/spf13/cobra"

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "delete resource",
}

func init() {
	deleteCmd.AddCommand(deleteServerCmd)
	deleteCmd.AddCommand(deleteDeploymentCmd)
}
