#!/usr/bin/env bash
set -e

# This script fetches and templates Helm charts for llm-d integration
# It creates a kustomize-compatible structure in opt/manifests/llmd

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
MANIFESTS_DIR="${REPO_ROOT}/opt/manifests/llmd"

# Helm chart configurations
MODELSERVICE_CHART_URL="https://llm-d-incubation.github.io/llm-d-modelservice/"
MODELSERVICE_VERSION="0.2.11"
MODELSERVICE_NAME="llm-d-modelservice"

INFRA_CHART_URL="https://llm-d-incubation.github.io/llm-d-infra/"
INFRA_VERSION="1.3.3"
INFRA_NAME="llm-d-infra"

GATEWAY_API_CHART_URL="https://github.com/kubernetes-sigs/gateway-api-inference-extension/releases/download/v1.0.1"
GATEWAY_API_VERSION="1.0.1"
GATEWAY_API_NAME="inferencepool"

# Check if helm is installed
if ! command -v helm &> /dev/null; then
    echo "ERROR: helm is not installed. Please install helm to proceed."
    exit 1
fi

# Create manifests directory structure
echo "Creating manifests directory structure..."
mkdir -p "${MANIFESTS_DIR}"/{modelservice,infra,gateway-api,overlays/default}

# Function to template and save helm chart
template_helm_chart() {
    local chart_url=$1
    local chart_name=$2
    local version=$3
    local output_dir=$4
    local release_name=$5

    echo "Templating ${chart_name} version ${version}..."

    # Add helm repo if it's a repo URL (not a direct chart URL)
    if [[ $chart_url == https://*.github.io/* ]]; then
        local repo_name="llmd-${chart_name}"
        helm repo add "${repo_name}" "${chart_url}" 2>/dev/null || true
        helm repo update "${repo_name}"
        helm template "${release_name}" "${repo_name}/${chart_name}" \
            --version "${version}" \
            --namespace llm-d \
            > "${output_dir}/resources.yaml"
    else
        # For GitHub releases or direct chart URLs
        local temp_dir=$(mktemp -d)
        cd "${temp_dir}"
        wget -q "${chart_url}/${chart_name}-${version}.tgz" -O chart.tgz || {
            echo "ERROR: Failed to download chart from ${chart_url}"
            cd - > /dev/null
            rm -rf "${temp_dir}"
            return 1
        }
        tar -xzf chart.tgz
        helm template "${release_name}" "./${chart_name}" \
            --namespace llm-d \
            > "${output_dir}/resources.yaml"
        cd - > /dev/null
        rm -rf "${temp_dir}"
    fi
}

# Template each Helm chart
template_helm_chart "${MODELSERVICE_CHART_URL}" "${MODELSERVICE_NAME}" "${MODELSERVICE_VERSION}" "${MANIFESTS_DIR}/modelservice" "llmd-modelservice"
template_helm_chart "${INFRA_CHART_URL}" "${INFRA_NAME}" "${INFRA_VERSION}" "${MANIFESTS_DIR}/infra" "llmd-infra"

# For Gateway API, we need to handle it differently as it's from GitHub releases
echo "Templating Gateway API Inference Extension..."
GATEWAY_TEMP_DIR=$(mktemp -d)
cd "${GATEWAY_TEMP_DIR}"
git clone --depth 1 --branch v${GATEWAY_API_VERSION} https://github.com/kubernetes-sigs/gateway-api-inference-extension.git
cd gateway-api-inference-extension/config/charts/inferencepool
helm template llmd-gateway-api . --namespace llm-d > "${MANIFESTS_DIR}/gateway-api/resources.yaml"
cd - > /dev/null
rm -rf "${GATEWAY_TEMP_DIR}"

# Create kustomization.yaml for each component
cat > "${MANIFESTS_DIR}/modelservice/kustomization.yaml" << EOF
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

namespace: llm-d

resources:
  - resources.yaml

commonLabels:
  app.kubernetes.io/part-of: llm-d
  app.kubernetes.io/component: modelservice
EOF

cat > "${MANIFESTS_DIR}/infra/kustomization.yaml" << EOF
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

namespace: llm-d

resources:
  - resources.yaml

commonLabels:
  app.kubernetes.io/part-of: llm-d
  app.kubernetes.io/component: infra
EOF

cat > "${MANIFESTS_DIR}/gateway-api/kustomization.yaml" << EOF
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

namespace: llm-d

resources:
  - resources.yaml

commonLabels:
  app.kubernetes.io/part-of: llm-d
  app.kubernetes.io/component: gateway-api
EOF

# Create default overlay that includes all components
cat > "${MANIFESTS_DIR}/overlays/default/kustomization.yaml" << EOF
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

namespace: llm-d

resources:
  - ../../infra
  - ../../modelservice
  - ../../gateway-api

commonLabels:
  app.kubernetes.io/managed-by: opendatahub-operator
EOF

echo "Successfully templated llm-d Helm charts to ${MANIFESTS_DIR}"
