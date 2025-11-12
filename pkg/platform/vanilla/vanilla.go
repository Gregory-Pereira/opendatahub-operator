package vanilla

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	ctrlmetrics "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	ctrlwebhook "sigs.k8s.io/controller-runtime/pkg/webhook"

	"github.com/opendatahub-io/opendatahub-operator/v2/internal/controller/components/kserve"
	cr "github.com/opendatahub-io/opendatahub-operator/v2/internal/controller/components/registry"
	"github.com/opendatahub-io/opendatahub-operator/v2/internal/controller/services/certconfigmapgenerator"
	sr "github.com/opendatahub-io/opendatahub-operator/v2/internal/controller/services/registry"
	"github.com/opendatahub-io/opendatahub-operator/v2/internal/webhook"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/cluster"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/platform"
	vanillaSupport "github.com/opendatahub-io/opendatahub-operator/v2/pkg/platform/support/vanilla"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/upgrade"
)

const Type = cluster.Vanilla

var _ platform.Platform = (*Vanilla)(nil)

// Vanilla implements platform-specific behavior for vanilla Kubernetes deployments.
type Vanilla struct {
	setupClient       client.Client
	config            *cluster.OperatorConfig
	scheme            *runtime.Scheme
	componentRegistry *cr.Registry
	serviceRegistry   *sr.Registry
	meta              platform.Meta
}

// New creates a new Vanilla platform instance.
func New(scheme *runtime.Scheme, oconfig *cluster.OperatorConfig) (platform.Platform, error) {
	// Create uncached setup client
	setupClient, err := client.New(oconfig.RestConfig, client.Options{Scheme: scheme})
	if err != nil {
		return nil, fmt.Errorf("failed to create setup client: %w", err)
	}

	return &Vanilla{
		setupClient: setupClient,
		config:      oconfig,
		scheme:      scheme,
		componentRegistry: cr.NewRegistry(
			&kserve.ComponentHandler{},
		),
		serviceRegistry: sr.NewRegistry(
			&certconfigmapgenerator.ServiceHandler{},
		),
	}, nil
}

// Upgrade performs platform-specific upgrade operations for vanilla Kubernetes deployments.
func (v *Vanilla) Upgrade(ctx context.Context) error {
	rel, _ := upgrade.GetDeployedRelease(ctx, v.setupClient)
	if rel.Version.Major == 0 && rel.Version.Minor == 0 && rel.Version.Patch == 0 {
		return nil
	}

	if err := upgrade.CleanupExistingResource(ctx, v.setupClient, Type, rel); err != nil {
		return fmt.Errorf("failed to cleanup existing resources: %w", err)
	}
	return nil
}

// Init performs platform-specific initialization for vanilla Kubernetes deployments.
func (v *Vanilla) Init(ctx context.Context) error {
	// Discover cluster metadata
	meta, err := vanillaSupport.DiscoverMeta(ctx, v.setupClient, v.config.RestConfig, Type)
	if err != nil {
		return fmt.Errorf("failed to discover cluster metadata: %w", err)
	}
	v.meta = meta

	// Update global cluster config for backwards compatibility
	cluster.SetMeta(Type, meta.Version, meta.DistributionVersion, meta.Distribution, meta.FIPSEnabled)

	// Initialize services
	if err := v.serviceRegistry.ForEach(func(sh sr.ServiceHandler) error {
		return sh.Init(Type)
	}); err != nil {
		return fmt.Errorf("unable to init services: %w", err)
	}

	// Initialize components
	if err := v.componentRegistry.ForEach(func(ch cr.ComponentHandler) error {
		return ch.Init(Type)
	}); err != nil {
		return fmt.Errorf("unable to init components: %w", err)
	}

	return nil
}

// Run executes platform-specific runtime logic for vanilla Kubernetes deployments.
// This creates the controller-runtime manager, registers webhooks and controllers,
// and starts the manager (blocking until shutdown).
func (v *Vanilla) Run(ctx context.Context) error {
	log := ctrl.LoggerFrom(ctx)

	// Create cache configuration
	cacheOptions, err := CreateCacheOptions(v.scheme)
	if err != nil {
		return fmt.Errorf("failed to create cache options: %w", err)
	}

	// Create manager
	mgr, err := ctrl.NewManager(v.config.RestConfig, ctrl.Options{
		Scheme:  v.scheme,
		Metrics: ctrlmetrics.Options{BindAddress: v.config.MetricsAddr},
		WebhookServer: ctrlwebhook.NewServer(ctrlwebhook.Options{
			Port: 9443,
		}),
		PprofBindAddress:       v.config.PprofAddr,
		HealthProbeBindAddress: v.config.HealthProbeAddr,
		Cache:                  cacheOptions,
		LeaderElection:         v.config.LeaderElection,
		LeaderElectionID:       platform.LeaderElectionID,
		Client: client.Options{
			Cache: &client.CacheOptions{
				DisableFor: []client.Object{
					&corev1.Pod{},
				},
				Unstructured: true,
			},
		},
	})
	if err != nil {
		return fmt.Errorf("unable to create manager: %w", err)
	}

	// Register webhooks
	if err := webhook.RegisterAllWebhooks(mgr, v); err != nil {
		return fmt.Errorf("unable to register webhooks: %w", err)
	}

	// Setup reconcilers
	if err := platform.SetupCoreReconcilers(ctx, mgr, v.componentRegistry); err != nil {
		return err
	}

	if err := platform.SetupServiceReconcilers(ctx, mgr, v.serviceRegistry, v.componentRegistry); err != nil {
		return err
	}

	if err := platform.SetupComponentReconcilers(ctx, mgr, v.componentRegistry); err != nil {
		return err
	}

	// Add platform-specific setup
	if err := mgr.Add(manager.RunnableFunc(v.setupResources)); err != nil {
		return fmt.Errorf("failed to add setup resources to manager: %w", err)
	}

	// Add health checks
	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		return fmt.Errorf("unable to set up health check: %w", err)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		return fmt.Errorf("unable to set up ready check: %w", err)
	}

	// Start manager (blocking)
	log.Info("starting manager", "platform", Type)
	if err := mgr.Start(ctx); err != nil {
		return fmt.Errorf("problem running manager: %w", err)
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

// Meta returns a copy of the platform's cluster metadata.
func (v *Vanilla) Meta() platform.Meta {
	return v.meta
}
