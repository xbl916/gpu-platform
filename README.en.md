# GPU Platform - GPU Container Service Management Platform

<div align="center">

![License](https://img.shields.io/badge/License-MIT-blue.svg)
![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat-square&logo=go)
![React](https://img.shields.io/badge/React-18.2-61DAFB?style=flat-square&logo=react)
![Kubernetes](https://img.shields.io/badge/Kubernetes-1.28-326CE5?style=flat-square&logo=kubernetes)
![Docker](https://img.shields.io/badge/Docker-24.0+-2496ED?style=flat-square&logo=docker)

A GPU container service management platform similar to AutoDL, supporting GPU resource pool management, container instance lifecycle management, and container template customization.

[English](README.en.md) | [简体中文](README.md)

</div>

## ✨ Features

### 🎯 Core Features

| Feature | Description |
|--------|-------------|
| **User Management** | User registration, login, JWT authentication, role-based access control |
| **GPU Resource Management** | GPU server management, resource pool configuration, real-time status monitoring |
| **Container Management** | Create, start, stop, restart, and delete container instances |
| **Template Marketplace** | Pre-built container templates (PyTorch, TensorFlow, etc.), custom templates |
| **Resource Quotas** | User resource quota management to prevent resource abuse |
| **Monitoring & Alerts** | GPU temperature, memory usage, CPU/memory monitoring, alert notifications |
| **Project Management** | Multi-tenant isolation, project member management, resource sharing |

---

## 📋 Table of Contents

- [Quick Start](#quick-start)
- [Environment Requirements](#environment-requirements)
- [Pre-deployment Preparation](#pre-deployment-preparation)
- [Database Initialization](#database-initialization)
- [Configuration](#configuration)
- [Starting Services](#starting-services)
- [Development Environment](#development-environment)
- [Production Deployment](#production-deployment)
- [FAQ](#faq)

---

## 🚀 Quick Start

### Environment Requirements

| Component | Minimum | Recommended |
|-----------|---------|-------------|
| Kubernetes | 1.20 | 1.28+ |
| Go | 1.20 | 1.21+ |
| Node.js | 16 | 20 LTS |
| PostgreSQL | 12 | 15+ |
| Redis | 6 | 7+ |
| NVIDIA GPU Driver | 450.x | 535.x+ |

---

## 🔧 Pre-deployment Preparation

### 1. Prepare Kubernetes Cluster

```bash
# Check cluster connection
kubectl cluster-info

# Check node status
kubectl get nodes

# Create dedicated namespace
kubectl create namespace gpu-platform
```

### 2. Install NVIDIA GPU Operator

```bash
# Add NVIDIA Helm repository
helm repo add nvidia https://nvidia.github.io/gpu-operator
helm repo update

# Create namespace
kubectl create namespace gpu-operator-ns

# Install GPU Operator
helm install gpu-operator nvidia/gpu-operator \
  --namespace gpu-operator-ns \
  --version 24.3.0 \
  --set driver.enabled=true \
  --set toolkit.enabled=true \
  --set devicePlugin.enabled=true \
  --set dcgmExporter.enabled=true

# Monitor installation progress
kubectl get pods -n gpu-operator-ns -w

# Verify installation
kubectl get nodes -l nvidia.com/gpu.product
```

### 3. Prepare Storage Backend

```bash
# Option 1: MinIO (recommended for testing)
docker run -d \
  --name minio \
  -p 9000:9000 \
  -p 9001:9001 \
  -e MINIO_ROOT_USER=admin \
  -e MINIO_ROOT_PASSWORD=password123 \
  quay.io/minio/minio server /data --console-address ":9001"
```

---

## 📦 Database Initialization

### 1. Install PostgreSQL

```bash
# Docker method (quick start)
docker run -d \
  --name postgres \
  -e POSTGRES_USER=gpu_admin \
  -e POSTGRES_PASSWORD=your_password \
  -e POSTGRES_DB=gpu_platform \
  -p 5432:5432 \
  postgres:15
```

### 2. Run Database Migrations

```bash
# Navigate to project root
cd /path/to/gpu-platform

# Build migration tool
go build -o bin/migrate ./scripts/migrate.go

# Run migrations
./bin/migrate -config=config.yaml -action=up
./bin/migrate -config=config.yaml -action=seed
```

---

## ⚙️ Configuration

### Configuration File Structure

```yaml
# config.yaml

# Application Configuration
app:
  host: "0.0.0.0"
  port: 8080
  name: "GPU Platform"
  domain: "gpu-platform.local"

# Database Configuration
database:
  host: "localhost"
  port: 5432
  username: "gpu_admin"
  password: "your_password"
  name: "gpu_platform"
  sslmode: "disable"
  max_open_conns: 25
  max_idle_conns: 5
  conn_max_lifetime: 300

# Redis Configuration
redis:
  host: "localhost"
  port: 6379
  password: ""
  db: 0

# Storage Configuration (MinIO/S3)
storage:
  endpoint: "localhost:9000"
  access_key: "minioadmin"
  secret_key: "miniopassword"
  use_ssl: false
  bucket_prefix: "gpu-platform"

# Kubernetes Configuration
kubernetes:
  kubeconfig: ""
  namespace: "gpu-platform"
  gpu_pool_label: "gpu-pool"
  instance_label: "gpu-instance"

# JWT Configuration
jwt:
  secret: "your-secret-key-change-in-production"
  access_token_expire: 3600
  refresh_token_expire: 604800

# GPU Default Configuration
gpu:
  default_gpu_model: "a100"
  max_gpu_per_instance: 8
  default_cpu_cores: 4
  max_cpu_cores: 64
  default_memory_mb: 16384
  max_memory_mb: 262144
  default_storage_gb: 100
  max_storage_gb: 2000

# Container Default Configuration
container:
  default_image: "nvidia/cuda:12.1-runtime-ubuntu22.04"
  max_instances_per_user: 10
  max_running_hours: 24
  workspace_size_gb: 50
  data_size_gb: 200

# Monitoring Configuration
monitoring:
  enabled: true
  metrics_port: 9090
  alert_threshold_gpu_temp: 85
  alert_threshold_gpu_memory: 90

# Notification Configuration
notification:
  enabled: true
  email:
    smtp_host: "smtp.example.com"
    smtp_port: 587
    smtp_user: "noreply@example.com"
    smtp_password: "your-smtp-password"
    from_address: "noreply@example.com"
  webhook:
    enabled: false
    url: "https://your-webhook.com/notify"
```

---

## ▶️ Starting Services

### Option 1: Direct Run (Development)

```bash
# Copy configuration template
cp config.yaml.example config.yaml

# Edit configuration
vim config.yaml

# Build and start API service
cd cmd/api-server
go run main.go

# Start frontend in new terminal
cd frontend
npm install
npm run dev
```

### Option 2: Docker Compose (Quick Testing)

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop all services
docker-compose down
```

### Verification

```bash
# API health check
curl http://localhost:8080/health

# Expected output:
# {"status":"healthy"}

# Access web interface
# Open browser to http://localhost:5173
```

---

## 💻 Development Environment

### 1. Environment Setup

```bash
# Install Go
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin

# Install Node.js
curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
sudo apt-get install -y nodejs

# Install kubectl
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
chmod +x kubectl
sudo mv kubectl /usr/local/bin/

# Configure kubectl
export KUBECONFIG=/path/to/your/kubeconfig
```

### 2. Start Dependent Services

```bash
# PostgreSQL
docker run -d \
  --name postgres-dev \
  -e POSTGRES_USER=gpu_admin \
  -e POSTGRES_PASSWORD=password \
  -e POSTGRES_DB=gpu_platform \
  -p 5432:5432 \
  postgres:15

# Redis
docker run -d \
  --name redis-dev \
  -p 6379:6379 \
  redis:7-alpine \
  redis-server --requirepass password

# MinIO
docker run -d \
  --name minio-dev \
  -p 9000:9000 \
  -p 9001:9001 \
  -e MINIO_ROOT_USER=minioadmin \
  -e MINIO_ROOT_PASSWORD=miniopassword \
  quay.io/minio/minio server /data --console-address ":9001"
```

---

## 🏭 Production Deployment

### 1. Prepare Production Environment

```bash
# Create production namespace
kubectl create namespace gpu-platform-prod

# Create secrets
kubectl create secret generic gpu-platform-secrets \
  --from-literal=database-password="your-secure-password" \
  --from-literal=jwt-secret="your-super-secure-jwt-secret" \
  --from-literal=redis-password="your-redis-password" \
  --from-literal=minio-access-key="your-access-key" \
  --from-literal=minio-secret-key="your-secret-key" \
  -n gpu-platform-prod
```

### 2. Deploy to Kubernetes

```bash
# Using Helm
helm install gpu-platform ./deployments/helm/gpu-platform \
  --namespace gpu-platform-prod \
  --create-namespace \
  --values gpu-platform-values.yaml

# Verify deployment
kubectl get pods -n gpu-platform-prod
kubectl logs -n gpu-platform-prod -l app=gpu-platform --tail=100
```

### 3. Configure HTTPS

```bash
# Install cert-manager
helm install cert-manager jetstack/cert-manager \
  --namespace cert-manager \
  --create-namespace \
  --version v1.14.0

# Apply production ingress configuration with TLS
# See deployment documentation for details
```

---

## ❓ FAQ

### Q1: GPU nodes not found?

```bash
# Check node labels
kubectl get nodes --show-labels | grep nvidia

# Manually add label if missing
kubectl label nodes <gpu-node-name> nvidia.com/gpu.product=A100-SXM4-40GB --overwrite

# Verify GPU resources
kubectl get nodes -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.status.allocatable.nvidia\.com/gpu}{"\n"}'
```

### Q2: GPU Operator installation failed?

```bash
# Check operator logs
kubectl logs -n gpu-operator-ns deployment/gpu-operator

# Uninstall and reinstall
helm uninstall gpu-operator -n gpu-operator-ns
kubectl delete crd $(kubectl get crd | grep nvidia | awk '{print $1}')
helm install gpu-operator nvidia/gpu-operator -n gpu-operator-ns
```

### Q3: GPU memory not visible?

```bash
# Check device plugin status
kubectl get ds nvidia-device-plugin-daemonset -n gpu-operator-ns

# Restart device plugin
kubectl rollout restart ds nvidia-device-plugin-daemonset -n gpu-operator-ns
```

---

## 📚 Reference Documentation

| Documentation | Link |
|---------------|------|
| Kubernetes | https://kubernetes.io/docs/home/ |
| NVIDIA GPU Operator | https://docs.nvidia.com/datacenter/cloud-native/gpu-operator/ |
| Go Documentation | https://go.dev/doc/ |
| React Documentation | https://react.dev/ |
| Gin Framework | https://gin-gonic.com/docs/ |

---

## 📄 License

MIT License

---

<div align="center">
Made with ❤️ by GPU Platform Team
</div>
