package openshift

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	ctrlmetrics "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	ctrlwebhook "sigs.k8s.io/controller-runtime/pkg/webhook"

	"github.com/opendatahub-io/opendatahub-operator/v2/internal/controller/components/dashboard"
	"github.com/opendatahub-io/opendatahub-operator/v2/internal/controller/components/datasciencepipelines"
	"github.com/opendatahub-io/opendatahub-operator/v2/internal/controller/components/feastoperator"
	"github.com/opendatahub-io/opendatahub-operator/v2/internal/controller/components/kserve"
	"github.com/opendatahub-io/opendatahub-operator/v2/internal/controller/components/kueue"
	"github.com/opendatahub-io/opendatahub-operator/v2/internal/controller/components/llamastackoperator"
	"github.com/opendatahub-io/opendatahub-operator/v2/internal/controller/components/modelcontroller"
	"github.com/opendatahub-io/opendatahub-operator/v2/internal/controller/components/modelregistry"
	"github.com/opendatahub-io/opendatahub-operator/v2/internal/controller/components/ray"
	cr "github.com/opendatahub-io/opendatahub-operator/v2/internal/controller/components/registry"
	"github.com/opendatahub-io/opendatahub-operator/v2/internal/controller/components/trainingoperator"
	"github.com/opendatahub-io/opendatahub-operator/v2/internal/controller/components/trustyai"
	"github.com/opendatahub-io/opendatahub-operator/v2/internal/controller/components/workbenches"
	"github.com/opendatahub-io/opendatahub-operator/v2/internal/controller/services/auth"
	"github.com/opendatahub-io/opendatahub-operator/v2/internal/controller/services/certconfigmapgenerator"
	"github.com/opendatahub-io/opendatahub-operator/v2/internal/controller/services/gateway"
	"github.com/opendatahub-io/opendatahub-operator/v2/internal/controller/services/monitoring"
	sr "github.com/opendatahub-io/opendatahub-operator/v2/internal/controller/services/registry"
	"github.com/opendatahub-io/opendatahub-operator/v2/internal/controller/services/setup"
	"github.com/opendatahub-io/opendatahub-operator/v2/internal/webhook"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/cluster"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/platform"
	openshiftsupport "github.com/opendatahub-io/opendatahub-operator/v2/pkg/platform/support/openshift"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/upgrade"
)

var _ platform.Platform = (*OpenShift)(nil)

// OpenShift implements platform-specific behavior for OpenShift-based deployments.
// It consolidates common logic for Managed, SelfManaged, and OpenDataHub variants,
// with variant-specific behavior delegated to Variant.PreRun hooks.
type OpenShift struct {
	setupClient       client.Client
	operatorConfig    *cluster.OperatorConfig
	variant           Variant
	scheme            *runtime.Scheme
	componentRegistry *cr.Registry
	serviceRegistry   *sr.Registry
	meta              platform.Meta
}

// New creates a new OpenShift platform instance with the specified variant configuration.
func New(
	scheme *runtime.Scheme,
	oconfig *cluster.OperatorConfig,
	variant Variant,
) (platform.Platform, error) {
	// Create uncached setup client
	setupClient, err := client.New(oconfig.RestConfig, client.Options{Scheme: scheme})
	if err != nil {
		return nil, fmt.Errorf("failed to create setup client: %w", err)
	}

	// All OpenShift variants have identical component and service registries
	componentRegistry := cr.NewRegistry(
		&dashboard.ComponentHandler{},
		&datasciencepipelines.ComponentHandler{},
		&feastoperator.ComponentHandler{},
		&kserve.ComponentHandler{},
		&kueue.ComponentHandler{},
		&llamastackoperator.ComponentHandler{},
		&modelcontroller.ComponentHandler{},
		&modelregistry.ComponentHandler{},
		&ray.ComponentHandler{},
		&trainingoperator.ComponentHandler{},
		&trustyai.ComponentHandler{},
		&workbenches.ComponentHandler{},
	)

	serviceRegistry := sr.NewRegistry(
		&auth.ServiceHandler{},
		&certconfigmapgenerator.ServiceHandler{},
		&gateway.ServiceHandler{},
		&monitoring.ServiceHandler{},
		&setup.ServiceHandler{},
	)

	return &OpenShift{
		setupClient:       setupClient,
		operatorConfig:    oconfig,
		variant:           variant,
		scheme:            scheme,
		componentRegistry: componentRegistry,
		serviceRegistry:   serviceRegistry,
	}, nil
}

// Upgrade performs platform-specific upgrade operations.
func (p *OpenShift) Upgrade(ctx context.Context) error {
	rel, _ := upgrade.GetDeployedRelease(ctx, p.setupClient)
	if rel.Version.Major == 0 && rel.Version.Minor == 0 && rel.Version.Patch == 0 {
		return nil
	}

	if err := upgrade.CleanupExistingResource(ctx, p.setupClient, p.variant.Type, rel); err != nil {
		return fmt.Errorf("failed to cleanup existing resources: %w", err)
	}
	return nil
}

