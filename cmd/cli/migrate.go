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

		for _, deployment := range deployments {
			if deployment.Type != "docker" {
				continue
			}
			projectId, err := uuid.NewRandom()
			if err != nil {
				return err
			}

			serviceId, err := uuid.NewRandom()
			if err != nil {
				return err
			}

			if err := app.sp.CreateProject(projectId.String(), deployment.UserID, deployment.Name); err != nil {
				return err
			}

			if err := app.dsvc.CreateProject("proj-" + projectId.String()); err != nil {
				log.Println(err)
			}

			if err := app.sp.CreateService(serviceId.String(), projectId.String(), deployment.Name, deployment.Source); err != nil {
				return err
			}

			if err := app.dsvc.Create(deploy.Deployment{
				Namespace: "proj-" + projectId.String(),
				Name:      "svc-" + serviceId.String(),
				Hostname:  "tc-" + serviceId.String(),
				Env:       map[string][]byte{},
				Image:     deployment.Source,
				Port:      3000,
			}); err != nil {
				log.Println(err)
			}
		}

		return nil
	},
}
