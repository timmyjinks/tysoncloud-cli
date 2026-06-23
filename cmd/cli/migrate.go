package main

import (
	"github.com/spf13/cobra"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "migrate infra to other machines",
	Args:  cobra.NoArgs,
	Annotations: map[string]string{
		"deploy":   "true",
		"supabase": "true",
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		deployments, err := app.sp.GetDeployments()
		if err != nil {
			return err
		}

		projects, err := app.sp.GetProjects()
		if err != nil {
			return err
		}

		projectMap := make(map[string]bool)
		for _, project := range projects {
			projectMap[project.ID] = true
		}

		services, err := app.sp.GetServices()
		if err != nil {
			return err
		}

		serviceMap := make(map[string]bool)
		for _, service := range services {
			serviceMap[service.ProjectId] = true
		}

		for _, deployment := range deployments {
			if deployment.Type != "docker" {
				continue
			}

			if !projectMap[deployment.ID] {
				if err := app.sp.CreateProject(deployment.ID, deployment.UserID, deployment.Name); err != nil {
					return err
				}
			}

			if !serviceMap[deployment.ID] {
				if err := app.sp.CreateService(deployment.ID, deployment.Name, deployment.Source, deployment.Status); err != nil {
					return err
				}
			}
		}

		return nil
	},
}
