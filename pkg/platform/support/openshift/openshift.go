package openshift

import (
	"context"
	"errors"
	"fmt"
	"os"

	configv1 "github.com/openshift/api/config/v1"
	ofapiv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	"github.com/opendatahub-io/opendatahub-operator/v2/api/common"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/cluster"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/cluster/gvk"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/platform"
)

const (
	openshiftVersionObj    = "version"
	distributionOpenShift  = "OpenShift"
	clusterConfigConfigMap = "cluster-config-v1"
	kubeSystemNamespace    = "kube-system"
	defaultVersion         = "0.0.0"
)

// DiscoverMeta discovers and returns cluster metadata for OpenShift-based platforms.
// This includes operator version, distribution version, and FIPS status.
// It also sets the operator namespace in the global cluster config.
func DiscoverMeta(
	ctx context.Context,
	cli client.Client,
	platformType common.Platform,
) (platform.Meta, error) {
	logger := ctrl.LoggerFrom(ctx)

	// Set operator namespace
	if err := cluster.SetOperatorNamespace(); err != nil {
		return platform.Meta{}, fmt.Errorf("failed to set operator namespace: %w", err)
	}

	operatorNs, err := cluster.GetOperatorNamespace()
	if err != nil {
		return platform.Meta{}, fmt.Errorf("failed to get operator namespace: %w", err)
	}

	// Discover operator version from CSV
	operatorVersion, err := discoverOperatorVersion(ctx, cli, operatorNs)
	if err != nil {
		logger.Info("unable to discover operator version, using default", "error", err)
		operatorVersion = defaultVersion
	}

	// Discover OpenShift cluster version
	distributionVersion, err := discoverClusterVersion(ctx, cli)
	if err != nil {
		logger.Info("unable to discover cluster version", "error", err)
		distributionVersion = ""
	}

	// Discover FIPS status
	fipsEnabled, err := discoverFIPSEnabled(ctx, cli)
	if err != nil {
		logger.Info("unable to determine FIPS status, defaulting to false", "error", err)
		fipsEnabled = false
	}

	// Return populated metadata
	return platform.Meta{
		Type:                platformType,
		Version:             operatorVersion,
		DistributionVersion: distributionVersion,
		Distribution:        distributionOpenShift,
		FIPSEnabled:         fipsEnabled,
	}, nil
}

// discoverClusterVersion discovers the OpenShift cluster version.
func discoverClusterVersion(ctx context.Context, cli client.Client) (string, error) {
	clusterVersion := &configv1.ClusterVersion{}

	err := cli.Get(ctx, client.ObjectKey{Name: openshiftVersionObj}, clusterVersion)
	switch {
	case k8serr.IsNotFound(err), meta.IsNoMatchError(err):
		// Not OpenShift, return empty version
		return "", nil
	case err != nil:
		return "", fmt.Errorf("unable to get cluster version: %w", err)
	}

	if len(clusterVersion.Status.History) == 0 {
		return "", errors.New("cluster version history is empty")
	}

	return clusterVersion.Status.History[0].Version, nil
}

// discoverFIPSEnabled determines if FIPS is enabled on the OpenShift cluster.
func discoverFIPSEnabled(ctx context.Context, cli client.Client) (bool, error) {
	cm := &corev1.ConfigMap{}
	namespacedName := types.NamespacedName{
		Name:      clusterConfigConfigMap,
		Namespace: kubeSystemNamespace,
	}

	if err := cli.Get(ctx, namespacedName, cm); err != nil {
		return false, client.IgnoreNotFound(err)
	}

	installConfigStr := cm.Data["install-config"]
	if installConfigStr == "" {
		return false, nil
	}

	installConfig := &cluster.InstallConfig{}
	if err := yaml.Unmarshal([]byte(installConfigStr), installConfig); err != nil {
		return false, fmt.Errorf("unable to parse install-config: %w", err)
	}

	return installConfig.FIPS, nil
}

// discoverOperatorVersion discovers the operator version from CSV.
func discoverOperatorVersion(ctx context.Context, cli client.Client, operatorNamespace string) (string, error) {
	// For unit-tests
	if os.Getenv("CI") == "true" {
		return defaultVersion, nil
	}

	csv, err := getClusterServiceVersion(ctx, cli, operatorNamespace)
	switch {
	case k8serr.IsNotFound(err), meta.IsNoMatchError(err):
		return defaultVersion, nil
	case err != nil:
		return "", err
	default:
		return csv.Spec.Version.String(), nil
	}
}

// getClusterServiceVersion retrieves CSV only from the defined namespace.
func getClusterServiceVersion(ctx context.Context, c client.Client, namespace string) (*ofapiv1alpha1.ClusterServiceVersion, error) {
	clusterServiceVersionList := &ofapiv1alpha1.ClusterServiceVersionList{}
	paginateListOption := &client.ListOptions{
		Limit:     100,
		Namespace: namespace,
	}
	for {
		if err := c.List(ctx, clusterServiceVersionList, paginateListOption); err != nil {
			return nil, fmt.Errorf("failed listing cluster service versions for %s: %w", namespace, err)
		}
		for _, csv := range clusterServiceVersionList.Items {
			for _, operatorCR := range csv.Spec.CustomResourceDefinitions.Owned {
				if operatorCR.Kind == "DataScienceCluster" {
					return &csv, nil
				}
			}
		}
		if paginateListOption.Continue = clusterServiceVersionList.GetContinue(); paginateListOption.Continue == "" {
			break
		}
	}

	return nil, k8serr.NewNotFound(schema.GroupResource{Group: gvk.ClusterServiceVersion.Group}, gvk.ClusterServiceVersion.Kind)
}
