# OpenDataHub Operator Helm Chart

This Helm chart deploys the OpenDataHub Operator on vanilla Kubernetes clusters.

## Prerequisites

- Kubernetes 1.28+
- Helm 3.13+
- cert-manager 1.13+ (installed separately or via dependencies)

## Installing the Chart

### Install cert-manager (if not already installed)

```bash
helm repo add jetstack https://charts.jetstack.io
helm repo update
helm install cert-manager jetstack/cert-manager \
  --namespace cert-manager \
  --create-namespace \
  --set crds.enabled=true
```

### Install the OpenDataHub Operator

```bash
helm install opendatahub-operator ./config/charts/opendatahub-operator \
  --namespace opendatahub-system \
  --create-namespace
```

## Configuration

The following table lists the configurable parameters of the chart and their default values.

| Parameter | Description | Default |
|-----------|-------------|---------|
| `components.namespace` | Application namespace for ODH components | `opendatahub` |
| `controller.replicaCount` | Number of controller replicas | `1` |
| `controller.resources.limits.cpu` | CPU limit | `500m` |
| `controller.resources.limits.memory` | Memory limit | `4Gi` |
| `controller.resources.requests.cpu` | CPU request | `100m` |
| `controller.resources.requests.memory` | Memory request | `780Mi` |

**Note**: The operator namespace is controlled by the `--namespace` flag in `helm install`, not by a chart value.

## Examples

### Install with custom namespaces

```bash
helm install opendatahub-operator ./config/charts/opendatahub-operator \
  --namespace my-operator-ns \
  --create-namespace \
  --set components.namespace=my-apps-ns
```

### Install with custom resources

```bash
helm install opendatahub-operator ./config/charts/opendatahub-operator \
  --set controller.resources.limits.cpu=1000m \
  --set controller.resources.limits.memory=8Gi
```

## Uninstalling the Chart

```bash
helm uninstall opendatahub-operator --namespace opendatahub-system
```

## Development

### Package the chart

```bash
make helm-package
```

### Push to OCI registry

```bash
make helm-push HELM_OCI_IMG=quay.io/my-org/opendatahub-operator-chart:v1.0.0
```

### Install from local chart

```bash
make helm-install
```

### Uninstall

```bash
make helm-uninstall
```