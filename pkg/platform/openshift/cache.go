package openshift

import (
	"fmt"

	configv1 "github.com/openshift/api/config/v1"
	operatorv1 "github.com/openshift/api/operator/v1"
	routev1 "github.com/openshift/api/route/v1"
	userv1 "github.com/openshift/api/user/v1"
	ofapiv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	promv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	appsv1 "k8s.io/api/apps/v1"
	authorizationv1 "k8s.io/api/authorization/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/cluster"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/cluster/gvk"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/resources"
)

// getCommonCacheNamespaces returns the base set of namespaces to cache for OpenShift.
func getCommonCacheNamespaces(variant Variant) (map[string]cache.Config, error) {
	namespaceConfigs := map[string]cache.Config{}

	operatorNs, err := cluster.GetOperatorNamespace()
	if err != nil {
		return nil, err
	}
	namespaceConfigs[operatorNs] = cache.Config{}

	// Use variant-specific monitoring namespace
	namespaceConfigs[variant.MonitoringNamespace] = cache.Config{}

	appNamespace := cluster.GetApplicationNamespace()
	namespaceConfigs[appNamespace] = cache.Config{}

	// Add console link namespace if specified (Managed variant only)
	if variant.ConsoleNamespace != "" {
		namespaceConfigs[variant.ConsoleNamespace] = cache.Config{}
	}

	return namespaceConfigs, nil
}

// createSecretCacheNamespaces returns namespaces where secrets should be cached.
func createSecretCacheNamespaces(variant Variant) (map[string]cache.Config, error) {
	namespaceConfigs, err := getCommonCacheNamespaces(variant)
	if err != nil {
		return nil, err
	}
	namespaceConfigs[cluster.OpenshiftIngressNamespace] = cache.Config{}
	return namespaceConfigs, nil
}

// createGeneralCacheNamespaces returns namespaces where general resources should be cached.
func createGeneralCacheNamespaces(variant Variant) (map[string]cache.Config, error) {
	namespaceConfigs, err := getCommonCacheNamespaces(variant)
	if err != nil {
		return nil, err
	}
	namespaceConfigs[cluster.OpenshiftOperatorsNamespace] = cache.Config{}
	namespaceConfigs[cluster.OpenshiftIngressNamespace] = cache.Config{}
	return namespaceConfigs, nil
}

// CreateCacheOptions creates platform-specific cache configuration for OpenShift.
func CreateCacheOptions(scheme *runtime.Scheme, variant Variant) (cache.Options, error) {
	secretCache, err := createSecretCacheNamespaces(variant)
	if err != nil {
		return cache.Options{}, fmt.Errorf("unable to create secret cache config: %w", err)
	}

	generalCache, err := createGeneralCacheNamespaces(variant)
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

// CreateClientOptions creates platform-specific client configuration for OpenShift.
// All OpenShift variants use the same client cache exclusions.
func CreateClientOptions() client.Options {
	return client.Options{
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
	}
}
