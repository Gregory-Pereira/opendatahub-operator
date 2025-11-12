package platform

import (
	"context"
	"fmt"

	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	cr "github.com/opendatahub-io/opendatahub-operator/v2/internal/controller/components/registry"
	dscctrl "github.com/opendatahub-io/opendatahub-operator/v2/internal/controller/datasciencecluster"
	dscictrl "github.com/opendatahub-io/opendatahub-operator/v2/internal/controller/dscinitialization"
	sr "github.com/opendatahub-io/opendatahub-operator/v2/internal/controller/services/registry"
)

// SetupCoreReconcilers sets up DSCI and DSC reconcilers.
// These reconcilers are required for all platforms.
func SetupCoreReconcilers(ctx context.Context, mgr ctrl.Manager, componentRegistry *cr.Registry) error {
	if err := dscictrl.NewDSCInitializationReconciler(ctx, mgr); err != nil {
		return fmt.Errorf("unable to create DSCI controller: %w", err)
	}

	if err := dscctrl.NewDataScienceClusterReconciler(ctx, mgr, componentRegistry); err != nil {
		return fmt.Errorf("unable to create DSC controller: %w", err)
	}

	return nil
}

// SetupServiceReconcilers sets up reconcilers for all services in the provided registry.
// Only services registered in the platform's service registry will be set up.
// Some services (like monitoring) need access to the component registry.
func SetupServiceReconcilers(ctx context.Context, mgr ctrl.Manager, serviceRegistry *sr.Registry, componentRegistry *cr.Registry) error {
	l := logf.FromContext(ctx)

	return serviceRegistry.ForEach(func(sh sr.ServiceHandler) error {
		l.Info("creating reconciler", "type", "service", "name", sh.GetName())
		if err := sh.NewReconciler(ctx, mgr, componentRegistry); err != nil {
			return fmt.Errorf("error creating %s service reconciler: %w", sh.GetName(), err)
		}
		return nil
	})
}

// SetupComponentReconcilers sets up reconcilers for all components in the provided registry.
// Only components registered in the platform's component registry will be set up.
func SetupComponentReconcilers(ctx context.Context, mgr ctrl.Manager, componentRegistry *cr.Registry) error {
	l := logf.FromContext(ctx)

	return componentRegistry.ForEach(func(ch cr.ComponentHandler) error {
		l.Info("creating reconciler", "type", "component", "name", ch.GetName())
		if err := ch.NewComponentReconciler(ctx, mgr); err != nil {
			return fmt.Errorf("error creating %s component reconciler: %w", ch.GetName(), err)
		}
		return nil
	})
}
