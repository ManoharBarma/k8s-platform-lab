# Kubernetes GitOps Platform Lab

Welcome to the Kubernetes GitOps Platform Lab. This repository is the source of truth for a complete, scalable, production-style GitOps architecture managed entirely by Argo CD. It provisions Istio Ambient mesh, observability (Prometheus, Grafana, Kiali), and a full-scale **banking-platform** microservices application (`banking-app`).

---

## 🏗️ Architecture

This lab strictly adheres to the GitOps philosophy: **Argo CD** is the only component that directly applies state to the Kubernetes cluster. 

```text
Bootstrap Script
       |
       v
   [Argo CD] <-------- (Detects Changes & Reconciles)
       |                            |
       |                            | (Source of Truth)
       |                            v
       |                        [GitHub Repository]
       |
       v
+-------------------------------------------------------------+
|                     Kubernetes Cluster                      |
|                                                             |
|  [Istio Core]      [Observability]      [Banking App]       |
|  (Helm Sources)    (Helm Sources)       (Helm Chart)        |
|  - istio-base      - Prometheus         - api-gateway       |
|  - istiod          - Grafana            - auth-service      |
|  - istio-cni       - Kiali              - account-service   |
|  - ztunnel                              - transaction-svc   |
|                                         - 30+ microservices |
|  [Istio Config] (Plain YAML)                                |
|  - Ambient labels, Gateways, Waypoints                      |
|  - PeerAuthentication, AuthorizationPolicy                  |
|  - VirtualService, DestinationRule                          |
+-------------------------------------------------------------+
```

### GitOps Flow
1. **Developer Change**: You modify a file (e.g., scale replicas, change timeout) and run `git commit` -> `git push`.
2. **GitHub**: The repository reflects the new desired state.
3. **Argo CD Detects Change**: Argo CD continuously polls GitHub (or receives webhooks).
4. **Argo CD Reconciles**: It compares the desired state in Git against the live cluster state.
5. **Kubernetes Desired State Updated**: Argo CD synchronizes the cluster, applying the exact configuration defined in Git.

### Centralized Helm Strategy
This repository demonstrates a scalable consumption model for Kubernetes configuration:
- **Platform/Third-Party Software (Istio, Prometheus, Kiali)**: Consumed directly through Argo CD **Helm references**. We *pin* the chart version and declare configuration overrides (values). We **DO NOT** commit huge, generated Helm manifests to Git, keeping the repository clean, reviewable, and easily upgradable.
- **Application Workloads (banking-app)**: Deployed via a **Helm chart** (`workloads/banking-app/`) managed by Argo CD. The chart packages 30+ microservices, infrastructure components (Postgres, Redis, Kafka), and ingress routing.

---

## 🚀 Getting Started

### 1. Bootstrap Argo CD
The bootstrap script is strictly responsible for one thing: installing Argo CD and exposing it. It does not install any lab workloads.
```bash
chmod +x bootstrap/bootstrap-argocd.sh
./bootstrap/bootstrap-argocd.sh
```

### 2. Connect Argo CD
The script will output your Argo CD URL (`http://<NodeIP>:30080/gitops`) and your initial admin credentials.
Once Argo CD is running, you configure it to point to this repository. Argo CD will then read the `argocd/applications/` directory and begin the Sync Wave process:
- **Wave 0**: Installs Istio Core via Helm (istio-base, istiod, istio-cni, ztunnel).
- **Wave 1**: Deploys the `banking-app` namespace and all banking platform microservices.
- **Wave 2**: Applies Istio routing and security configurations.
- **Wave 3**: Deploys the observability stack (kube-prometheus-stack).
- **Wave 4**: Deploys Kiali (Optional/Manual sync by default).

---

## 🕸️ Istio Ambient Mesh

Istio Ambient is a sidecar-less service mesh architecture dividing responsibilities into L4 and L7 components:

- **ztunnel (Zero Trust Tunnel)**: A lightweight, node-level proxy. It provides **L4 secure transport** by wrapping traffic in **HBONE** (HTTP-Based Overlay Network Encapsulation). It transparently enforces **mTLS** and basic workload identity via SPIFFE without injecting sidecars into your application pods.
- **Waypoint Proxy**: A namespace-level Envoy proxy handling **L7 processing**. While ztunnel handles fast L4 routing, Waypoints are dynamically provisioned only when advanced features are needed (e.g., HTTP routing, L7 authorization, retries, traffic splitting). *Note: Not all workloads use a waypoint automatically; it requires the `istio.io/use-waypoint` label.*

### Traffic Flow
**External Traffic (Ingress)**
```text
Client -> NodeIP:30080 -> nginx-ingress -> api-gateway (Port 8080) / frontend (Port 8080)
```
**Internal Mesh Traffic (L4 / L7)**
```text
api-gateway -> node iptables (istio-cni) -> ztunnel (encapsulates HBONE) -> target node ztunnel -> backend-service
```
If L7 policies are applied, ztunnel securely routes traffic to the Waypoint proxy first, which evaluates HTTP rules before delivering it to the backend.

### Traffic & Security Policies
- **Gateway**: Integrates Istio ingress/mesh entrypoints using the modern Kubernetes Gateway API (`gateway.networking.k8s.io`).
- **Waypoint**: Gateway API configuration instructing Istio to deploy an L7 proxy for `banking-app`.
- **PeerAuthentication**: Enforces STRICT mutual TLS (mTLS) across the namespace, rejecting plaintext traffic.
- **AuthorizationPolicy**: Demonstrates Zero-Trust access control, allowing only specific ports to the API gateway, leveraging Ambient's SPIFFE workload identities.
- **VirtualService**: Demonstrates HTTP path matching, custom header matching, timeouts, and automatic 5xx retries.
- **DestinationRule**: Configures connection pool limits to prevent overload and outlier detection (circuit breaking).

---

## 📊 Observability

- **Prometheus**: Deployed via Argo CD Helm source. Scrapes Kubernetes metrics (cAdvisor, kubelet) and Istio Ambient telemetry (ztunnel TCP metrics on `15020`, Waypoint HTTP metrics on `15090`).
- **Grafana**: Visualizes Prometheus data. Provides dashboards for request rates, error rates, latencies (RED metrics), and CPU/Memory usage.
- **Kiali**: (Optional) Visualizes the real-time Ambient service mesh topology, showing L4/L7 traffic edges and validating Istio configurations.

---

## 🛠️ Testing & Validation Commands

### Cluster Status
```bash
kubectl get pods -A
kubectl get svc -A
kubectl get ingress -A
kubectl -n banking-app get pods -o wide
```

### Istio Ambient Verification
```bash
istioctl ztunnel-config workloads
istioctl ztunnel-config certificates
istioctl ztunnel-config connections
```

### External Application Traffic
*(Replace `<NODE_IP>` with your lab's node IP address)*
```bash
curl -i http://<NODE_IP>:30080/
curl -i http://<NODE_IP>:30080/api/healthz
```

### Ambient Mesh Internal Traffic (From inside the mesh)
*Exec into one of the banking-app pods to test service-to-service communication:*
```bash
# Test basic connectivity to the API gateway
kubectl -n banking-app exec deploy/api-gateway-deployment -- curl -s http://auth-service:8080/healthz

# Test account service connectivity
kubectl -n banking-app exec deploy/api-gateway-deployment -- curl -s http://account-service:8080/healthz

# Test transaction service connectivity
kubectl -n banking-app exec deploy/api-gateway-deployment -- curl -s http://transaction-service:8080/healthz
```
