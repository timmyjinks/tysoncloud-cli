package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var getDeploymentCmd = &cobra.Command{
	Use:   "projects",
	Short: "get deployments",
	Annotations: map[string]string{
		"supabase": "true",
	},
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		deployments, err := app.sp.GetProjects()
		if err != nil {
			return err
		}

		for _, deployment := range deployments {
			fmt.Printf("%s %s\n", deployment.ID, deployment.Namespace)
		}
		return nil
	},
}

var deleteDeploymentCmd = &cobra.Command{
	Use:   "deployment [userId] [name]",
	Short: "get deployments",
	Annotations: map[string]string{
		"deploy": "true",
	},
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		namespace := args[0]
		err := app.dsvc.Delete(namespace)
		if err != nil {
			return err
		}
		return nil
	},
}
