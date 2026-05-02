# HelixPlay Deployment Guide

## Prerequisites

- Docker 24.0+ or Podman 4.0+
- NVIDIA Container Toolkit (for GPU hosts)
- Kubernetes 1.28+ (for production)
- TLS certificates (Let's Encrypt or custom CA)

## Quick Start (Docker Compose)

```bash
# Clone with submodules
git clone --recurse-submodules git@github.com:HelixDevelopment/HelixPlay.git
cd HelixPlay

# Start local topology
docker compose up -d

# Verify health
curl http://localhost:8080/health
curl http://localhost:8053/health
```

## GPU Host Agent Setup

### NVIDIA (Linux)

```bash
# Install NVIDIA Container Toolkit
distribution=$(. /etc/os-release;echo $ID$VERSION_ID)
curl -s -L https://nvidia.github.io/nvidia-docker/gpgkey | sudo apt-key add -
curl -s -L https://nvidia.github.io/nvidia-docker/$distribution/nvidia-docker.list | \
  sudo tee /etc/apt/sources.list.d/nvidia-docker.list
sudo apt-get update && sudo apt-get install -y nvidia-docker2
sudo systemctl restart docker

# Run host agent with GPU passthrough
docker run --gpus all --privileged \
  -v /dev:/dev \
  -e HELIXPLAY_GPU_VENDOR=nvidia \
  helixplay/host-agent:latest
```

### AMD (Linux)

```bash
docker run --device /dev/dri \
  --group-add video \
  -e HELIXPLAY_GPU_VENDOR=amd \
  helixplay/host-agent:latest
```

### Intel (Linux)

```bash
docker run --device /dev/dri \
  --group-add video \
  -e HELIXPLAY_GPU_VENDOR=intel \
  helixplay/host-agent:latest
```

## TLS Certificate Setup

```bash
# Generate mTLS certificates for host agents
mkdir -p certs
cd certs

# CA
cfssl gencert -initca ca-csr.json | cfssljson -bare ca

# Host agent certificate
cfssl gencert -ca=ca.pem -ca-key=ca-key.pem -config=ca-config.json \
  -profile=client host-agent-csr.json | cfssljson -bare host-agent

# Core backend certificate
cfssl gencert -ca=ca.pem -ca-key=ca-key.pem -config=ca-config.json \
  -profile=server core-csr.json | cfssljson -bare core
```

## Kubernetes Deployment

```bash
# Apply namespace and core services
kubectl apply -f deploy/k8s/namespace.yaml
kubectl apply -f deploy/k8s/core/
kubectl apply -f deploy/k8s/discovery/

# Deploy host agent DaemonSet (one per GPU node)
kubectl apply -f deploy/k8s/host-agent/daemonset.yaml

# Verify
kubectl get pods -n helixplay
kubectl logs -n helixplay -l app=helixplay-core
```

## Multi-Region Setup

1. Deploy core backend in each region
2. Configure GeoDNS for rendezvous endpoints
3. Set `HELIXPLAY_REGION` environment variable per deployment
4. Enable cross-region NATS event replication

## Monitoring

- Prometheus metrics: `:9090/metrics`
- Health endpoints: `:8080/health`, `:8053/health`
- Logs: Structured JSON via `slog`
