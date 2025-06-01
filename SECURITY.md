# 🔒 Security Policy

## Supported Versions

We actively support the following versions of GreenLedger with security updates:

| Version | Supported          |
| ------- | ------------------ |
| 1.x.x   | :white_check_mark: |
| 0.x.x   | :x:                |

## Reporting a Vulnerability

**Please do not report security vulnerabilities through public GitHub issues.**

We take security seriously and appreciate your efforts to responsibly disclose vulnerabilities. To report a security vulnerability, please use one of the following methods:

### Primary Contact

- **Email**: [truonglevinhphuc2006@gmail.com](mailto:truonglevinhphuc2006@gmail.com)
- **Subject**: `[SECURITY] Brief description of the vulnerability`

### What to Include

When reporting a vulnerability, please include:

1. **Description** of the vulnerability
2. **Steps to reproduce** the issue
3. **Potential impact** and attack scenarios
4. **Affected versions** (if known)
5. **Suggested fix** (if you have one)
6. **Your contact information** for follow-up

### Response Timeline

- **Initial Response**: Within 24 hours
- **Vulnerability Assessment**: Within 72 hours
- **Fix Development**: Depends on severity (see below)
- **Public Disclosure**: After fix is deployed

### Severity Levels

| Severity | Response Time | Description |
|----------|---------------|-------------|
| **Critical** | 24 hours | Remote code execution, data breach |
| **High** | 72 hours | Privilege escalation, authentication bypass |
| **Medium** | 1 week | Information disclosure, DoS |
| **Low** | 2 weeks | Minor security improvements |

## Security Best Practices

### For Users

#### Authentication & Authorization

- Use strong, unique passwords
- Enable two-factor authentication when available
- Regularly rotate API keys and tokens
- Follow the principle of least privilege for user roles

#### Data Protection

- Use HTTPS for all communications
- Encrypt sensitive data at rest
- Regularly backup your data
- Monitor access logs for suspicious activity

#### Infrastructure Security

- Keep Docker images and dependencies updated
- Use secure network configurations
- Implement proper firewall rules
- Regular security audits and penetration testing

### For Developers

#### Secure Coding Practices

- Validate all input data
- Use parameterized queries to prevent SQL injection
- Implement proper error handling (don't expose sensitive information)
- Use secure random number generation
- Follow OWASP security guidelines

#### Dependencies

- Regularly update dependencies
- Use dependency scanning tools
- Monitor for known vulnerabilities
- Pin dependency versions in production

#### API Security

- Implement rate limiting
- Use proper authentication and authorization
- Validate and sanitize all inputs
- Implement CORS properly
- Use security headers

## Security Features

### Authentication & Authorization

- **JWT-based authentication** with configurable expiration
- **Role-based access control (RBAC)** with granular permissions
- **Session management** with secure token handling
- **Password hashing** using bcrypt with salt

### Data Protection

- **Input validation** on all API endpoints
- **SQL injection prevention** through GORM ORM
- **XSS protection** through proper output encoding
- **CSRF protection** for state-changing operations

### Infrastructure Security

- **TLS encryption** for all communications
- **Secure headers** middleware
- **Rate limiting** to prevent abuse
- **Health checks** for monitoring
- **Audit logging** for security events

### Container Security

- **Non-root containers** for all services
- **Minimal base images** (Alpine Linux)
- **Security scanning** in CI/CD pipeline
- **Resource limits** to prevent DoS

## Vulnerability Disclosure Policy

### Coordinated Disclosure

We follow a coordinated disclosure process:

1. **Report received** and acknowledged
2. **Vulnerability confirmed** and assessed
3. **Fix developed** and tested
4. **Security advisory** prepared
5. **Fix deployed** to production
6. **Public disclosure** with credit to reporter

### Public Disclosure Timeline

- **Critical/High**: 90 days after initial report
- **Medium**: 120 days after initial report
- **Low**: 180 days after initial report

We may request an extension if:

- The fix is complex and requires more time
- The vulnerability affects multiple systems
- Coordinating with other vendors is necessary

### Recognition

We believe in recognizing security researchers who help improve our security:

- **Hall of Fame** listing on our website
- **Public acknowledgment** in security advisories
- **Swag and rewards** for significant findings (when budget allows)

## Security Contacts

### Security Team

- **Lead Security Engineer**: [truonglevinhphuc2006@gmail.com](mailto:truonglevinhphuc2006@gmail.com)
- **Development Team**: [truonglevinhphuc2006@gmail.com](mailto:truonglevinhphuc2006@gmail.com)

### Emergency Contact

For critical security issues requiring immediate attention:

- **Email**: [truonglevinhphuc2006@gmail.com](mailto:truonglevinhphuc2006@gmail.com) (monitored 24/7 for critical issues)
- **GitHub Security Advisory**: Use GitHub's private vulnerability reporting feature

## Security Resources

### Documentation

- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [Go Security Checklist](https://github.com/securego/gosec)
- [Docker Security Best Practices](https://docs.docker.com/engine/security/)

### Tools We Use

- **Static Analysis**: gosec, golangci-lint
- **Dependency Scanning**: Snyk, GitHub Dependabot
- **Container Scanning**: Trivy, Clair
- **Dynamic Testing**: OWASP ZAP

### Security Audits

We conduct regular security audits:

- **Internal audits**: Quarterly
- **External audits**: Annually
- **Penetration testing**: Bi-annually

## Compliance

GreenLedger is designed to comply with:

- **GDPR** (General Data Protection Regulation)
- **SOC 2 Type II** (in progress)
- **ISO 27001** (planned)

## Security Updates

### Notification Channels

- **Security Advisories**: GitHub Security Advisories
- **GitHub Releases**: Security updates included in release notes
- **GitHub Watch**: Subscribe to repository notifications for security updates

### Update Process

1. **Security patch** developed and tested
2. **Advisory published** with details
3. **Automated updates** for critical issues
4. **Manual updates** for non-critical issues

## Questions?

If you have questions about this security policy or GreenLedger's security practices, please contact us at [truonglevinhphuc2006@gmail.com](mailto:truonglevinhphuc2006@gmail.com).

---

**Last Updated**: May 2025
**Version**: 1.0
