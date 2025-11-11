//go:build !nowebhook

package v2

import (
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// RegisterWebhooks registers the webhooks for DataScienceCluster v2.
// The platformValidator is injected to enable platform-specific validation.
func RegisterWebhooks(mgr ctrl.Manager, platformValidator admission.Handler) error {
	// Register the validating webhook
	if err := (&Validator{
		Client:            mgr.GetAPIReader(),
		Name:              "datasciencecluster-v2-validating",
		PlatformValidator: platformValidator,
	}).SetupWithManager(mgr); err != nil {
		return err
	}

	// Register the defaulting webhook
	if err := (&Defaulter{
		Name: "datasciencecluster-v2-defaulter",
	}).SetupWithManager(mgr); err != nil {
		return err
	}

	return nil
}
