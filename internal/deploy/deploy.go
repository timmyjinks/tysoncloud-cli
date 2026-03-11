package deploy

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func Deploy() {
	config, err := clientcmd.BuildConfigFromFlags("", "/etc/rancher/k3s/k3s.yaml")
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	pods, err := clientset.CoreV1().Pods("test").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("There are %d pods in the cluster: %v\n", len(pods.Items))
	for _, pod := range pods.Items {
		fmt.Println(pod.Name)
	}
}

func Create(name string) error {
	config, err := clientcmd.BuildConfigFromFlags("", "/etc/rancher/k3s/k3s.yaml")
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	pod, err := clientset.CoreV1().Pods("default").Create(context.TODO(), &v1.Pod{
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{Name: name, Image: "nginx"},
			},
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	fmt.Println(pod)
	return nil
}
