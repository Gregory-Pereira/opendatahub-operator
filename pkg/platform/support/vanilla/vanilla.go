package vanilla

import (
	"context"
	"fmt"
	"os"

	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/opendatahub-io/opendatahub-operator/v2/api/common"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/cluster"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/platform"
)

const (
	distributionVanilla = "Vanilla"
	defaultVersion      = "0.0.0"
)

// DiscoverMeta discovers and returns cluster metadata for vanilla Kubernetes platforms.
// This includes operator version and Kubernetes version from the discovery API.
// It also sets the operator namespace in the global cluster config.
func DiscoverMeta(
	ctx context.Context,
	cli client.Client,
	restConfig *rest.Config,
	platformType common.Platform,
) (platform.Meta, error) {
	logger := ctrl.LoggerFrom(ctx)

	// Set operator namespace
	if err := cluster.SetOperatorNamespace(); err != nil {
		return platform.Meta{}, fmt.Errorf("failed to set operator namespace: %w", err)
	}

	_, err := cluster.GetOperatorNamespace()
	if err != nil {
		return platform.Meta{}, fmt.Errorf("failed to get operator namespace: %w", err)
	}

	// Discover operator version from environment variable
	operatorVersion := discoverOperatorVersion()

	// Discover Kubernetes version using discovery client
	distributionVersion, err := discoverKubernetesVersion(restConfig)
	if err != nil {
		logger.Info("unable to discover Kubernetes version", "error", err)
		distributionVersion = ""
	}

	// Return populated metadata
	return platform.Meta{
		Type:                platformType,
		Version:             operatorVersion,
		DistributionVersion: distributionVersion,
		Distribution:        distributionVanilla,
		FIPSEnabled:         false, // Not applicable for vanilla K8s
	}, nil
}

// discoverKubernetesVersion discovers the Kubernetes version using the discovery client.
func discoverKubernetesVersion(restConfig *rest.Config) (string, error) {
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(restConfig)
	if err != nil {
		return "", fmt.Errorf("failed to create discovery client: %w", err)
	}

	serverVersion, err := discoveryClient.ServerVersion()
	if err != nil {
		return "", fmt.Errorf("failed to get server version: %w", err)
	}

	return serverVersion.GitVersion, nil
}

// discoverOperatorVersion discovers the operator version from ODH_PLATFORM_VERSION environment variable.
func discoverOperatorVersion() string {
	version := os.Getenv("ODH_PLATFORM_VERSION")
	if version == "" {
		version = defaultVersion
	}

	return version
}
