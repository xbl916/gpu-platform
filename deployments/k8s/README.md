# GPU Platform - 完整 K8s 部署配置
# 使用方法: kubectl apply -k deployments/k8s/overlays/production

## 目录结构

deployments/
├── base/                          # 基础配置
│   ├── namespace.yaml              # Namespace
│   ├── configmap.yaml            # ConfigMap
│   ├── deployment.yaml           # Deployment
│   ├── service.yaml             # Service
│   ├── ingress.yaml             # Ingress
│   └── kustomization.yaml       # Kustomization
└── overlays/
    ├── development/              # 开发环境
    │   └── kustomization.yaml
    └── production/               # 生产环境
        ├── replica-patch.yaml     # 副本数调整
        └── kustomization.yaml

## 快速部署

```bash
# 部署到开发环境
kubectl apply -k deployments/k8s/overlays/development

# 部署到生产环境
kubectl apply -k deployments/k8s/overlays/production
```