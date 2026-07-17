package deploy

import (
	"context"
	"fmt"

	"github.com/timmyjinks/tysoncloud-cli/util"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	appsv1apply "k8s.io/client-go/applyconfigurations/apps/v1"
	v2 "k8s.io/client-go/applyconfigurations/autoscaling/v2"
	appcorev1 "k8s.io/client-go/applyconfigurations/core/v1"
	appmetav1 "k8s.io/client-go/applyconfigurations/meta/v1"
)

func (d *DeployService) createDeployment(deployment Deployment) error {
	container := []appcorev1.ContainerApplyConfiguration{
		{
			Name:  &deployment.Name,
			Image: &deployment.Image,
			Resources: &appcorev1.ResourceRequirementsApplyConfiguration{
				Limits: &corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("500m"),
					corev1.ResourceMemory: resource.MustParse("256Mi"),
				},
				Requests: &corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("100m"),
					corev1.ResourceMemory: resource.MustParse("1Mi"),
				},
			},
			Ports: []appcorev1.ContainerPortApplyConfiguration{
				{
					Protocol:      (*corev1.Protocol)(util.StringPtr(string(corev1.ProtocolTCP))),
					ContainerPort: &deployment.Port,
				},
			},
		},
	}

	if len(deployment.Env) != 0 {
		container[0].EnvFrom = []appcorev1.EnvFromSourceApplyConfiguration{
			{
				SecretRef: &appcorev1.SecretEnvSourceApplyConfiguration{
					LocalObjectReferenceApplyConfiguration: appcorev1.LocalObjectReferenceApplyConfiguration{
						Name: &deployment.Name,
					},
				},
			},
		}
	}

	spec := &appsv1apply.DeploymentSpecApplyConfiguration{
		Selector: &appmetav1.LabelSelectorApplyConfiguration{
			MatchLabels: map[string]string{
				"app": deployment.Name,
			},
		},
		Template: &appcorev1.PodTemplateSpecApplyConfiguration{
			ObjectMetaApplyConfiguration: &appmetav1.ObjectMetaApplyConfiguration{
				Name: &deployment.Name,
				Labels: map[string]string{
					"app": deployment.Name,
				},
			},
			Spec: &appcorev1.PodSpecApplyConfiguration{
				Containers: container,
			},
		},
	}

	if deployment.Volume != nil {
		spec.Template.Spec.Containers[0].VolumeMounts = []appcorev1.VolumeMountApplyConfiguration{
			{
				Name:      &deployment.Name,
				MountPath: &deployment.Volume.MountPath,
			},
		}
		container[0].VolumeMounts = []appcorev1.VolumeMountApplyConfiguration{
			{
				Name:      &deployment.Name,
				MountPath: &deployment.Volume.MountPath,
			},
		}
		spec.Template.Spec.Volumes = []appcorev1.VolumeApplyConfiguration{
			{
				Name: &deployment.Name,
				VolumeSourceApplyConfiguration: appcorev1.VolumeSourceApplyConfiguration{
					PersistentVolumeClaim: &appcorev1.PersistentVolumeClaimVolumeSourceApplyConfiguration{
						ClaimName: &deployment.Name,
					},
				},
			},
		}
	}

	_, err := d.clientset.AppsV1().Deployments(deployment.Namespace).Apply(context.TODO(), &appsv1apply.DeploymentApplyConfiguration{
		TypeMetaApplyConfiguration: appmetav1.TypeMetaApplyConfiguration{
			Kind:       util.StringPtr("Deployment"),
			APIVersion: util.StringPtr("apps/v1"),
		},
		ObjectMetaApplyConfiguration: &appmetav1.ObjectMetaApplyConfiguration{
			Name: &deployment.Name,
			Labels: map[string]string{
				"app": deployment.Name,
			},
			Annotations: map[string]string{
				"reloader.stakater.com/auto": "true",
			},
		},
		Spec: spec,
	}, metav1.ApplyOptions{
		FieldManager: "controller",
	})
	if err != nil {
		return err
	}
	return nil
}

