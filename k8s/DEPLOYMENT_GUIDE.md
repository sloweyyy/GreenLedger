# 🚀 GreenLedger Kubernetes Deployment Guide

This guide provides step-by-step instructions for deploying GreenLedger to Kubernetes environments.

## 📋 Prerequisites

### Required Tools
- `kubectl` (v1.25+)
- `helm` (v3.10+) - optional, for easier management
- Access to a Kubernetes cluster (EKS, GKE, AKS, or local)

### Required Cluster Components
- **Ingress Controller** (NGINX recommended)
- **Cert-Manager** (for SSL certificates)
- **Metrics Server** (for HPA)
- **Storage Class** (for persistent volumes)

### Container Images
Ensure all service images are built and pushed to your container registry:
- `ghcr.io/sloweyyy/greenledger/calculator:latest`
- `ghcr.io/sloweyyy/greenledger/tracker:latest`
- `ghcr.io/sloweyyy/greenledger/wallet:latest`
- `ghcr.io/sloweyyy/greenledger/user-auth:latest`
- `ghcr.io/sloweyyy/greenledger/reporting:latest`

## 🔧 Pre-Deployment Setup

### 1. Install Required Cluster Components

```bash
# Install NGINX Ingress Controller
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.8.2/deploy/static/provider/cloud/deploy.yaml

# Install Cert-Manager
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.0/cert-manager.yaml

# Verify installations
kubectl get pods -n ingress-nginx
kubectl get pods -n cert-manager
```

### 2. Create ClusterIssuer for SSL Certificates

```bash
# Create staging ClusterIssuer
cat <<EOF | kubectl apply -f -
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-staging
spec:
  acme:
    server: https://acme-staging-v02.api.letsencrypt.org/directory
    email: your-email@example.com
    privateKeySecretRef:
      name: letsencrypt-staging
    solvers:
    - http01:
        ingress:
          class: nginx
EOF

# Create production ClusterIssuer
cat <<EOF | kubectl apply -f -
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
```

## 🏗️ Deployment Steps

### Staging Environment

#### 1. Create Namespace and Base Resources
```bash
# Create namespaces
kubectl apply -f k8s/base/namespace.yaml

# Verify namespaces
kubectl get namespaces | grep greenledger
```

#### 2. Create Secrets
```bash
# Generate strong passwords
CALC_DB_PASS=$(openssl rand -base64 32)
TRACKER_DB_PASS=$(openssl rand -base64 32)
WALLET_DB_PASS=$(openssl rand -base64 32)
USERAUTH_DB_PASS=$(openssl rand -base64 32)
REPORTING_DB_PASS=$(openssl rand -base64 32)
JWT_SECRET=$(openssl rand -base64 64)
REDIS_PASS=$(openssl rand -base64 32)

# Create database secrets
kubectl create secret generic db-secrets \
  --from-literal=calculator-db-password="$CALC_DB_PASS" \
  --from-literal=tracker-db-password="$TRACKER_DB_PASS" \
  --from-literal=wallet-db-password="$WALLET_DB_PASS" \
  --from-literal=userauth-db-password="$USERAUTH_DB_PASS" \
  --from-literal=reporting-db-password="$REPORTING_DB_PASS" \
  --namespace=greenledger-staging

# Create JWT secret
kubectl create secret generic jwt-secret \
  --from-literal=jwt-secret-key="$JWT_SECRET" \
  --namespace=greenledger-staging

# Create external service secrets
kubectl create secret generic external-secrets \
  --from-literal=redis-password="$REDIS_PASS" \
  --from-literal=email-api-key="your-sendgrid-api-key" \
  --from-literal=blockchain-rpc-url="https://sepolia.infura.io/v3/your-project-id" \
  --from-literal=blockchain-private-key="your-private-key" \
  --from-literal=carbon-api-key="your-carbon-api-key" \
  --from-literal=weather-api-key="your-weather-api-key" \
  --namespace=greenledger-staging
```

#### 3. Deploy Infrastructure
```bash
# Deploy databases, Redis, and Kafka
kubectl apply -f k8s/staging/infrastructure.yaml

# Wait for infrastructure to be ready
kubectl wait --for=condition=ready pod -l app=postgres --timeout=300s -n greenledger-staging
kubectl wait --for=condition=ready pod -l app=redis --timeout=300s -n greenledger-staging
kubectl wait --for=condition=ready pod -l app=kafka --timeout=300s -n greenledger-staging
```

