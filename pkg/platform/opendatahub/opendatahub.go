package opendatahub

import (
	"context"
	"fmt"
	"os"

	userv1 "github.com/openshift/api/user/v1"
	ofapiv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	authorizationv1 "k8s.io/api/authorization/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log"
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
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/cluster/gvk"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/platform"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/platform/support/openshift"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/resources"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/upgrade"
)

const Type = cluster.OpenDataHub

var _ platform.Platform = (*OpenDataHub)(nil)

// OpenDataHub implements platform-specific behavior for OpenDataHub deployments.
type OpenDataHub struct {
	setupClient       client.Client
	config            *cluster.OperatorConfig
	scheme            *runtime.Scheme
	componentRegistry *cr.Registry
	serviceRegistry   *sr.Registry
	meta              platform.Meta
}

// New creates a new OpenDataHub platform instance.
func New(scheme *runtime.Scheme, oconfig *cluster.OperatorConfig) (platform.Platform, error) {
	// Create uncached setup client
	setupClient, err := client.New(oconfig.RestConfig, client.Options{Scheme: scheme})
	if err != nil {
		return nil, fmt.Errorf("failed to create setup client: %w", err)
	}

	return &OpenDataHub{
		setupClient: setupClient,
		config:      oconfig,
		scheme:      scheme,
		componentRegistry: cr.NewRegistry(
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
		),
		serviceRegistry: sr.NewRegistry(
			&auth.ServiceHandler{},
			&certconfigmapgenerator.ServiceHandler{},
			&gateway.ServiceHandler{},
			&monitoring.ServiceHandler{},
			&setup.ServiceHandler{},
		),
	}, nil
}

// Upgrade performs platform-specific upgrade operations for OpenDataHub deployments.
func (p *OpenDataHub) Upgrade(ctx context.Context) error {
	rel, _ := upgrade.GetDeployedRelease(ctx, p.setupClient)
	if rel.Version.Major == 0 && rel.Version.Minor == 0 && rel.Version.Patch == 0 {
		return nil
	}

	if err := upgrade.CleanupExistingResource(ctx, p.setupClient, Type, rel); err != nil {
		return fmt.Errorf("failed to cleanup existing resources: %w", err)
	}
	return nil
}

// Init performs platform-specific initialization for OpenDataHub deployments.
func (p *OpenDataHub) Init(ctx context.Context) error {
	// Discover cluster metadata
	meta, err := openshift.DiscoverMeta(ctx, p.setupClient, Type)
	if err != nil {
		return fmt.Errorf("failed to discover cluster metadata: %w", err)
	}
	p.meta = meta

	// Update global cluster config for backwards compatibility
	cluster.SetMeta(Type, meta.Version, meta.DistributionVersion, meta.Distribution, meta.FIPSEnabled)

	// Set application namespace
	if err := cluster.SetApplicationNamespace(ctx, p.setupClient, Type); err != nil {
		return fmt.Errorf("failed to set application namespace: %w", err)
	}

	// Set monitoring namespace
	cluster.SetManagedMonitoringNamespace(Type)

	// Initialize services
	if err := p.serviceRegistry.ForEach(func(sh sr.ServiceHandler) error {
		return sh.Init(Type)
	}); err != nil {
		return fmt.Errorf("unable to init services: %w", err)
	}

	// Initialize components
	if err := p.componentRegistry.ForEach(func(ch cr.ComponentHandler) error {
		return ch.Init(Type)
	}); err != nil {
		return fmt.Errorf("unable to init components: %w", err)
	}

	return nil
}

// Run executes platform-specific runtime logic for OpenDataHub deployments.
// This creates the controller-runtime manager, registers webhooks and controllers,
// and starts the manager (blocking until shutdown).
func (p *OpenDataHub) Run(ctx context.Context) error {
	logger := ctrl.LoggerFrom(ctx)

	// Create cache configuration
	cacheOptions, err := CreateCacheOptions(p.scheme)
	if err != nil {
		return fmt.Errorf("failed to create cache options: %w", err)
	}

	// Create manager
	mgr, err := ctrl.NewManager(p.config.RestConfig, ctrl.Options{
		Scheme:  p.scheme,
		Metrics: ctrlmetrics.Options{BindAddress: p.config.MetricsAddr},
		WebhookServer: ctrlwebhook.NewServer(ctrlwebhook.Options{
			Port: 9443,
		}),
		PprofBindAddress:       p.config.PprofAddr,
		HealthProbeBindAddress: p.config.HealthProbeAddr,
		Cache:                  cacheOptions,
		LeaderElection:         p.config.LeaderElection,
		LeaderElectionID:       platform.LeaderElectionID,
		Client: client.Options{
			Cache: &client.CacheOptions{
				DisableFor: []client.Object{
					resources.GvkToUnstructured(gvk.OpenshiftIngress),
					&ofapiv1alpha1.Subscription{},
					&authorizationv1.SelfSubjectRulesReview{},
					&corev1.Pod{},
					&userv1.Group{},
					&ofapiv1alpha1.CatalogSource{},
				},
				Unstructured: true,
			},
		},
	})
	if err != nil {
		return fmt.Errorf("unable to create manager: %w", err)
	}

	// Register webhooks
	if err := webhook.RegisterAllWebhooks(mgr, p); err != nil {
		return fmt.Errorf("unable to register webhooks: %w", err)
	}

	// Setup reconcilers
	if err := platform.SetupCoreReconcilers(ctx, mgr, p.componentRegistry); err != nil {
		return err
	}

	if err := platform.SetupServiceReconcilers(ctx, mgr, p.serviceRegistry, p.componentRegistry); err != nil {
		return err
	}

	if err := platform.SetupComponentReconcilers(ctx, mgr, p.componentRegistry); err != nil {
		return err
	}

	// Add platform-specific setup
	if err := mgr.Add(manager.RunnableFunc(p.setupResources)); err != nil {
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
	logger.Info("starting manager", "platform", Type)
	if err := mgr.Start(ctx); err != nil {
		return fmt.Errorf("problem running manager: %w", err)
	}

	return nil
}

// setupResources creates default resources (DSCI and Gateway) for OpenDataHub deployments.
// Errors are logged but don't block startup (non-blocking).
func (p *OpenDataHub) setupResources(ctx context.Context) error {
	l := log.FromContext(ctx)

	if os.Getenv("DISABLE_DSC_CONFIG") != "true" {
		l.Info("Creating default DSCInitialization")
		if err := upgrade.CreateDefaultDSCI(ctx, p.setupClient, Type, p.config.MonitoringNamespace); err != nil {
			l.Error(err, "unable to create default DSCInitialization")
			// Non-blocking: log error but don't return it
		}
	}

	l.Info("Creating default GatewayConfig")
	if err := cluster.CreateGatewayConfig(ctx, p.setupClient); err != nil {
		l.Error(err, "unable to create default GatewayConfig")
		return err
	}

	return nil
}

// Validator returns the platform-specific validator for OpenDataHub deployments.
func (p *OpenDataHub) Validator() platform.Validator {
	return &validator{}
}

// Meta returns a copy of the platform's cluster metadata.
func (p *OpenDataHub) Meta() platform.Meta {
	return p.meta
}
