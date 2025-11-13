package setup

import (
	"context"
	"fmt"

	operatorv1 "github.com/openshift/api/operator/v1"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/opendatahub-io/opendatahub-operator/v2/api/common"
	dsciv2 "github.com/opendatahub-io/opendatahub-operator/v2/api/dscinitialization/v2"
	cr "github.com/opendatahub-io/opendatahub-operator/v2/internal/controller/components/registry"
)

const (
	ServiceName = "setupcontroller"
)

type ServiceHandler struct {
}

func (h *ServiceHandler) Init(_ common.Platform) error {
	return nil
}

func (h *ServiceHandler) GetName() string {
	return ServiceName
}

func (h *ServiceHandler) GetManagementState(_ common.Platform, _ *dsciv2.DSCInitialization) operatorv1.ManagementState {
	return operatorv1.Managed
}

func (h *ServiceHandler) NewReconciler(_ context.Context, mgr ctrl.Manager, _ *cr.Registry) error {
	rec := &SetupControllerReconciler{
		Client: mgr.GetClient(),
	}

	if err := rec.SetupWithManager(mgr); err != nil {
		return fmt.Errorf("could not create the %s controller: %w", ServiceName, err)
	}

	return nil
}
