package main

import (
	"log"

	"github.com/spf13/cobra"
)

var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "pull infra to current machine",
	Annotations: map[string]string{
		"supabase": "true",
		"deploy":   "true",
	},
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		deployments, err := app.sp.GetDeployments()
		if err != nil {
			return err
		}
		for _, deployment := range deployments {
			if deployment.Type == "docker" {
				err := app.dsvc.Create(deployment.Name, deployment.Source, deployment.Port)
				if err != nil {
					log.Println(err)
				}
			}
		}
		return nil
	},
}