// Init performs platform-specific initialization.
func (p *OpenShift) Init(ctx context.Context) error {
	// Discover cluster metadata
	meta, err := openshiftsupport.DiscoverMeta(ctx, p.setupClient, p.variant.Type)
	if err != nil {
		return fmt.Errorf("failed to discover cluster metadata: %w", err)
	}
	p.meta = meta

	// Update global cluster config for backwards compatibility
	cluster.SetMeta(
		p.variant.Type,
		meta.Version,
		meta.DistributionVersion,
		meta.Distribution,
		meta.FIPSEnabled,
	)

	// Set application namespace
	if err := cluster.SetApplicationNamespace(ctx, p.setupClient, p.variant.Type); err != nil {
		return fmt.Errorf("failed to set application namespace: %w", err)
	}

	// Set monitoring namespace
	cluster.SetManagedMonitoringNamespace(p.variant.Type)

	// Initialize services
	if err := p.serviceRegistry.ForEach(func(sh sr.ServiceHandler) error {
		return sh.Init(p.variant.Type)
	}); err != nil {
		return fmt.Errorf("unable to init services: %w", err)
	}

	// Initialize components
	if err := p.componentRegistry.ForEach(func(ch cr.ComponentHandler) error {
		return ch.Init(p.variant.Type)
	}); err != nil {
		return fmt.Errorf("unable to init components: %w", err)
	}

	return nil
}

// Run executes platform-specific runtime logic.
// This creates the controller-runtime manager, registers webhooks and controllers,
// and starts the manager (blocking until shutdown).
func (p *OpenShift) Run(ctx context.Context) error {
	logger := ctrl.LoggerFrom(ctx)

	// Create cache configuration
	cacheOptions, err := CreateCacheOptions(p.scheme, p.variant)
	if err != nil {
		return fmt.Errorf("failed to create cache options: %w", err)
	}

	// Create manager
	mgr, err := ctrl.NewManager(p.operatorConfig.RestConfig, ctrl.Options{
		Scheme:  p.scheme,
		Metrics: ctrlmetrics.Options{BindAddress: p.operatorConfig.MetricsAddr},
		WebhookServer: ctrlwebhook.NewServer(ctrlwebhook.Options{
			Port: 9443,
		}),
		PprofBindAddress:       p.operatorConfig.PprofAddr,
		HealthProbeBindAddress: p.operatorConfig.HealthProbeAddr,
		Cache:                  cacheOptions,
		LeaderElection:         p.operatorConfig.LeaderElection,
		LeaderElectionID:       platform.LeaderElectionID,
		Client:                 CreateClientOptions(),
	})
	if err != nil {
		return fmt.Errorf("unable to create manager: %w", err)
	}

	// Register webhooks
	if err := webhook.RegisterAllWebhooks(mgr, p); err != nil {
		return fmt.Errorf("unable to register webhooks: %w", err)
	}

	// Setup reconcilers
	if err := platform.SetupCoreReconcilers(ctx, mgr, p); err != nil {
		return err
	}

	if err := platform.SetupServiceReconcilers(ctx, mgr, p); err != nil {
		return err
	}

	if err := platform.SetupComponentReconcilers(ctx, mgr, p); err != nil {
		return err
	}

	// Add platform-specific setup via variant PreRun hook
	if p.variant.PreRun != nil {
		if err := mgr.Add(manager.RunnableFunc(func(ctx context.Context) error {
			return p.variant.PreRun(ctx, p)
		})); err != nil {
			return fmt.Errorf("failed to add variant setup to manager: %w", err)
		}
	}

	// Add health checks
	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		return fmt.Errorf("unable to set up health check: %w", err)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		return fmt.Errorf("unable to set up ready check: %w", err)
	}

	// Start manager (blocking)
	logger.Info("starting manager",
		"platform", p.variant.Type,
		"variant", p.variant.Name,
	)
	if err := mgr.Start(ctx); err != nil {
		return fmt.Errorf("problem running manager: %w", err)
	}

	return nil
}

// Validator returns the platform-specific validator.
func (p *OpenShift) Validator() platform.Validator {
	return &validator{}
}

// Meta returns a copy of the platform's cluster metadata.
func (p *OpenShift) Meta() platform.Meta {
	return p.meta
}

// ComponentRegistry returns the platform's component registry.
func (p *OpenShift) ComponentRegistry() *cr.Registry {
	return p.componentRegistry
}

// ServiceRegistry returns the platform's service registry.
func (p *OpenShift) ServiceRegistry() *sr.Registry {
	return p.serviceRegistry
}
