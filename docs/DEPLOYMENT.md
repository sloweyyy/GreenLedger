# 🚀 GreenLedger Deployment Guide

## Prerequisites

### Local Development Setup
- Go 1.21+
- Docker & Docker Compose
- PostgreSQL client (for manual DB operations)
- Make (for build automation)

### Production
- Kubernetes cluster (1.25+)
- kubectl configured
- Helm 3.x (optional, for package management)
- Container registry access

## Local Development Setup

### 1. Clone and Setup
```bash
git clone https://github.com/sloweyyy/GreenLedger.git
cd GreenLedger

# Install Go dependencies
go mod download
```

### 2. Environment Configuration
Create `.env` file in the root directory:
```bash
# Database Configuration
DB_HOST=localhost
DB_USER=postgres
DB_PASSWORD=password
DB_SSLMODE=disable

# Redis Configuration
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# Kafka Configuration
KAFKA_BROKERS=localhost:9092

# JWT Configuration
JWT_SECRET=your-super-secret-jwt-key-change-in-production

# Service Ports
CALCULATOR_PORT=8081
TRACKER_PORT=8082
WALLET_PORT=8083
USER_AUTH_PORT=8084
REPORTING_PORT=8085
CERTIFIER_PORT=8086

# Monitoring
PROMETHEUS_PORT=9090
GRAFANA_PORT=3000

# Log Level
LOG_LEVEL=info
ENVIRONMENT=development
```

### 3. Start Infrastructure Services
```bash
# Start PostgreSQL, Redis, and Kafka
docker-compose up -d postgres-calculator postgres-tracker postgres-wallet postgres-userauth redis kafka

# Wait for services to be ready
sleep 30

# Verify services are running
docker-compose ps
```

### 4. Run Database Migrations
```bash
# Run migrations for all services
make migrate-up

# Or run individually
cd services/calculator && migrate -path migrations -database "postgres://postgres:password@localhost:5432/calculator_db?sslmode=disable" up
cd services/tracker && migrate -path migrations -database "postgres://postgres:password@localhost:5433/tracker_db?sslmode=disable" up
cd services/wallet && migrate -path migrations -database "postgres://postgres:password@localhost:5434/wallet_db?sslmode=disable" up
cd services/user-auth && migrate -path migrations -database "postgres://postgres:password@localhost:5435/userauth_db?sslmode=disable" up
```

### 5. Start Services

#### Option A: Using Make (Recommended)
```bash
# Start all services
make run-all

# Or start individual services
make run-calculator
make run-tracker
make run-wallet
```

#### Option B: Manual Start
```bash
# Terminal 1: Calculator Service
cd services/calculator
go run cmd/main.go

# Terminal 2: Tracker Service (when implemented)
cd services/tracker
go run cmd/main.go

# Terminal 3: Wallet Service (when implemented)
cd services/wallet
go run cmd/main.go
```

### 6. Verify Deployment
```bash
# Test calculator service
curl http://localhost:8081/health

# Run comprehensive tests
./scripts/test-calculator.sh

# Check API documentation
open http://localhost:8081/swagger
```

## Docker Deployment

### 1. Build Docker Images
```bash
# Build all services
make docker-build

# Or build individual services
docker build -t greenledger/calculator:latest -f services/calculator/Dockerfile .
docker build -t greenledger/tracker:latest -f services/tracker/Dockerfile .
docker build -t greenledger/wallet:latest -f services/wallet/Dockerfile .
```

### 2. Start with Docker Compose
```bash
# Start all services with Docker Compose
docker-compose up -d

# Check service status
docker-compose ps

# View logs
docker-compose logs -f calculator-service
```

### 3. Scale Services
```bash
# Scale calculator service
docker-compose up -d --scale calculator-service=3

# Scale with load balancer
docker-compose -f docker-compose.yml -f docker-compose.scale.yml up -d
```

## Production Deployment (Kubernetes)

### 1. Prepare Kubernetes Manifests

Create `k8s/` directory structure:
```
k8s/
├── namespace.yaml
├── configmaps/
├── secrets/
├── databases/
├── services/
│   ├── calculator/
│   ├── tracker/
│   ├── wallet/
│   └── user-auth/
├── ingress/
└── monitoring/
```

### 2. Create Namespace
```yaml
# k8s/namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: greenledger
  labels:
    name: greenledger
```

### 3. Database Deployment
```yaml
# k8s/databases/postgres-calculator.yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: postgres-calculator
  namespace: greenledger
spec:
  serviceName: postgres-calculator
  replicas: 1
  selector:
    matchLabels:
      app: postgres-calculator
  template:
    metadata:
      labels:
        app: postgres-calculator
    spec:
      containers:
      - name: postgres
        image: postgres:15-alpine
        env:
        - name: POSTGRES_DB
          value: calculator_db
        - name: POSTGRES_USER
          value: postgres
        - name: POSTGRES_PASSWORD
          valueFrom:
            secretKeyRef:
              name: postgres-secret
              key: password
        ports:
        - containerPort: 5432
        volumeMounts:
        - name: postgres-storage
          mountPath: /var/lib/postgresql/data
  volumeClaimTemplates:
  - metadata:
      name: postgres-storage
    spec:
      accessModes: ["ReadWriteOnce"]
      resources:
        requests:
          storage: 10Gi
```

