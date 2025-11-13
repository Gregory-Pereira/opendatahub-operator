//go:build nowebhook

package webhook

import (
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/platform"
)

// RegisterWebhooks is a no-op stub for builds without webhooks.
func RegisterAllWebhooks(_ ctrl.Manager, _ platform.Platform) error {
	return nil
}
