#!/usr/bin/env bash

set -euo pipefail

# ============================================================
# Kubernetes GitOps Lab - Minimal Argo CD Bootstrap
#
# Responsibility:
#   1. Verify Kubernetes
#   2. Fix mount propagation required by the lab environment
#   3. Install Argo CD
#   4. Expose Argo CD through existing nginx-ingress at /gitops
#
# Everything else is managed by Argo CD from Git.
#
# NOT installed here:
#   - Istio
#   - Istio CNI
#   - ztunnel
#   - Prometheus
#   - Grafana
#   - Kiali
#   - Sample applications
#   - Application ingress
#   - Istio policies
#   - App-of-Apps / ApplicationSets
# ============================================================

ARGOCD_NAMESPACE="argocd"
ARGOCD_INSTALL_URL="https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml"

GITOPS_PATH="/gitops"

# ============================================================
# Helpers
# ============================================================

log() {
    echo
    echo "============================================================"
    echo "$1"
    echo "============================================================"
}

fail() {
    echo
    echo "ERROR: $1"
    exit 1
}

command -v kubectl >/dev/null 2>&1 || \
    fail "kubectl is not installed or not available in PATH."

# ============================================================
# 1. Kubernetes connectivity
# ============================================================

log "Checking Kubernetes connectivity"

kubectl cluster-info >/dev/null 2>&1 || \
    fail "Cannot connect to Kubernetes cluster."

echo "Kubernetes cluster is reachable."

# ============================================================
# 2. Wait for node
# ============================================================

log "Waiting for Kubernetes node"

NODE=""

for i in {1..60}; do
    NODE="$(kubectl get nodes \
        --no-headers \
        -o custom-columns=':metadata.name' 2>/dev/null | head -n1 || true)"

    if [[ -n "$NODE" ]]; then
        break
    fi

    echo "Waiting for node... ($i/60)"
    sleep 2
done

[[ -n "$NODE" ]] || fail "No Kubernetes node found."

echo "Node: $NODE"

kubectl wait \
    --for=condition=Ready \
    "node/$NODE" \
    --timeout=180s

echo "Node is Ready."

# ============================================================
# 3. Fix mount propagation
#
# The KodeKloud/containerized lab environment previously
# required the root filesystem to be shared for Istio CNI.
#
# This does not install Istio.
# It only prepares the node environment.
# ============================================================

log "Checking mount propagation"

CURRENT_PROPAGATION="$(
    findmnt -n -o PROPAGATION / 2>/dev/null || true
)"

echo "Current / propagation: ${CURRENT_PROPAGATION:-unknown}"

if [[ "$CURRENT_PROPAGATION" != "shared" ]]; then

    echo "Root filesystem is not shared."
    echo "Applying shared mount propagation..."

    sudo mount --make-rshared /

    CURRENT_PROPAGATION="$(
        findmnt -n -o PROPAGATION / 2>/dev/null || true
    )"

    echo "New / propagation: ${CURRENT_PROPAGATION:-unknown}"

    [[ "$CURRENT_PROPAGATION" == "shared" ]] || \
        fail "Failed to configure shared mount propagation."
else
    echo "Root filesystem is already shared."
fi

# ============================================================
# 4. Create Argo CD namespace
# ============================================================

log "Preparing Argo CD namespace"

kubectl get namespace "$ARGOCD_NAMESPACE" >/dev/null 2>&1 || \
    kubectl create namespace "$ARGOCD_NAMESPACE"

echo "Namespace: $ARGOCD_NAMESPACE"

# ============================================================
# 5. Install Argo CD
#
# Server-side apply avoids the large-annotation issue that
# can occur with the Argo CD installation manifest.
# ============================================================

log "Installing Argo CD"

kubectl apply \
    --server-side \
    --force-conflicts \
    -n "$ARGOCD_NAMESPACE" \
    -f "$ARGOCD_INSTALL_URL"

echo "Argo CD manifests applied."

# ============================================================
# 6. Wait for Argo CD
# ============================================================

log "Waiting for Argo CD components"

kubectl rollout status \
    deployment/argocd-server \
    -n "$ARGOCD_NAMESPACE" \
    --timeout=300s

kubectl rollout status \
    deployment/argocd-repo-server \
    -n "$ARGOCD_NAMESPACE" \
    --timeout=300s

kubectl rollout status \
    deployment/argocd-applicationset-controller \
    -n "$ARGOCD_NAMESPACE" \
    --timeout=300s

