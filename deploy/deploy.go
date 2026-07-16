package deploy

import (
	"context"
	"errors"
	"fmt"

	"github.com/timmyjinks/tysoncloud-cli/util"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	appcorev1 "k8s.io/client-go/applyconfigurations/core/v1"
	appmetav1 "k8s.io/client-go/applyconfigurations/meta/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	gatewayclient "sigs.k8s.io/gateway-api/pkg/client/clientset/versioned"
)

type DeployService struct {
	clusterIP     string
	clientset     *kubernetes.Clientset
	gatewayClient *gatewayclient.Clientset
	dynamicClient *dynamic.DynamicClient
}

func NewDeployService(kubeconfigPath string) (*DeployService, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		panic(err)
	}

	clientset := kubernetes.NewForConfigOrDie(config)
	gatewayClient := gatewayclient.NewForConfigOrDie(config)
	dynamicClient := dynamic.NewForConfigOrDie(config)

	svc, err := clientset.CoreV1().
		Services("default").
		Get(context.Background(), "kubernetes", metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	ip := svc.Spec.ClusterIP

	return &DeployService{
		clusterIP:     ip,
		clientset:     clientset,
		gatewayClient: gatewayClient,
		dynamicClient: dynamicClient,
	}, nil
}

func (d *DeployService) Get() {
	pods, err := d.clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("There are %d pods in the cluster (all): \n", len(pods.Items))
	for _, pod := range pods.Items {
		fmt.Println(pod.Name)
	}
}

func (d *DeployService) BatchCreate(deployments ...Deployment) error {
	for _, deployment := range deployments {
		err := d.ensureNamespace(deployment.Namespace)
		if err != nil {
			return err
		}

		if err := d.createSecret(deployment); err != nil {
			return err
		}

		if err := d.createService(deployment); err != nil {
			return err
		}

		if deployment.Volume != nil {
			if err := d.createPVC(deployment); err != nil {
				return err
			}
		}

		if err := d.createDeployment(deployment); err != nil {
			return err
		}

		if err := d.createHPA(deployment); err != nil {
			return err
		}

		if err := d.createHTTPRoute(deployment); err != nil {
			return err
		}
	}
	return nil
}

func (d *DeployService) CreateProject(namespace string) error {
	err := d.ensureNamespace(namespace)
	if err != nil {
		return err
	}

	if err := d.ensureNetworkPolicy(namespace, d.clusterIP); err != nil {
		return err
	}
	return nil
}

func (d *DeployService) Create(deployment Deployment) error {
	err := d.ensureNamespace(deployment.Namespace)
	if err != nil {
		return err
	}

	if err := d.createService(deployment); err != nil {
		return err
	}

	if len(deployment.Env) != 0 {
		if err := d.createSecret(deployment); err != nil {
			return err
		}
	}

	if deployment.Volume != nil {
		if err := d.createPVC(deployment); err != nil {
			return err
		}
	}

	if err := d.createDeployment(deployment); err != nil {
		return err
	}

	if err := d.createHPA(deployment); err != nil {
		return err
	}

	if err := d.createHTTPRoute(deployment); err != nil {
		return err
	}

	return nil
}

func (d *DeployService) CreateDatabase(database Database) error {
	switch database.Engine {
	case "postgres":
		_, err := d.clientset.CoreV1().Secrets(database.Namespace).Apply(context.TODO(), &appcorev1.SecretApplyConfiguration{
			TypeMetaApplyConfiguration: appmetav1.TypeMetaApplyConfiguration{
				Kind:       util.StringPtr("Secret"),
				APIVersion: util.StringPtr("v1"),
			},
			ObjectMetaApplyConfiguration: &appmetav1.ObjectMetaApplyConfiguration{
				Namespace: &database.Namespace,
				Name:      &database.Name,
			},
			Type: (*v1.SecretType)(util.StringPtr("kubernetes.io/basic-auth")),
			Data: map[string][]byte{
				"username": []byte(database.Name),
				"password": []byte("password"),
			},
		}, metav1.ApplyOptions{
			FieldManager: "controller",
		})
		if err != nil {
			return err
		}
		return d.CreatePostgresDatabase(database)
	default:
		return errors.New("DB engine not found")
	}
}

func (d *DeployService) Update(deployment Deployment, newName string) error {
	err := d.Delete(deployment.Namespace)
	if err != nil {
		return err
	}

	err = d.Create(Deployment{
		Namespace: deployment.Namespace,
		Name:      newName,
		Image:     deployment.Image,
		Port:      deployment.Port,
	})
	if err != nil {
		return err
	}

	return nil
}

func (d *DeployService) Delete(namespace string) error {
	return d.clientset.CoreV1().Namespaces().Delete(context.TODO(), namespace, metav1.DeleteOptions{})
}

func (d *DeployService) DeleteService(deployment Deployment, envs bool, volume bool) error {
	if err := d.clientset.CoreV1().Services(deployment.Namespace).Delete(context.TODO(), deployment.Name, metav1.DeleteOptions{}); err != nil {
		return err
	}

	if envs {
		if err := d.clientset.CoreV1().Secrets(deployment.Namespace).Delete(context.TODO(), deployment.Name, metav1.DeleteOptions{}); err != nil {
			return err
		}
	}

	if err := d.clientset.AppsV1().Deployments(deployment.Namespace).Delete(context.TODO(), deployment.Name, metav1.DeleteOptions{}); err != nil {
		return err
	}

	if volume {
		if err := d.clientset.CoreV1().PersistentVolumeClaims(deployment.Namespace).Delete(context.TODO(), deployment.Name, metav1.DeleteOptions{}); err != nil {
			return err
		}
	}

	if err := d.clientset.AutoscalingV1().HorizontalPodAutoscalers(deployment.Namespace).Delete(context.TODO(), deployment.Name, metav1.DeleteOptions{}); err != nil {
		return err
	}

	if err := d.gatewayClient.GatewayV1().HTTPRoutes(deployment.Namespace).Delete(context.TODO(), deployment.Name, metav1.DeleteOptions{}); err != nil {
		return err
	}

	return nil
}
