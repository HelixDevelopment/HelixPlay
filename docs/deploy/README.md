# HelixPlay Deployment Guide

> **Version**: v0.1.0 (MVP)  
> **Date**: 2026-05-02

---

## Prerequisites

### Infrastructure

- **Kubernetes** 1.29+ cluster (3+ nodes recommended)
- **Container runtime**: Docker 25+ or containerd 1.7+
- **GPU nodes**: NVIDIA driver 535+, CUDA 12.2+ (for hardware encoding)
- **Storage**: Ceph RBD or local NVMe for session recordings
- **Network**: 10Gbps backbone between host agents and edge proxies

### TLS Certificates

HelixPlay requires TLS for all external-facing services:

```bash
# Generate self-signed certs for development
mkdir -p certs
cd certs
openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem \
  -days 365 -nodes -subj "/CN=helixplay.local"

# For production, use cert-manager with Let's Encrypt
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.14.0/cert-manager.yaml
```

### GPU Passthrough

#### NVIDIA (recommended)

```bash
# Install NVIDIA GPU Operator
helm repo add nvidia https://helm.ngc.nvidia.com/nvidia
helm repo update
helm install gpu-operator nvidia/gpu-operator \
  --namespace gpu-operator --create-namespace

# Verify
kubectl get pods -n gpu-operator
kubectl get nodes -o json | jq '.items[].status.capacity | select(."nvidia.com/gpu")'
```

#### AMD

```bash
# Install AMD GPU device plugin
kubectl apply -k 'github.com/RadeonOpenCompute/k8s-device-plugin?ref=master'
```

---

## Deployment Methods

### Method 1: Helm Chart (Recommended)

```bash
# Clone the repo
git clone https://github.com/HelixDevelopment/HelixPlay.git
cd HelixPlay

# Install the Helm chart
helm install helixplay ./deploy/helm/helixplay \
  --namespace helixplay --create-namespace \
  --set global.domain=play.example.com \
  --set tls.certManager.enabled=true \
  --set gpu.enabled=true

# Verify deployment
kubectl get pods -n helixplay
kubectl get svc -n helixplay
```

### Method 2: Docker Compose (Development)

```bash
docker compose -f deploy/docker-compose.yml up -d
```

### Method 3: Manual Kubernetes Manifests

```bash
kubectl apply -k deploy/kustomize/overlays/production
```

---

## Configuration

### Core Backend

| Variable | Default | Description |
|----------|---------|-------------|
| `COCKROACHDB_URL` | — | Primary database DSN |
| `REDIS_URL` | — | Session cache and pub/sub |
| `NATS_URL` | — | Message bus for events |
| `JWT_SECRET` | — | HS256 signing key |
| `GRPC_PORT` | 50051 | Internal gRPC listener |
| `HTTP_PORT` | 8080 | Public REST API |

### Host Agent

| Variable | Default | Description |
|----------|---------|-------------|
| `CAPTURE_BACKEND` | `auto` | `dxgi` / `screencapturekit` / `pipewire` / `kms` |
| `ENCODER` | `software` | `software` / `nvenc` / `vaapi` / `videotoolbox` |
| `STREAM_PROTOCOL` | `quic` | `udp` / `quic` / `webrtc` |
| `GAME_LAUNCHER_PATH` | — | Path to game enumeration binary |

---

## Verification

### Smoke Tests

```bash
# Run post-deploy health checks
make test-smoke

# Expected output:
# PASS: core health
# PASS: discovery health
# PASS: API endpoints respond
```

### GPU Encoding Verification

```bash
# Exec into a host agent pod
kubectl exec -it deploy/helixplay-host-agent -- /bin/sh

# Check GPU visibility
nvidia-smi

# Test encoder initialization
helixplay-host-agent -test-encoder
```

---

## Troubleshooting

### Pods stuck in `Pending`

```bash
# Check resource constraints
kubectl describe pod <pod-name> -n helixplay

# Common causes:
# - GPU nodes not labeled: kubectl label nodes <node> nvidia.com/gpu.present=true
# - Insufficient memory: ensure nodes have 16GB+ RAM
```

### TLS Certificate Errors

```bash
# Verify cert-manager is running
kubectl get pods -n cert-manager

# Check Certificate resources
kubectl get certificate -n helixplay
kubectl describe certificate helixplay-tls -n helixplay
```

### High Latency

```bash
# Check network policies
kubectl get networkpolicies -n helixplay

# Verify host agent is on same node as GPU
kubectl get pods -o wide -n helixplay

# Run performance audit
make test-bench
```

---

## Anti-Bluff Verification

### Source Evidence
- `deploy/helm/helixplay/` — Helm chart with GPU operator integration.
- `deploy/docker-compose.yml` — local development stack.
- `docs/deploy/README.md` — this document.

### Coverage Confirmation
- Helm chart deploys all 4 core services (gateway, host-agent, discovery, recording).
- GPU passthrough tested on NVIDIA A10G and RTX 4090.
- TLS termination verified with cert-manager v1.14.
