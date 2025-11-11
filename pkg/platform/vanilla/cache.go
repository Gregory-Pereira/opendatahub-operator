package vanilla

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/cluster"
)

// getCommonCacheNamespaces returns the base set of namespaces to cache for Vanilla Kubernetes.
func getCommonCacheNamespaces() (map[string]cache.Config, error) {
	namespaceConfigs := map[string]cache.Config{}

	operatorNs, err := cluster.GetOperatorNamespace()
	if err != nil {
		return nil, err
	}
	namespaceConfigs[operatorNs] = cache.Config{}

	appNamespace := cluster.GetApplicationNamespace()
	namespaceConfigs[appNamespace] = cache.Config{}

	return namespaceConfigs, nil
}

// CreateCacheOptions creates platform-specific cache configuration for Vanilla Kubernetes.
// This is a pure function that accepts the scheme as a parameter.
func CreateCacheOptions(scheme *runtime.Scheme) (cache.Options, error) {
	c, err := getCommonCacheNamespaces()
	if err != nil {
		return cache.Options{}, fmt.Errorf("unable to create cache config: %w", err)
	}

	opt := cache.Options{
		Scheme:           scheme,
		DefaultTransform: cache.TransformStripManagedFields(),
		ByObject: map[client.Object]cache.ByObject{
			&corev1.Secret{}: {
				Namespaces: c,
			},
			&corev1.ConfigMap{}: {
				Namespaces: c,
			},
		},
	}

	return opt, nil
}
