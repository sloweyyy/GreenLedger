---
name: 🔒 Security Issue
about: Report a security vulnerability (use private reporting for sensitive issues)
title: '[SECURITY] '
labels: ['security', 'bug']
assignees: ''
---

# 🔒 Security Issue

> **⚠️ IMPORTANT:** For sensitive security vulnerabilities, please use GitHub's private vulnerability reporting feature or email truonglevinhphuc2006@gmail.com directly instead of creating a public issue.

## 🚨 Security Issue Type

<!-- Mark the relevant option with an "x" -->

- [ ] 🔓 Authentication bypass
- [ ] 🔑 Authorization issues
- [ ] 💉 SQL injection
- [ ] 🌐 Cross-site scripting (XSS)
- [ ] 🔄 Cross-site request forgery (CSRF)
- [ ] 📊 Information disclosure
- [ ] 🔐 Cryptographic issues
- [ ] 📦 Dependency vulnerabilities
- [ ] 🐳 Container security issues
- [ ] ☸️ Kubernetes security issues
- [ ] 🔧 Configuration issues
- [ ] 🎯 Other: ___________

## 🎯 Affected Components

<!-- Mark all affected components -->

- [ ] 🧮 Calculator Service
- [ ] 📊 Tracker Service
- [ ] 💰 Wallet Service
- [ ] 🔐 User Auth Service
- [ ] 📈 Reporting Service
- [ ] 🏆 Certificate Service
- [ ] 🌐 API Gateway
- [ ] 🗄️ Database
- [ ] 🔄 Message Queue
- [ ] 💾 Cache
- [ ] 🐳 Docker containers
- [ ] ☸️ Kubernetes manifests
- [ ] 🔧 Configuration files

## 📊 Severity Assessment

### CVSS Score (if known)

**Base Score:** ___/10

**Vector String:** CVSS:3.1/AV:_/AC:_/PR:_/UI:_/S:_/C:_/I:_/A:_

### Impact Level

- [ ] 🔥 Critical (immediate action required)
- [ ] ⚡ High (significant security risk)
- [ ] 📋 Medium (moderate security risk)
- [ ] 🔍 Low (minor security concern)

### Exploitability

- [ ] 🎯 Easily exploitable
- [ ] 🔧 Requires specific conditions
- [ ] 🔒 Difficult to exploit
- [ ] 📚 Theoretical only

## 🔍 Vulnerability Details

### Description

<!-- Provide a clear description of the security issue -->

### Attack Vector

<!-- How can this vulnerability be exploited? -->

### Prerequisites

<!-- What conditions are needed for exploitation? -->

- [ ] 🔓 No authentication required
- [ ] 👤 Valid user account required
- [ ] 🔑 Admin privileges required
- [ ] 🌐 Network access required
- [ ] 💻 Local access required
- [ ] 🎯 Other: ___________

## 🧪 Proof of Concept

<!-- Provide steps to reproduce (if safe to do so) -->

### Steps to Reproduce

1. 
2. 
3. 
4. 

### Expected Behavior

<!-- What should happen normally? -->

### Actual Behavior

<!-- What actually happens? -->

### Evidence

<!-- Include screenshots, logs, or other evidence (redact sensitive info) -->

```
# Example request/response (sanitized)
curl -X POST https://api.example.com/endpoint \
  -H "Content-Type: application/json" \
  -d '{"payload": "..."}'
```

## 💥 Impact Analysis

### Confidentiality Impact

- [ ] 🔓 No impact
- [ ] 📊 Limited data exposure
- [ ] 🗄️ Significant data exposure
- [ ] 💾 Complete data exposure

### Integrity Impact

- [ ] 🔒 No impact
- [ ] ✏️ Limited data modification
- [ ] 📝 Significant data modification
- [ ] 💥 Complete data corruption

### Availability Impact

- [ ] ✅ No impact
- [ ] ⏳ Limited service disruption
- [ ] 🚫 Significant service disruption
- [ ] 💥 Complete service unavailability

### Business Impact

<!-- How does this affect the business? -->

## 🔧 Environment Details

**Environment:** 
<!-- e.g., production, staging, development -->

**Version:** 
<!-- e.g., v1.2.3 -->

**Deployment:** 
<!-- e.g., Docker, Kubernetes, bare metal -->

**Configuration:**
<!-- Any relevant configuration details (sanitized) -->

## 💡 Suggested Remediation

### Immediate Actions

<!-- What should be done immediately? -->

1. 
2. 
3. 

### Long-term Solutions

<!-- What permanent fixes are needed? -->

1. 
2. 
3. 

### Workarounds

<!-- Any temporary workarounds available? -->

## 🛡️ Security Controls

### Current Controls

<!-- What security controls are currently in place? -->

- [ ] 🔐 Authentication
- [ ] 🔑 Authorization
- [ ] 🔍 Input validation
- [ ] 🛡️ Output encoding
- [ ] 🔒 Encryption
- [ ] 📊 Logging/monitoring
- [ ] 🚫 Rate limiting
- [ ] 🌐 Network security

### Missing Controls

<!-- What security controls are missing? -->

## 📋 Compliance Impact

<!-- Does this affect any compliance requirements? -->

- [ ] 🇪🇺 GDPR
- [ ] 🏥 HIPAA
- [ ] 💳 PCI DSS
- [ ] 📊 SOC 2
- [ ] 🔒 ISO 27001
- [ ] 🎯 Other: ___________

## 🔗 References

<!-- Include any relevant references -->

### CVE References

<!-- If this relates to known CVEs -->

- CVE-YYYY-NNNN

### Security Advisories

<!-- Any related security advisories -->

### Documentation

<!-- Relevant security documentation -->

## 📝 Additional Information

### Timeline

<!-- When was this discovered? -->

**Discovery Date:** 
**Reporter:** 
**Affected Since:** 

### Related Issues

<!-- Link any related security issues -->

- Related to #
- Blocks #
- Blocked by #

### Testing Notes

<!-- Any testing considerations -->

---

## 🔒 For Security Team

### Triage Checklist

- [ ] 📊 Severity assessed
- [ ] 🎯 Impact analyzed
- [ ] 🔍 Reproducibility confirmed
- [ ] 💡 Remediation plan created
- [ ] 📅 Timeline established
- [ ] 👥 Stakeholders notified

### Response Actions

- [ ] 🚨 Incident response activated
- [ ] 🔧 Hotfix deployed
- [ ] 📢 Security advisory published
- [ ] 🧪 Security tests added
- [ ] 📚 Documentation updated

---

**Thank you for helping keep GreenLedger secure! 🔒**

> Remember: For sensitive vulnerabilities, please use private reporting channels.
