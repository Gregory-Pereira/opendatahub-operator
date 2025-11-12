# LLM-D Component

## Overview

The llm-d component integrates LLM Distribution capabilities into OpenDataHub using a Helm-based approach.

## Structure

```
llmd/
├── llmd.go                       # Component handler (registry integration)
├── llmd_controller.go            # Reconciliation controller
├── llmd_controller_actions.go   # Custom reconciliation actions
├── llmd_support.go               # Helper functions and manifest paths
├── monitoring/
│   └── llmd-prometheusrules.tmpl.yaml  # Prometheus monitoring rules
└── README.md                     # This file
```

## Helm Charts Deployed

1. **llm-d-modelservice** (v0.2.11)
   - Core model serving functionality
   - Deployed from https://llm-d-incubation.github.io/llm-d-modelservice/

2. **llm-d-infra** (v1.3.3)
   - Infrastructure components
   - Deployed from https://llm-d-incubation.github.io/llm-d-infra/

3. **inferencepool** (v1.0.1)
   - Gateway API Inference Extension
   - Deployed from https://github.com/kubernetes-sigs/gateway-api-inference-extension

## Manifest Generation

Manifests are generated using `hack/fetch-helm-charts.sh` which:
1. Downloads the Helm charts
2. Templates them with default namespace `llm-d`
3. Creates Kustomize-compatible structure in `opt/manifests/llmd/`

## Component Lifecycle

The component follows the standard OpenDataHub component pattern:

1. **Registration**: Registered in `init()` with the component registry
2. **Initialization**: `Init()` is called during operator startup
3. **CR Creation**: `NewCRObject()` creates Llmd CR from DataScienceCluster
4. **Reconciliation**: Controller watches Llmd CR and deploys manifests
5. **Status Updates**: `UpdateDSCStatus()` syncs status back to DataScienceCluster

## Development

### Local Testing

```bash
# Generate manifests
make get-manifests

# Run controller locally
make run

# In another terminal, apply test CR
kubectl apply -f config/samples/components_v1alpha1_llmd.yaml
```

### Modifying Chart Versions

1. Edit `hack/fetch-helm-charts.sh` with new versions
2. Update default versions in `api/components/v1alpha1/llmd_types.go`
3. Run `make get-manifests`
4. Test the changes

## Configuration

### Default Configuration

```yaml
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

### Custom Values (Future Enhancement)

```yaml
spec:
  modelService:
    enabled: true
    version: "0.2.11"
    values:
      customKey: customValue
```

## Monitoring

Prometheus rules are automatically deployed with the component. See `monitoring/llmd-prometheusrules.tmpl.yaml`.

## See Also

- [LLM-D Integration Guide](../../../../docs/llmd-integration.md)
- [Component Integration Guide](../../../../docs/COMPONENT_INTEGRATION.md)
