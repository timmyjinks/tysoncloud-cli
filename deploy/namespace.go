package deploy

import (
	"context"

	"github.com/timmyjinks/tysoncloud-cli/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	appcorev1 "k8s.io/client-go/applyconfigurations/core/v1"
	appmetav1 "k8s.io/client-go/applyconfigurations/meta/v1"
)

func (d *DeployService) ensureNamespace(ctx context.Context, namespace string) error {
	_, err := d.clientset.CoreV1().Namespaces().Apply(ctx, &appcorev1.NamespaceApplyConfiguration{
		TypeMetaApplyConfiguration: appmetav1.TypeMetaApplyConfiguration{
			APIVersion: util.StringPtr("v1"),
			Kind:       util.StringPtr("Namespace"),
		},
		ObjectMetaApplyConfiguration: &appmetav1.ObjectMetaApplyConfiguration{
			Name: util.StringPtr(namespace),
			Labels: map[string]string{
				"managed-by": "tysoncloud",
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
