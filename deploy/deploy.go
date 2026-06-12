package deploy

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type DeployService struct {
	namespace string
	clientset *kubernetes.Clientset
}

func NewDeployService(namespace string, kubeconfigPath string) *DeployService {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	return &DeployService{
		namespace: namespace,
		clientset: clientset,
	}
}

func (d *DeployService) Get() {
	pods, err := d.clientset.CoreV1().Pods(d.namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("There are %d pods in the cluster (test): \n", len(pods.Items))
	for _, pod := range pods.Items {
		fmt.Println(pod.Name)
	}
}

func (d *DeployService) Create(name string, image string) error {
	var replicas int32 = 1
	_, err := d.clientset.AppsV1().Deployments(d.namespace).Create(context.TODO(), &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"app": name,
			},
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": name,
				},
			},
			Replicas: &replicas,
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
					Labels: map[string]string{
						"app": name,
					},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  name,
							Image: image,
							Ports: []v1.ContainerPort{
								{
									Protocol:      v1.ProtocolTCP,
									ContainerPort: 3000,
								},
							},
						},
					},
				},
			},
		},
	}, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	if _, err := d.clientset.AutoscalingV2().HorizontalPodAutoscalers(d.namespace).Create(context.TODO(), &autoscalingv2.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: autoscalingv2.CrossVersionObjectReference{
				APIVersion: "apps/v1",
				Kind:       "Deployment",
				Name:       name,
			},
			MaxReplicas: 10,
		},
	}, metav1.CreateOptions{}); err != nil {
		return err
	}

	if _, err := d.clientset.CoreV1().Services(d.namespace).Create(context.TODO(), &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"app": name,
			},
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Protocol: v1.ProtocolTCP,
					Port:     3000,
				},
			},
			Selector: map[string]string{
				"app": name,
			},
		},
	}, metav1.CreateOptions{}); err != nil {
		return err
	}

	return nil
}

func (d *DeployService) Update(name string, newName string) error {
	err := d.Delete(name)
	if err != nil {
		return err
	}

	err = d.Create(newName, "nginx")
	if err != nil {
		return err
	}

	return nil
}

func (d *DeployService) Delete(name string) error {
	err := d.clientset.CoreV1().Pods(d.namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	return nil
}
