package main

import (
	"fmt"
	"strings"

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
		var builder strings.Builder
		infraFile, err := util.ReadFile("infra.txt")
		if err != nil {
			return err
		}

		projects, err := app.sp.GetProjects()
		if err != nil {
			return err
		}

		for _, project := range projects {
			fmt.Fprintf(&builder, "%s\n", project.Namespace)
			services, err := app.sp.GetServicesByID(project.ID)
			if err != nil {
				return err
			}

			for _, service := range services {
				fmt.Fprintf(&builder, "%s %s\n", service.ResourceName, project.Namespace)
				environments, err := app.sp.GetEnvironments(service.ID)
				if err != nil {
					return err
				}

				environmentsString := util.ToEnvString(environments)

				if err := app.dsvc.Create(deploy.Deployment{
					Namespace: project.Namespace,
					Name:      service.ResourceName,
					Hostname:  service.PublicDomain,
					Env:       environmentsString,
					Image:     service.Image,
					Port:      service.Port,
				}); err != nil {
					return err
				}
			}
		}

		removed, added := util.CompareDiff(infraFile, builder.String())

		for _, id := range added {
			fmt.Println("+", id)
		}

		for _, id := range removed {
			fmt.Println("-", id)
			resource := strings.Split(id, "-")

			switch resource[0] {
			case "proj":
				err := app.dsvc.Delete(resource[0])
				if err != nil {
					return err
				}
			case "svc":
				split := strings.Split(id, " ")
				err := app.dsvc.DeleteService(deploy.Deployment{
					Name:      split[0],
					Namespace: split[1],
				})
				if err != nil {
					return err
				}
			}
		}
		return util.WriteFile(builder.String(), "infra.txt")
	},
}
