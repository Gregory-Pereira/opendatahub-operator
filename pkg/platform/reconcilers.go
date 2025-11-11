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
func SetupCoreReconcilers(ctx context.Context, mgr ctrl.Manager) error {
	if err := dscictrl.NewDSCInitializationReconciler(ctx, mgr); err != nil {
		return fmt.Errorf("unable to create DSCI controller: %w", err)
	}

	if err := dscctrl.NewDataScienceClusterReconciler(ctx, mgr); err != nil {
		return fmt.Errorf("unable to create DSC controller: %w", err)
	}

	return nil
}

// SetupServiceReconcilers sets up reconcilers for all registered services.
// Only services registered via blank imports in the platform package will be set up.
func SetupServiceReconcilers(ctx context.Context, mgr ctrl.Manager) error {
	l := logf.FromContext(ctx)

	return sr.ForEach(func(sh sr.ServiceHandler) error {
		l.Info("creating reconciler", "type", "service", "name", sh.GetName())
		if err := sh.NewReconciler(ctx, mgr); err != nil {
			return fmt.Errorf("error creating %s service reconciler: %w", sh.GetName(), err)
		}
		return nil
	})
}

// SetupComponentReconcilers sets up reconcilers for all registered components.
// Only components registered via blank imports in the platform package will be set up.
func SetupComponentReconcilers(ctx context.Context, mgr ctrl.Manager) error {
	l := logf.FromContext(ctx)

	return cr.ForEach(func(ch cr.ComponentHandler) error {
		l.Info("creating reconciler", "type", "component", "name", ch.GetName())
		if err := ch.NewComponentReconciler(ctx, mgr); err != nil {
			return fmt.Errorf("error creating %s component reconciler: %w", ch.GetName(), err)
		}
		return nil
	})
}
