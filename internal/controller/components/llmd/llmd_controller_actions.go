package llmd

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/controller/actions/render"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/controller/types"
)

func initialize(ctx context.Context, rr *types.ReconciliationRequest) error {
	rr.Manifests = []types.ManifestInfo{
		llmdManifestInfo(llmdManifestSourcePath),
	}

	return nil
}

var _ render.ResourceCustomizer = customizeLlmdResources

func customizeLlmdResources(_ context.Context, _ client.Object, _ *render.GVRKey, _ map[string]any) error {
	// Add any resource customization logic here if needed
	// For now, we rely on the Helm-templated manifests
	return nil
}