func (d *DeployService) createPVC(deployment Deployment) error {
	_, err := d.clientset.CoreV1().PersistentVolumeClaims(deployment.Namespace).Apply(context.TODO(), &appcorev1.PersistentVolumeClaimApplyConfiguration{
		TypeMetaApplyConfiguration: appmetav1.TypeMetaApplyConfiguration{
			Kind:       util.StringPtr("PersistentVolumeClaim"),
			APIVersion: util.StringPtr("v1"),
		},
		ObjectMetaApplyConfiguration: &appmetav1.ObjectMetaApplyConfiguration{
			Name:      &deployment.Name,
			Namespace: &deployment.Namespace,
		},
		Spec: &appcorev1.PersistentVolumeClaimSpecApplyConfiguration{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			Resources: &appcorev1.VolumeResourceRequirementsApplyConfiguration{
				Requests: &corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse(fmt.Sprintf("%dGi", deployment.Volume.StorageGB)),
				},
			},
		},
	}, metav1.ApplyOptions{
		FieldManager: "controller",
	})
	if err != nil {
		return err
	}
	return nil
}

func (d *DeployService) createService(deployment Deployment) error {
	_, err := d.clientset.CoreV1().Services(deployment.Namespace).Apply(context.TODO(), &appcorev1.ServiceApplyConfiguration{
		TypeMetaApplyConfiguration: appmetav1.TypeMetaApplyConfiguration{
			Kind:       util.StringPtr("Service"),
			APIVersion: util.StringPtr("v1"),
		},
		ObjectMetaApplyConfiguration: &appmetav1.ObjectMetaApplyConfiguration{
			Name: &deployment.Name,
			Labels: map[string]string{
				"app": deployment.Name,
			},
		},
		Spec: &appcorev1.ServiceSpecApplyConfiguration{
			Ports: []appcorev1.ServicePortApplyConfiguration{
				{
					Protocol:   (*corev1.Protocol)(util.StringPtr(string(corev1.ProtocolTCP))),
					Port:       util.IntPtr(80),
					TargetPort: &intstr.IntOrString{IntVal: deployment.Port},
				},
			},
			Selector: map[string]string{
				"app": deployment.Name,
			},
		},
	}, metav1.ApplyOptions{
		FieldManager: "controller",
	})
	if err != nil {
		return err
	}
	return nil
}

func (d *DeployService) createHPA(deployment Deployment) error {
	_, err := d.clientset.AutoscalingV2().HorizontalPodAutoscalers(deployment.Namespace).Apply(context.TODO(), &v2.HorizontalPodAutoscalerApplyConfiguration{
		TypeMetaApplyConfiguration: appmetav1.TypeMetaApplyConfiguration{
			Kind:       util.StringPtr("HorizontalPodAutoscaler"),
			APIVersion: util.StringPtr("autoscaling/v2"),
		},
		ObjectMetaApplyConfiguration: &appmetav1.ObjectMetaApplyConfiguration{
			Name: &deployment.Name,
			Labels: map[string]string{
				"app.kubernetes.io/component": "service",
			},
		},
		Spec: &v2.HorizontalPodAutoscalerSpecApplyConfiguration{
			ScaleTargetRef: &v2.CrossVersionObjectReferenceApplyConfiguration{
				Kind:       util.StringPtr("Deployment"),
				APIVersion: util.StringPtr("apps/v1"),
				Name:       &deployment.Name,
			},
			MinReplicas: util.IntPtr(1),
			MaxReplicas: util.IntPtr(10),
			Metrics: []v2.MetricSpecApplyConfiguration{
				{
					Type: (*autoscalingv2.MetricSourceType)(util.StringPtr(string(autoscalingv2.ResourceMetricSourceType))),
					Resource: &v2.ResourceMetricSourceApplyConfiguration{
						Name: (*corev1.ResourceName)(util.StringPtr(string(corev1.ResourceCPU))),
						Target: &v2.MetricTargetApplyConfiguration{
							Type:               (*autoscalingv2.MetricTargetType)(util.StringPtr(string(autoscalingv2.UtilizationMetricType))),
							AverageUtilization: util.IntPtr(50),
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
	return nil
}

func (d *DeployService) createSecret(deployment Deployment) error {
	_, err := d.clientset.CoreV1().Secrets(deployment.Namespace).Apply(context.TODO(), &appcorev1.SecretApplyConfiguration{
		TypeMetaApplyConfiguration: appmetav1.TypeMetaApplyConfiguration{
			Kind:       util.StringPtr("Secret"),
			APIVersion: util.StringPtr("v1"),
		},
		ObjectMetaApplyConfiguration: &appmetav1.ObjectMetaApplyConfiguration{
			Namespace: &deployment.Namespace,
			Name:      &deployment.Name,
		},
		Data: deployment.Env,
	}, metav1.ApplyOptions{
		FieldManager: "controller",
	})
	if err != nil {
		return err
	}
	return nil
}
