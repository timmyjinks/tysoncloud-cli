package deploy

import (
	"context"

	"github.com/timmyjinks/tysoncloud-cli/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	appmetav1 "k8s.io/client-go/applyconfigurations/meta/v1"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	v1 "sigs.k8s.io/gateway-api/applyconfiguration/apis/v1"
)

func (d *DeployService) createHTTPRoute(ctx context.Context, deployment Deployment) error {
	_, err := d.gatewayClient.GatewayV1().HTTPRoutes(deployment.Namespace).Apply(ctx, &v1.HTTPRouteApplyConfiguration{
		TypeMetaApplyConfiguration: appmetav1.TypeMetaApplyConfiguration{
			Kind:       util.StringPtr("HTTPRoute"),
			APIVersion: util.StringPtr("gateway.networking.k8s.io/v1"),
		},
		ObjectMetaApplyConfiguration: &appmetav1.ObjectMetaApplyConfiguration{
			Namespace: util.StringPtr(deployment.Namespace),
			Name:      util.StringPtr(deployment.Name),
		},
		Spec: &v1.HTTPRouteSpecApplyConfiguration{
			CommonRouteSpecApplyConfiguration: v1.CommonRouteSpecApplyConfiguration{
				ParentRefs: []v1.ParentReferenceApplyConfiguration{
					{
						Name:      (*gatewayv1.ObjectName)(util.StringPtr("tysoncloud-gateway")),
						Namespace: (*gatewayv1.Namespace)(util.StringPtr("tc-system")),
					},
				},
			},
			Hostnames: []gatewayv1.Hostname{
				gatewayv1.Hostname(deployment.Hostname),
			},
			Rules: []v1.HTTPRouteRuleApplyConfiguration{
				{
					Matches: []v1.HTTPRouteMatchApplyConfiguration{
						{
							Path: &v1.HTTPPathMatchApplyConfiguration{
								Type:  (*gatewayv1.PathMatchType)(util.StringPtr("PathPrefix")),
								Value: util.StringPtr("/"),
							},
						},
					},
					BackendRefs: []v1.HTTPBackendRefApplyConfiguration{
						{
							BackendRefApplyConfiguration: v1.BackendRefApplyConfiguration{
								BackendObjectReferenceApplyConfiguration: v1.BackendObjectReferenceApplyConfiguration{
									Name: (*gatewayv1.ObjectName)(util.StringPtr(deployment.Name)),
									Port: util.IntPtr(80),
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
	return nil
}
