package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "pull infra to current machine",
	Annotations: map[string]string{
		"supabase": "true",
	},
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		deployments, err := app.sp.GetDeployments()
		if err != nil {
			return err
		}
		fmt.Println(deployments)

		// possible issues with env
		return nil
	},
}
