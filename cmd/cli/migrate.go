package main

import (
	"log"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/timmyjinks/tysoncloud-cli/deploy"
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

				if err := app.dsvc.CreateProject("proj-" + deployment.ID); err != nil {
					log.Println(err)
				}
			}

			if !serviceMap[deployment.ID] {
				serviceId, err := uuid.NewRandom()
				if err != nil {
					return err
				}

				if err := app.sp.CreateService(serviceId.String(), deployment.ID, deployment.Name, deployment.Source, deployment.Status); err != nil {
					return err
				}

				if err := app.dsvc.Create(deploy.Deployment{
					Namespace: "proj-" + deployment.ID,
					Name:      "svc-" + serviceId.String(),
					Hostname:  "tc-" + serviceId.String(),
					Env:       map[string][]byte{},
					Image:     deployment.Source,
					Port:      3000,
				}); err != nil {
					log.Println(err)
				}
			}
		}

		return nil
	},
}
