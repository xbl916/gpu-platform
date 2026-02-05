# GPU Platform - GPU 容器化服务管理平台

<div align="center">

![License](https://img.shields.io/badge/License-MIT-blue.svg)
![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat-square&logo=go)
![React](https://img.shields.io/badge/React-18.2-61DAFB?style=flat-square&logo=react)
![Kubernetes](https://img.shields.io/badge/Kubernetes-1.28-326CE5?style=flat-square&logo=kubernetes)
![Docker](https://img.shields.io/badge/Docker-24.0+-2496ED?style=flat-square&logo=docker)

一个类似 AutoDL 的 GPU 容器化服务管理平台，支持 GPU 资源池管理、容器实例生命周期管理、容器模板定制等功能。

[English](README.en.md) | 简体中文

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
| **项目管理** | 多租户隔离、项目成员管理、资源共享 |

### 🏗️ 技术架构

```
┌─────────────────────────────────────────────────────────────┐
│                      前端 (React + TypeScript)               │
│                   Ant Design Pro 组件库                      │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    API 网关 (Gin + JWT)                    │
│                 认证授权 │ 限流 │ 日志                    │
└─────────────────────────────────────────────────────────────┘
                              │
        ┌─────────────────────┼─────────────────────┐
        ▼                     ▼                     ▼
┌─────────────┐      ┌─────────────┐      ┌─────────────┐
│  用户服务   │      │  容器服务   │      │ 资源服务   │
└─────────────┘      └─────────────┘      └─────────────┘
        │                     │                     │
        ▼                     ▼                     ▼
┌─────────────────────────────────────────────────────────────┐
│                    数据层                                     │
│  PostgreSQL │ Redis │ MinIO │ Prometheus            │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│               Kubernetes 集群                               │
│     GPU 节点池 │ 容器编排 │ 存储卷 │ 网络                     │
└─────────────────────────────────────────────────────────────┘
```

## 📋 目录

- [快速开始](#快速开始)
- [环境要求](#环境要求)
- [部署前准备](#部署前准备)
- [数据库初始化](#数据库初始化)
- [配置说明](#配置说明)
- [启动服务](#启动服务)
- [开发环境](#开发环境)
- [生产部署](#生产部署)
- [常见问题](#常见问题)

---

## 🚀 快速开始

### 环境要求

| 组件 | 最低版本 | 推荐版本 |
|------|---------|---------|
| Kubernetes | 1.20 | 1.28+ |
| Go | 1.20 | 1.21+ |
| Node.js | 16 | 20 LTS |
| PostgreSQL | 12 | 15+ |
| Redis | 6 | 7+ |
| NVIDIA GPU Driver | 450.x | 535.x+ |

---

## 🔧 部署前准备

### 1. 准备 Kubernetes 集群

```bash
# 检查集群连接
kubectl cluster-info

# 检查节点状态
kubectl get nodes

# 检查 GPU 节点（如果有 GPU）
kubectl get nodes -l accelerator=nvidia

# 创建专用 namespace
kubectl create namespace gpu-platform
```

### 2. 安装 NVIDIA GPU Operator

GPU Operator 负责在 GPU 节点上安装和管理 NVIDIA 驱动、容器工具包等。

```bash
# 添加 NVIDIA Helm 仓库
helm repo add nvidia https://nvidia.github.io/gpu-operator
helm repo update

# 创建 namespace
kubectl create namespace gpu-operator-ns

# 安装 GPU Operator
helm install gpu-operator nvidia/gpu-operator \
  --namespace gpu-operator-ns \
  --version 24.3.0 \
  --set driver.enabled=true \
  --set toolkit.enabled=true \
  --set devicePlugin.enabled=true \
  --set dcgmExporter.enabled=true

# 监控安装进度
kubectl get pods -n gpu-operator-ns -w

# 验证安装
kubectl get nodes -l nvidia.com/gpu.product
```

### 3. 准备存储后端

```bash
# 方式一：使用 MinIO（推荐用于测试）
docker run -d \
  --name minio \
  -p 9000:9000 \
  -p 9001:9001 \
  -e MINIO_ROOT_USER=admin \
  -e MINIO_ROOT_PASSWORD=password123 \
  quay.io/minio/minio server /data --console-address ":9001"

# 方式二：使用 S3 兼容存储（如阿里云 OSS、AWS S3）
# 直接在配置中指定 endpoint 即可
```

---

## 📦 数据库初始化

### 1. 安装 PostgreSQL

```bash
# Docker 方式（快速启动）
docker run -d \
  --name postgres \
  -e POSTGRES_USER=gpu_admin \
  -e POSTGRES_PASSWORD=your_password \
  -e POSTGRES_DB=gpu_platform \
  -p 5432:5432 \
  postgres:15

# 或者使用云数据库（RDS、Cloud SQL 等）
```

### 2. 运行数据库迁移

```bash
# 进入项目根目录
cd /path/to/gpu-platform

# 编译迁移工具
go build -o bin/migrate ./scripts/migrate.go

# 运行迁移（up-升级，down-回滚，seed-初始化测试数据）
./bin/migrate -config=config.yaml -action=up
./bin/migrate -config=config.yaml -action=seed
```

### 迁移脚本说明

| 参数 | 说明 |
|------|------|
| `-action=up` | 执行所有未执行的迁移 |
| `-action=down` | 回滚最后一个迁移 |
| `-action=status` | 查看迁移状态 |
| `-action=seed` | 插入测试数据 |

---

## ⚙️ 配置说明

### 配置文件结构

```yaml
# config.yaml

# 应用配置
app:
  host: "0.0.0.0"
  port: 8080
  name: "GPU Platform"
  domain: "gpu-platform.local"

# 数据库配置
database:
  host: "localhost"
  port: 5432
  username: "gpu_admin"
  password: "your_password"
  name: "gpu_platform"
  sslmode: "disable"
  max_open_conns: 25
  max_idle_conns: 5
  conn_max_lifetime: 300  # 秒

# Redis 配置
redis:
  host: "localhost"
  port: 6379
  password: ""
  db: 0

# 存储配置（MinIO/S3）
storage:
  endpoint: "localhost:9000"
  access_key: "minioadmin"
  secret_key: "miniopassword"
  use_ssl: false
  bucket_prefix: "gpu-platform"

# Kubernetes 配置
kubernetes:
  kubeconfig: ""  # 留空使用 in-cluster 配置
  namespace: "gpu-platform"
  gpu_pool_label: "gpu-pool"
  instance_label: "gpu-instance"

# JWT 配置
jwt:
  secret: "your-secret-key-change-in-production"
  access_token_expire: 3600      # 秒（1小时）
  refresh_token_expire: 604800    # 秒（7天）

# GPU 默认配置
gpu:
  default_gpu_model: "a100"
  max_gpu_per_instance: 8
  default_cpu_cores: 4
  max_cpu_cores: 64
  default_memory_mb: 16384    # 16GB
  max_memory_mb: 262144       # 256GB
  default_storage_gb: 100
  max_storage_gb: 2000

# 容器默认配置
container:
  default_image: "nvidia/cuda:12.1-runtime-ubuntu22.04"
  max_instances_per_user: 10
  max_running_hours: 24
  workspace_size_gb: 50
  data_size_gb: 200

# 监控配置
monitoring:
  enabled: true
  metrics_port: 9090
  alert_threshold_gpu_temp: 85      # 摄氏度
  alert_threshold_gpu_memory: 90     # 百分比

# 通知配置
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

# 会话配置
session:
  max_session_duration: 86400      # 秒（24小时）
  max_sessions_per_user: 5
```

---

## ▶️ 启动服务

### 方式一：直接运行（开发环境）

```bash
# 1. 确保配置文件正确
cp config.yaml.example config.yaml
vim config.yaml

# 2. 编译并启动 API 服务
cd cmd/api-server
go run main.go

# 3. 新终端启动前端
cd frontend
npm install
npm run dev
```

### 方式二：使用 Makefile

```bash
# 查看所有可用命令
make help

# 运行开发环境
make dev

# 构建生产镜像
make build

# 运行测试
make test
```

### 方式三：Docker Compose（推荐用于快速测试）

```bash
# 启动所有服务
docker-compose up -d

# 查看日志
docker-compose logs -f

# 停止所有服务
docker-compose down
```

### 验证服务

```bash
# API 健康检查
curl http://localhost:8080/health

# 预期输出:
# {"status":"healthy"}

# 访问 Web 界面
# 打开浏览器访问 http://localhost:5173
```

---

## 💻 开发环境

### 1. 环境准备

```bash
# 安装 Go
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin

# 安装 Node.js
curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
sudo apt-get install -y nodejs

# 安装 kubectl
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
chmod +x kubectl
sudo mv kubectl /usr/local/bin/

# 配置 kubectl
export KUBECONFIG=/path/to/your/kubeconfig
```

### 2. 启动依赖服务

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

### 3. 开发工作流

```bash
# 1. 克隆项目
git clone https://github.com/xbl916/gpu-platform.git
cd gpu-platform

# 2. 创建分支（遵循规范）
git checkout -b 260205-feat-add-new-feature

# 3. 开发完成后提交
git add -A
git commit -m "feat: 添加新功能描述"
git push origin 260205-feat-add-new-feature
```

---

## 🏭 生产部署

### 1. 准备生产环境

```bash
# 1. 准备 Kubernetes 集群
# 确保已安装 GPU Operator，参考前述步骤

# 2. 创建生产命名空间
kubectl create namespace gpu-platform-prod

# 3. 创建 Secret
kubectl create secret generic gpu-platform-secrets \
  --from-literal=database-password="your-secure-password" \
  --from-literal=jwt-secret="your-super-secure-jwt-secret" \
  --from-literal=redis-password="your-redis-password" \
  --from-literal=minio-access-key="your-access-key" \
  --from-literal=minio-secret-key="your-secret-key" \
  -n gpu-platform-prod
```

### 2. 配置生产参数

```yaml
# gpu-platform-values.yaml

# 副本配置
replicaCount: 3

# 资源限制
resources:
  limits:
    cpu: "1000m"
    memory: "1Gi"
  requests:
    cpu: "500m"
    memory: "512Mi"

# Ingress 配置
ingress:
  enabled: true
  className: nginx
  hosts:
    - host: gpu-platform.example.com
      paths:
        - path: /
          pathType: Prefix
  tls:
    - secretName: gpu-platform-tls
      hosts:
        - gpu-platform.example.com

# 高可用配置
autoscaling:
  enabled: true
  minReplicas: 3
  maxReplicas: 10
```

### 3. 部署到 Kubernetes

```bash
# 方式一：使用 Helm
helm install gpu-platform ./deployments/helm/gpu-platform \
  --namespace gpu-platform-prod \
  --create-namespace \
  --values gpu-platform-values.yaml

# 方式二：使用 Kustomize
kubectl apply -k deployments/k8s/overlays/production

# 验证部署
kubectl get pods -n gpu-platform-prod
kubectl logs -n gpu-platform-prod -l app=gpu-platform --tail=100
```

### 4. 配置 HTTPS

```bash
# 使用 cert-manager 自动管理证书
helm install cert-manager jetstack/cert-manager \
  --namespace cert-manager \
  --create-namespace \
  --version v1.14.0

# 创建 ClusterIssuer
kubectl apply -f - <<EOF
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: your-email@example.com
    privateKeySecretRef:
      name: letsencrypt-prod
    solvers:
    - http01:
        ingress:
          class: nginx
EOF

# 更新 Ingress 配置启用 HTTPS
# 参考前述 Ingress 配置中的 tls 部分
```

---

## ❓ 常见问题

### Q1: 找不到 GPU 节点？

```bash
# 检查节点标签
kubectl get nodes --show-labels | grep nvidia

# 如果没有标签，手动添加
kubectl label nodes <gpu-node-name> nvidia.com/gpu.product=A100-SXM4-40GB --overwrite

# 验证 GPU 资源
kubectl get nodes -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.status.allocatable.nvidia\.com/gpu}{"\n"}'
```

### Q2: GPU Operator 安装失败？

```bash
# 检查 Operator 日志
kubectl logs -n gpu-operator-ns deployment/gpu-operator

# 完全卸载后重新安装
helm uninstall gpu-operator -n gpu-operator-ns
kubectl delete crd $(kubectl get crd | grep nvidia | awk '{print $1}')
helm install gpu-operator nvidia/gpu-operator -n gpu-operator-ns
```

### Q3: GPU 内存不可见？

```bash
# 检查设备插件状态
kubectl get ds nvidia-device-plugin-daemonset -n gpu-operator-ns

# 重启设备插件
kubectl rollout restart ds nvidia-device-plugin-daemonset -n gpu-operator-ns

# 检查节点资源分配
kubectl describe node <gpu-node> | grep -A10 "Allocated resources"
```

### Q4: 数据库连接失败？

```bash
# 测试数据库连接
kubectl exec -n gpu-platform-prod <api-pod-name> -- \
  PGPASSWORD=your_password psql -h postgres-host -U gpu_admin -d gpu_platform -c "\dt"

# 检查连接字符串配置
# 确保 config.yaml 中的数据库配置正确
```

### Q5: 前端无法连接 API？

```bash
# 检查 API 地址配置
# frontend/.env 文件中
VITE_API_URL=http://localhost:8080/api/v1

# 生产环境改为实际的 API 地址
VITE_API_URL=https://gpu-platform.example.com/api/v1

# 重启前端开发服务器
npm run dev
```

---

## 📚 参考文档

| 文档 | 链接 |
|------|------|
| Kubernetes 官方文档 | https://kubernetes.io/docs/home/ |
| NVIDIA GPU Operator | https://docs.nvidia.com/datacenter/cloud-native/gpu-operator/ |
| Go 语言文档 | https://go.dev/doc/ |
| React 官方文档 | https://react.dev/ |
| Gin 框架文档 | https://gin-gonic.com/docs/ |
| Ant Design Pro | https://procomponents.ant.design/ |

---

## 📄 许可证

MIT License

---

<div align="center">
Made with ❤️ by GPU Platform Team
</div>
