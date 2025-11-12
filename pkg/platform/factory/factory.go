package factory

import (
	"errors"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"

	"github.com/opendatahub-io/opendatahub-operator/v2/api/common"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/cluster"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/platform"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/platform/managed"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/platform/opendatahub"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/platform/selfmanaged"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/platform/vanilla"
)

// New creates a Platform instance based on the provided platform type.
// The scheme parameter is used to create the controller-runtime manager with proper type registration.
// The oconfig parameter contains operator configuration including rest.Config for cluster connection.
// Supported platforms: OpenDataHub, SelfManagedRhoai, ManagedRhoai, Vanilla.
// Returns an error if the platform type is unknown or if platform initialization fails.
func New(platformType common.Platform, scheme *runtime.Scheme, oconfig *cluster.OperatorConfig) (platform.Platform, error) {
	if oconfig.RestConfig == nil {
		return nil, errors.New("RestConfig must be set in OperatorConfig")
	}

	switch platformType {
	case cluster.SelfManagedRhoai:
		return selfmanaged.New(scheme, oconfig)
	case cluster.OpenDataHub:
		return opendatahub.New(scheme, oconfig)
	case cluster.ManagedRhoai:
		return managed.New(scheme, oconfig)
	case cluster.Vanilla:
		return vanilla.New(scheme, oconfig)
	default:
		return nil, fmt.Errorf("unknown platform type: %s (valid types: %s, %s, %s, %s)",
			platformType,
			cluster.SelfManagedRhoai,
			cluster.OpenDataHub,
			cluster.ManagedRhoai,
			cluster.Vanilla,
		)
	}
}
