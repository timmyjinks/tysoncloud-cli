package deploy

import (
	"context"
	"fmt"

	"github.com/timmyjinks/tysoncloud-cli/util"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	appsv1apply "k8s.io/client-go/applyconfigurations/apps/v1"
	v2 "k8s.io/client-go/applyconfigurations/autoscaling/v2"
	appcorev1 "k8s.io/client-go/applyconfigurations/core/v1"
	appmetav1 "k8s.io/client-go/applyconfigurations/meta/v1"
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

func (d *DeployService) Create(name string, image string, port int32) error {
	var replicas int32 = 1
	_, err := d.clientset.AppsV1().Deployments(d.namespace).Apply(context.TODO(), &appsv1apply.DeploymentApplyConfiguration{
		TypeMetaApplyConfiguration: appmetav1.TypeMetaApplyConfiguration{
			Kind:       util.StringPtr("Deployment"),
			APIVersion: util.StringPtr("apps/v1"),
		},
		ObjectMetaApplyConfiguration: &appmetav1.ObjectMetaApplyConfiguration{
			Name: &name,
			Labels: map[string]string{
				"app": name,
			},
		},
		Spec: &appsv1apply.DeploymentSpecApplyConfiguration{
			Selector: &appmetav1.LabelSelectorApplyConfiguration{
				MatchLabels: map[string]string{
					"app": name,
				},
			},
			Replicas: &replicas,
			Template: &appcorev1.PodTemplateSpecApplyConfiguration{
				ObjectMetaApplyConfiguration: &appmetav1.ObjectMetaApplyConfiguration{
					Name: &name,
					Labels: map[string]string{
						"app": name,
					},
				},
				Spec: &appcorev1.PodSpecApplyConfiguration{
					Containers: []appcorev1.ContainerApplyConfiguration{
						{
							Name:  &name,
							Image: &image,
							Ports: []appcorev1.ContainerPortApplyConfiguration{
								{
									Protocol:      (*corev1.Protocol)(util.StringPtr(string(corev1.ProtocolTCP))),
									ContainerPort: &port,
								},
							},
						},
					},
				},
			},
		},
	}, metav1.ApplyOptions{
		FieldManager: "controller",
	})
	if err != nil {
		return err
	}

	if _, err := d.clientset.CoreV1().Services(d.namespace).Apply(context.TODO(), &appcorev1.ServiceApplyConfiguration{
		TypeMetaApplyConfiguration: appmetav1.TypeMetaApplyConfiguration{
			Kind:       util.StringPtr("Service"),
			APIVersion: util.StringPtr("v1"),
		},
		ObjectMetaApplyConfiguration: &appmetav1.ObjectMetaApplyConfiguration{
			Name: &name,
			Labels: map[string]string{
				"app": name,
			},
		},
		Spec: &appcorev1.ServiceSpecApplyConfiguration{
			Ports: []appcorev1.ServicePortApplyConfiguration{
				{
					Protocol:   (*corev1.Protocol)(util.StringPtr(string(corev1.ProtocolTCP))),
					Port:       util.IntPtr(80),
					TargetPort: &intstr.IntOrString{IntVal: port},
				},
			},
			Selector: map[string]string{
				"app": name,
			},
		},
	}, metav1.ApplyOptions{
		FieldManager: "controller",
	}); err != nil {
		return err
	}

	if _, err := d.clientset.AutoscalingV2().HorizontalPodAutoscalers(d.namespace).Apply(context.TODO(), &v2.HorizontalPodAutoscalerApplyConfiguration{
		TypeMetaApplyConfiguration: appmetav1.TypeMetaApplyConfiguration{
			Kind:       util.StringPtr("HorizontalPodAutoscaler"),
			APIVersion: util.StringPtr("autoscaling/v2"),
		},
		ObjectMetaApplyConfiguration: &appmetav1.ObjectMetaApplyConfiguration{
			Name: &name,
		},
		Spec: &v2.HorizontalPodAutoscalerSpecApplyConfiguration{
			ScaleTargetRef: &v2.CrossVersionObjectReferenceApplyConfiguration{
				Kind:       util.StringPtr("Deployment"),
				APIVersion: util.StringPtr("apps/v1"),
				Name:       &name,
			},
			MaxReplicas: util.IntPtr(10),
		},
	}, metav1.ApplyOptions{
		FieldManager: "controller",
	}); err != nil {
		return err
	}

	return nil
}

func (d *DeployService) Update(name string, newName string) error {
	err := d.Delete(name)
	if err != nil {
		return err
	}

	err = d.Create(newName, "nginx", 3000)
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
