//go:build !nowebhook

package v2

import (
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// RegisterWebhooks registers the webhooks for DSCInitialization v2.
// The platformValidator is injected to enable platform-specific validation.
func RegisterWebhooks(mgr ctrl.Manager, platformValidator admission.Handler) error {
	if err := (&Validator{
		Client:            mgr.GetAPIReader(),
		Name:              "dscinitialization-v2-validating",
		PlatformValidator: platformValidator,
	}).SetupWithManager(mgr); err != nil {
		return err
	}

	return nil
}
