#!/bin/bash

helm install prom \
  oci://ghcr.io/prometheus-community/charts/kube-prometheus-stack \
  --set kubeStateMetrics.enabled=false \
  --set nodeExporter.enabled=false \
  --set grafana.enabled=false

helm install cert-manager \
  oci://quay.io/jetstack/charts/cert-manager \
  --version v1.19.1 \
  --namespace cert-manager \
  --create-namespace \
  --set crds.enabled=true