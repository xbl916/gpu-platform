# GPU Platform - GPU 容器化服务管理平台

<div align="center">

![License](https://img.shields.io/badge/License-MIT-blue.svg)
![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat-square&logo=go)
![React](https://img.shields.io/badge/React-18.2-61DAFB?style=flat-square&logo=react)
![Kubernetes](https://img.shields.io/badge/Kubernetes-1.28-326CE5?style=flat-square&logo=kubernetes)
![Docker](https://img.shields.io/badge/Docker-24.0+-2496ED?style=flat-square&logo=docker)

一个类似 AutoDL 的 GPU 容器化服务管理平台，支持 GPU 资源池管理、容器实例生命周期管理、容器模板定制等功能。

[English](README.md) | 简体中文

</div>

## ✨ 功能特性

### 🎯 核心功能

| 功能 | 描述 |
|------|------|
| **用户管理** | 用户注册、登录、JWT 认证、角色权限控制 |
| **GPU 资源管理** | GPU 服务器管理、资源池配置、实时状态监控 |
| **容器管理** | 创建、启动、停止、重启、删除容器实例 |
| **模板市场** | 预置容器模板（PyTorch、TensorFlow 等）、自定义模板 |
| **资源配额** | 用户资源配额管理、防止资源滥用 |
| **监控告警** | GPU 温度、显存使用率、CPU/内存监控、告警通知 |
| **文件管理** | 基于 MinIO 的用户文件存储管理 |
| **Web 终端** | 基于 WebSocket 的 Web SSH 终端 |

## 📋 目录

- [快速开始](#快速开始)
- [环境要求](#环境要求)
- [Kubernetes 集群部署](#kubernetes-集群部署)
- [平台部署](#平台部署)
- [本地开发](#本地开发)
- [生产环境配置](#生产环境配置)
- [API 文档](#api-文档)
- [项目结构](#项目结构)
- [常见问题](#常见问题)

---

## 🚀 快速开始

### 前置条件

在开始部署之前，请确保你已经具备以下环境：

1. **Kubernetes 集群**（至少 1 个 master 节点）
2. **GPU 节点**（至少 1 个带有 NVIDIA GPU 的节点）
3. **Helm 3** - 用于安装 GPU Operator
4. **Docker/Docker Compose** - 用于本地开发

### 步骤 1：准备 Kubernetes 集群

首先，确保你有一个运行中的 Kubernetes 集群。如果还没有，请参考 [官方文档](https://kubernetes.io/docs/setup/) 创建集群。

推荐方案：
- **轻量级**: [k3s](https://k3s.io/) - 适合测试和小规模部署
- **生产级**: [ kubeadm](https://kubernetes.io/docs/reference/setup-tools/kubeadm/) - 生产环境推荐

### 步骤 2：安装 NVIDIA GPU Operator

NVIDIA GPU Operator 是必须在 GPU 节点上安装的关键组件，它负责：

- 自动安装 NVIDIA 容器运行时
- 管理 NVIDIA 设备插件
- 自动配置 GPU 监控
- 管理 DCGM Exporter

#### 使用 Helm 安装 GPU Operator

```bash
# 1. 添加 NVIDIA Helm 仓库
helm repo add nvidia https://nvidia.github.io/gpu-operator
helm repo update

# 2. 创建 namespace
kubectl create namespace gpu-operator

# 3. 安装 GPU Operator（选择合适的版本）
helm install gpu-operator nvidia/gpu-operator \
  --namespace gpu-operator \
  --version 23.9.0 \
  --set driver.enabled=false \  # 如果已安装驱动，设置为 false
  --set toolkit.enabled=true

# 4. 验证安装
kubectl get pods -n gpu-operator

# 等待所有 Pod 运行正常
# NAME                                             READY   STATUS    RESTARTS   AGE
# gpu-operator-7f5b5b6b8c-abcde                    1/1     Running   0          2m
# nvidia-container-toolkit-daemonset-xxxxx           1/1     Running   0          1m
# nvidia-device-plugin-daemonset-xxxxx              1/1     Running   0          1m
```

#### 验证 GPU 节点配置

```bash
# 检查 GPU 节点状态
kubectl get nodes -l accelerator=nvidia

# 检查 GPU 设备插件
kubectl get nodes -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.status.allocatable.nvidia\.com/gpu}{"\n"}'

# 预期输出示例:
# gpu-node-1    2
# gpu-node-2    4
```

#### （可选）手动安装 NVIDIA 驱动

如果你的节点还没有安装 NVIDIA 驱动，可以使用以下方式：

**Ubuntu/Debian:**
```bash
# 添加 NVIDIA 驱动仓库
wget https://developer.download.nvidia.com/compute/cuda/repos/ubuntu2204/x86_64/cuda-keyring_1.1-1_all.deb
sudo dpkg -i cuda-keyring_1.1-1_all.deb
sudo apt-get update
sudo apt-get install -y nvidia-driver-535 nvidia-dkms-535
```

**CentOS/RHEL:**
```bash
# 添加 NVIDIA 驱动仓库
sudo dnf config-manager --add-repo https://developer.download.nvidia.com/compute/cuda/repos/rhel9/x86_64/cuda-rhel9.repo
sudo dnf module install nvidia-driver:535
```

> **重要**: 重启节点使驱动生效

### 步骤 3：准备存储后端

GPU Platform 需要以下存储：

1. **PostgreSQL** - 主数据库
2. **Redis** - 缓存和会话
3. **MinIO** - 对象存储（用户文件）

你可以选择：

#### 方式 A：使用自建服务（生产推荐）

```bash
# PostgreSQL
kubectl apply -f https://raw.githubusercontent.com/postgresql/postgres-operator/master/deploy/manifests/postgres.yaml

# 或使用 Helm
helm install postgresql bitnami/postgresql \
  --set auth.postgresPassword=your_password \
  --set auth.database=gpu_platform

# MinIO
helm install minio bitnami/minio \
  --set auth.rootUser=minioadmin \
  --set auth.rootPassword=minio_password
```

#### 方式 B：使用云服务（快速开始）

- **PostgreSQL**: [Supabase](https://supabase.com/)、[Neon](https://neon.tech/)
- **Redis**: [Redis Cloud](https://redis.com/)、[Upstash](https://upstash.com/)
- **MinIO**: [Backblaze B2](https://www.backblaze.com/)、[AWS S3](https://aws.amazon.com/s3/)

### 步骤 4：配置平台

#### 克隆项目

```bash
git clone https://github.com/xbl916/gpu-platform.git
cd gpu-platform
```

#### 配置数据库连接

编辑 `config.yaml`:

```yaml
database:
  host: "your-postgres-host"     # PostgreSQL 地址
  port: 5432
  username: "gpu_admin"
  password: "your_secure_password"
  name: "gpu_platform"
  sslmode: "require"               # 生产环境必须启用
```

#### 配置 Kubernetes 连接

```yaml
kubernetes:
  # 方式 1: 使用 kubeconfig 文件
  kubeconfig: "/path/to/your/kubeconfig"
  
  # 方式 2: 使用集群内配置（推荐）
  # kubeconfig: ""  # 留空表示使用 in-cluster 配置
  
  namespace: "gpu-platform"
```

### 步骤 5：部署 GPU Platform

#### 方式 A：Helm 部署（推荐）

```bash
# 创建 namespace
kubectl create namespace gpu-platform

# 安装 GPU Platform
helm install gpu-platform ./deployments/helm/gpu-platform \
  --namespace gpu-platform \
  --set image.tag=latest \
  --set replicaCount=3
```

#### 方式 B：Kustomize 部署

```bash
kubectl apply -k deployments/k8s/overlays/production
```

#### 方式 C：手动部署

```bash
# 创建命名空间
kubectl create namespace gpu-platform

# 应用所有配置
kubectl apply -f deployments/k8s/base/

# 检查状态
kubectl get pods -n gpu-platform
```

### 步骤 6：验证部署

```bash
# 检查 API 服务
kubectl get svc -n gpu-platform
# 预期输出:
# NAME             TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)          AGE
# gpu-platform     ClusterIP   10.96.0.1     <none>        8080/TCP         2m

# 检查日志
kubectl logs -n gpu-platform -l app=gpu-platform-api --tail=100
```

---

## 🔧 本地开发

### 环境要求

- **Go 1.21+**
- **Node.js 18+**
- **PostgreSQL 14+**
- **Redis 7+**
- **MinIO**（或使用本地 S3 兼容存储）
- **kubectl** - 用于连接 K8s 集群

### 1. 启动依赖服务

```bash
# 使用 Docker Compose 启动基础服务
docker-compose -f docker-compose.dev.yml up -d postgres redis minio

# 验证服务
docker ps
```

### 2. 配置本地连接

创建 `.env` 文件：

```bash
# 数据库
DB_HOST=localhost
DB_PORT=5432
DB_USER=gpu_admin
DB_PASSWORD=your_password
DB_NAME=gpu_platform

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379

# MinIO
MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minio_password

# Kubernetes（可选，用于本地开发）
KUBECONFIG=~/.kube/config
```

### 3. 启动后端服务

```bash
# 安装依赖
go mod download

# 运行服务
make run

# 后端运行在 http://localhost:8080
```

### 4. 启动前端开发服务器

```bash
cd frontend

# 安装依赖
npm install

# 启动开发服务器
npm run dev

# 前端运行在 http://localhost:3000
```

---

## 📊 Kubernetes 集群详细部署

### 1. 创建 GPU 节点池

**阿里云 ACK:**
```yaml
# 在 ACK 控制台创建节点池
# 选择 GPU 实例类型: ecs.gn6v-c8g1.2xlarge (V100 * 1)
```

**AWS EKS:**
```bash
# 创建节点组
aws eks create-nodegroup \
  --cluster-name my-gpu-cluster \
  --nodegroup-name gpu-nodes \
  --node-role arn:aws:iam::123456789:role/EKSNodeRole \
  --subnets subnet-xxx \
  --instance-types p4d.24xlarge \
  --scaling-config minSize=1,maxSize=10,desiredSize=2
```

**GCP GKE:**
```bash
# 创建 GPU 节点池
gcloud container node-pools create gpu-pool \
  --cluster my-gpu-cluster \
  --machine-type a2-highgpu-1g \
  --accelerator type=nvidia-tesla-a100,count=1 \
  --num-nodes 2 \
  --enable-autoscaling
```

### 2. 安装 NVIDIA 驱动

**使用 Node Feature Discovery (NFD):**

```bash
# 安装 NFD
helm install nfd nfd/nfd \
  --namespace node-feature-discovery \
  --create-namespace

# NFD 会自动检测 GPU 节点
kubectl get nfd cr
```

### 3. 配置设备插件

GPU Operator 会自动安装设备插件，验证：

```bash
# 检查设备插件
kubectl get ds nvidia-device-plugin-daemonset -n gpu-operator

# 验证 GPU 资源
kubectl get nodes -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.status.allocatable}{"\n"}' | grep nvidia
```

### 4. 配置资源限制

创建 `gpu-platform` namespace：

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: gpu-platform
  labels:
    gpu-platform.io/enabled: "true"
```

配置资源配额：

```yaml
apiVersion: v1
kind: ResourceQuota
metadata:
  name: gpu-quota
  namespace: gpu-platform
spec:
  hard:
    nvidia.com/gpu: "100"
    cpu: "1000"
    memory: 4Ti
```

---

## 🔐 生产环境配置

### 1. 启用 TLS

```yaml
# ingress 配置示例
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: gpu-platform-ingress
  annotations:
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    nginx.ingress.kubernetes.io/proxy-body-size: "50m"
spec:
  tls:
    - hosts:
        - gpu-platform.example.com
      secretName: gpu-platform-tls
  rules:
    - host: gpu-platform.example.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: gpu-platform
                port:
                  number: 8080
```

### 2. 配置监控

```bash
# 安装 Prometheus Operator
helm install prometheus-stack prometheus-community/kube-prometheus-stack \
  --namespace monitoring \
  --set grafana.adminPassword=your_grafana_password

# 安装 GPU 监控
helm install gpu-monitoring nvidia/gpu-monitoring \
  --namespace monitoring
```

### 3. 配置告警通知

```yaml
# alertmanager-config.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: alertmanager-config
  namespace: monitoring
data:
  alertmanager.yml: |
    route:
      group_by: ['alertname']
      receiver: 'default-receiver'
    receivers:
      - name: 'default-receiver'
        email_configs:
          to: admin@example.com
          send_resolved: true
        webhook_configs:
          url: http://your-webhook-endpoint
```

---

## 📡 API 文档

### 认证

```bash
# 注册
POST /api/v1/auth/register
Content-Type: application/json
{
  "email": "user@example.com",
  "password": "SecurePassword123!",
  "name": "Test User"
}

# 登录
POST /api/v1/auth/login
{
  "email": "user@example.com",
  "password": "SecurePassword123!"
}

# 响应
{
  "accessToken": "eyJ...",
  "refreshToken": "eyJ...",
  "expiresIn": 28800
}
```

### 容器管理

```bash
# 创建容器
POST /api/v1/containers
Authorization: Bearer <token>
{
  "name": "pytorch-training",
  "resources": {
    "gpuCount": 1,
    "gpuModel": "a100",
    "cpuCores": 8,
    "memoryMb": 32768,
    "storageGb": 200,
    "image": "nvidia/cuda:12.1-runtime-ubuntu22.04"
  }
}

# 响应
{
  "instance": {
    "id": "uuid-...",
    "name": "pytorch-training",
    "status": "running",
    "gpuServer": "gpu-node-1",
    "webUrl": "http://gpu-node-1:8888",
    "sshInfo": {
      "host": "gpu-node-1",
      "port": 2222,
      "user": "root"
    }
  }
}

# 容器操作
POST /api/v1/containers/{id}/start  # 启动
POST /api/v1/containers/{id}/stop   # 停止
DELETE /api/v1/containers/{id}     # 删除
GET /api/v1/containers/{id}/logs   # 查看日志
```

---

## 📁 项目结构

```
gpu-platform/
├── cmd/                          # 程序入口
│   └── api-server/               # API 服务入口
├── internal/                      # 内部包
│   ├── auth/                      # 认证服务
│   ├── config/                    # 配置管理
│   ├── handlers/                   # HTTP 处理器
│   ├── k8s/                      # Kubernetes 客户端
│   ├── middleware/                 # 中间件
│   ├── models/                    # 数据模型
│   ├── monitor/                   # 监控服务
│   ├── notification/              # 通知服务
│   ├── repository/               # 数据访问层
│   ├── scheduler/                  # 资源调度器
│   ├── services/                  # 业务逻辑层
│   └── storage/                   # 存储服务
├── frontend/                     # React 前端
│   ├── src/
│   │   ├── components/          # 公共组件
│   │   ├── pages/                # 页面组件
│   │   ├── services/            # API 服务
│   │   └── store/               # 状态管理
│   └── package.json
├── deployments/                  # 部署配置
│   ├── k8s/                     # K8s 配置
│   └── helm/                    # Helm Chart
├── docker-compose.yml            # Docker Compose
├── config.yaml                   # 配置文件
└── Makefile                    # 构建脚本
```

---

## ❓ 常见问题

### Q1: GPU 节点不显示 GPU 数量？

```bash
# 检查 NFD 状态
kubectl get nfd cr

# 检查设备插件日志
kubectl logs -n gpu-operator -l app=nvidia-device-plugin

# 常见原因：
# 1. NVIDIA 驱动未安装
# 2. 驱动版本与 CUDA 版本不兼容
# 3. 节点标签未正确设置
```

### Q2: Pod 无法调度到 GPU 节点？

```bash
# 检查节点污点
kubectl describe node <gpu-node-name> | grep Taints

# 添加容忍
kubectl taint nodes <gpu-node-name> nvidia.com/gpu=:NoSchedule-
```

### Q3: GPU 内存不足？

```yaml
# 调整 Pod 资源限制
resources:
  limits:
    nvidia.com/gpu: 1
    memory: 64Gi
  requests:
    nvidia.com/gpu: 1
    memory: 32Gi
```

### Q4: 如何升级 GPU Operator？

```bash
# 查看当前版本
helm list -n gpu-operator

# 升级到新版本
helm upgrade gpu-operator nvidia/gpu-operator \
  --namespace gpu-operator \
  --version 24.0.0
```

### Q5: 监控数据不显示？

```bash
# 检查 Prometheus 配置
kubectl get prometheus -n monitoring

# 检查 ServiceMonitor
kubectl get servicemonitor -n monitoring

# 检查 GPU metrics 端点
kubectl exec -it <prometheus-pod> -n monitoring -- wget -qO- http://gpu-exporter:9400/metrics
```

---

## 🔗 参考链接

- [NVIDIA GPU Operator 文档](https://docs.nvidia.com/datacenter/cloud-native/gpu-operator/)
- [NVIDIA Device Plugin](https://github.com/NVIDIA/k8s-device-plugin)
- [Kubernetes 官方文档](https://kubernetes.io/docs/home/)
- [Helm 官方文档](https://helm.sh/docs/)
- [Prometheus 监控](https://prometheus.io/docs/introduction/overview/)

---

## 📄 许可证

本项目基于 MIT 许可证开源。

---

<div align="center">
Made with ❤️ by GPU Platform Team
</div>
