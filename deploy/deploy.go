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

func (d *DeployService) CreateProject(deployment Deployment) error {
	err := d.ensureNamespace(deployment.Namespace)
	if err != nil {
		return err
	}
	return nil
}

func (d *DeployService) Create(deployment Deployment) error {
	err := d.ensureNamespace(deployment.Namespace)
	if err != nil {
		return err
	}

	if err := d.ensureNetworkPolicy(deployment); err != nil {
		return err
	}

	if err := d.createService(deployment); err != nil {
		return err
	}

	if err := d.createSecret(deployment); err != nil {
		return err
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

func (d *DeployService) DeleteService(deployment Deployment) error {
	if err := d.clientset.CoreV1().Services(deployment.Namespace).Delete(context.TODO(), deployment.Name, metav1.DeleteOptions{}); err != nil {
		return err
	}

	if err := d.clientset.CoreV1().Secrets(deployment.Namespace).Delete(context.TODO(), deployment.Name, metav1.DeleteOptions{}); err != nil {
		return err
	}

	if err := d.clientset.AppsV1().Deployments(deployment.Namespace).Delete(context.TODO(), deployment.Name, metav1.DeleteOptions{}); err != nil {
		return err
	}

	if err := d.clientset.AutoscalingV1().HorizontalPodAutoscalers(deployment.Namespace).Delete(context.TODO(), deployment.Name, metav1.DeleteOptions{}); err != nil {
		return err
	}

	if err := d.gatewayClient.GatewayV1().HTTPRoutes(deployment.Namespace).Delete(context.TODO(), deployment.Name, metav1.DeleteOptions{}); err != nil {
		return err
	}

	return nil
}
