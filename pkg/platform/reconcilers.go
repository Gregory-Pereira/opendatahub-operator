package platform

import (
	"context"
	"fmt"
	"os"

	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	cr "github.com/opendatahub-io/opendatahub-operator/v2/internal/controller/components/registry"
	dscctrl "github.com/opendatahub-io/opendatahub-operator/v2/internal/controller/datasciencecluster"
	dscictrl "github.com/opendatahub-io/opendatahub-operator/v2/internal/controller/dscinitialization"
	sr "github.com/opendatahub-io/opendatahub-operator/v2/internal/controller/services/registry"
)

// SetupCoreReconcilers sets up DSCI and DSC reconcilers.
// DSCI reconciler is always created. DSC reconciler is created unless DISABLE_DSC_CONTROLLER=true.
func SetupCoreReconcilers(ctx context.Context, mgr ctrl.Manager, plat Platform) error {
	if os.Getenv("DISABLE_DSCI_CONTROLLER") != "true" {
		if err := SetupDSCIReconciler(ctx, mgr, plat.Meta().Distribution); err != nil {
			return err
		}
	}

	if os.Getenv("DISABLE_DSC_CONTROLLER") != "true" {
		if err := SetupDSCReconciler(ctx, mgr, plat.ComponentRegistry()); err != nil {
			return err
		}
	}

	return nil
}

// SetupDSCIReconciler creates the DSCInitialization reconciler.
func SetupDSCIReconciler(ctx context.Context, mgr ctrl.Manager, dist string) error {
	if err := dscictrl.NewDSCInitializationReconciler(ctx, mgr, dist); err != nil {
		return fmt.Errorf("unable to create DSCI controller: %w", err)
	}
	return nil
}

// SetupDSCReconciler creates the DataScienceCluster reconciler.
func SetupDSCReconciler(ctx context.Context, mgr ctrl.Manager, compRegistry *cr.Registry) error {
	if err := dscctrl.NewDataScienceClusterReconciler(ctx, mgr, compRegistry); err != nil {
		return fmt.Errorf("unable to create DSC controller: %w", err)
	}
	return nil
}

// SetupServiceReconcilers sets up reconcilers for all services in the platform's service registry.
// Some services (like monitoring) need access to the component registry.
func SetupServiceReconcilers(ctx context.Context, mgr ctrl.Manager, plat Platform) error {
	l := logf.FromContext(ctx)

	return plat.ServiceRegistry().ForEach(func(sh sr.ServiceHandler) error {
		l.Info("creating reconciler", "type", "service", "name", sh.GetName())
		if err := sh.NewReconciler(ctx, mgr, plat.ComponentRegistry()); err != nil {
			return fmt.Errorf("error creating %s service reconciler: %w", sh.GetName(), err)
		}
		return nil
	})
}

// SetupComponentReconcilers sets up reconcilers for all components in the platform's component registry.
func SetupComponentReconcilers(ctx context.Context, mgr ctrl.Manager, plat Platform) error {
	l := logf.FromContext(ctx)

	return plat.ComponentRegistry().ForEach(func(ch cr.ComponentHandler) error {
		l.Info("creating reconciler", "type", "component", "name", ch.GetName())
		if err := ch.NewComponentReconciler(ctx, mgr); err != nil {
			return fmt.Errorf("error creating %s component reconciler: %w", ch.GetName(), err)
		}
		return nil
	})
}
