# Vanilla Kubernetes Overlay

This overlay configures the OpenDataHub operator for deployment on vanilla Kubernetes (non-OpenShift) clusters using cert-manager for webhook certificate management.

## Prerequisites

- Kubernetes cluster (vanilla, not OpenShift)
- [cert-manager](https://cert-manager.io/) installed on the cluster
- `kubectl` with kustomize support

## Installation

### 1. Install cert-manager

If cert-manager is not already installed:

```bash
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.14.0/cert-manager.yaml
```

Verify cert-manager is running:

```bash
kubectl get pods -n cert-manager
```

### 2. Deploy the operator

Build and apply the manifests:

```bash
kubectl apply -k config/overlays/vanilla --server-side --load-restrictor LoadRestrictionsNone
```

Or generate the manifests first for review:

```bash
kubectl kustomize config/overlays/vanilla --load-restrictor LoadRestrictionsNone > operator-manifests.yaml
kubectl apply -f operator-manifests.yaml --server-side
```

### 3. Verify the deployment

Check that the operator is running:

```bash
kubectl get pods -n opendatahub-operator-system
```

Verify the certificate was issued:

```bash
kubectl get certificate -n opendatahub-operator-system
kubectl get secret opendatahub-operator-controller-webhook-cert -n opendatahub-operator-system
```

## Key Differences from Default Configuration

This overlay makes the following changes compared to the default OpenShift configuration:

### Certificate Management
- **Uses cert-manager** instead of OpenShift service-ca-operator
- Creates a self-signed `Issuer` and `Certificate` resource
- Certificate is stored in secret `opendatahub-operator-controller-webhook-cert`

### Webhook Configuration
- **Removes OpenShift annotations**:
  - `service.beta.openshift.io/serving-cert-secret-name` (removed from Service)
  - `service.beta.openshift.io/inject-cabundle` (removed from webhook configurations)
- **Adds cert-manager CA injection**:
  - `cert-manager.io/inject-ca-from` annotation on CRDs for automatic CA bundle injection
  - `cert-manager.io/inject-ca-from` annotation on MutatingWebhookConfiguration for automatic CA bundle injection
  - `cert-manager.io/inject-ca-from` annotation on ValidatingWebhookConfiguration for automatic CA bundle injection

### CRD Conversion Webhooks
- **Removed conversion webhooks** for DataScienceCluster and DSCInitialization CRDs
- Both v1 and v2 API versions remain available
- Users should use v2 APIs directly as there is no automatic conversion

## Configuration Options

### Namespace
The default namespace is `opendatahub-operator-system`. To customize:

Edit `config/overlays/vanilla/kustomization.yaml`:
```yaml
namespace: your-custom-namespace
```

### Manager Image
The overlay is configured to use a custom operator image:
- Image: `quay.io/lburgazzoli/opendatahub-operator:vanilla`

To change the image, edit `config/overlays/vanilla/kustomization.yaml`:
```yaml
images:
- name: controller
  newName: your-registry/your-operator
  newTag: your-tag
```

### Replica Count
The manager deployment is configured with 1 replica for vanilla Kubernetes deployments.

To change the replica count, edit `config/overlays/vanilla/kustomization.yaml`:
```yaml
replicas:
- name: controller-manager
  count: 3  # or your desired count
```

### ConfigMap Name Suffix
ConfigMap name suffix hashing is disabled to ensure consistent naming:
```yaml
generatorOptions:
  disableNameSuffixHash: true
```

### Certificate Issuer
The default uses a self-signed issuer. For production, consider using a proper CA:

Edit `config/overlays/vanilla/certmanager/certificate.yaml`:
```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: selfsigned-issuer
  namespace: system
spec:
  ca:  # Use CA issuer instead
    secretName: your-ca-secret
```

## Troubleshooting

### Certificate not ready
Check cert-manager logs:
```bash
kubectl logs -n cert-manager deploy/cert-manager
```

Check certificate status:
```bash
kubectl describe certificate opendatahub-operator-serving-cert -n opendatahub-operator-system
```

### Webhook failures
Verify the webhook service is accessible:
```bash
kubectl get svc opendatahub-operator-webhook-service -n opendatahub-operator-system
```

Check webhook configurations:
```bash
kubectl get validatingwebhookconfigurations | grep opendatahub
kubectl get mutatingwebhookconfigurations | grep opendatahub
```

### CA bundle not injected
Verify cert-manager CA injector is running:
```bash
kubectl get pods -n cert-manager -l app.kubernetes.io/component=cainjector
```

Check CRD annotations:
```bash
kubectl get crd datascienceclusters.datasciencecluster.opendatahub.io -o yaml | grep cert-manager.io/inject-ca-from
```

## Directory Structure

```
config/overlays/vanilla/
├── README.md                           # This file
├── kustomization.yaml                  # Main overlay configuration
├── manager_auth_proxy_patch.yaml      # Manager auth proxy configuration
├── manager_webhook_patch.yaml         # Manager webhook volume mounts
├── certmanager/
│   ├── kustomization.yaml
│   ├── certificate.yaml                # Self-signed Issuer + Certificate
│   └── kustomizeconfig.yaml            # Variable substitution config
├── crd/
│   └── patches/
│       ├── cainjection_datasciencecluster.yaml    # cert-manager CA injection for CRD
│       └── cainjection_dscinitialization.yaml     # cert-manager CA injection for CRD
└── webhook/
    ├── kustomization.yaml
    ├── manifests.yaml                  # Webhook configurations (copied)
    ├── service.yaml                    # Webhook service (copied)
    ├── service_patch.yaml              # Remove OpenShift annotations
    ├── cainjection_mutatingwebhook.yaml   # cert-manager CA injection for mutating webhooks
    └── cainjection_validatingwebhook.yaml # cert-manager CA injection for validating webhooks
```

## References

- [cert-manager documentation](https://cert-manager.io/docs/)
- [Kubebuilder webhooks](https://book.kubebuilder.io/cronjob-tutorial/webhook-implementation.html)
- [OpenDataHub documentation](https://opendatahub.io/)
