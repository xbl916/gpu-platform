# GPU 容器化服务管理平台 - 使用指南

## 1. 环境准备

### 1.1 系统要求

- **操作系统**: Linux (Ubuntu 18.04+/CentOS 7+), macOS
- **Go 版本**: 1.21+
- **数据库**: PostgreSQL 14+
- **Kubernetes**: 1.28+ (可选，用于容器编排)
- **GPU 环境**: NVIDIA Driver 450.x+

### 1.2 安装依赖

```bash
# 安装 Go 1.21+
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin

# 验证安装
go version
```

### 1.3 安装 PostgreSQL

```bash
# Ubuntu/Debian
sudo apt update
sudo apt install -y postgresql postgresql-contrib

# 启动服务
sudo systemctl start postgresql
sudo systemctl enable postgresql

# 创建数据库
sudo -u postgres psql
CREATE DATABASE gpu_platform;
CREATE USER gpu_admin WITH PASSWORD 'your_password';
GRANT ALL PRIVILEGES ON DATABASE gpu_platform TO gpu_admin;
\q
```

### 1.4 安装 kubectl (可选)

```bash
# Linux
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
chmod +x kubectl
sudo mv kubectl /usr/local/bin/
```

## 2. 配置说明

### 2.1 修改配置文件

编辑 `config.yaml`:

```yaml
# 应用配置
app:
  name: "gpu-platform"
  host: "0.0.0.0"
  port: 8080
  env: "development"  # 改为 production

# JWT 配置
jwt:
  secret: "your-256-bit-secret-key-change-this"
  access_token_expire: 28800  # 8小时

# 数据库配置
database:
  host: "localhost"
  port: 5432
  username: "gpu_admin"
  password: "your_password"
  name: "gpu_platform"

# Redis 配置
redis:
  host: "localhost"
  port: 6379

# GPU 配置
gpu:
  default_gpu_model: "a100"
  default_gpu_count: 1
  max_gpu_per_instance: 8

# Kubernetes 配置 (可选)
kubernetes:
  kubeconfig: ""  # 留空使用 in-cluster 配置
  namespace: "gpu-platform"
```

### 2.2 设置环境变量

```bash
# 创建环境变量文件
cat > .env <<EOF
export DB_HOST=localhost
export DB_PASSWORD=your_password
export REDIS_HOST=localhost
export JWT_SECRET_KEY=your-256-bit-secret-key-change-this
export MINIO_ENDPOINT=localhost:9000
export SMTP_HOST=smtp.example.com
EOF

source .env
```

## 3. 编译和运行

### 3.1 编译项目

```bash
# 方式一: 使用 Makefile
make build

# 方式二: 手动编译
go mod download
go build -o bin/api-server ./cmd/api-server
```

### 3.2 运行服务

```bash
# 开发模式运行
make run

# 或直接运行二进制文件
./bin/api-server
```

### 3.3 验证服务

```bash
# 健康检查
curl http://localhost:8080/health

# 预期输出:
# {"status":"healthy"}

# API 文档
curl http://localhost:8080/
```

## 4. API 使用指南

### 4.1 用户注册

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePassword123!",
    "name": "Test User"
  }'
```

**响应示例**:
```json
{
  "user": {
    "id": "uuid-...",
    "email": "user@example.com",
    "name": "Test User",
    "role": "developer",
    "status": "active"
  },
  "accessToken": "eyJhbG...",
  "refreshToken": "eyJhbG...",
  "expiresIn": 28800
}
```

### 4.2 用户登录

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePassword123!"
  }'
```

### 4.3 使用 API_TOKEN 访问受保护接口

```bash
# 设置环境变量
export ACCESS_TOKEN="your-access-token"

# 获取当前用户信息
curl http://localhost:8080/api/v1/users/me \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

### 4.4 查看 GPU 资源可用性

```bash
curl http://localhost:8080/api/v1/resources/availability \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

**响应示例**:
```json
{
  "totalServers": 5,
  "onlineServers": 4,
  "totalGpuCount": 20,
  "availableGpuCount": 12,
  "gpuByModel": {
    "a100": 8,
    "rtx3090": 4
  }
}
```

### 4.5 创建 GPU 容器实例

```bash
curl -X POST http://localhost:8080/api/v1/containers \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "my-gpu-container",
    "resources": {
      "gpuCount": 1,
      "gpuModel": "a100",
      "cpuCores": 4,
      "memoryMb": 16384,
      "storageGb": 100,
      "image": "nvidia/cuda:12.1-runtime-ubuntu22.04",
      "command": "sleep infinity"
    }
  }'
```

**响应示例**:
```json
{
  "instance": {
    "id": "uuid-...",
    "name": "my-gpu-container",
    "status": "pending",
    "resources": {
      "gpuCount": 1,
      "cpuCores": 4,
      "memoryMb": 16384
    },
    "gpuServer": "10.0.0.5"
  },
  "webUrl": "https://10.0.0.5:8888",
  "sshInfo": {
    "host": "10.0.0.5",
    "port": 2222,
    "user": "root"
  }
}
```

### 4.6 管理容器生命周期

