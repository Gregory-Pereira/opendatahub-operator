package vanilla

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/platform"
)

var _ platform.Validator = (*validator)(nil)
var _ admission.Handler = (*dsciHandler)(nil)
var _ admission.Handler = (*dscHandler)(nil)

type validator struct{}

func (p *validator) DSCInitializationValidator() admission.Handler {
	return &dsciHandler{}
}

func (p *validator) DataScienceClusterValidator() admission.Handler {
	return &dscHandler{}
}

type dsciHandler struct {
	decoder admission.Decoder
}

// InjectDecoder injects the decoder into the dsciHandler.
func (h *dsciHandler) InjectDecoder(d admission.Decoder) error {
	h.decoder = d
	return nil
}

func (h *dsciHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	return admission.Allowed("DSCInitialization validation passed")
}

type dscHandler struct {
	decoder admission.Decoder
}

// InjectDecoder injects the decoder into the dscHandler.
func (h *dscHandler) InjectDecoder(d admission.Decoder) error {
	h.decoder = d
	return nil
}

func (h *dscHandler) Handle(_ context.Context, req admission.Request) admission.Response {
	return admission.Allowed("DataScienceCluster validation passed")
}
