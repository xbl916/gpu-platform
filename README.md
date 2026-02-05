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

## 📋 目录

- [快速开始](#快速开始)
- [连接到 Kubernetes 集群](#连接到-kubernetes-集群)
- [安装 NVIDIA GPU Operator](#安装-nvidia-gpu-operator)
- [部署 GPU Platform](#部署-gpu-platform)
- [本地开发](#本地开发)
- [生产环境配置](#生产环境配置)
- [常见问题](#常见问题)

---

## 🚀 快速开始

### 环境要求

| 组件 | 要求 |
|------|------|
| **Kubernetes** | 1.20+ |
| **kubectl** | 与集群版本兼容 |
| **Helm** | 3.0+ |
| **GPU 节点** | 至少 1 个带有 NVIDIA GPU 的节点 |
| **存储** | PostgreSQL 14+, Redis 7+, MinIO/S3 |

---

## 🔗 连接到 Kubernetes 集群

在使用 GPU Platform 之前，你需要先配置好 kubectl 以连接到你的 Kubernetes 集群。

### 方式一：云服务（阿里云、AWS、GCP）

#### 阿里云 ACK

```bash
# 1. 安装 kubectl 和 ack 插件
curl -LO https://aliyuncli.github.io/aliyun-cli-linux/latest/aliyun-cli-linux/latestPackages/aliyun-cli_3.0.196_linux_amd64.tar.gz
tar -xzf aliyun-cli_3.0.196_linux_amd64.tar.gz
sudo mv acs /usr/local/bin/

# 2. 登录阿里云
aliyun configure \
  --mode StsToken \
  --profile default

# 3. 获取 kubeconfig
aliyuncs ack get-kubeconfig --clusterId <your-cluster-id> > ~/.kube/config

# 4. 验证连接
kubectl get nodes
```

#### AWS EKS

```bash
# 1. 安装 eksctl
curl --silent --location "https://github.com/weaveworks/eksctl/releases/latest/download/eksctl_$(uname -s)_amd64.tar.gz" | tar xz -C /tmp
sudo mv /tmp/eksctl /usr/local/bin/

# 2. 配置 AWS 凭证
aws configure
# 或使用 IAM Role

# 3. 更新 kubeconfig
aws eks update-kubeconfig --region <region> --name <cluster-name>

# 4. 验证连接
kubectl get nodes
```

#### GCP GKE

```bash
# 1. 安装 gcloud CLI
curl https://sdk.cloud.google.com | bash
exec -l $SHELL

# 2. 登录
gcloud auth login
gcloud container clusters get-credentials <cluster-name> --zone <zone>

# 3. 验证连接
kubectl get nodes
```

### 方式二：自建集群（k3s、kubespray）

#### k3s（轻量级推荐）

```bash
# 在 master 节点安装
curl -sfL https://get.k3s.io | INSTALL_K3S_VERSION=v1.28.4+k3s1 sh -

# 获取 kubeconfig
sudo cat /etc/rancher/k3s/k3s.yaml > ~/.kube/config

# 在其他节点加入集群
curl -sfL https://get.k3s.io | K3S_URL=https://<master-ip>:6443 K3S_TOKEN=<node-token> sh -
```

#### kubespray（生产级）

```bash
# 克隆仓库
git clone https://github.com/kubernetes-sigs/kubespray.git
cd kubespray

# 配置 inventory
cp -r inventory/sample inventory/mycluster
vim inventory/mycluster/hosts.yaml

# 配置 SSH 密钥
ssh-copy-id user@<master-ip>

# 部署集群
ansible-playbook -i inventory/mycluster/hosts.yaml cluster.yml -b -v
```

### 方式三：minikube（仅测试用）

```bash
# 安装 minikube
curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64
sudo install minikube-linux-amd64 /usr/local/bin/minikube

# 启动集群（需要 GPU 支持）
minikube start --driver=nvidia-docker2

# 或不使用 GPU
minikube start --cpus=4 --memory=8192
```

### 验证集群连接

```bash
# 1. 检查节点
kubectl get nodes -o wide

# 预期输出示例:
# NAME           STATUS   ROLES    AGE   VERSION   INTERNAL-IP    EXTERNAL-IP
# gpu-node-1     Ready    master  10d   v1.28.4   10.0.0.11    <none>

# 2. 检查集群信息
kubectl cluster-info

# 预期输出:
# Kubernetes control plane is running at https://<api-server>:6443

# 3. 检查集群版本
kubectl version --client
kubectl get nodes -o jsonpath='{.items[0].status.nodeInfo.kubeletVersion}'

# 4. 创建 namespace
kubectl create namespace gpu-platform

# 5. 检查 GPU 节点（如果有 GPU）
kubectl get nodes -l accelerator=nvidia
```

### 配置 kubeconfig（多集群场景）

```bash
# 查看当前配置的集群
kubectl config get-contexts

# 添加新集群
kubectl config set-cluster new-cluster \
  --server=https://<api-server>:6443 \
  --certificate-authority=/path/to/ca.crt

# 设置凭证
kubectl config set-credentials admin \
  --token=<bearer-token>

# 或使用 client-certificate
kubectl config set-cluster new-cluster \
  --server=https://<api-server>:6443 \
  --certificate-authority=/path/to/ca.crt
kubectl config set-credentials admin \
  --client-certificate=/path/to/client.crt \
  --client-key=/path/to/client.key

# 切换集群
kubectl config use-context new-cluster

# 合并 kubeconfig（从其他集群复制）
cat ~/.kube/config >> /shared/kubeconfig
KUBECONFIG=/shared/kubeconfig kubectl config view --flatten > /merged-kubeconfig
```

---

## 🖥️ 安装 NVIDIA GPU Operator

GPU Operator 是必需的组件，它负责在 GPU 节点上安装和管理 NVIDIA 驱动、容器工具包等。

### 前置检查

```bash
# 1. 确认你有 GPU 节点
kubectl get nodes -o wide | grep -i gpu

# 或检查标签
kubectl get nodes --show-labels | grep -i nvidia

# 2. 如果没有 GPU 标签，需要安装 Node Feature Discovery (NFD)
kubectl apply -f https://raw.githubusercontent.com/kubernetes-sigs/node-feature-discovery/v0.11.0/deployment/manifests/nfd.yaml

# 3. 验证 NFD
kubectl get pods -n node-feature-discovery
```

### 安装 GPU Operator（两种方式）

#### 方式一：Helm 安装（推荐）

```bash
# 1. 添加 NVIDIA Helm 仓库
helm repo add nvidia https://nvidia.github.io/gpu-operator
helm repo update

# 2. 创建 namespace
kubectl create namespace gpu-operator

# 3. 安装 GPU Operator
helm install gpu-operator nvidia/gpu-operator \
  --namespace gpu-operator \
  --version 24.3.0 \
  --set driver.enabled=true \     # 如果节点没有驱动，设为 true
  --set toolkit.enabled=true \
  --set devicePlugin.enabled=true \
  --set dcgmExporter.enabled=true

# 4. 监控安装进度
kubectl get pods -n gpu-operator -w

# 等待所有 Pod 运行正常
# NAME                                     READY   STATUS      RESTARTS   AGE
# gpu-operator-6f8d5b7b9c-xkq5w            1/1     Running     0          2m
# nvidia-container-toolkit-daemonset-xxxxx     1/1     Running     0          1m
# nvidia-device-plugin-daemonset-xxxxx        1/1     Running     0          1m
# nvidia-cuda-exporter-xxxxx               1/1     Running     0          1m
```

#### 方式二：YAML 直接安装

```bash
# 直接应用所有资源
kubectl apply -f https://raw.githubusercontent.com/NVIDIA/gpu-operator/v24.3.0/deployments/gpu-operator/manifests/gpu-operator.yaml
```

### 验证 GPU Operator

```bash
# 1. 检查 GPU 节点标签
kubectl get nodes -l nvidia.com/gpu.product

# 预期输出:
# NAME           LABELS                                          Taints
# gpu-node-1    nvidia.com/gpu.product=A100-SXM4-40GB   <none>

# 2. 检查 GPU 资源
kubectl get nodes -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.status.allocatable.nvidia\.com/gpu}{"\n"}'

# 预期输出:
# gpu-node-1    2

# 3. 检查 Device Plugin
kubectl get ds nvidia-device-plugin-daemonset -n gpu-operator

# 4. 验证 NVIDIA 设备插件可以列出 GPU
kubectl get nodes -o jsonpath='{range .items[*]}{.status.allocatable.nvidia\.com/gpu}{"\n"}'
```

### 常见问题解决

```bash
# 问题 1: Pod 一直处于 Pending
kubectl describe pod <pod-name> -n gpu-operator

# 检查是否有节点选择器/污点
kubectl get nodes
kubectl describe node <gpu-node> | grep -A5 Taints

# 如果有污点，需要容忍
kubectl taint nodes <gpu-node> nvidia.com/gpu=:NoSchedule-

# 问题 2: GPU 驱动版本不兼容
# 检查节点驱动版本
kubectl exec -n gpu-operator nvidia-driver-daemonset-<xxx> -- nvidia-smi | head -5

# 问题 3: 设备插件不工作
# 重启设备插件
kubectl rollout restart ds nvidia-device-plugin-daemonset -n gpu-operator
```

---

## 📦 部署 GPU Platform

### 1. 创建配置文件

创建 `gpu-platform-values.yaml`:

```yaml
# gpu-platform-values.yaml

# 数据库连接
database:
  host: "your-postgres-host"
  port: 5432
  username: "gpu_admin"
  password: "your-secure-password"
  name: "gpu_platform"

# Redis 连接
redis:
  host: "your-redis-host"
  port: 6379

# MinIO 配置
storage:
  endpoint: "your-minio-host:9000"
  accessKey: "minioadmin"
  secretKey: "your-minio-password"

# Kubernetes 配置
kubernetes:
  kubeconfig: ""  # 留空使用 in-cluster 配置
  namespace: "gpu-platform"
  gpuPoolLabel: "gpu-pool"
  instanceLabel: "gpu-instance"

# GPU 默认配置
gpu:
  defaultGPUModel: "a100"
  maxGPUPerInstance: 8
  defaultCPUCores: 4
  maxCPUCores: 64
  defaultMemoryMB: 16384
  maxMemoryMB: 262144

# 容器默认配置
container:
  defaultImage: "nvidia/cuda:12.1-runtime-ubuntu22.04"
  maxInstancesPerUser: 10
  maxRunningHours: 24
  workspaceSizeGB: 100
  dataSizeGB: 200

# 监控配置
monitoring:
  enabled: true
  metricsPort: 9090
  alertThresholdGPUTemp: 85

# 镜像仓库（可选）
registry:
  address: "your-registry.example.com"
  username: "registry-user"
  password: "registry-password"
```

### 2. 安装 GPU Platform

```bash
# 方式一：Helm 安装（推荐）
helm install gpu-platform ./deployments/helm/gpu-platform \
  --namespace gpu-platform \
  --create-namespace \
  --values gpu-platform-values.yaml

# 方式二：Kustomize 安装
kubectl apply -k deployments/k8s/overlays/production
```

### 3. 验证部署

```bash
# 1. 检查 Pod 状态
kubectl get pods -n gpu-platform

# 2. 检查服务
kubectl get svc -n gpu-platform

# 3. 查看日志
kubectl logs -n gpu-platform -l app=gpu-platform-api --tail=100

# 4. 检查 API 健康
kubectl port-forward -n gpu-platform svc/gpu-platform 8080:8080 &
curl http://localhost:8080/health
```

### 4. 访问 Web 界面

```bash
# 方法一：Port Forward（临时访问）
kubectl port-forward -n gpu-platform svc/gpu-platform 3000:80 &

# 打开浏览器访问 http://localhost:3000

# 方法二：Ingress（生产访问）
kubectl apply -f - <<EOF
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: gpu-platform-ingress
  namespace: gpu-platform
spec:
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
              number: 80
EOF
```

---

## 💻 本地开发

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

# 配置 kubectl 连接到集群
# 使用 kubeconfig
export KUBECONFIG=/path/to/your/kubeconfig

# 或合并到默认配置
cat /path/to/your/kubeconfig >> ~/.kube/config
```

### 2. 启动依赖服务

```bash
# PostgreSQL
docker run -d \
  --name postgres \
  -e POSTGRES_USER=gpu_admin \
  -e POSTGRES_PASSWORD=password \
  -e POSTGRES_DB=gpu_platform \
  -p 5432:5432 \
  postgres:15

# Redis
docker run -d \
  --name redis \
  -p 6379:6379 \
  redis:7-alpine \
  redis-server --requirepass password

# MinIO
docker run -d \
  --name minio \
  -p 9000:9000 \
  -p 9001:9001 \
  -e MINIO_ROOT_USER=minioadmin \
  -e MINIO_ROOT_PASSWORD=miniopassword \
  quay.io/minio/minio server /data --console-address ":9001"
```

### 3. 配置和运行

```bash
# 复制配置模板
cp config.yaml.example config.yaml

# 编辑配置
vim config.yaml

# 运行后端
cd cmd/api-server
go run main.go

# 新终端运行前端
cd frontend
npm install
npm run dev
```

---

## 🏭 生产环境配置

### TLS/HTTPS 配置

```yaml
# values.yaml
ingress:
  enabled: true
  className: nginx
  annotations:
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    nginx.ingress.kubernetes.io/proxy-body-size: "50m"
  hosts:
    - host: gpu-platform.example.com
      paths:
        - path: /
          pathType: Prefix
  tls:
    - secretName: gpu-platform-tls
      hosts:
        - gpu-platform.example.com
```

### 高可用配置

```yaml
# values.yaml
replicaCount: 3

autoscaling:
  enabled: true
  minReplicas: 3
  maxReplicas: 10

resources:
  limits:
    cpu: "1000m"
    memory: "1Gi"
  requests:
    cpu: "500m"
    memory: "512Mi"
```

### 资源限制配置

```yaml
# values.yaml
resources:
  limits:
    nvidia.com/gpu: 1
    cpu: "2000m"
    memory: "2Gi"
  requests:
    nvidia.com/gpu: 1
    cpu: "1000m"
    memory: "1Gi"
```

---

## ❓ 常见问题

### Q1: 找不到 GPU 节点？

```bash
# 检查节点标签
kubectl get nodes --show-labels

# 如果没有 nvidia 标签，重新安装 NFD
kubectl apply -f https://raw.githubusercontent.com/kubernetes-sigs/node-feature-discovery/v0.14.0/deployment/manifests/nfd.yaml

# 手动标签节点
kubectl label nodes <gpu-node> nvidia.com/gpu.product=A100-SXM4-40GB --overwrite
```

### Q2: GPU Operator 安装失败？

```bash
# 检查 operator 日志
kubectl logs -n gpu-operator deployment/gpu-operator

# 检查特定组件日志
kubectl logs -n gpu-operator -l app=nvidia-container-toolkit-daemonset

# 完全卸载后重新安装
helm uninstall gpu-operator -n gpu-operator
kubectl delete crd $(kubectl get crd | grep nvidia | awk '{print $1}')
helm install gpu-operator nvidia/gpu-operator -n gpu-operator
```

### Q3: GPU 内存不可见？

```bash
# 检查设备插件状态
kubectl get ds nvidia-device-plugin-daemonset -n gpu-operator -o yaml

# 重启设备插件
kubectl rollout restart ds nvidia-device-plugin-daemonset -n gpu-operator

# 检查节点资源
kubectl describe node <gpu-node> | grep -A10 "Allocated resources"
```

---

## 📚 参考链接

- [Kubernetes 官方文档](https://kubernetes.io/docs/home/)
- [NVIDIA GPU Operator 文档](https://docs.nvidia.com/datacenter/cloud-native/gpu-operator/)
- [kubectl 安装配置](https://kubernetes.io/docs/reference/kubectl/)
- [Helm 官方文档](https://helm.sh/docs/)
- [阿里云 ACK 文档](https://help.aliyun.com/document_detail/86589.html)
- [AWS EKS 文档](https://docs.aws.amazon.com/eks/latest/userguide/what-is-eks.html)
- [GCP GKE 文档](https://cloud.google.com/kubernetes-engine/docs)

---

## 📄 许可证

MIT License

---

<div align="center">
Made with ❤️ by GPU Platform Team
</div>
