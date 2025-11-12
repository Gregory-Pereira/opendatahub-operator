# LLM-D Integration Guide

## Overview

This document describes the integration of llm-d (LLM Distribution) as a component in the OpenDataHub operator. This is a POC implementation using an operator SDK + Helm-based approach.

## Architecture

The llm-d integration consists of three main Helm charts:

1. **ModelService** (v0.2.11) - Core model serving functionality
   - Chart URL: https://llm-d-incubation.github.io/llm-d-modelservice/

2. **Infra** (v1.3.3) - Infrastructure components
   - Chart URL: https://llm-d-incubation.github.io/llm-d-infra/

3. **Gateway API Inference Extension** (v1.0.1) - Gateway API integration
   - Chart URL: https://github.com/kubernetes-sigs/gateway-api-inference-extension

## Components

### API Types

The llm-d component is defined in `api/components/v1alpha1/llmd_types.go` with the following structure:

- **Llmd**: The main CRD for llm-d component
- **LlmdSpec**: Configuration for all three Helm charts
- **LlmdStatus**: Status information for the component

### Controller

The controller is implemented in `internal/controller/components/llmd/` with:

- **llmd.go**: Component handler implementing the registry interface
- **llmd_controller.go**: Reconciliation logic
- **llmd_controller_actions.go**: Custom actions during reconciliation
- **llmd_support.go**: Helper functions

### Manifest Management

Helm charts are fetched and templated using `hack/fetch-helm-charts.sh`, which:

1. Downloads the specified Helm chart versions
2. Templates them with default values
3. Creates Kustomize-compatible manifests in `opt/manifests/llmd/`

## Installation

### Prerequisites

- OpenShift 4.19 or higher
- Helm CLI installed (for manifest generation)
- OpenDataHub operator installed

### Deploying llm-d Component

#### Method 1: Using DataScienceCluster

Create a DataScienceCluster CR with llm-d enabled:

```yaml
apiVersion: datasciencecluster.opendatahub.io/v2
kind: DataScienceCluster
metadata:
  name: default-dsc
spec:
  components:
    llmd:
      managementState: Managed
      modelService:
        enabled: true
        version: "0.2.11"
      infra:
        enabled: true
        version: "1.3.3"
      gatewayAPI:
        enabled: true
        version: "1.0.1"
```

#### Method 2: Using Llmd Component Directly

Create an Llmd CR:

```yaml
apiVersion: components.platform.opendatahub.io/v1alpha1
kind: Llmd
metadata:
  name: default-llm-d
spec:
  modelService:
    enabled: true
    version: "0.2.11"
  infra:
    enabled: true
    version: "1.3.3"
  gatewayAPI:
    enabled: true
    version: "1.0.1"
```

### Building the Operator with llm-d Support

1. Fetch manifests (including llm-d Helm charts):
   ```bash
   make get-manifests
   ```

2. Generate CRDs:
   ```bash
   make manifests
   ```

3. Build the operator image:
   ```bash
   make image IMG=quay.io/<username>/opendatahub-operator:<tag>
   ```

4. Deploy the operator:
   ```bash
   make deploy IMG=quay.io/<username>/opendatahub-operator:<tag>
   ```

## Configuration

### Customizing Helm Values

You can override default Helm chart values using the `values` field in each component:

```yaml
spec:
  modelService:
    enabled: true
    version: "0.2.11"
    values:
      replicas: "3"
      resourceLimits: "high"
```

### Disabling Components

Individual components can be disabled:

```yaml
spec:
  modelService:
    enabled: true
  infra:
    enabled: false  # Disable infra component
  gatewayAPI:
    enabled: true
```

## Monitoring

The llm-d component includes Prometheus monitoring rules defined in:
`internal/controller/components/llmd/monitoring/llmd-prometheusrules.tmpl.yaml`

These rules monitor:
- ModelService availability
- Infra component availability
- Overall component health

## Troubleshooting

### Check Component Status

```bash
oc get llmd default-llm-d -o yaml
```

Look for the status section to see the current state.

### Check DataScienceCluster Status

```bash
oc get datasciencecluster default-dsc -o jsonpath='{.status.components.llmd}'
```

### View Deployed Resources

```bash
oc get all -n llm-d -l app.kubernetes.io/part-of=llm-d
```

### Common Issues

1. **Helm charts not found**
   - Ensure `hack/fetch-helm-charts.sh` has been executed
   - Check that Helm CLI is installed

2. **CRD validation errors**
   - Run `make manifests` to regenerate CRDs
   - Ensure API types are properly defined

3. **Deployment failures**
   - Check operator logs: `oc logs -n opendatahub-operator-system deployment/opendatahub-operator-controller-manager`
   - Verify namespace exists: `oc get ns llm-d`

## Development

### Adding New Helm Chart Versions

1. Update versions in `hack/fetch-helm-charts.sh`
2. Update default versions in `api/components/v1alpha1/llmd_types.go`
3. Run `make get-manifests` to fetch new versions

### Modifying Component Behavior

Component reconciliation logic can be modified in:
- `internal/controller/components/llmd/llmd_controller.go`
- `internal/controller/components/llmd/llmd_controller_actions.go`

### Testing Changes

```bash
# Run unit tests
make unit-test

# Run e2e tests
make e2e-test
```

## Limitations

This is a POC implementation with the following limitations:

1. **Helm Chart Management**: Charts are templated at build time, not dynamically
2. **Version Upgrades**: Changing chart versions requires rebuilding the operator
3. **Value Overrides**: Limited support for custom Helm values
4. **Namespace**: Currently deploys to a fixed `llm-d` namespace

## Future Enhancements

Potential improvements for production readiness:

1. **Dynamic Helm Integration**: Use Helm SDK to deploy charts dynamically
2. **Advanced Configuration**: Support full Helm values override
3. **Multi-Namespace Support**: Allow deployment to custom namespaces
4. **Upgrade Strategy**: Implement proper Helm chart upgrade logic
5. **Rollback Support**: Add ability to rollback failed deployments

## References

- [LLM-D ModelService Chart](https://llm-d-incubation.github.io/llm-d-modelservice/)
- [LLM-D Infra Chart](https://llm-d-incubation.github.io/llm-d-infra/)
- [Gateway API Inference Extension](https://github.com/kubernetes-sigs/gateway-api-inference-extension)
- [OpenDataHub Operator Documentation](../README.md)