```bash
# 查看容器状态
curl http://localhost:8080/api/v1/containers/{id} \
  -H "Authorization: Bearer $ACCESS_TOKEN"

# 启动容器
curl -X POST http://localhost:8080/api/v1/containers/{id}/start \
  -H "Authorization: Bearer $ACCESS_TOKEN"

# 停止容器
curl -X POST http://localhost:8080/api/v1/containers/{id}/stop \
  -H "Authorization: Bearer $ACCESS_TOKEN"

# 删除容器
curl -X DELETE http://localhost:8080/api/v1/containers/{id} \
  -H "Authorization: Bearer $ACCESS_TOKEN"

# 获取容器日志
curl http://localhost:8080/api/v1/containers/{id}/logs \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

### 4.7 容器模板管理

```bash
# 创建模板
curl -X POST http://localhost:8080/api/v1/templates \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "pytorch-training",
    "description": "PyTorch training environment",
    "config": {
      "baseImage": "nvidia/cuda:12.1-runtime-ubuntu22.04",
      "cudaVersion": "12.1",
      "pythonVersion": "3.10",
      "packages": ["python3-dev", "gcc"],
      "pipPackages": ["torch", "torchvision", "tensorboard"]
    }
  }'

# 列出所有模板
curl http://localhost:8080/api/v1/templates \
  -H "Authorization: Bearer $ACCESS_TOKEN"

# 使用模板创建容器
curl -X POST http://localhost:8080/api/v1/containers \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "training-job-1",
    "templateId": "template-uuid"
  }'
```

### 4.8 查看监控仪表盘

```bash
curl http://localhost:8080/api/v1/monitor/dashboard \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

## 5. Kubernetes 部署

### 5.1 创建 Kubernetes 命名空间

```bash
kubectl create namespace gpu-platform
```

### 5.2 部署 GPU Operator (用于管理 GPU 节点)

```bash
# 安装 NVIDIA GPU Operator
kubectl apply -f https://nvidia.github.io/gpu-operator/nvidia-driver.yaml
kubectl rollout status -n gpu-operator-state ds/nvidia-driver-daemonset
```

### 5.3 部署 API 服务

```bash
# 使用 Helm 部署
helm upgrade --install gpu-platform ./deployments/helm/gpu-platform \
  --namespace gpu-platform \
  --set image.tag=latest \
  --set replicaCount=3
```

### 5.4 验证部署

```bash
# 检查 Pod 状态
kubectl get pods -n gpu-platform

# 查看日志
kubectl logs -n gpu-platform -l app=gpu-platform-api
```

## 6. Docker 本地开发

### 6.1 构建镜像

```bash
make docker-build
```

### 6.2 运行容器

```bash
docker run -d \
  --name gpu-platform-api \
  -p 8080:8080 \
  -e DB_HOST=host.docker.internal \
  -e DB_PASSWORD=your_password \
  gpu-platform/api-server:latest
```

## 7. 测试

### 7.1 运行单元测试

```bash
make test
```

### 7.2 运行测试覆盖率

```bash
make test-coverage
```

## 8. 常见问题

### Q1: 数据库连接失败

```bash
# 检查 PostgreSQL 服务状态
sudo systemctl status postgresql

# 检查连接
psql -h localhost -U gpu_admin -d gpu_platform
```

### Q2: GPU 不可用

```bash
# 检查 NVIDIA 驱动
nvidia-smi

# 检查 Kubernetes GPU 节点
kubectl get nodes -l accelerator=nvidia
```

### Q3: JWT token 过期

```bash
# 使用 refresh token 获取新的 access token
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{"refreshToken": "your-refresh-token"}'
```

## 9. API 完整列表

| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| POST | /api/v1/auth/register | 用户注册 | 否 |
| POST | /api/v1/auth/login | 用户登录 | 否 |
| POST | /api/v1/auth/logout | 登出 | 是 |
| POST | /api/v1/auth/refresh | 刷新 Token | 否 |
| GET | /api/v1/users/me | 获取当前用户 | 是 |
| PUT | /api/v1/users/me | 更新用户信息 | 是 |
| GET | /api/v1/resources/gpu-servers | 列出 GPU 服务器 | 是 |
| GET | /api/v1/resources/availability | 资源可用性 | 是 |
| GET | /api/v1/containers | 列出容器 | 是 |
| POST | /api/v1/containers | 创建容器 | 是 |
| GET | /api/v1/containers/:id | 获取容器详情 | 是 |
| DELETE | /api/v1/containers/:id | 删除容器 | 是 |
| POST | /api/v1/containers/:id/start | 启动容器 | 是 |
| POST | /api/v1/containers/:id/stop | 停止容器 | 是 |
| GET | /api/v1/containers/:id/logs | 获取日志 | 是 |
| GET | /api/v1/monitor/dashboard | 监控仪表盘 | 是 |
| GET | /admin/users | 列出用户 (管理员) | 是 |

## 10. 后续步骤

1. **配置 HTTPS/TLS** - 使用 nginx 或 Traefik 配置 SSL
2. **配置监控** - 集成 Prometheus + Grafana
3. **配置告警** - 设置邮件和 Slack 通知
4. **配置日志** - 集成 ELK Stack 或 Loki
5. **备份策略** - 设置数据库和配置备份
