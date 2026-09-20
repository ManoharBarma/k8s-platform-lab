#!/bin/bash
set -e

echo "🚀 Bootstrapping Kubernetes Platform Lab with Argo CD..."

echo "📂 1. Applying Argo CD Projects..."
kubectl apply -f argocd/projects/

echo "🛠️  2. Applying Platform Applications (Prometheus, Grafana, Istio, etc.)..."
kubectl apply -f argocd/applications/platform/

echo "📦 3. Applying Workload Applications (Banking App)..."
kubectl apply -f argocd/applications/workloads/

echo "✅ Done! Argo CD will now detect the changes and begin syncing the configurations."
echo "🔍 You can monitor the progress by running: kubectl get applications -n argocd"
