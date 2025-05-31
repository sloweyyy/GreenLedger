# 🌱 GreenLedger Kubernetes Configuration

This directory contains Kubernetes manifests for deploying GreenLedger microservices to different environments.

## 📁 Directory Structure

```
k8s/
├── README.md                    # This file
├── base/                        # Base configurations (shared)
│   ├── namespace.yaml          # Namespace definitions
│   ├── configmap.yaml          # Common configuration
│   ├── secrets.yaml            # Secret templates
│   └── ingress.yaml            # Ingress controller setup
├── staging/                     # Staging environment
│   ├── calculator-deployment.yaml
│   ├── tracker-deployment.yaml
│   ├── wallet-deployment.yaml
│   ├── user-auth-deployment.yaml
│   ├── reporting-deployment.yaml
│   ├── services.yaml           # All service definitions
│   ├── configmap.yaml          # Staging-specific config
│   ├── secrets.yaml            # Staging secrets
│   ├── ingress.yaml            # Staging ingress
│   └── infrastructure.yaml     # PostgreSQL, Redis, Kafka
└── production/                  # Production environment
    ├── calculator-deployment.yaml
    ├── tracker-deployment.yaml
    ├── wallet-deployment.yaml
    ├── user-auth-deployment.yaml
    ├── reporting-deployment.yaml
    ├── services.yaml           # All service definitions
    ├── configmap.yaml          # Production-specific config
    ├── secrets.yaml            # Production secrets
    ├── ingress.yaml            # Production ingress
    └── infrastructure.yaml     # PostgreSQL, Redis, Kafka
```

## 🚀 Deployment Instructions

### Prerequisites

1. **Kubernetes cluster** (EKS, GKE, AKS, or local)
2. **kubectl** configured with cluster access
3. **Docker images** built and pushed to registry
4. **Secrets configured** (see secrets.yaml templates)

### Quick Start

```bash
# 1. Create namespaces
kubectl apply -f k8s/base/namespace.yaml

# 2. Deploy to staging
kubectl apply -f k8s/staging/ --namespace=greenledger-staging

# 3. Deploy to production (when ready)
kubectl apply -f k8s/production/ --namespace=greenledger-production
```

### Environment-Specific Deployment

#### Staging Environment

```bash
# Deploy infrastructure first
kubectl apply -f k8s/staging/infrastructure.yaml --namespace=greenledger-staging

# Deploy configuration
kubectl apply -f k8s/staging/configmap.yaml --namespace=greenledger-staging
kubectl apply -f k8s/staging/secrets.yaml --namespace=greenledger-staging

# Deploy services
kubectl apply -f k8s/staging/calculator-deployment.yaml --namespace=greenledger-staging
kubectl apply -f k8s/staging/tracker-deployment.yaml --namespace=greenledger-staging
kubectl apply -f k8s/staging/wallet-deployment.yaml --namespace=greenledger-staging
kubectl apply -f k8s/staging/user-auth-deployment.yaml --namespace=greenledger-staging
kubectl apply -f k8s/staging/reporting-deployment.yaml --namespace=greenledger-staging

# Deploy service definitions and ingress
kubectl apply -f k8s/staging/services.yaml --namespace=greenledger-staging
kubectl apply -f k8s/staging/ingress.yaml --namespace=greenledger-staging
```

#### Production Environment

```bash
# Deploy infrastructure first
kubectl apply -f k8s/production/infrastructure.yaml --namespace=greenledger-production

# Deploy configuration
kubectl apply -f k8s/production/configmap.yaml --namespace=greenledger-production
kubectl apply -f k8s/production/secrets.yaml --namespace=greenledger-production

# Deploy services
kubectl apply -f k8s/production/calculator-deployment.yaml --namespace=greenledger-production
kubectl apply -f k8s/production/tracker-deployment.yaml --namespace=greenledger-production
kubectl apply -f k8s/production/wallet-deployment.yaml --namespace=greenledger-production
kubectl apply -f k8s/production/user-auth-deployment.yaml --namespace=greenledger-production
kubectl apply -f k8s/production/reporting-deployment.yaml --namespace=greenledger-production

# Deploy service definitions and ingress
kubectl apply -f k8s/production/services.yaml --namespace=greenledger-production
kubectl apply -f k8s/production/ingress.yaml --namespace=greenledger-production
```

## 🔧 Configuration

### Environment Variables

Each environment has its own ConfigMap with environment-specific settings:

- **Database connections** (per service)
- **Redis configuration**
- **Kafka brokers**
- **Service discovery**
- **Feature flags**

### Secrets Management

Before deploying, create the required secrets:

```bash
# Database passwords
kubectl create secret generic db-secrets \
  --from-literal=calculator-db-password=your-password \
  --from-literal=tracker-db-password=your-password \
  --from-literal=wallet-db-password=your-password \
  --from-literal=userauth-db-password=your-password \
  --from-literal=reporting-db-password=your-password \
  --namespace=greenledger-staging

# JWT secrets
kubectl create secret generic jwt-secret \
  --from-literal=jwt-secret-key=your-super-secret-jwt-key \
  --namespace=greenledger-staging

# External service credentials
kubectl create secret generic external-secrets \
  --from-literal=redis-password=your-redis-password \
  --from-literal=kafka-password=your-kafka-password \
  --namespace=greenledger-staging
```

## 📊 Monitoring

### Health Checks

All services include:
- **Liveness probes** - Restart unhealthy containers
- **Readiness probes** - Route traffic only to ready containers
- **Startup probes** - Handle slow-starting containers

### Resource Management

Each service has:
- **Resource requests** - Guaranteed resources
- **Resource limits** - Maximum resource usage
- **Horizontal Pod Autoscaler** - Scale based on CPU/memory

### Observability

- **Prometheus metrics** exposed on `/metrics`
- **Health endpoints** on `/health`
- **Grafana dashboards** for monitoring
- **Distributed tracing** with Jaeger

## 🔒 Security

### Network Policies

- **Service-to-service** communication restrictions
- **Database access** limited to specific services
- **External traffic** controlled via ingress

### Pod Security

- **Non-root containers**
- **Read-only root filesystem**
- **Security contexts** with minimal privileges
- **Resource quotas** per namespace

## 🔄 CI/CD Integration

These manifests are designed to work with the GitHub Actions deployment workflow:

1. **Image tags** are updated automatically during deployment
2. **Rolling updates** ensure zero-downtime deployments
3. **Health checks** verify deployment success
4. **Rollback** capabilities for failed deployments

## 📚 Additional Resources

- [Kubernetes Documentation](https://kubernetes.io/docs/)
- [EKS Best Practices](https://aws.github.io/aws-eks-best-practices/)
- [Helm Charts](https://helm.sh/) (for more complex deployments)
- [Kustomize](https://kustomize.io/) (for configuration management)

---

**Last Updated**: January 2025
**Maintained by**: GreenLedger Team
