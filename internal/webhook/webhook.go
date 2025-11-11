//go:build !nowebhook

package webhook

import (
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/opendatahub-io/opendatahub-operator/v2/internal/webhook/dashboard"
	dscv1webhook "github.com/opendatahub-io/opendatahub-operator/v2/internal/webhook/datasciencecluster/v1"
	dscv2webhook "github.com/opendatahub-io/opendatahub-operator/v2/internal/webhook/datasciencecluster/v2"
	dsciv1webhook "github.com/opendatahub-io/opendatahub-operator/v2/internal/webhook/dscinitialization/v1"
	dsciv2webhook "github.com/opendatahub-io/opendatahub-operator/v2/internal/webhook/dscinitialization/v2"
	hardwareprofilewebhook "github.com/opendatahub-io/opendatahub-operator/v2/internal/webhook/hardwareprofile"
	kueuewebhook "github.com/opendatahub-io/opendatahub-operator/v2/internal/webhook/kueue"
	notebookwebhook "github.com/opendatahub-io/opendatahub-operator/v2/internal/webhook/notebook"
	serving "github.com/opendatahub-io/opendatahub-operator/v2/internal/webhook/serving"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/platform"
)

// RegisterAllWebhooks registers all webhook setup functions with the given manager.
// Platform validators are injected into DSCI and DSC webhooks for platform-specific validation.
// Returns the first error encountered during registration, or nil if all succeed.
func RegisterAllWebhooks(mgr ctrl.Manager, plat platform.Platform) error {
	// Get platform-specific validators
	platformValidator := plat.Validator()
	dsciValidator := platformValidator.DSCInitializationValidator()
	dscValidator := platformValidator.DataScienceClusterValidator()

	// Register webhooks with platform validators injected where needed
	if err := dscv1webhook.RegisterWebhooks(mgr, dscValidator); err != nil {
		return err
	}
	if err := dscv2webhook.RegisterWebhooks(mgr, dscValidator); err != nil {
		return err
	}
	if err := dsciv1webhook.RegisterWebhooks(mgr, dsciValidator); err != nil {
		return err
	}
	if err := dsciv2webhook.RegisterWebhooks(mgr, dsciValidator); err != nil {
		return err
	}

	// Register other webhooks (no platform validation needed)
	webhookRegistrations := []func(ctrl.Manager) error{
		hardwareprofilewebhook.RegisterWebhooks,
		kueuewebhook.RegisterWebhooks,
		serving.RegisterWebhooks,
		notebookwebhook.RegisterWebhooks,
		dashboard.RegisterWebhooks,
	}
	for _, reg := range webhookRegistrations {
		if err := reg(mgr); err != nil {
			return err
		}
	}

	return nil
}
