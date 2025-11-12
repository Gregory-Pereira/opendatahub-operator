package selfmanaged

import (
	"fmt"

	configv1 "github.com/openshift/api/config/v1"
	operatorv1 "github.com/openshift/api/operator/v1"
	routev1 "github.com/openshift/api/route/v1"
	promv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/cluster"
)

// getCommonCacheNamespaces returns the base set of namespaces to cache for SelfManaged RHOAI.
func getCommonCacheNamespaces() (map[string]cache.Config, error) {
	namespaceConfigs := map[string]cache.Config{}

	operatorNs, err := cluster.GetOperatorNamespace()
	if err != nil {
		return nil, err
	}
	namespaceConfigs[operatorNs] = cache.Config{}
	namespaceConfigs[cluster.DefaultMonitoringNamespaceRHOAI] = cache.Config{}

	appNamespace := cluster.GetApplicationNamespace()
	namespaceConfigs[appNamespace] = cache.Config{}

	return namespaceConfigs, nil
}

// createSecretCacheNamespaces returns namespaces where secrets should be cached.
func createSecretCacheNamespaces() (map[string]cache.Config, error) {
	namespaceConfigs, err := getCommonCacheNamespaces()
	if err != nil {
		return nil, err
	}
	namespaceConfigs[cluster.OpenshiftIngressNamespace] = cache.Config{}
	return namespaceConfigs, nil
}

// createGeneralCacheNamespaces returns namespaces where general resources should be cached.
func createGeneralCacheNamespaces() (map[string]cache.Config, error) {
	namespaceConfigs, err := getCommonCacheNamespaces()
	if err != nil {
		return nil, err
	}
	namespaceConfigs[cluster.OpenshiftOperatorsNamespace] = cache.Config{}
	namespaceConfigs[cluster.OpenshiftIngressNamespace] = cache.Config{}
	return namespaceConfigs, nil
}

// CreateCacheOptions creates platform-specific cache configuration for SelfManaged RHOAI.
// This is a pure function that accepts the scheme as a parameter.
func CreateCacheOptions(scheme *runtime.Scheme) (cache.Options, error) {
	secretCache, err := createSecretCacheNamespaces()
	if err != nil {
		return cache.Options{}, fmt.Errorf("unable to create secret cache config: %w", err)
	}

	generalCache, err := createGeneralCacheNamespaces()
	if err != nil {
		return cache.Options{}, fmt.Errorf("unable to create general cache config: %w", err)
	}

	return cache.Options{
		Scheme: scheme,
		ByObject: map[client.Object]cache.ByObject{
			&corev1.Secret{}: {
				Namespaces: secretCache,
			},
			&corev1.ConfigMap{}: {
				Namespaces: generalCache,
			},
			&operatorv1.IngressController{}: {
				Field: fields.Set{"metadata.name": "default"}.AsSelector(),
			},
			&configv1.Authentication{}: {
				Field: fields.Set{"metadata.name": cluster.ClusterAuthenticationObj}.AsSelector(),
			},
			&appsv1.Deployment{}: {
				Namespaces: generalCache,
			},
			&promv1.PrometheusRule{}: {
				Namespaces: generalCache,
			},
			&promv1.ServiceMonitor{}: {
				Namespaces: generalCache,
			},
			&routev1.Route{}: {
				Namespaces: generalCache,
			},
			&networkingv1.NetworkPolicy{}: {
				Namespaces: generalCache,
			},
			&rbacv1.Role{}: {
				Namespaces: generalCache,
			},
			&rbacv1.RoleBinding{}: {
				Namespaces: generalCache,
			},
		},
		DefaultTransform: cache.TransformStripManagedFields(),
	}, nil
}
