package registry

import (
	"context"

	"github.com/hashicorp/go-multierror"
	operatorv1 "github.com/openshift/api/operator/v1"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/opendatahub-io/opendatahub-operator/v2/api/common"
	dsciv2 "github.com/opendatahub-io/opendatahub-operator/v2/api/dscinitialization/v2"
	cr "github.com/opendatahub-io/opendatahub-operator/v2/internal/controller/components/registry"
)

// ServiceHandler is an interface to manage a service
// Every method should accept ctx since it contains the logger.
type ServiceHandler interface {
	Init(platform common.Platform) error
	GetName() string
	GetManagementState(platform common.Platform, dsci *dsciv2.DSCInitialization) operatorv1.ManagementState
	NewReconciler(ctx context.Context, mgr ctrl.Manager, componentRegistry *cr.Registry) error
}

// Registry is a struct that maintains a list of registered ServiceHandlers.
type Registry struct {
	handlers []ServiceHandler
}

// NewRegistry creates a new service registry instance.
// Accepts optional ServiceHandlers to register during initialization.
func NewRegistry(handlers ...ServiceHandler) *Registry {
	return &Registry{
		handlers: append([]ServiceHandler{}, handlers...),
	}
}

// Add registers a new ServiceHandler to the registry.
// not thread safe, supposed to be called during initialization.
func (r *Registry) Add(ch ServiceHandler) {
	r.handlers = append(r.handlers, ch)
}

// ForEach iterates over all registered ServiceHandlers and applies the given function.
// If any handler returns an error, that error is collected and returned at the end.
// With go1.23 probably https://go.dev/blog/range-functions can be used.
func (r *Registry) ForEach(f func(ch ServiceHandler) error) error {
	var errs *multierror.Error
	for _, ch := range r.handlers {
		errs = multierror.Append(errs, f(ch))
	}

	return errs.ErrorOrNil()
}
