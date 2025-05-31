# 🌱 GreenLedger Pull Request

## 📋 Description

<!-- Provide a brief description of the changes in this PR -->

### 🎯 Type of Change

<!-- Mark the relevant option with an "x" -->

- [ ] 🐛 Bug fix (non-breaking change which fixes an issue)
- [ ] ✨ New feature (non-breaking change which adds functionality)
- [ ] 💥 Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] 📚 Documentation update
- [ ] 🔧 Refactoring (no functional changes)
- [ ] ⚡ Performance improvement
- [ ] 🧪 Test improvements
- [ ] 🔒 Security enhancement
- [ ] 📦 Dependency update
- [ ] 🚀 CI/CD improvements

### 🎯 Affected Services

<!-- Mark all services affected by this change -->

- [ ] 🧮 Calculator Service
- [ ] 📊 Tracker Service  
- [ ] 💰 Wallet Service
- [ ] 🔐 User Auth Service
- [ ] 📈 Reporting Service
- [ ] 🏆 Certificate Service
- [ ] 🌐 API Gateway
- [ ] 🔧 Shared Components
- [ ] 📚 Documentation
- [ ] 🐳 Infrastructure/Docker
- [ ] ☸️ Kubernetes/Deployment

## 🔗 Related Issues

<!-- Link to related issues using keywords like "Fixes #123" or "Closes #456" -->

- Fixes #
- Related to #

## 🧪 Testing

### ✅ Test Coverage

- [ ] Unit tests added/updated
- [ ] Integration tests added/updated
- [ ] Load tests added/updated (if applicable)
- [ ] Manual testing completed

### 🔍 Test Results

<!-- Describe the testing performed and results -->

```bash
# Example test commands and results
make test
make test-coverage
make load-test
```

**Coverage:** X% (target: >80%)

### 🧪 Manual Testing Steps

<!-- Describe manual testing steps for reviewers -->

1. 
2. 
3. 

## 📸 Screenshots/Recordings

<!-- If applicable, add screenshots or recordings to help explain your changes -->

## 🔒 Security Considerations

<!-- Address any security implications -->

- [ ] No sensitive data exposed
- [ ] Input validation implemented
- [ ] Authentication/authorization checked
- [ ] SQL injection prevention verified
- [ ] XSS prevention verified
- [ ] Security tests pass

## 📊 Performance Impact

<!-- Describe any performance implications -->

- [ ] No performance impact
- [ ] Performance improved
- [ ] Performance impact analyzed and acceptable
- [ ] Load testing completed

**Benchmark Results:**
<!-- Include before/after performance metrics if applicable -->

## 🔄 Database Changes

<!-- If this PR includes database changes -->

- [ ] No database changes
- [ ] Migration scripts included
- [ ] Backward compatibility maintained
- [ ] Data migration tested
- [ ] Rollback plan documented

### Migration Details

<!-- Describe database changes -->

## 🐳 Docker/Infrastructure Changes

<!-- If this PR includes infrastructure changes -->

- [ ] No infrastructure changes
- [ ] Dockerfile updated
- [ ] Docker Compose updated
- [ ] Kubernetes manifests updated
- [ ] Environment variables updated
- [ ] Configuration changes documented

## 📚 Documentation

<!-- Documentation updates -->

- [ ] README updated
- [ ] API documentation updated
- [ ] Code comments added/updated
- [ ] Architecture documentation updated
- [ ] Deployment guide updated
- [ ] CHANGELOG updated

## ✅ Pre-Merge Checklist

### 🔍 Code Quality

- [ ] Code follows project style guidelines
- [ ] Self-review completed
- [ ] Code is self-documenting or well-commented
- [ ] No debugging code left in
- [ ] Error handling implemented
- [ ] Logging added where appropriate

### 🧪 Testing

- [ ] All tests pass locally
- [ ] New tests cover the changes
- [ ] Edge cases considered and tested
- [ ] Integration tests pass
- [ ] Load tests pass (if applicable)

### 🔒 Security

- [ ] Security review completed
- [ ] No hardcoded secrets
- [ ] Input validation implemented
- [ ] Authentication/authorization verified
- [ ] Security tests pass

### 📋 Process

- [ ] Branch is up to date with main
- [ ] Commit messages follow conventional commits
- [ ] PR title follows conventional commits
- [ ] Breaking changes documented
- [ ] Migration guide provided (if needed)

### 🚀 Deployment

- [ ] Environment variables documented
- [ ] Configuration changes noted
- [ ] Deployment steps documented
- [ ] Rollback plan available
- [ ] Monitoring/alerting considered

## 🎯 Reviewer Focus Areas

<!-- Guide reviewers on what to focus on -->

Please pay special attention to:

- [ ] Business logic correctness
- [ ] Error handling
- [ ] Performance implications
- [ ] Security considerations
- [ ] API contract changes
- [ ] Database schema changes
- [ ] Configuration changes

## 📝 Additional Notes

<!-- Any additional information for reviewers -->

### 🔄 Follow-up Tasks

<!-- List any follow-up tasks or future improvements -->

- [ ] 
- [ ] 
- [ ] 

### 🤔 Questions for Reviewers

<!-- Any specific questions or concerns -->

1. 
2. 
3. 

---

## 📋 For Maintainers

### 🏷️ Labels to Add

<!-- Maintainers: Add appropriate labels -->

- `service/calculator` `service/tracker` `service/wallet` `service/user-auth` `service/reporting` `service/certifier`
- `type/bug` `type/feature` `type/docs` `type/refactor` `type/performance` `type/security`
- `priority/low` `priority/medium` `priority/high` `priority/critical`
- `size/xs` `size/s` `size/m` `size/l` `size/xl`

### 🎯 Merge Strategy

- [ ] Squash and merge (recommended for feature branches)
- [ ] Rebase and merge (for clean history)
- [ ] Create merge commit (for release branches)

---

**Thank you for contributing to GreenLedger! 🌱**

By submitting this pull request, I confirm that my contribution is made under the terms of the MIT license and I have read and agree to the [Contributing Guidelines](CONTRIBUTING.md) and [Code of Conduct](CODE_OF_CONDUCT.md).
