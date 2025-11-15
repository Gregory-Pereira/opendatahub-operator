# Minikube Setup Guide

This guide provides step-by-step instructions for setting up a Minikube cluster with cert-manager, Prometheus, Istio, and Kubernetes Gateway API.

## Prerequisites

Before starting, ensure you have the following installed:

- **Minikube** - Local Kubernetes cluster
- **kubectl**
- **Helm**
- **istioctl 1.27.x** - Download from https://istio.io/latest/docs/setup/getting-started/
- **System Resources** - At least 16GB RAM available on your machine

## Installation Steps

### 1. Start Minikube Cluster

Create a Minikube cluster with sufficient resources:

```bash
minikube start --driver=docker --cpus=4 --memory=16g --disk-size=30g
```

**Resource Notes:**
- Minimum: 8GB RAM, 4 CPUs
- Recommended: 16GB RAM, 4 CPUs for stability with all components

### 2. Install Kubernetes Gateway API CRDs

**CRITICAL: Install Gateway API CRDs before Istio**

Istio's control plane needs the Gateway API CRDs to exist when it starts up.

```bash
# Install standard channel (recommended for production)
kubectl kustomize "github.com/kubernetes-sigs/gateway-api/config/crd?ref=v1.4.0" | kubectl apply -f -

# Verify installation
kubectl get crd | grep gateway
```

Expected output should include:
- `gateways.gateway.networking.k8s.io`
- `httproutes.gateway.networking.k8s.io`
- `grpcroutes.gateway.networking.k8s.io`
- And other Gateway API CRDs

### 3. Install cert-manager

cert-manager manages TLS certificates and can integrate with Istio for both gateway and mesh certificates.

```bash
# Install cert-manager with CRDs using OCI registry
helm install cert-manager \
  oci://quay.io/jetstack/charts/cert-manager \
  --namespace cert-manager \
  --create-namespace \
  --set crds.enabled=true

# Verify installation
kubectl get pods -n cert-manager
```

All pods should be in `Running` state.

### 4. Install Prometheus (kube-prometheus-stack)

```bash
# Install kube-prometheus-stack (minimal installation) using OCI registry
helm install prometheus \
  oci://ghcr.io/prometheus-community/charts/kube-prometheus-stack \
  --namespace monitoring \
  --create-namespace \
  --set kubeStateMetrics.enabled=false \
  --set nodeExporter.enabled=false \
  --set grafana.enabled=false

# Verify installation
kubectl get pods -n monitoring
```

### 5. Install Istio with Minimal Profile

When using Gateway API, use the minimal profile since Gateway API auto-provisions gateway deployments.

```bash
# Install Istio control plane
istioctl install -y --set profile=minimal --set values.pilot.env.ENABLE_GATEWAY_API_INFERENCE_EXTENSION=true

# Verify installation
istioctl verify-install
kubectl get pods -n istio-system
```

**Why minimal profile?** Gateway API resources automatically provision gateway deployments, so you don't need the default `istio-ingressgateway`.

### 6. Start Minikube Tunnel

**Required for LoadBalancer Support**

Run this in a separate terminal and keep it running:

```bash
minikube tunnel
```

This enables LoadBalancer services to get external IPs in Minikube. Without this, gateway services will remain in `<pending>` state.

**Note:** This requires administrative privileges on your machine.

### 7. Install OpenDataHub Operator

Install the OpenDataHub operator using the Helm chart from OCI registry:

```bash
helm install opendatahub-operator \
  oci://quay.io/lburgazzoli/opendatahub-operator-chart:0.1.0 \
  --namespace opendatahub-system \
  --create-namespace

# Verify installation
kubectl get pods -n opendatahub-system
```

Wait for the operator pod to be in `Running` state before proceeding.

### 8. Deploy KServe Component

Create the KServe component:

```bash
kubectl apply -f - <<EOF
apiVersion: components.platform.opendatahub.io/v1alpha1
kind: Kserve
metadata:
  name: default-kserve
spec:
  nim:
    managementState: Removed
EOF

# Verify KServe resource
kubectl get kserve
```

### 9. Deploy CPU-based LLM Example

Deploy a sample LLM inference service running on CPU:

```bash
kubectl apply -f - <<EOF
apiVersion: serving.kserve.io/v1alpha1
kind: LLMInferenceService
metadata:
  name: facebook-opt-125m-single
  annotations:
    security.opendatahub.io/enable-auth: "false"
spec:
  model:
    uri: hf://facebook/opt-125m
    name: facebook/opt-125m
  replicas: 1
  router:
    scheduler: { }
    route: { }
    gateway: {}
  template:
    containers:
      - name: main
        image: quay.io/pierdipi/vllm-cpu:latest
        securityContext:
          runAsNonRoot: false
        env:
          - name: VLLM_LOGGING_LEVEL
            value: DEBUG
        resources:
          limits:
            cpu: '1'
            memory: 10Gi
          requests:
            cpu: '100m'
            memory: 8Gi
        livenessProbe:
          initialDelaySeconds: 30
          periodSeconds: 30
          timeoutSeconds: 30
          failureThreshold: 5
EOF

# Verify deployment
kubectl get llminferenceservice
kubectl get pods
```

This deploys the Facebook OPT-125M model using vLLM on CPU.

**Test the LLM inference service:**

Wait for the service to be ready, then test it with a completion request:

```bash
# Extract the service URL from the LLMInferenceService status
URL=$(kubectl get llminferenceservice facebook-opt-125m-single -o jsonpath='{.status.addresses[0].url}')
echo "Service URL: $URL"

# Test the service with a completion request
curl -v ${URL}/v1/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "facebook/opt-125m",
    "prompt": "San Francisco is a"
  }'
```

## Installation Order Summary

The correct installation order is critical:

1. **Minikube cluster** - Start cluster with sufficient resources
2. **Gateway API CRDs** - Must be installed before Istio
3. **cert-manager** - No dependencies
4. **Prometheus** - No dependencies
5. **Istio** - Requires Gateway API CRDs
6. **Minikube tunnel** - Start after Istio for LoadBalancer support
7. **OpenDataHub Operator** - Install via Helm chart
8. **KServe Component** - Deploy after operator is running
9. **LLM Example** - Deploy sample inference service

## Additional Resources

- [Kubernetes Gateway API Documentation](https://gateway-api.sigs.k8s.io/)
- [Istio Gateway API Documentation](https://istio.io/latest/docs/tasks/traffic-management/ingress/gateway-api/)
- [cert-manager Documentation](https://cert-manager.io/docs/)
- [kube-prometheus-stack Chart](https://github.com/prometheus-community/helm-charts/tree/main/charts/kube-prometheus-stack)
- [Minikube Documentation](https://minikube.sigs.k8s.io/docs/)
