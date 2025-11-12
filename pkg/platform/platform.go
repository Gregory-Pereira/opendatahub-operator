package platform

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/opendatahub-io/opendatahub-operator/v2/api/common"
)

const (
	// LeaderElectionID is the resource name used for leader election across all platforms.
	LeaderElectionID = "07ed84f7.opendatahub.io"
)

// Meta contains cluster metadata discovered during platform initialization.
type Meta struct {
	Type                common.Platform // Platform type (Vanilla, OpenDataHub, ManagedRHOAI, SelfManagedRHOAI)
	Version             string          // Operator version (from CSV)
	DistributionVersion string          // K8s distribution version (e.g., "1.28.0", "4.15.0")
	Distribution        string          // K8s distribution (OpenShift, Vanilla, GKE, EKS, AKS)
	FIPSEnabled         bool            // FIPS status
}

// Platform defines platform-specific behavior for the operator.
// Different platform variants (SelfManaged, OpenDataHub, Managed, Vanilla) can implement
// custom logic for upgrades, initialization, runtime behavior, and validation.
// Platform implements fmt.Stringer to provide the canonical platform display name.
type Platform interface {
	// Upgrade performs platform-specific upgrade operations.
	// This is called when the operator detects a version change.
	Upgrade(ctx context.Context) error

	// Init performs platform-specific initialization.
	// This is called during operator startup before controllers run.
	Init(ctx context.Context) error

	// Run executes platform-specific runtime logic.
	// This creates the controller-runtime manager, registers webhooks and controllers,
	// and starts the manager (blocking until shutdown).
	Run(ctx context.Context) error

	// Validator returns the platform-specific validator for webhooks.
	// The validator provides admission handlers for DSCInitialization and DataScienceCluster.
	Validator() Validator

	// Meta returns a copy of the platform's cluster metadata.
	// Metadata is populated during Init() and contains cluster characteristics.
	Meta() Meta
}

// Validator provides platform-specific admission webhook handlers.
// Different platforms may enforce different validation rules for CRs.
type Validator interface {
	// DSCInitializationValidator returns the admission handler for DSCInitialization validation.
	DSCInitializationValidator() admission.Handler

	// DataScienceClusterValidator returns the admission handler for DataScienceCluster validation.
	DataScienceClusterValidator() admission.Handler
}
