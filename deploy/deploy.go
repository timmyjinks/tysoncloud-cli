package deploy

import (
	"context"
	"errors"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
		return nil, err
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

func (d *DeployService) Get(ctx context.Context) error {
	pods, err := d.clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}
	fmt.Printf("There are %d pods in the cluster (all): \n", len(pods.Items))
	for _, pod := range pods.Items {
		fmt.Println(pod.Name)
	}
	return nil
}

func (d *DeployService) BatchCreate(ctx context.Context, deployments ...Deployment) error {
	for _, deployment := range deployments {
		err := d.ensureNamespace(ctx, deployment.Namespace)
		if err != nil {
			return err
		}

		if err := d.createSecret(ctx, deployment); err != nil {
			return err
		}

		if err := d.createService(ctx, deployment); err != nil {
			return err
		}

		if deployment.Volume != nil {
			if err := d.createPVC(ctx, deployment); err != nil {
				return err
			}
		}

		if err := d.createDeployment(ctx, deployment); err != nil {
			return err
		}

		if err := d.createHPA(ctx, deployment); err != nil {
			return err
		}

		if err := d.createHTTPRoute(ctx, deployment); err != nil {
			return err
		}
	}
	return nil
}

func (d *DeployService) CreateProject(ctx context.Context, namespace string) error {
	err := d.ensureNamespace(ctx, namespace)
	if err != nil {
		return err
	}

	if err := d.ensureNetworkPolicy(ctx, namespace, d.clusterIP); err != nil {
		return err
	}
	return nil
}

func (d *DeployService) Create(ctx context.Context, deployment Deployment) error {
	err := d.ensureNamespace(ctx, deployment.Namespace)
	if err != nil {
		return err
	}

	if err := d.createService(ctx, deployment); err != nil {
		return err
	}

	if len(deployment.Env) != 0 {
		if err := d.createSecret(ctx, deployment); err != nil {
			return err
		}
	}

	if deployment.Volume != nil {
		if err := d.createPVC(ctx, deployment); err != nil {
			return err
		}
	}

	if err := d.createDeployment(ctx, deployment); err != nil {
		return err
	}

	if err := d.createHPA(ctx, deployment); err != nil {
		return err
	}

	if err := d.createHTTPRoute(ctx, deployment); err != nil {
		return err
	}

	return nil
}

func (d *DeployService) CreateDatabase(ctx context.Context, database Database) error {
	switch database.Engine {
	case "postgres":
		return d.CreatePostgresDatabase(ctx, database)
	default:
		return errors.New("DB engine not found")
	}
}

func (d *DeployService) Delete(ctx context.Context, namespace string) error {
	return d.clientset.CoreV1().Namespaces().Delete(ctx, namespace, metav1.DeleteOptions{})
}

func (d *DeployService) DeleteService(ctx context.Context, deployment Deployment, envs bool, volume bool) error {
	if err := d.clientset.CoreV1().Services(deployment.Namespace).Delete(ctx, deployment.Name, metav1.DeleteOptions{}); err != nil {
		return err
	}

	if envs {
		if err := d.clientset.CoreV1().Secrets(deployment.Namespace).Delete(ctx, deployment.Name, metav1.DeleteOptions{}); err != nil {
			return err
		}
	}

	if err := d.clientset.AppsV1().Deployments(deployment.Namespace).Delete(ctx, deployment.Name, metav1.DeleteOptions{}); err != nil {
		return err
	}

	if volume {
		if err := d.clientset.CoreV1().PersistentVolumeClaims(deployment.Namespace).Delete(ctx, deployment.Name, metav1.DeleteOptions{}); err != nil {
			return err
		}
	}

	if err := d.clientset.AutoscalingV1().HorizontalPodAutoscalers(deployment.Namespace).Delete(ctx, deployment.Name, metav1.DeleteOptions{}); err != nil {
		return err
	}

	if err := d.gatewayClient.GatewayV1().HTTPRoutes(deployment.Namespace).Delete(ctx, deployment.Name, metav1.DeleteOptions{}); err != nil {
		return err
	}

	return nil
}
