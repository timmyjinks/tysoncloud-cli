package main

import (
	"github.com/spf13/cobra"
	"github.com/timmyjinks/tysoncloud-cli/deploy"
	"github.com/timmyjinks/tysoncloud-cli/util"
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
				environments, err := app.sp.GetEnvironments(service.ID)
				if err != nil {
					return err
				}

				environmentsMap := util.ToEnvString(environments)

				if err := app.dsvc.Create(deploy.Deployment{
					Namespace: project.Namespace,
					Name:      service.K8sName,
					Hostname:  service.Hostname,
					Env:       environmentsMap,
					Image:     service.Image,
					Port:      service.Port,
				}); err != nil {
					return err
				}
			}
		}
		return nil
	},
}
