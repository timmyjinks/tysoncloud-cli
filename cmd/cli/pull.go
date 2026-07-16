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
	projects           []store.ProjectsTable
	services           []store.ServicesTable
	environments       []store.EnvironmentsTable
	volumes            []store.VolumesTable
	databases          []store.DatabasesTable
	servicesMap        map[string][]store.ServicesTable
	servicesByNameMap  map[string]store.ServicesTable
	databasesMap       map[string][]store.DatabasesTable
	databasesByNameMap map[string]store.DatabasesTable
	volumesMap         map[string]*store.VolumesTable
	environmentsMap    map[string][]store.EnvironmentsTable
}

var projectDir = fmt.Sprintf("%s/.tysoncloud", util.GetEnv("HOME", "./diff.txt"))
var diffFileLocation = path.Join(projectDir, "diff.txt")

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
		var wg sync.WaitGroup

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

		var (
			mu         sync.Mutex
			errorEdits = make(map[string]bool)
		)

		for _, d := range deletions {
			wg.Go(func() {
				err := destroy(d)
				if err != nil {
					mu.Lock()
					errorEdits[strings.Split(d, " ")[0]] = true
					mu.Unlock()
					log.Println(err)
				}
			})
		}

		wg.Wait()

		targets := additions
		if force {
			targets = strings.Split(strings.TrimRight(newDiffFile, "\n"), "\n")
		}

		for _, t := range targets {
			wg.Go(func() {
				err := create(t, state)
				if err != nil {
					mu.Lock()
					errorEdits[strings.Split(t, " ")[0]] = true
					mu.Unlock()
					log.Println(err)
				}
			})
		}

		wg.Wait()

		if len(errorEdits) != 0 {
			newDiffFile = getCurrentStateFailed(state, errorEdits)
		}

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
		databases    []store.DatabasesTable
	)

	eg.Go(func() (err error) { projects, err = app.sp.GetProjects(); return })
	eg.Go(func() (err error) { services, err = app.sp.GetServices(); return })
	eg.Go(func() (err error) { databases, err = app.sp.GetDatabases(); return })
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

	databasesMap := make(map[string][]store.DatabasesTable)
	databasesByNameMap := make(map[string]store.DatabasesTable)
	for _, database := range databases {
		databasesMap[database.ProjectId] = append(databasesMap[database.ProjectId], database)
		databasesByNameMap[database.ResourceName] = database
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
		projects:           projects,
		services:           services,
		environments:       environments,
		volumes:            volumes,
		databases:          databases,
		servicesMap:        servicesMap,
		servicesByNameMap:  servicesByNameMap,
		databasesMap:       databasesMap,
		databasesByNameMap: databasesByNameMap,
		volumesMap:         volumesMap,
		environmentsMap:    environmentsMap,
	}, nil
}

func getCurrentState(s *infraState) string {
	var currentState strings.Builder

	for _, project := range s.projects {
		fmt.Fprintln(&currentState, project.Namespace)
		services := s.servicesMap[project.ID]
		databases := s.databasesMap[project.ID]
		for _, service := range services {
			environments := util.ToEnvString(s.environmentsMap[service.ID])
			volume := s.volumesMap[service.ID]

			if volume == nil {
				fmt.Fprintln(&currentState, service.ResourceName, project.Namespace, service.Port, service.Image, ":", ":", environments)
			} else {
				fmt.Fprintln(&currentState, service.ResourceName, project.Namespace, service.Port, service.Image, volume.MountPath, volume.StorageGB, environments)
			}
		}

		for _, database := range databases {
			fmt.Fprintln(&currentState, database.ResourceName, project.Namespace, database.Port, database.Engine, database.StorageGB)
		}
	}
	return currentState.String()
}

func getCurrentStateFailed(s *infraState, exclude map[string]bool) string {
	var currentState strings.Builder

	for _, project := range s.projects {
		if exclude[project.Namespace] {
			continue
		}
		fmt.Fprintln(&currentState, project.Namespace)
		services := s.servicesMap[project.ID]
		databases := s.databasesMap[project.ID]
		for _, service := range services {
			if exclude[service.ResourceName] {
				continue
			}
			environments := util.ToEnvString(s.environmentsMap[service.ID])
			volume := s.volumesMap[service.ID]

			if volume == nil {
				fmt.Fprintln(&currentState, service.ResourceName, project.Namespace, service.Port, service.Image, ":", ":", environments)
			} else {
				fmt.Fprintln(&currentState, service.ResourceName, project.Namespace, service.Port, service.Image, volume.MountPath, volume.StorageGB, environments)
			}
		}

		for _, database := range databases {
			if exclude[database.ResourceName] {
				fmt.Println("DLKSDJKSDJLKSDJFKDLJJJJJJJJJJJj")
				continue
			}
			fmt.Fprintln(&currentState, database.ResourceName, project.Namespace, database.Port, database.Engine, database.StorageGB)
		}
	}
	return currentState.String()
}

func create(id string, state *infraState) error {
	fields := strings.Fields(id)
	if len(fields) == 0 {
		return fmt.Errorf("invalid empty diff row")
	}
	resource, _, _ := strings.Cut(fields[0], "-")

	switch resource {
	case "proj":
		err := app.dsvc.CreateProject(id)
		if err != nil {
			return err
		}
	case "svc":
		if len(fields) < 2 {
			return fmt.Errorf("invalid service diff row %q", id)
		}
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
	case "db":
		if len(fields) < 5 {
			return fmt.Errorf("invalid database diff row %q", id)
		}
		name, namespace := fields[0], fields[1]

		database := state.databasesByNameMap[name]

		err := app.dsvc.CreateDatabase(deploy.Database{
			Namespace: namespace,
			Name:      name,
			Engine:    database.Engine,
			StorageGB: database.StorageGB,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func destroy(id string) error {
	fields := strings.Fields(id)
	if len(fields) == 0 {
		return fmt.Errorf("invalid empty diff row")
	}
	resource, _, _ := strings.Cut(fields[0], "-")

	switch resource {
	case "proj":
		err := app.dsvc.Delete(id)
		if err != nil {
			return err
		}
	case "svc":
		if len(fields) < 6 {
			return fmt.Errorf("invalid service diff row %q", id)
		}
		name, namespace := fields[0], fields[1]

		envs := false
		if len(fields) == 7 {
			envs = true
		}

		hasVolume := fields[4] != ":"

		err := app.dsvc.DeleteService(deploy.Deployment{
			Namespace: namespace,
			Name:      name,
		}, envs, hasVolume)
		if err != nil {
			return err
		}
	case "db":
		if len(fields) < 5 {
			return fmt.Errorf("invalid database diff row %q", id)
		}
		name, namespace := fields[0], fields[1]

		err := app.dsvc.DeleteDatabase(deploy.Database{
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
