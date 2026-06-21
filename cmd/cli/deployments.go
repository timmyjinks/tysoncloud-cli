package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var getDeploymentCmd = &cobra.Command{
	Use:   "deployments",
	Short: "get deployments",
	Annotations: map[string]string{
		"supabase": "true",
	},
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		deployments, err := app.sp.GetDeployments()
		if err != nil {
			return err
		}

		for _, deployment := range deployments {
			fmt.Printf("%s %s %s\n", deployment.ID, deployment.UserID, deployment.Name)
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
	Args: cobra.MaximumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		userId := args[0]
		name := args[1]
		err := app.dsvc.Delete(userId + "-" + name)
		if err != nil {
			return err
		}
		return nil
	},
}
