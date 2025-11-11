package cluster

import (
	"context"

	k8serr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	infrav1 "github.com/opendatahub-io/opendatahub-operator/v2/api/infrastructure/v1"
	serviceApi "github.com/opendatahub-io/opendatahub-operator/v2/api/services/v1alpha1"
)

func CreateGatewayConfig(
	ctx context.Context,
	cli client.Client,
) error {
	l := log.FromContext(ctx)

	gw := &serviceApi.GatewayConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name: serviceApi.GatewayInstanceName,
		},
		Spec: serviceApi.GatewayConfigSpec{
			Certificate: &infrav1.CertificateSpec{
				Type:       infrav1.OpenshiftDefaultIngress,
				SecretName: "default-gateway-tls",
			},
		},
	}

	err := cli.Get(ctx, client.ObjectKeyFromObject(gw), gw)
	switch {
	case k8serr.IsNotFound(err):
		if createErr := cli.Create(ctx, gw); createErr != nil {
			l.Error(createErr, "unable to create default Gateway CR")
			return createErr
		}
		l.Info("Created default Gateway CR", "name", serviceApi.GatewayInstanceName)
		return nil
	case err != nil:
		l.Error(err, "error checking for existing Gateway CR")
		return err
	default:
		l.Info("Default Gateway CR already exists", "name", serviceApi.GatewayInstanceName)
		return nil
	}
}
