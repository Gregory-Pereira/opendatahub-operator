package openshift

import (
	"context"
	"os"

	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/opendatahub-io/opendatahub-operator/v2/api/common"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/cluster"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/upgrade"
)

// Variant holds configuration for OpenShift platform variants.
// Different variants (Managed, SelfManaged, OpenDataHub) are distinguished
// primarily by naming conventions and initialization behavior.
type Variant struct {
	// Type is the platform type identifier
	Type common.Platform

	// Name is the internal name of the variant
	Name string

	// DisplayName is the human-readable name
	DisplayName string

	// MonitoringNamespace is the namespace where monitoring resources are deployed
	MonitoringNamespace string

	// AdminGroupName is the default admin group name for this variant
	AdminGroupName string

	// ConsoleNamespace is the namespace for console link resources
	// Empty for non-managed variants
	ConsoleNamespace string

	// SubscriptionName is the operator subscription name for uninstall
	// Empty means skip subscription deletion (managed variant)
	SubscriptionName string

	// PreRun is called during platform Run() before starting the controller manager
	// It performs variant-specific initialization like creating default resources
	PreRun func(ctx context.Context, p *OpenShift) error
}

var (
	// Managed is the configuration for Red Hat OpenShift AI (Managed) variant.
	Managed = Variant{
		Type:                cluster.ManagedRhoai,
		Name:                "managed-rhoai",
		DisplayName:         "Red Hat OpenShift AI (Managed)",
		MonitoringNamespace: cluster.DefaultMonitoringNamespaceRHOAI,
		AdminGroupName:      cluster.DefaultAdminGroupManaged,
		ConsoleNamespace:    cluster.DefaultConsoleLinkNamespace,
		SubscriptionName:    "", // Don't delete subscription on uninstall
		PreRun:              managedPreRun,
	}

	// SelfManaged is the configuration for Red Hat OpenShift AI (Self-Managed) variant.
	SelfManaged = Variant{
		Type:                cluster.SelfManagedRhoai,
		Name:                "selfmanaged-rhoai",
		DisplayName:         "Red Hat OpenShift AI (Self-Managed)",
		MonitoringNamespace: cluster.DefaultMonitoringNamespaceRHOAI,
		AdminGroupName:      cluster.DefaultAdminGroupSelfManaged,
		ConsoleNamespace:    "",
		SubscriptionName:    cluster.SubscriptionNameRHOAI,
		PreRun:              nonManagedPreRun,
	}

	// OpenDataHub is the configuration for Open Data Hub variant.
	OpenDataHub = Variant{
		Type:                cluster.OpenDataHub,
		Name:                "opendatahub",
		DisplayName:         "Open Data Hub",
		MonitoringNamespace: cluster.DefaultMonitoringNamespaceODH,
		AdminGroupName:      cluster.DefaultAdminGroupODH,
		ConsoleNamespace:    "",
		SubscriptionName:    cluster.SubscriptionNameODH,
		PreRun:              nonManagedPreRun,
	}
)

// managedPreRun creates default resources for the Managed variant.
// All resource creation errors block startup.
func managedPreRun(ctx context.Context, p *OpenShift) error {
	log := ctrl.LoggerFrom(ctx)
	cli := p.setupClient
	variant := p.variant

	log.Info("Creating default DSCInitialization")
	if err := upgrade.CreateDefaultDSCI(ctx, cli, variant.Type, variant.MonitoringNamespace); err != nil {
		log.Error(err, "unable to create default DSCInitialization")
		return err
	}

	log.Info("Creating default DataScienceCluster")
	if err := upgrade.CreateDefaultDSC(ctx, cli); err != nil {
		log.Error(err, "unable to create default DataScienceCluster")
		return err
	}

	log.Info("Creating default GatewayConfig")
	if err := cluster.CreateGatewayConfig(ctx, cli); err != nil {
		log.Error(err, "unable to create default GatewayConfig")
		return err
	}

	return nil
}

// nonManagedPreRun creates default resources for non-managed variants (SelfManaged and OpenDataHub).
// DSCI creation respects DISABLE_DSC_CONFIG environment variable and errors are non-blocking.
// GatewayConfig errors block startup.
func nonManagedPreRun(ctx context.Context, p *OpenShift) error {
	log := ctrl.LoggerFrom(ctx)
	cli := p.setupClient
	variant := p.variant

	if os.Getenv("DISABLE_DSC_CONFIG") != "true" {
		log.Info("Creating default DSCInitialization")
		if err := upgrade.CreateDefaultDSCI(ctx, cli, variant.Type, variant.MonitoringNamespace); err != nil {
			log.Error(err, "unable to create default DSCInitialization")
			// Non-blocking: log but continue
		}
	}

	log.Info("Creating default GatewayConfig")
	if err := cluster.CreateGatewayConfig(ctx, cli); err != nil {
		log.Error(err, "unable to create default GatewayConfig")
		return err
	}

	return nil
}
