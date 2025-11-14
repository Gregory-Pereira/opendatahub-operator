package openshift

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/platform"
)

var _ platform.Validator = (*validator)(nil)
var _ admission.Handler = (*dsciHandler)(nil)
var _ admission.Handler = (*dscHandler)(nil)

type validator struct{}

func (v *validator) DSCInitializationValidator() admission.Handler {
	return &dsciHandler{}
}

func (v *validator) DataScienceClusterValidator() admission.Handler {
	return &dscHandler{}
}

type dsciHandler struct{}

func (h *dsciHandler) Handle(_ context.Context, _ admission.Request) admission.Response {
	return admission.Allowed("validation not yet implemented")
}

type dscHandler struct{}

func (h *dscHandler) Handle(_ context.Context, _ admission.Request) admission.Response {
	return admission.Allowed("validation not yet implemented")
}
