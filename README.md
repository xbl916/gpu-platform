# GPU Platform - GPU 容器化服务管理平台

<div align="center">

![License](https://img.shields.io/badge/License-MIT-blue.svg)
![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat-square&logo=go)
![React](https://img.shields.io/badge/React-18.2-61DAFB?style=flat-square&logo=react)
![Kubernetes](https://img.shields.io/badge/Kubernetes-1.28-326CE5?style=flat-square&logo=kubernetes)

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

### 📊 资源调度

- **多种调度策略**: BinPack（打满）、Spread（分散）、按 GPU 数量、按内存
- **智能分配**: 根据资源可用性和用户配额自动分配
- **实时监控**: GPU 利用率、温度、显存使用实时采集

### 🔐 安全特性

- JWT Token 认证（Access Token + Refresh Token）
- 密码加密存储（bcrypt）
- 会话管理
- API 访问控制

## 🏗️ 技术架构

```
┌─────────────────────────────────────────────────────────────┐
│                        前端 (React + TypeScript)             │
│  Ant Design + Recharts + Zustand                          │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                      API Gateway (Gin)                       │
│                JWT Auth + Rate Limit + CORS                  │
└─────────────────────────────────────────────────────────────┘
                            │
        ┌───────────────────┼───────────────────┐
        ▼                   ▼                   ▼
┌───────────────┐  ┌───────────────┐  ┌───────────────┐
│  User Service │  │ Resource Svc  │  │ Container Svc  │
│  (用户认证)    │  │ (资源管理)    │  │ (容器管理)    │
└───────────────┘  └───────────────┘  └───────────────┘
        │                   │                   │
        └───────────────────┼───────────────────┘
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                   Kubernetes / Docker                         │
│              GPU Pod 管理 + 容器编排                          │
└─────────────────────────────────────────────────────────────┘
                            │
        ┌───────────────────┼───────────────────┐
        ▼                   ▼                   ▼
┌───────────────┐  ┌───────────────┐  ┌───────────────┐
│  PostgreSQL   │  │    Redis      │  │   MinIO       │
│  (数据持久化)  │  │  (缓存/会话)   │  │  (对象存储)    │
└───────────────┘  └───────────────┘  └───────────────┘
```

### 技术栈

| 层次 | 技术选型 |
|------|---------|
| **前端** | React 18, TypeScript, Ant Design 5, Recharts, Zustand |
| **后端** | Go 1.21, Gin, GORM, JWT, bcrypt |
| **容器** | Docker 20.x, Kubernetes 1.28, NVIDIA GPU Operator |
| **数据库** | PostgreSQL 15, Redis 7, MinIO |
| **监控** | Prometheus, Grafana |
| **工具** | Make, Vite, Docker |

## 🚀 快速开始

### 环境要求

- **操作系统**: Ubuntu 18.04+ / CentOS 7+ / macOS
- **Go**: 1.21+
- **数据库**: PostgreSQL 14+
- **可选**: Kubernetes 1.28+, Docker 24+

### 1. 克隆项目

```bash
git clone https://github.com/xbl916/gpu-platform.git
cd gpu-platform
```

### 2. 配置数据库

```bash
# 创建数据库和用户
sudo -u postgres psql

CREATE DATABASE gpu_platform;
CREATE USER gpu_admin WITH PASSWORD 'your_password';
GRANT ALL PRIVILEGES ON DATABASE gpu_platform TO gpu_admin;
\q
```

### 3. 配置环境

```bash
# 复制配置文件
cp config.yaml.example config.yaml

# 编辑配置
vim config.yaml
```

```yaml
# config.yaml 关键配置
app:
  host: "0.0.0.0"
  port: 8080

database:
  host: "localhost"
  port: 5432
  username: "gpu_admin"
  password: "your_password"
  name: "gpu_platform"

jwt:
  secret: "your-256-bit-secret-key"
```

### 4. 编译运行

```bash
# 编译后端
make build

# 运行服务
./bin/api-server

# 或开发模式运行
make run
```

### 5. 启动前端

```bash
# 安装前端依赖
cd frontend
npm install

# 启动前端开发服务器
npm run dev
```

### 6. 访问应用

- **前端界面**: http://localhost:3000
- **后端 API**: http://localhost:8080
- **健康检查**: http://localhost:8080/health

## 📖 使用指南

### 用户注册

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePassword123!",
    "name": "Test User"
  }'
```

### 创建 GPU 容器

```bash
# 登录获取 Token
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com", "password": "SecurePassword123!"}'

# 创建容器 (替换 ACCESS_TOKEN)
curl -X POST http://localhost:8080/api/v1/containers \
  -H "Authorization: Bearer ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "my-gpu-container",
    "resources": {
      "gpuCount": 1,
      "gpuModel": "a100",
      "cpuCores": 4,
      "memoryMb": 16384,
      "storageGb": 100,
      "image": "nvidia/cuda:12.1-runtime-ubuntu22.04"
    }
  }'
```

### 容器生命周期

```bash
# 查看容器列表
GET /api/v1/containers

# 启动容器
POST /api/v1/containers/{id}/start

# 停止容器
POST /api/v1/containers/{id}/stop

# 删除容器
DELETE /api/v1/containers/{id}
```

## 📁 项目结构

```
gpu-platform/
├── cmd/
│   └── api-server/          # API 服务入口
├── config.yaml             # 配置文件
├── Makefile               # 构建脚本
├── frontend/              # React 前端
│   ├── src/
│   │   ├── components/    # 公共组件
│   │   ├── pages/        # 页面组件
│   │   ├── services/     # API 服务
│   │   └── store/       # 状态管理
│   └── package.json
├── internal/
│   ├── auth/             # 认证服务
│   ├── config/           # 配置管理
│   ├── handlers/         # HTTP 处理器
│   ├── k8s/             # K8s 客户端
│   ├── middleware/       # 中间件
│   ├── models/           # 数据模型
│   ├── monitor/          # 监控服务
│   ├── notification/     # 通知服务
│   ├── repository/       # 数据访问层
│   └── services/          # 业务逻辑层
├── test/                 # 测试文件
└── docs/                  # 文档
```

## 🛠️ 开发指南

### Make 命令

```bash
# 编译项目
make build

# 运行服务
make run

# 运行测试
make test

# 测试覆盖率
make test-coverage

# 清理构建产物
make clean

# 构建 Docker 镜像
make docker-build
```

### 添加新功能

1. 在 `internal/models/` 添加数据模型
2. 在 `internal/repository/` 添加数据访问层
3. 在 `internal/services/` 添加业务逻辑
4. 在 `internal/handlers/` 添加 API 接口
5. 在 `frontend/src/pages/` 添加前端页面

### 代码规范

- 遵循 Go 代码规范（gofmt）
- 使用 TypeScript 严格模式
- 编写单元测试
- 使用 Git Commit Message 规范

## 🐳 Docker 部署

```bash
# 构建镜像
docker build -t gpu-platform/api-server:latest .

# 运行容器
docker run -d \
  --name gpu-platform \
  -p 8080:8080 \
  -e DB_HOST=db_host \
  -e DB_PASSWORD=your_password \
  gpu-platform/api-server:latest
```

## ☸️ Kubernetes 部署

```bash
# 部署到 Kubernetes
kubectl apply -f deployments/k8s/
```

## 📝 API 文档

| 方法 | 路径 | 描述 |
|------|------|------|
| POST | /api/v1/auth/register | 用户注册 |
| POST | /api/v1/auth/login | 用户登录 |
| GET | /api/v1/users/me | 获取当前用户 |
| GET | /api/v1/resources/availability | 资源可用性 |
| GET | /api/v1/containers | 容器列表 |
| POST | /api/v1/containers | 创建容器 |
| GET | /api/v1/containers/:id | 容器详情 |
| DELETE | /api/v1/containers/:id | 删除容器 |
| POST | /api/v1/containers/:id/start | 启动容器 |
| POST | /api/v1/containers/:id/stop | 停止容器 |
| GET | /api/v1/templates | 模板列表 |
| POST | /api/v1/templates | 创建模板 |
| GET | /api/v1/monitor/dashboard | 监控仪表盘 |

## 🤝 贡献指南

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/xxx`)
3. 提交更改 (`git commit -m 'feat: xxx'`)
4. 推送到分支 (`git push origin feature/xxx`)
5. 创建 Pull Request

## 📄 许可证

本项目基于 MIT 许可证开源。

## 🙏 感谢

- [NVIDIA GPU Operator](https://github.com/NVIDIA/gpu-operator)
- [Kubernetes](https://kubernetes.io/)
- [Gin](https://gin-gonic.com/)
- [Ant Design](https://ant.design/)

---

<div align="center">
Made with ❤️ by GPU Platform Team
</div>
