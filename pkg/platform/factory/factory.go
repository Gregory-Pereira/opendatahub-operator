package factory

import (
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/opendatahub-io/opendatahub-operator/v2/api/common"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/cluster"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/platform"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/platform/managed"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/platform/opendatahub"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/platform/selfmanaged"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/platform/vanilla"
)

// New creates a Platform instance based on the provided platform type.
// The client parameter provides access to the Kubernetes API for platform initialization.
// The oconfig parameter contains operator configuration including monitoring namespace.
// Supported platforms: OpenDataHub, SelfManagedRhoai, ManagedRhoai, Vanilla.
// Returns an error if the platform type is unknown or if platform initialization fails.
func New(platformType common.Platform, cli client.Client, oconfig *cluster.OperatorConfig) (platform.Platform, error) {
	switch platformType {
	case cluster.SelfManagedRhoai:
		return selfmanaged.New(cli, oconfig)
	case cluster.OpenDataHub:
		return opendatahub.New(cli, oconfig)
	case cluster.ManagedRhoai:
		return managed.New(cli, oconfig)
	case cluster.Vanilla:
		return vanilla.New(cli, oconfig)
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
