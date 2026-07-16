package deploy

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	gatewayclient "sigs.k8s.io/gateway-api/pkg/client/clientset/versioned"
)

type DeployService struct {
	clientset     *kubernetes.Clientset
	gatewayClient *gatewayclient.Clientset
}

func NewDeployService(kubeconfigPath string) *DeployService {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		panic(err.Error())
	}

	clientset := kubernetes.NewForConfigOrDie(config)
	gatewayClient := gatewayclient.NewForConfigOrDie(config)

	return &DeployService{
		clientset:     clientset,
		gatewayClient: gatewayClient,
	}
}

func (d *DeployService) Get(ctx context.Context) {
	pods, err := d.clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("There are %d pods in the cluster (all): \n", len(pods.Items))
	for _, pod := range pods.Items {
		fmt.Println(pod.Name)
	}
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
	return nil
}

func (d *DeployService) Create(ctx context.Context, deployment Deployment) error {
	err := d.ensureNamespace(ctx, deployment.Namespace)
	if err != nil {
		return err
	}

	if err := d.ensureNetworkPolicy(ctx, deployment); err != nil {
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

func (d *DeployService) Update(ctx context.Context, deployment Deployment, newName string) error {
	err := d.Delete(ctx, deployment.Namespace)
	if err != nil {
		return err
	}

	err = d.Create(ctx, Deployment{
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
