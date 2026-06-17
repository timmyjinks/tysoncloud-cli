package main

import (
	"github.com/spf13/cobra"
)

var getDeploymentCmd = &cobra.Command{
	Use:   "deployments",
	Short: "get deployments",
	Annotations: map[string]string{
		"deploy": "true",
	},
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		app.dsvc.Get()
		return nil
	},
}

var deleteDeploymentCmd = &cobra.Command{
	Use:   "deployment [name]",
	Short: "get deployments",
	Annotations: map[string]string{
		"deploy": "true",
	},
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		err := app.dsvc.Delete(name)
		if err != nil {
			return err
		}
		return nil
	},
}
