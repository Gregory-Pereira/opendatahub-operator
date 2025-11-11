package vanilla

import (
	"context"
	"net/http"

	operatorv1 "github.com/openshift/api/operator/v1"
	admissionv1 "k8s.io/api/admission/v1"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/opendatahub-io/opendatahub-operator/v2/api/common"
	dscv1 "github.com/opendatahub-io/opendatahub-operator/v2/api/datasciencecluster/v1"
	dsciv2 "github.com/opendatahub-io/opendatahub-operator/v2/api/dscinitialization/v2"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/platform"
)

var _ platform.Validator = (*validator)(nil)
var _ admission.Handler = (*dsciHandler)(nil)
var _ admission.Handler = (*dscHandler)(nil)

// isComponentRemoved returns true if the component's ManagementState is empty or Removed.
// This indicates the component is not enabled.
func isComponentRemoved(spec common.ManagementSpec) bool {
	return spec.ManagementState == "" || spec.ManagementState == operatorv1.Removed
}

type validator struct{}

// DSCInitializationValidator returns the admission handler for DSCInitialization validation
// in vanilla Kubernetes deployments.
func (p *validator) DSCInitializationValidator() admission.Handler {
	return &dsciHandler{}
}

// DataScienceClusterValidator returns the admission handler for DataScienceCluster validation
// in vanilla Kubernetes deployments.
func (p *validator) DataScienceClusterValidator() admission.Handler {
	return &dscHandler{}
}

// dsciHandler validates DSCInitialization resources for vanilla Kubernetes deployments.
// Only allows spec.applicationsNamespace and spec.trustedCABundle to be set.
// spec.monitoring must be empty or Removed, spec.devFlags must be nil.
type dsciHandler struct {
	decoder admission.Decoder
}

// InjectDecoder injects the decoder into the dsciHandler.
func (h *dsciHandler) InjectDecoder(d admission.Decoder) error {
	h.decoder = d
	return nil
}

func (h *dsciHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	if req.Operation == admissionv1.Delete {
		return admission.Allowed("DELETE operation is allowed")
	}

	dsci := &dsciv2.DSCInitialization{}
	if err := h.decoder.DecodeRaw(req.Object, dsci); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	if !isComponentRemoved(dsci.Spec.Monitoring.ManagementSpec) {
		return admission.Denied("spec.monitoring cannot be enabled for vanilla Kubernetes deployments (must be empty or Removed)")
	}

	if dsci.Spec.DevFlags != nil {
		return admission.Denied("spec.devFlags cannot be set for vanilla Kubernetes deployments")
	}

	return admission.Allowed("DSCInitialization validation passed")
}

// dscHandler validates DataScienceCluster resources for vanilla Kubernetes deployments.
// Only allows spec.components.kserve to be set.
// All other components must have ManagementState empty or Removed.
type dscHandler struct {
	decoder admission.Decoder
}

// InjectDecoder injects the decoder into the dscHandler.
func (h *dscHandler) InjectDecoder(d admission.Decoder) error {
	h.decoder = d
	return nil
}

func (h *dscHandler) Handle(_ context.Context, req admission.Request) admission.Response {
	if req.Operation == admissionv1.Delete {
		return admission.Allowed("DELETE operation is allowed")
	}

	dsc := &dscv1.DataScienceCluster{}
	if err := h.decoder.DecodeRaw(req.Object, dsc); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	if !isComponentRemoved(dsc.Spec.Components.Dashboard.ManagementSpec) {
		return admission.Denied("spec.components.dashboard cannot be enabled for vanilla Kubernetes deployments (must be empty or Removed)")
	}

	if !isComponentRemoved(dsc.Spec.Components.Workbenches.ManagementSpec) {
		return admission.Denied("spec.components.workbenches cannot be enabled for vanilla Kubernetes deployments (must be empty or Removed)")
	}

	if !isComponentRemoved(dsc.Spec.Components.ModelMeshServing.ManagementSpec) {
		return admission.Denied("spec.components.modelmeshserving cannot be enabled for vanilla Kubernetes deployments (must be empty or Removed)")
	}

	if !isComponentRemoved(dsc.Spec.Components.DataSciencePipelines.ManagementSpec) {
		return admission.Denied("spec.components.datasciencepipelines cannot be enabled for vanilla Kubernetes deployments (must be empty or Removed)")
	}

	if dsc.Spec.Components.Kueue.ManagementState != "" && dsc.Spec.Components.Kueue.ManagementState != operatorv1.Removed {
		return admission.Denied("spec.components.kueue cannot be enabled for vanilla Kubernetes deployments (must be empty or Removed)")
	}

	if !isComponentRemoved(dsc.Spec.Components.CodeFlare.ManagementSpec) {
		return admission.Denied("spec.components.codeflare cannot be enabled for vanilla Kubernetes deployments (must be empty or Removed)")
	}

	if !isComponentRemoved(dsc.Spec.Components.Ray.ManagementSpec) {
		return admission.Denied("spec.components.ray cannot be enabled for vanilla Kubernetes deployments (must be empty or Removed)")
	}

	if !isComponentRemoved(dsc.Spec.Components.TrustyAI.ManagementSpec) {
		return admission.Denied("spec.components.trustyai cannot be enabled for vanilla Kubernetes deployments (must be empty or Removed)")
	}

	if !isComponentRemoved(dsc.Spec.Components.ModelRegistry.ManagementSpec) {
		return admission.Denied("spec.components.modelregistry cannot be enabled for vanilla Kubernetes deployments (must be empty or Removed)")
	}

	if !isComponentRemoved(dsc.Spec.Components.TrainingOperator.ManagementSpec) {
		return admission.Denied("spec.components.trainingoperator cannot be enabled for vanilla Kubernetes deployments (must be empty or Removed)")
	}

	if !isComponentRemoved(dsc.Spec.Components.FeastOperator.ManagementSpec) {
		return admission.Denied("spec.components.feastoperator cannot be enabled for vanilla Kubernetes deployments (must be empty or Removed)")
	}

	if !isComponentRemoved(dsc.Spec.Components.LlamaStackOperator.ManagementSpec) {
		return admission.Denied("spec.components.llamastackoperator cannot be enabled for vanilla Kubernetes deployments (must be empty or Removed)")
	}

	return admission.Allowed("DataScienceCluster validation passed")
}
