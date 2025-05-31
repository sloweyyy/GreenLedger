# 🔄 GitHub Workflows Documentation

This document provides an overview of all GitHub workflows configured for the GreenLedger project.

## 📋 Workflow Overview

### 1. 🔄 Continuous Integration (`ci.yml`)
**Triggers**: Push to main/develop, Pull Requests
**Purpose**: Comprehensive testing and quality assurance

**Jobs**:
- **Code Quality & Linting**: golangci-lint, gosec security scanning
- **Unit Tests**: Matrix testing across all services with coverage reporting
- **Integration Tests**: Full service testing with PostgreSQL, Redis, Kafka
- **Docker Build & Test**: Multi-service Docker image building
- **Load Testing**: Performance testing for production readiness
- **Security Scanning**: Vulnerability detection with multiple tools

### 2. 🚀 Deployment (`deploy.yml`)
**Triggers**: Push to main, Tags (v*), Manual dispatch
**Purpose**: Automated deployment to staging and production

**Jobs**:
- **Build Production Images**: Multi-architecture Docker images
- **Deploy to Staging**: Automated staging deployment with health checks
- **Deploy to Production**: Controlled production deployment
- **Rollback**: Manual rollback capability for emergencies

### 3. 📦 Release Management (`release.yml`)
**Triggers**: Tags (v*), Manual dispatch
**Purpose**: Automated release creation and artifact management

**Jobs**:
- **Release Validation**: Version format and changelog verification
- **Full Test Suite**: Comprehensive testing before release
- **Build Artifacts**: Multi-platform binary builds
- **Docker Images**: Tagged release images
- **Release Notes**: Automated release note generation
- **Documentation Updates**: Version reference updates

### 4. 🔒 Security Scanning (`security.yml`)
**Triggers**: Push, Pull Requests, Daily schedule, Manual dispatch
**Purpose**: Comprehensive security analysis

**Jobs**:
- **SAST**: Static application security testing
- **Dependency Scanning**: Vulnerability detection in dependencies
- **Container Security**: Multi-layer container security scanning
- **Infrastructure Security**: Docker/Kubernetes security analysis
- **Secrets Detection**: Credential and secret scanning
- **License Compliance**: License verification
- **Security Reporting**: Comprehensive security summary

### 5. 📦 Dependency Updates (`dependency-update.yml`)
**Triggers**: Weekly schedule, Manual dispatch
**Purpose**: Automated dependency management

**Jobs**:
- **Go Dependencies**: Automated Go module updates
- **Docker Images**: Base image update recommendations
- **GitHub Actions**: Action version updates
- **Security Patches**: Automated vulnerability patching
- **Pull Request Creation**: Automated PR with detailed descriptions

### 6. 📊 Workflow Status (`workflow-status.yml`)
**Triggers**: Workflow completion, Daily schedule, Manual dispatch
**Purpose**: Monitor and report on workflow health

**Jobs**:
- **Workflow Health**: Overall workflow status monitoring
- **Security Status**: Security workflow monitoring
- **Performance Monitor**: Workflow execution time analysis
- **Dependency Health**: Dependency status checking

## 🔧 Setup Instructions

### Required Secrets

Add these secrets in your GitHub repository settings (`Settings > Secrets and variables > Actions`):

#### AWS Deployment (if using AWS)
```
AWS_ACCESS_KEY_ID=your-access-key-id
AWS_SECRET_ACCESS_KEY=your-secret-access-key
AWS_REGION=us-west-2
EKS_CLUSTER_NAME_STAGING=your-staging-cluster
EKS_CLUSTER_NAME_PRODUCTION=your-production-cluster
```

#### Security Scanning
```
SNYK_TOKEN=your-snyk-token (optional)
SEMGREP_APP_TOKEN=your-semgrep-token (optional)
```

#### Container Registry
```
GITHUB_TOKEN=automatically-provided
```

### Required Environments

Configure these environments in your repository (`Settings > Environments`):

#### Staging Environment
- **Protection rules**: None (auto-deploy)
- **Environment secrets**: Staging-specific configurations

#### Production Environment
- **Protection rules**: Required reviewers
- **Environment secrets**: Production-specific configurations

### Branch Protection Rules

Configure branch protection for `main` branch:

1. Go to `Settings > Branches`
2. Add rule for `main` branch
3. Enable:
   - Require a pull request before merging
   - Require status checks to pass before merging
   - Require branches to be up to date before merging
   - Include administrators

### Required Status Checks

Add these status checks to branch protection:
- `🔍 Code Quality & Linting`
- `🧪 Unit Tests (calculator)`
- `🧪 Unit Tests (tracker)`
- `🧪 Unit Tests (wallet)`
- `🧪 Unit Tests (user-auth)`
- `🔗 Integration Tests`
- `🐳 Docker Build & Test (calculator)`
- `🐳 Docker Build & Test (tracker)`
- `🐳 Docker Build & Test (wallet)`
- `🐳 Docker Build & Test (user-auth)`

## 🎯 Workflow Triggers

### Automatic Triggers
- **Push to main/develop**: CI workflow
- **Pull Requests**: CI and Security workflows
- **Tags (v*)**: Release and Deploy workflows
- **Daily at 2 AM UTC**: Security scanning
- **Weekly on Mondays**: Dependency updates
- **Daily at 8 AM UTC**: Workflow status check

### Manual Triggers
All workflows can be triggered manually via:
1. Go to `Actions` tab
2. Select the workflow
3. Click `Run workflow`
4. Choose branch and parameters

## 📊 Monitoring and Alerts

### Workflow Status
- Monitor workflow health via the Status workflow
- Automatic issue creation for persistent failures
- Performance monitoring and analysis

### Security Alerts
- SARIF uploads to GitHub Security tab
- Automated security issue creation
- Daily security status reports

### Notifications
- Failed workflow notifications via GitHub
- Security alert notifications
- Dependency update PR notifications

## 🔧 Customization

### Service Configuration
Update service lists in workflows when adding/removing services:
- `ci.yml`: Update matrix.service arrays
- `deploy.yml`: Update service deployment configurations
- `security.yml`: Update container scanning services

### Environment Variables
Customize environment variables in `.env.example` and workflow files:
- Database configurations
- Service ports
- Security settings
- Feature flags

### Workflow Schedules
Modify cron schedules in workflow files:
- Security scanning frequency
- Dependency update schedule
- Status check intervals

## 🚨 Troubleshooting

### Common Issues

#### 1. Docker Build Failures
- Check if Dockerfile exists for the service
- Verify go.mod and go.sum files are present
- Check for missing dependencies

#### 2. Test Failures
- Ensure all required services are running
- Check database connectivity
- Verify environment variables

#### 3. Security Scan Failures
- Review SARIF files in Security tab
- Check for new vulnerabilities
- Update dependencies if needed

#### 4. Deployment Failures
- Verify AWS credentials and permissions
- Check Kubernetes cluster connectivity
- Review deployment manifests

### Getting Help

1. **Check workflow logs**: Click on failed jobs for detailed logs
2. **Review documentation**: Check service-specific documentation
3. **Create an issue**: Use issue templates for bug reports
4. **Contact maintainers**: Reach out via email or discussions

## 📚 Additional Resources

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Docker Best Practices](https://docs.docker.com/develop/dev-best-practices/)
- [Go Testing Guide](https://golang.org/doc/tutorial/add-a-test)
- [Security Best Practices](../SECURITY.md)
- [Contributing Guidelines](../CONTRIBUTING.md)

---

**Last Updated**: May 2025
**Version**: 1.0
