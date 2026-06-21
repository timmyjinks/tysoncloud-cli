package main

import (
	"github.com/spf13/cobra"
	"github.com/timmyjinks/tysoncloud-cli/deploy"
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
		projects, err := app.sp.GetProjects()
		if err != nil {
			return err
		}

		for _, project := range projects {
			services, err := app.sp.GetServicesByID(project.ID)
			if err != nil {
				return err
			}

			for _, service := range services {
				err := app.dsvc.Create(deploy.Deployment{
					Namespace: project.Namespace,
					Name:      service.Name,
					Image:     service.Image,
					Port:      service.Port,
				})
				if err != nil {
					return err
				}
			}
		}
		return nil
	},
}
