package deploy

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

type Database struct {
	Namespace string
	Name      string
	Engine    string
}

func (d *DeployService) CreateDatabase(database Database) error {
	var clusterGVR = schema.GroupVersionResource{
		Group:    "postgresql.cnpg.io",
		Version:  "v1",
		Resource: "clusters",
	}

	cluster := &unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": "postgresql.cnpg.io/v1",
			"kind":       "Cluster",
			"metadata": map[string]any{
				"name": database.Name,
			},
			"spec": map[string]any{
				"instances": 1,
				"imageName": "ghcr.io/cloudnative-pg/postgresql:17",
				"storage": map[string]any{
					"size": "10Gi",;
				},
			},
		},
	}

	_, err := d.Dynamic.
		Resource(clusterGVR).
		Namespace("databases").
		Create(ctx, cluster, metav1.CreateOptions{})

	return err

	settings := cli.New()

	install := action.NewInstall(d.helmClient)
	install.ReleaseName = database.Name
	install.Namespace = database.Namespace
	install.CreateNamespace = true

	chartPath, err := install.ChartPathOptions.LocateChart(
		database.Engine,
		settings,
	)
	if err != nil {
		return err
	}

	chart, err := loader.Load(chartPath)
	if err != nil {
		return err
	}

	values := map[string]any{}

	_, err = install.RunWithContext(context.Background(), chart, values)
	if err != nil {
		return err
	}

	return nil
}

func (d *DeployService) UpdateDatabase(deployment Deployment, newName string) error {
	return nil
}

func (d *DeployService) DeleteDatabase(database Database) error {
	uninstall := action.NewUninstall(d.helmClient)

	_, err := uninstall.Run(database.Name)
	return err
}
