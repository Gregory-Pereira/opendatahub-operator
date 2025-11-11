package vanilla

import (
	"context"
	"fmt"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/opendatahub-io/opendatahub-operator/v2/api/common"
	cr "github.com/opendatahub-io/opendatahub-operator/v2/internal/controller/components/registry"
	sr "github.com/opendatahub-io/opendatahub-operator/v2/internal/controller/services/registry"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/cluster"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/platform"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/upgrade"

	_ "github.com/opendatahub-io/opendatahub-operator/v2/internal/controller/components/kserve"
	_ "github.com/opendatahub-io/opendatahub-operator/v2/internal/controller/services/certconfigmapgenerator"
)

const Type = cluster.Vanilla

var _ platform.Platform = (*Vanilla)(nil)

// Vanilla implements platform-specific behavior for vanilla Kubernetes deployments.
type Vanilla struct {
	client client.Client
	config *cluster.OperatorConfig
}

// New creates a new Vanilla platform instance.
func New(cli client.Client, oconfig *cluster.OperatorConfig) (platform.Platform, error) {
	return &Vanilla{
		client: cli,
		config: oconfig,
	}, nil
}

// Upgrade performs platform-specific upgrade operations for vanilla Kubernetes deployments.
func (v *Vanilla) Upgrade(ctx context.Context) error {
	rel, _ := upgrade.GetDeployedRelease(ctx, v.client)
	if rel.Version.Major == 0 && rel.Version.Minor == 0 && rel.Version.Patch == 0 {
		return nil
	}

	if err := upgrade.CleanupExistingResource(ctx, v.client, Type, rel); err != nil {
		return fmt.Errorf("failed to cleanup existing resources: %w", err)
	}
	return nil
}

// Init performs platform-specific initialization for vanilla Kubernetes deployments.
func (v *Vanilla) Init(ctx context.Context) error {
	// Initialize services
	if err := sr.ForEach(func(sh sr.ServiceHandler) error {
		return sh.Init(Type)
	}); err != nil {
		return fmt.Errorf("unable to init services: %w", err)
	}

	// Initialize components
	if err := cr.ForEach(func(ch cr.ComponentHandler) error {
		return ch.Init(Type)
	}); err != nil {
		return fmt.Errorf("unable to init components: %w", err)
	}

	return nil
}

// Run executes platform-specific runtime logic for vanilla Kubernetes deployments.
func (v *Vanilla) Run(ctx context.Context, mgr ctrl.Manager) error {
	if err := platform.SetupCoreReconcilers(ctx, mgr); err != nil {
		return err
	}

	if err := platform.SetupServiceReconcilers(ctx, mgr); err != nil {
		return err
	}

	if err := platform.SetupComponentReconcilers(ctx, mgr); err != nil {
		return err
	}

	if err := mgr.Add(manager.RunnableFunc(v.setupResources)); err != nil {
		return fmt.Errorf("failed to add setup resources to manager: %w", err)
	}

	return nil
}

// setupResources creates default resources for vanilla Kubernetes deployments.
// Errors are logged but don't block startup (non-blocking).
func (v *Vanilla) setupResources(_ context.Context) error {
	return nil
}

// Validator returns the platform-specific validator for vanilla Kubernetes deployments.
func (v *Vanilla) Validator() platform.Validator {
	return &validator{}
}

// Type returns the platform type identifier for vanilla Kubernetes deployments.
func (v *Vanilla) Type() common.Platform {
	return Type
}

// String returns the canonical platform display name for vanilla Kubernetes deployments.
func (v *Vanilla) String() string {
	return string(Type)
}