#### 4. Deploy Configuration
```bash
# Deploy ConfigMap
kubectl apply -f k8s/staging/configmap.yaml

# Verify configuration
kubectl get configmap greenledger-config -n greenledger-staging -o yaml
```

#### 5. Deploy Services
```bash
# Deploy all microservices
kubectl apply -f k8s/staging/calculator-deployment.yaml
kubectl apply -f k8s/staging/tracker-deployment.yaml
kubectl apply -f k8s/staging/wallet-deployment.yaml
kubectl apply -f k8s/staging/user-auth-deployment.yaml
kubectl apply -f k8s/staging/reporting-deployment.yaml

# Deploy service definitions
kubectl apply -f k8s/staging/services.yaml

# Wait for services to be ready
kubectl wait --for=condition=available deployment --all --timeout=600s -n greenledger-staging
```

#### 6. Deploy Ingress
```bash
# Deploy ingress
kubectl apply -f k8s/staging/ingress.yaml

# Check ingress status
kubectl get ingress -n greenledger-staging
kubectl describe ingress greenledger-ingress -n greenledger-staging
```

### Production Environment

Follow the same steps as staging, but use the `k8s/production/` directory and `greenledger-production` namespace.

```bash
# Quick production deployment (after staging is working)
kubectl apply -f k8s/production/
```

## 🔍 Verification and Testing

### Check Deployment Status
```bash
# Check all resources in staging
kubectl get all -n greenledger-staging

# Check pod logs
kubectl logs -f deployment/calculator-service -n greenledger-staging
kubectl logs -f deployment/tracker-service -n greenledger-staging

# Check service endpoints
kubectl get endpoints -n greenledger-staging
```

### Test API Endpoints
```bash
# Get ingress IP
INGRESS_IP=$(kubectl get ingress greenledger-ingress -n greenledger-staging -o jsonpath='{.status.loadBalancer.ingress[0].ip}')

# Test health endpoints
curl -k https://staging-api.greenledger.app/health
curl -k https://staging-api.greenledger.app/api/v1/calculator/health
curl -k https://staging-api.greenledger.app/api/v1/tracker/health
```

### Monitor Resources
```bash
# Check resource usage
kubectl top pods -n greenledger-staging
kubectl top nodes

# Check HPA status
kubectl get hpa -n greenledger-staging
```

## 🔧 Troubleshooting

### Common Issues

#### Pods Not Starting
```bash
# Check pod status and events
kubectl describe pod <pod-name> -n greenledger-staging
kubectl logs <pod-name> -n greenledger-staging --previous
```

#### Database Connection Issues
```bash
# Check database pods
kubectl get pods -l app=postgres -n greenledger-staging
kubectl logs <postgres-pod> -n greenledger-staging

# Test database connectivity
kubectl exec -it <service-pod> -n greenledger-staging -- nc -zv <db-host> 5432
```

#### Ingress Issues
```bash
# Check ingress controller
kubectl get pods -n ingress-nginx
kubectl logs -f deployment/ingress-nginx-controller -n ingress-nginx

# Check certificate status
kubectl get certificates -n greenledger-staging
kubectl describe certificate greenledger-staging-cert -n greenledger-staging
```

### Scaling Operations
```bash
# Manual scaling
kubectl scale deployment calculator-service --replicas=5 -n greenledger-staging

# Check HPA
kubectl get hpa -n greenledger-staging -w
```

## 🔄 Updates and Rollbacks

### Rolling Updates
```bash
# Update image tag
kubectl set image deployment/calculator-service calculator=ghcr.io/sloweyyy/greenledger/calculator:v1.2.0 -n greenledger-staging

# Check rollout status
kubectl rollout status deployment/calculator-service -n greenledger-staging
```

### Rollbacks
```bash
# Check rollout history
kubectl rollout history deployment/calculator-service -n greenledger-staging

# Rollback to previous version
kubectl rollout undo deployment/calculator-service -n greenledger-staging
```

## 🧹 Cleanup

### Remove Staging Environment
```bash
kubectl delete namespace greenledger-staging
```

### Remove Production Environment
```bash
kubectl delete namespace greenledger-production
```

---

**Note**: Always test deployments in staging before applying to production. Monitor resource usage and adjust limits as needed based on actual workload patterns.
