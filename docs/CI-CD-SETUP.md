# 🔄 CI/CD Setup and Troubleshooting Guide

This guide provides comprehensive instructions for setting up and troubleshooting the CI/CD pipeline for GreenLedger.

## 🚀 Quick Start

### 1. Fix Dependencies (Most Common Issue)

If you're experiencing CI failures related to missing dependencies or go.sum files:

```bash
# Fix all dependencies automatically
make deps

# Or run the script directly
./scripts/fix-dependencies.sh

# Check dependency status
make deps-check
```

### 2. Run Local CI Pipeline

Test the CI pipeline locally before pushing:

```bash
# Run complete local CI pipeline
make ci-local

# Or run individual steps
make deps        # Fix dependencies
make lint        # Run linters
make test        # Run tests
make build       # Build services
```

## 🔧 Common CI Issues and Solutions

### Issue 1: Missing go.sum Files

**Error**: `go.sum: no such file or directory` or `missing go.sum entry`

**Solution**:

```bash
# Automatic fix
make deps

# Manual fix for specific service
cd services/calculator
go mod download
go mod tidy
cd ../..
```

### Issue 2: Incorrect gosec Action Reference

**Error**: `Unable to resolve action securecodewarrior/github-action-gosec@master`

**Solution**: ✅ **Already Fixed** - The workflow now installs gosec directly:

```yaml
- name: 🔒 Run gosec security scanner
  run: |
    go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
    gosec -fmt sarif -out gosec.sarif ./...
```

### Issue 3: Docker Build Failures

**Error**: `Dockerfile not found` or `failed to solve`

**Solution**: ✅ **Already Fixed** - The workflow now checks for Dockerfile existence:

```yaml
- name: 🔍 Check if Dockerfile exists
  id: dockerfile-check
  run: |
    if [ -f "./services/${{ matrix.service }}/Dockerfile" ]; then
      echo "dockerfile-exists=true" >> $GITHUB_OUTPUT
    else
      echo "dockerfile-exists=false" >> $GITHUB_OUTPUT
    fi
```

### Issue 4: Test Failures

**Error**: Tests failing due to missing dependencies or setup issues

**Solution**:

```bash
# Fix dependencies first
make deps

# Run tests locally to debug
make test

# Run tests with coverage
make test-coverage

# Check specific service
cd services/calculator
go test -v ./...
```

## 📋 CI/CD Workflow Overview

### Workflows Included

1. **🔄 Continuous Integration** (`ci.yml`)
   - Code quality and linting
   - Unit tests with coverage
   - Integration tests
   - Docker builds
   - Security scanning

2. **🚀 Deployment** (`deploy.yml`)
   - Production image building
   - Staging/production deployment
   - Health checks and rollback

3. **📦 Release Management** (`release.yml`)
   - Automated releases
   - Multi-platform builds
   - Release notes generation

4. **🔒 Security Scanning** (`security.yml`)
   - SAST, dependency scanning
   - Container security
   - Secrets detection

5. **📦 Dependency Updates** (`dependency-update.yml`)
   - Automated dependency updates
   - Security patches
   - PR creation

6. **📊 Workflow Status** (`workflow-status.yml`)
   - Health monitoring
   - Performance analysis
   - Automated issue creation

### Trigger Events

- **Push to main/develop**: CI workflow
- **Pull Requests**: CI and Security workflows
- **Tags (v*)**: Release and Deploy workflows
- **Daily/Weekly**: Security scans and dependency updates

## 🛠️ Development Workflow

### Before Committing

```bash
# 1. Fix dependencies
make deps

# 2. Format code
make format

# 3. Run linters
make lint

# 4. Run tests
make test

# 5. Build services
make build
```

### Creating a Pull Request

1. **Create feature branch**:

   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make changes and test locally**:

   ```bash
   make ci-local  # Run full CI pipeline locally
   ```

3. **Commit and push**:

   ```bash
   git add .
   git commit -m "feat: your feature description"
   git push origin feature/your-feature-name
   ```

4. **Create PR** using the provided template

### After PR Creation

The CI pipeline will automatically:

- ✅ Run code quality checks
- ✅ Execute all tests
- ✅ Build Docker images
- ✅ Perform security scans
- ✅ Generate coverage reports

## 🔒 Security Configuration

### Required Secrets

Add these in GitHub repository settings:

```bash
# AWS (if using AWS deployment)
AWS_ACCESS_KEY_ID=your-access-key
AWS_SECRET_ACCESS_KEY=your-secret-key
AWS_REGION=us-west-2

# Security scanning (optional)
SNYK_TOKEN=your-snyk-token
SEMGREP_APP_TOKEN=your-semgrep-token
```

### Security Scans Included

- **SAST**: gosec, Semgrep
- **Dependencies**: govulncheck, Snyk, Trivy
- **Containers**: Trivy, Grype, Docker Scout
- **Secrets**: TruffleHog, GitLeaks, detect-secrets
- **Infrastructure**: Checkov, Terrascan

## 📊 Monitoring and Alerts

### Workflow Health

The system automatically monitors:

- ✅ Workflow success rates
- ✅ Performance metrics
- ✅ Security scan results
- ✅ Dependency health

### Automatic Actions

- 🚨 Creates issues for persistent failures
- 📧 Sends notifications for security alerts
- 📝 Generates weekly dependency update PRs
- 📊 Provides health status reports

## 🐛 Troubleshooting Commands

### Dependency Issues

```bash
# Check all dependencies
make deps-check

# Fix all dependencies
make deps-fix

# Generate dependency report
./scripts/fix-dependencies.sh report
```

### Build Issues

```bash
# Clean and rebuild
make clean
make build

# Build specific service
cd services/calculator
go build ./cmd/main.go
```

### Test Issues

```bash
# Run tests with verbose output
cd services/calculator
go test -v ./...

# Run tests with race detection
go test -race ./...

# Run specific test
go test -run TestSpecificFunction ./...
```

### Docker Issues

```bash
# Check Docker setup
docker --version
docker-compose --version

# Build specific service
make docker-build-service SERVICE=calculator

# Check running containers
make docker-ps

# View logs
make docker-logs
```

## 📚 Additional Resources

- [GitHub Workflows Documentation](.github/WORKFLOWS.md)
- [Security Policy](../SECURITY.md)
- [Contributing Guidelines](../CONTRIBUTING.md)
- [Environment Setup](../.env.example)

## 🆘 Getting Help

1. **Check workflow logs** in GitHub Actions tab
2. **Run local CI** with `make ci-local`
3. **Review error messages** and apply suggested fixes
4. **Create an issue** using the provided templates
5. **Contact maintainers** via email or discussions

---

**Last Updated**: May 2025
**Maintained by**: Truong Le Vinh Phuc
