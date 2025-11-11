//go:build !nowebhook

package v1

import (
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	dscv1 "github.com/opendatahub-io/opendatahub-operator/v2/api/datasciencecluster/v1"
)

// RegisterWebhooks registers the webhooks for DataScienceCluster v1.
// The platformValidator is injected to enable platform-specific validation.
func RegisterWebhooks(mgr ctrl.Manager, platformValidator admission.Handler) error {
	// Register the conversion webhook
	if err := ctrl.NewWebhookManagedBy(mgr).For(&dscv1.DataScienceCluster{}).Complete(); err != nil {
		return err
	}

	// Register the validating webhook
	if err := (&Validator{
		Client:            mgr.GetAPIReader(),
		Name:              "datasciencecluster-v1-validating",
		Decoder:           admission.NewDecoder(mgr.GetScheme()),
		PlatformValidator: platformValidator,
	}).SetupWithManager(mgr); err != nil {
		return err
	}

	// Register the defaulting webhook
	if err := (&Defaulter{
		Name: "datasciencecluster-v1-defaulter",
	}).SetupWithManager(mgr); err != nil {
		return err
	}

	return nil
}