echo "Argo CD core components are Ready."

# ============================================================
# 7. Configure Argo CD for /gitops
# ============================================================

log "Configuring Argo CD under /gitops"

kubectl patch configmap argocd-cmd-params-cm \
    -n "$ARGOCD_NAMESPACE" \
    --type merge \
    -p '{
      "data": {
        "server.rootpath": "/gitops",
        "server.basehref": "/gitops"
      }
    }'

echo "Argo CD root path configured."

# Restart server so configuration is picked up.

kubectl rollout restart \
    deployment/argocd-server \
    -n "$ARGOCD_NAMESPACE"

kubectl rollout status \
    deployment/argocd-server \
    -n "$ARGOCD_NAMESPACE" \
    --timeout=300s

# ============================================================
# 8. Create nginx ingress for Argo CD
#
# Uses the EXISTING nginx ingress controller.
#
# Argo CD service speaks HTTPS on port 443.
# nginx communicates with Argo CD using HTTPS.
# ============================================================

log "Creating /gitops ingress"

kubectl apply -f - <<'EOF'
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: argocd
  namespace: argocd
  annotations:
    nginx.ingress.kubernetes.io/backend-protocol: "HTTPS"
    nginx.ingress.kubernetes.io/ssl-redirect: "false"
spec:
  ingressClassName: nginx
  rules:
    - http:
        paths:
          - path: /gitops
            pathType: Prefix
            backend:
              service:
                name: argocd-server
                port:
                  number: 443
EOF

echo "Argo CD ingress created."

# ============================================================
# 9. Wait briefly for ingress configuration
# ============================================================

sleep 5

# ============================================================
# 10. Get node IP
# ============================================================

NODE_IP="$(
    kubectl get node "$NODE" \
        -o jsonpath='{.status.addresses[?(@.type=="InternalIP")].address}'
)"

if [[ -z "$NODE_IP" ]]; then
    NODE_IP="$(
        kubectl get node "$NODE" \
            -o jsonpath='{.status.addresses[?(@.type=="ExternalIP")].address}'
    )"
fi

# ============================================================
# 11. Retrieve initial admin password
#
# Password remains the generated Argo CD password.
# Nothing is hardcoded into the script or Git repository.
# ============================================================

PASSWORD="$(
    kubectl -n "$ARGOCD_NAMESPACE" \
        get secret argocd-initial-admin-secret \
        -o jsonpath="{.data.password}" 2>/dev/null \
        | base64 -d 2>/dev/null || true
)"

# ============================================================
# 12. Final status
# ============================================================

log "Argo CD Bootstrap Complete"

echo
echo "Argo CD namespace:"
echo "  $ARGOCD_NAMESPACE"

echo
echo "Argo CD pods:"
kubectl get pods -n "$ARGOCD_NAMESPACE"

echo
echo "Argo CD services:"
kubectl get svc -n "$ARGOCD_NAMESPACE"

echo
echo "Argo CD ingress:"
kubectl get ingress -n "$ARGOCD_NAMESPACE"

echo
echo "------------------------------------------------------------"
echo "ACCESS"
echo "------------------------------------------------------------"

if [[ -n "$NODE_IP" ]]; then
    echo "Argo CD:"
    echo "  http://${NODE_IP}:30080${GITOPS_PATH}"
else
    echo "Node IP could not be detected."
    echo "Use the Kubernetes node IP manually."
fi

echo
echo "Username:"
echo "  admin"

echo
echo "Initial password:"
if [[ -n "$PASSWORD" ]]; then
    echo "  $PASSWORD"
else
    echo "  Could not retrieve automatically."
    echo
    echo "Retrieve it with:"
    echo "  kubectl -n argocd get secret argocd-initial-admin-secret \\"
    echo "    -o jsonpath=\"{.data.password}\" | base64 -d; echo"
fi

echo
echo "------------------------------------------------------------"
echo "NEXT STEP"
echo "------------------------------------------------------------"

echo
echo "Connect this Argo CD instance to your GitHub repository."
echo
echo "After that, Argo CD will manage:"
echo "  - Istio Ambient"
echo "  - Istio configuration"
echo "  - Prometheus"
echo "  - Grafana"
echo "  - Kiali"
echo "  - Sample applications"
echo "  - Application networking"
echo
echo "Bootstrap is complete."
echo "From this point onward, Git should be the source of truth."
echo
