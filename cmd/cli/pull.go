package main

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/spf13/cobra"
	"github.com/timmyjinks/tysoncloud-cli/deploy"
	"github.com/timmyjinks/tysoncloud-cli/store"
	"github.com/timmyjinks/tysoncloud-cli/util"
)

var projectDir = fmt.Sprintf("%s/.local/share/tysoncloud", util.GetEnv("HOME", "./diff.txt"))
var diffFileLocation = path.Join(projectDir, "diff.txt")

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
		if _, err := os.Stat(projectDir); errors.Is(err, os.ErrNotExist) {
			err := os.MkdirAll(projectDir, 0755)
			if err != nil {
				return err
			}
		}

		diffFile, err := util.ReadFile(diffFileLocation)
		if err != nil {
			return err
		}

		projects, err := app.sp.GetProjects()
		if err != nil {
			return err
		}

		services, err := app.sp.GetServices()
		if err != nil {
			return err
		}

		servicesMap := make(map[string][]store.ServicesTable)
		for _, service := range services {
			servicesMap[service.ProjectId] = append(servicesMap[service.ProjectId], service)
		}

		environments, err := app.sp.GetEnvironments()
		if err != nil {
			return err
		}

		environmentsMap := make(map[string][]store.EnvironmentsTable)
		for _, environment := range environments {
			environmentsMap[environment.ServiceId] = append(environmentsMap[environment.ServiceId], environment)
		}

		for _, project := range projects {
			fmt.Fprintf(&builder, "%s\n", project.Namespace)

			services := servicesMap[project.ID]
			for _, service := range services {
				environment := environmentsMap[service.ID]
				environmentString := util.ToEnvString(environment)
				fmt.Fprintf(&builder, "%s %s %d %s %s\n", service.ResourceName, project.Namespace, service.Port, service.Image, environmentString)
			}
		}

		removed, added := util.CompareDiff(diffFile, builder.String())
		printDiff(added, removed)

		for _, id := range removed {
			resource := strings.Split(id, "-")

			prefix := resource[0]

			switch prefix {
			case "proj":
				err := app.dsvc.Delete(id)
				if err != nil {
					return err
				}
			case "svc":
				serviceParts := strings.Fields(id)
				name, namespace := serviceParts[0], serviceParts[1]

				err := app.dsvc.DeleteService(deploy.Deployment{
					Name:      name,
					Namespace: namespace,
				})
				if err != nil {
					return err
				}
			}
		}

		for _, id := range added {
			resource := strings.Split(id, "-")
			prefix := resource[0]

			switch prefix {
			case "proj":
				if err := app.dsvc.CreateProject(id); err != nil {
					return err
				}
			case "svc":
				serviceParts := strings.Fields(id)
				name, namespace := serviceParts[0], serviceParts[1]

				for _, service := range services {
					if service.ResourceName == name {
						environment := environmentsMap[service.ID]
						environmentMap := util.ToEnvMap(environment)

						if err := app.dsvc.Create(deploy.Deployment{
							Namespace: namespace,
							Name:      name,
							Hostname:  service.PublicDomain,
							Env:       environmentMap,
							Image:     service.Image,
							Port:      service.Port,
						}); err != nil {
							return err
						}
					}
				}
			}
		}

		return util.WriteFile(builder.String(), diffFileLocation)
	},
}

func printDiff(added, removed []string) {
	for _, id := range added {
		fmt.Println("+", id)
	}

	for _, id := range removed {
		fmt.Println("-", id)
	}
}
