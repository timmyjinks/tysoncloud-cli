package main

import (
	"fmt"
	"log"
	"path"
	"strings"
	"sync"

	"github.com/spf13/cobra"
	"github.com/timmyjinks/tysoncloud-cli/deploy"
	"github.com/timmyjinks/tysoncloud-cli/store"
	"github.com/timmyjinks/tysoncloud-cli/util"
	"golang.org/x/sync/errgroup"
)

type infraState struct {
	projects          []store.ProjectsTable
	services          []store.ServicesTable
	environments      []store.EnvironmentsTable
	volumes           []store.VolumesTable
	servicesMap       map[string][]store.ServicesTable
	servicesByNameMap map[string]store.ServicesTable
	volumesMap        map[string]*store.VolumesTable
	environmentsMap   map[string][]store.EnvironmentsTable
}

var projectDir = fmt.Sprintf("%s/.local/share/tysoncloud", util.GetEnv("HOME", "./diff.txt"))
var diffFileLocation = path.Join(projectDir, "diff.txt")
var wg sync.WaitGroup

var force bool

var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "pull infra to current machine",
	Annotations: map[string]string{
		"supabase": "true",
		"deploy":   "true",
	},
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		err := util.EnsureDirExists(projectDir)
		if err != nil {
			return err
		}

		state, err := fetchAll()
		if err != nil {
			return err
		}

		diffFile, err := util.ReadFile(diffFileLocation)
		if err != nil {
			return err
		}
		newDiffFile := getCurrentState(state)

		deletions, additions := util.CompareDiff(diffFile, newDiffFile)
		util.PrintDiff(deletions, additions)

		for _, d := range deletions {
			wg.Go(func() {
				err := destroy(d)
				if err != nil {
					log.Println(err)
				}
			})
		}

		targets := additions
		if force {
			targets = strings.Split(strings.TrimRight(newDiffFile, "\n"), "\n")
		}

		for _, t := range targets {
			wg.Go(func() {
				err := create(t, state)
				if err != nil {
					log.Println(err)
				}
			})
		}

		wg.Wait()

		return util.WriteFile(newDiffFile, diffFileLocation)
	},
}

func fetchAll() (*infraState, error) {
	var (
		eg           errgroup.Group
		projects     []store.ProjectsTable
		services     []store.ServicesTable
		environments []store.EnvironmentsTable
		volumes      []store.VolumesTable
	)

	eg.Go(func() (err error) { projects, err = app.sp.GetProjects(); return })
	eg.Go(func() (err error) { services, err = app.sp.GetServices(); return })
	eg.Go(func() (err error) { volumes, err = app.sp.GetVolumes(); return })
	eg.Go(func() (err error) { environments, err = app.sp.GetEnvironments(); return })

	if err := eg.Wait(); err != nil {
		return nil, fmt.Errorf("fetching remote state: %w", err)
	}

	servicesMap := make(map[string][]store.ServicesTable)
	servicesByNameMap := make(map[string]store.ServicesTable)
	for _, service := range services {
		servicesMap[service.ProjectId] = append(servicesMap[service.ProjectId], service)
		servicesByNameMap[service.ResourceName] = service
	}

	environmentsMap := make(map[string][]store.EnvironmentsTable)
	for _, environment := range environments {
		environmentsMap[environment.ServiceId] = append(environmentsMap[environment.ServiceId], environment)
	}

	volumesMap := make(map[string]*store.VolumesTable)
	for _, volume := range volumes {
		volumesMap[volume.ServiceId] = &volume
	}

	return &infraState{
		projects:          projects,
		services:          services,
		environments:      environments,
		volumes:           volumes,
		servicesMap:       servicesMap,
		servicesByNameMap: servicesByNameMap,
		volumesMap:        volumesMap,
		environmentsMap:   environmentsMap,
	}, nil
}

func getCurrentState(s *infraState) string {
	var currentState strings.Builder

	for _, project := range s.projects {
		fmt.Fprintln(&currentState, project.Namespace)
		services := s.servicesMap[project.ID]
		for _, service := range services {
			environments := util.ToEnvString(s.environmentsMap[service.ID])
			volume := s.volumesMap[service.ID]

			if volume == nil {
				fmt.Fprintln(&currentState, service.ResourceName, project.Namespace, service.Port, service.Image, environments)
			} else {
				fmt.Fprintln(&currentState, service.ResourceName, project.Namespace, service.Port, service.Image, volume.MountPath, volume.StorageGB, environments)
			}
		}
	}
	return currentState.String()
}

func create(id string, state *infraState) error {
	fields := strings.Fields(id)
	resource, _, _ := strings.Cut(fields[0], "-")

	switch resource {
	case "proj":
		err := app.dsvc.CreateProject(id)
		if err != nil {
			return err
		}
	case "svc":
		name, namespace := fields[0], fields[1]

		service := state.servicesByNameMap[name]
		volume := state.volumesMap[service.ID]

		var v *deploy.Volume = nil
		if volume != nil {
			v = &deploy.Volume{
				MountPath: volume.MountPath,
				StorageGB: volume.StorageGB,
			}
		}

		environments := state.environmentsMap[service.ID]

		err := app.dsvc.Create(deploy.Deployment{
			Namespace: namespace,
			Name:      name,
			Hostname:  service.PublicDomain,
			Image:     service.Image,
			Port:      service.Port,
			Volume:    v,
			Env:       util.ToEnvMap(environments),
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func destroy(id string) error {
	fields := strings.Fields(id)
	resource := strings.Split(fields[0], "-")[0]

	switch resource {
	case "proj":
		err := app.dsvc.Delete(id)
		if err != nil {
			return err
		}
	case "svc":
		name, namespace := fields[0], fields[1]

		err := app.dsvc.DeleteService(deploy.Deployment{
			Namespace: namespace,
			Name:      name,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func init() {
	pullCmd.Flags().BoolVarP(&force, "force", "f", false, "force pull all")
}
