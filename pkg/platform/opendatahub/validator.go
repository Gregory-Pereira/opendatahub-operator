package opendatahub

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/platform"
)

var _ platform.Validator = (*validator)(nil)
var _ admission.Handler = (*dsciHandler)(nil)
var _ admission.Handler = (*dscHandler)(nil)

type validator struct{}

// DSCInitializationValidator returns the admission handler for DSCInitialization validation
// in OpenDataHub deployments.
func (v *validator) DSCInitializationValidator() admission.Handler {
	// TODO: implement platform-specific DSCInitialization validation logic
	return &dsciHandler{}
}

// DataScienceClusterValidator returns the admission handler for DataScienceCluster validation
// in OpenDataHub deployments.
func (v *validator) DataScienceClusterValidator() admission.Handler {
	// TODO: implement platform-specific DataScienceCluster validation logic
	return &dscHandler{}
}

// dsciHandler is a temporary admission handler for DSCInitialization validation.
// This will be replaced with actual platform-specific validation logic.
type dsciHandler struct{}

func (h *dsciHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	return admission.Allowed("validation not yet implemented")
}

// dscHandler is a temporary admission handler for DataScienceCluster validation.
// This will be replaced with actual platform-specific validation logic.
type dscHandler struct{}

func (h *dscHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	return admission.Allowed("validation not yet implemented")
}