### 4. Service Deployment
```yaml
# k8s/services/calculator/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: calculator-service
  namespace: greenledger
spec:
  replicas: 3
  selector:
    matchLabels:
      app: calculator-service
  template:
    metadata:
      labels:
        app: calculator-service
    spec:
      containers:
      - name: calculator
        image: greenledger/calculator:latest
        ports:
        - containerPort: 8081
        - containerPort: 9081
        env:
        - name: DB_HOST
          value: postgres-calculator
        - name: DB_NAME
          value: calculator_db
        - name: DB_USER
          value: postgres
        - name: DB_PASSWORD
          valueFrom:
            secretKeyRef:
              name: postgres-secret
              key: password
        - name: REDIS_HOST
          value: redis
        - name: KAFKA_BROKERS
          value: kafka:9092
        - name: JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: jwt-secret
              key: secret
        livenessProbe:
          httpGet:
            path: /health
            port: 8081
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8081
          initialDelaySeconds: 5
          periodSeconds: 5
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "512Mi"
            cpu: "500m"
```

### 5. Deploy to Kubernetes
```bash
# Create namespace
kubectl apply -f k8s/namespace.yaml

# Create secrets
kubectl create secret generic postgres-secret \
  --from-literal=password=your-secure-password \
  -n greenledger

kubectl create secret generic jwt-secret \
  --from-literal=secret=your-jwt-secret \
  -n greenledger

# Deploy databases
kubectl apply -f k8s/databases/

# Deploy services
kubectl apply -f k8s/services/

# Deploy ingress
kubectl apply -f k8s/ingress/

# Check deployment status
kubectl get pods -n greenledger
kubectl get services -n greenledger
```

### 6. Configure Ingress
```yaml
# k8s/ingress/ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: greenledger-ingress
  namespace: greenledger
  annotations:
    kubernetes.io/ingress.class: nginx
    cert-manager.io/cluster-issuer: letsencrypt-prod
spec:
  tls:
  - hosts:
    - api.greenledger.com
    secretName: greenledger-tls
  rules:
  - host: api.greenledger.com
    http:
      paths:
      - path: /calculator
        pathType: Prefix
        backend:
          service:
            name: calculator-service
            port:
              number: 8081
      - path: /tracker
        pathType: Prefix
        backend:
          service:
            name: tracker-service
            port:
              number: 8082
```

## Monitoring Setup

### 1. Prometheus Configuration
```bash
# Deploy Prometheus
kubectl apply -f k8s/monitoring/prometheus/

# Configure service monitors
kubectl apply -f k8s/monitoring/servicemonitors/
```

### 2. Grafana Setup
```bash
# Deploy Grafana
kubectl apply -f k8s/monitoring/grafana/

# Import dashboards
kubectl create configmap grafana-dashboards \
  --from-file=monitoring/grafana/dashboards/ \
  -n greenledger
```

## Security Configuration

### 1. Network Policies
```yaml
# k8s/security/network-policy.yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: greenledger-network-policy
  namespace: greenledger
spec:
  podSelector: {}
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: ingress-nginx
  egress:
  - to:
    - namespaceSelector:
        matchLabels:
          name: kube-system
```

### 2. Pod Security Standards
```yaml
# k8s/security/pod-security-policy.yaml
apiVersion: policy/v1beta1
kind: PodSecurityPolicy
metadata:
  name: greenledger-psp
spec:
  privileged: false
  allowPrivilegeEscalation: false
  requiredDropCapabilities:
    - ALL
  volumes:
    - 'configMap'
    - 'emptyDir'
    - 'projected'
    - 'secret'
    - 'downwardAPI'
    - 'persistentVolumeClaim'
  runAsUser:
    rule: 'MustRunAsNonRoot'
  seLinux:
    rule: 'RunAsAny'
  fsGroup:
    rule: 'RunAsAny'
```

## Backup and Recovery

### 1. Database Backup
```bash
# Create backup job
kubectl create job postgres-backup-$(date +%Y%m%d) \
  --from=cronjob/postgres-backup \
  -n greenledger

# Manual backup
kubectl exec -it postgres-calculator-0 -n greenledger -- \
  pg_dump -U postgres calculator_db > backup-$(date +%Y%m%d).sql
```

### 2. Disaster Recovery
```bash
# Restore from backup
kubectl exec -i postgres-calculator-0 -n greenledger -- \
  psql -U postgres calculator_db < backup-20231201.sql
```

## Troubleshooting

### Common Issues

#### 1. Service Not Starting
```bash
# Check pod logs
kubectl logs -f deployment/calculator-service -n greenledger

# Check events
kubectl get events -n greenledger --sort-by='.lastTimestamp'

# Check resource usage
kubectl top pods -n greenledger
```

#### 2. Database Connection Issues
```bash
# Test database connectivity
kubectl exec -it postgres-calculator-0 -n greenledger -- \
  psql -U postgres -d calculator_db -c "SELECT 1;"

# Check database logs
kubectl logs -f postgres-calculator-0 -n greenledger
```

#### 3. Performance Issues
```bash
# Check resource limits
kubectl describe pod calculator-service-xxx -n greenledger

# Monitor metrics
kubectl port-forward svc/prometheus 9090:9090 -n greenledger
open http://localhost:9090
```

### Health Checks
```bash
# Service health
curl http://api.greenledger.com/calculator/health

# Database health
kubectl exec -it postgres-calculator-0 -n greenledger -- \
  pg_isready -U postgres

# Overall system health
kubectl get pods -n greenledger
kubectl get services -n greenledger
kubectl get ingress -n greenledger
```

## Scaling

### Horizontal Pod Autoscaler
```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: calculator-hpa
  namespace: greenledger
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: calculator-service
  minReplicas: 2
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
```

### Vertical Pod Autoscaler
```yaml
apiVersion: autoscaling.k8s.io/v1
kind: VerticalPodAutoscaler
metadata:
  name: calculator-vpa
  namespace: greenledger
spec:
  targetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: calculator-service
  updatePolicy:
    updateMode: "Auto"
```
