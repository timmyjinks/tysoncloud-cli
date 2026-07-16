package deploy

import (
	"context"
	"errors"

	cnpgv1 "github.com/cloudnative-pg/cloudnative-pg/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type Database struct {
	Namespace string
	Name      string
	Engine    string
}

func (d *DeployService) CreateDatabase(database Database) error {
	switch database.Engine {
	case "postgres":
		return d.CreatePostgresDatabase(database)
	default:
		return errors.New("DB engine not found")
	}
}

func (d *DeployService) DeleteDatabase(database Database) error {
	gvr := schema.GroupVersionResource{
		Group:    "postgresql.cnpg.io",
		Version:  "v1",
		Resource: "clusters",
	}
	return d.dynamicClient.Resource(gvr).Namespace(database.Namespace).Delete(context.Background(), database.Name, metav1.DeleteOptions{})
}

func (d *DeployService) CreatePostgresDatabase(database Database) error {
	cluster := &cnpgv1.Cluster{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "postgresql.cnpg.io/v1",
			Kind:       "Cluster",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      database.Name,
			Namespace: database.Namespace,
		},
		Spec: cnpgv1.ClusterSpec{
			InheritedMetadata: &cnpgv1.EmbeddedObjectMetadata{
				Labels: map[string]string{
					"app.kubernetes.io/component": "database",
				},
			},
			Instances: 1,

			Managed: &cnpgv1.ManagedConfiguration{
				Roles: []cnpgv1.RoleConfiguration{
					{
						Name:  database.Name,
						Login: true,
						PasswordSecret: &api.LocalObjectReference{
							Name: database.Name,
						},
					},
				},
			},

			Bootstrap: &cnpgv1.BootstrapConfiguration{
				InitDB: &cnpgv1.BootstrapInitDB{
					Database: "app",
					Owner:    "app",
				},
			},
		},
	}

	if database.StorageGB <= 0 {
		cluster.Spec.StorageConfiguration = cnpgv1.StorageConfiguration{
			Size: fmt.Sprintf("%sGi", database.StorageGB),
		}
	}

	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(cluster)
	if err != nil {
		return err
	}

	gvr := schema.GroupVersionResource{
		Group:    "postgresql.cnpg.io",
		Version:  "v1",
		Resource: "clusters",
	}

	_, err = d.dynamicClient.Resource(gvr).
		Namespace(database.Namespace).
		Apply(
			context.Background(),
			database.Name,
			&unstructured.Unstructured{Object: obj},
			metav1.ApplyOptions{
				FieldManager: "controller",
			},
		)
	if err != nil {
		return err
	}
	return nil
}
