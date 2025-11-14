# Minikube Setup Guide

This guide provides step-by-step instructions for setting up a Minikube cluster with cert-manager, Prometheus, Istio, and Kubernetes Gateway API.

## Prerequisites

Before starting, ensure you have the following installed:

- **Minikube** - Local Kubernetes cluster
- **kubectl** 
- **Helm**
- **istioctl** - Download from https://istio.io/latest/docs/setup/getting-started/
- **System Resources** - At least 16GB RAM available on your machine

## Installation Steps

### 1. Start Minikube Cluster

Create a Minikube cluster with sufficient resources:

```bash
minikube start --driver=kvm2 --cpus=4 --memory=16g --disk-size=30g
```

**Resource Notes:**
- Minimum: 8GB RAM, 4 CPUs
- Recommended: 16GB RAM, 4 CPUs for stability with all components

### 2. Enable Metrics Server (Optional but Recommended)

```bash
minikube addons enable metrics-server
```

This enables resource metrics for monitoring and HPA (Horizontal Pod Autoscaling).

### 3. Install Kubernetes Gateway API CRDs

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

**Alternative: Experimental Channel** (includes experimental features):
```bash
kubectl apply -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.4.0/experimental-install.yaml
```

### 4. Install cert-manager

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

### 5. Install Prometheus (kube-prometheus-stack)

The kube-prometheus-stack includes Prometheus, Grafana, AlertManager, and the Prometheus Operator.

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

This provides:
- Prometheus Operator
- Prometheus instance with pre-configured rules
- AlertManager
- ServiceMonitor and PodMonitor CRDs for service discovery

**Note:** This minimal installation disables kube-state-metrics, node-exporter, and Grafana to reduce resource usage. If you need these components, remove the corresponding `--set` flags.

### 6. Install Istio with Minimal Profile

When using Gateway API, use the minimal profile since Gateway API auto-provisions gateway deployments.

```bash
# Install Istio control plane
istioctl install -y --set profile=minimal --set values.pilot.env.ENABLE_GATEWAY_API_INFERENCE_EXTENSION=true

# Verify installation
istioctl verify-install
kubectl get pods -n istio-system
```

**Why minimal profile?** Gateway API resources automatically provision gateway deployments, so you don't need the default `istio-ingressgateway`.

### 7. Create Gateway Namespace with Sidecar Injection

```bash
kubectl create namespace istio-ingress
kubectl label namespace istio-ingress istio-injection=enabled
```

This namespace will host your Gateway resources and their auto-provisioned deployments.

### 8. Start Minikube Tunnel

**Required for LoadBalancer Support**

Run this in a separate terminal and keep it running:

```bash
minikube tunnel
```

This enables LoadBalancer services to get external IPs in Minikube. Without this, gateway services will remain in `<pending>` state.

**Note:** This requires administrative privileges on your machine.

## Verification Steps

Verify each component is installed correctly:

```bash
# Verify Gateway API CRDs
kubectl get crd | grep gateway.networking.k8s.io

# Verify cert-manager
kubectl get pods -n cert-manager
# Expected: 3 pods running (controller, webhook, cainjector)

# Verify Prometheus
kubectl get pods -n monitoring
# Expected: Multiple pods including prometheus, grafana, alertmanager

# Verify Istio
kubectl get pods -n istio-system
# Expected: istiod pod running

# Verify gateway namespace
kubectl get namespace istio-ingress -o yaml
# Should have label: istio-injection: enabled
```

## Installation Order Summary

The correct installation order is critical:

1. **Minikube cluster** - Start cluster with sufficient resources
2. **Metrics Server** - Optional but recommended
3. **Gateway API CRDs** - Must be installed before Istio
4. **cert-manager** - No dependencies
5. **Prometheus** - No dependencies
6. **Istio** - Requires Gateway API CRDs
7. **Minikube tunnel** - Start after Istio for LoadBalancer support

## Additional Resources

- [Kubernetes Gateway API Documentation](https://gateway-api.sigs.k8s.io/)
- [Istio Gateway API Documentation](https://istio.io/latest/docs/tasks/traffic-management/ingress/gateway-api/)
- [cert-manager Documentation](https://cert-manager.io/docs/)
- [kube-prometheus-stack Chart](https://github.com/prometheus-community/helm-charts/tree/main/charts/kube-prometheus-stack)
- [Minikube Documentation](https://minikube.sigs.k8s.io/docs/)

## Quick Start Script

Here's a complete setup script for reference:

```bash
#!/bin/bash
set -e

echo "Starting Minikube..."
minikube start --cpus=4 --memory=16g --disk-size=30g --driver=docker --kubernetes-version=v1.26.1

echo "Enabling metrics-server..."
minikube addons enable metrics-server

echo "Installing Gateway API CRDs..."
kubectl kustomize "github.com/kubernetes-sigs/gateway-api/config/crd?ref=v1.4.0" | kubectl apply -f -

echo "Installing cert-manager..."
helm install cert-manager \
  oci://quay.io/jetstack/charts/cert-manager \
  --version v1.19.1 \
  --namespace cert-manager \
  --create-namespace \
  --set crds.enabled=true

echo "Installing Prometheus..."
helm install prom \
  oci://ghcr.io/prometheus-community/charts/kube-prometheus-stack \
  --namespace monitoring \
  --create-namespace \
  --set kubeStateMetrics.enabled=false \
  --set nodeExporter.enabled=false \
  --set grafana.enabled=false

echo "Installing Istio..."
istioctl install --set profile=minimal -y

echo "Creating gateway namespace..."
kubectl create namespace istio-ingress
kubectl label namespace istio-ingress istio-injection=enabled

echo "Setup complete!"
echo "Run 'minikube tunnel' in a separate terminal to enable LoadBalancer support"
```

Save this as `setup-minikube.sh` and run with `bash setup-minikube.sh`.
