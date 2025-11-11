package managed

import (
	"context"
	"fmt"

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

	"github.com/opendatahub-io/opendatahub-operator/v2/api/common"
	cr "github.com/opendatahub-io/opendatahub-operator/v2/internal/controller/components/registry"
	sr "github.com/opendatahub-io/opendatahub-operator/v2/internal/controller/services/registry"
	"github.com/opendatahub-io/opendatahub-operator/v2/internal/webhook"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/cluster"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/cluster/gvk"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/platform"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/resources"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/upgrade"

	_ "github.com/opendatahub-io/opendatahub-operator/v2/internal/controller/components/dashboard"
	_ "github.com/opendatahub-io/opendatahub-operator/v2/internal/controller/components/datasciencepipelines"
	_ "github.com/opendatahub-io/opendatahub-operator/v2/internal/controller/components/feastoperator"
	_ "github.com/opendatahub-io/opendatahub-operator/v2/internal/controller/components/kserve"
	_ "github.com/opendatahub-io/opendatahub-operator/v2/internal/controller/components/kueue"
	_ "github.com/opendatahub-io/opendatahub-operator/v2/internal/controller/components/llamastackoperator"
	_ "github.com/opendatahub-io/opendatahub-operator/v2/internal/controller/components/modelcontroller"
	_ "github.com/opendatahub-io/opendatahub-operator/v2/internal/controller/components/modelregistry"
	_ "github.com/opendatahub-io/opendatahub-operator/v2/internal/controller/components/ray"
	_ "github.com/opendatahub-io/opendatahub-operator/v2/internal/controller/components/trainingoperator"
	_ "github.com/opendatahub-io/opendatahub-operator/v2/internal/controller/components/trustyai"
	_ "github.com/opendatahub-io/opendatahub-operator/v2/internal/controller/components/workbenches"
	_ "github.com/opendatahub-io/opendatahub-operator/v2/internal/controller/services/auth"
	_ "github.com/opendatahub-io/opendatahub-operator/v2/internal/controller/services/certconfigmapgenerator"
	_ "github.com/opendatahub-io/opendatahub-operator/v2/internal/controller/services/gateway"
	_ "github.com/opendatahub-io/opendatahub-operator/v2/internal/controller/services/monitoring"
	_ "github.com/opendatahub-io/opendatahub-operator/v2/internal/controller/services/setup"
)

const Type = cluster.ManagedRhoai

var _ platform.Platform = (*Managed)(nil)

// Managed implements platform-specific behavior for managed OpenDataHub deployments.
type Managed struct {
	client client.Client
	config *cluster.OperatorConfig
	scheme *runtime.Scheme
}

// New creates a new Managed platform instance.
func New(cli client.Client, oconfig *cluster.OperatorConfig, scheme *runtime.Scheme) (platform.Platform, error) {
	return &Managed{
		client: cli,
		config: oconfig,
		scheme: scheme,
	}, nil
}

// Upgrade performs platform-specific upgrade operations for managed deployments.
func (p *Managed) Upgrade(ctx context.Context) error {
	rel, _ := upgrade.GetDeployedRelease(ctx, p.client)
	if rel.Version.Major == 0 && rel.Version.Minor == 0 && rel.Version.Patch == 0 {
		return nil
	}

	if err := upgrade.CleanupExistingResource(ctx, p.client, Type, rel); err != nil {
		return fmt.Errorf("failed to cleanup existing resources: %w", err)
	}
	return nil
}

// Init performs platform-specific initialization for managed deployments.
func (p *Managed) Init(ctx context.Context) error {
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

// Run executes platform-specific runtime logic for managed deployments.
// This creates the controller-runtime manager, registers webhooks and controllers,
// and starts the manager (blocking until shutdown).
func (p *Managed) Run(ctx context.Context) error {
	logger := ctrl.LoggerFrom(ctx)

	// Create cache configuration
	cacheOptions, err := CreateCacheOptions(p.scheme)
	if err != nil {
		return fmt.Errorf("failed to create cache options: %w", err)
	}

	// Create manager
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
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
	if err := platform.SetupCoreReconcilers(ctx, mgr); err != nil {
		return err
	}

	if err := platform.SetupServiceReconcilers(ctx, mgr); err != nil {
		return err
	}

	if err := platform.SetupComponentReconcilers(ctx, mgr); err != nil {
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
	logger.Info("starting manager", "platform", p.String())
	if err := mgr.Start(ctx); err != nil {
		return fmt.Errorf("problem running manager: %w", err)
	}

	return nil
}

// setupResources creates default resources (DSCI and DSC) for managed deployments.
// Errors block startup (blocking) - managed platform requires default resources.
func (p *Managed) setupResources(ctx context.Context) error {
	l := log.FromContext(ctx)

	l.Info("Creating default DSCInitialization")
	if err := upgrade.CreateDefaultDSCI(ctx, p.client, Type, p.config.MonitoringNamespace); err != nil {
		l.Error(err, "unable to create default DSCInitialization")
		return err // Blocking: return error
	}

	l.Info("Creating default DataScienceCluster")
	if err := upgrade.CreateDefaultDSC(ctx, p.client); err != nil {
		l.Error(err, "unable to create default DataScienceCluster")
		return err // Blocking: return error
	}

	l.Info("Creating default GatewayConfig")
	if err := cluster.CreateGatewayConfig(ctx, p.client); err != nil {
		l.Error(err, "unable to create default GatewayConfig")
		return err // Blocking: return error
	}

	return nil
}

// Validator returns the platform-specific validator for managed deployments.
func (p *Managed) Validator() platform.Validator {
	return &validator{}
}

// Type returns the platform type identifier for managed deployments.
func (p *Managed) Type() common.Platform {
	return Type
}

// String returns the canonical platform display name for managed deployments.
func (p *Managed) String() string {
	return string(Type)
}
