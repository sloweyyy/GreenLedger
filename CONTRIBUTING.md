# Contributing to GreenLedger

Thank you for your interest in contributing to GreenLedger! We welcome contributions from the community to help build a better carbon credit tracking system for a sustainable future.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Workflow](#development-workflow)
- [Code Style Guidelines](#code-style-guidelines)
- [Testing Requirements](#testing-requirements)
- [Commit Message Conventions](#commit-message-conventions)
- [Pull Request Process](#pull-request-process)
- [Issue Reporting](#issue-reporting)
- [Security Vulnerabilities](#security-vulnerabilities)

## Code of Conduct

This project and everyone participating in it is governed by our [Code of Conduct](CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code. Please report unacceptable behavior to [truonglevinhphuc2006@gmail.com](mailto:truonglevinhphuc2006@gmail.com).

## Getting Started

### Prerequisites

- Go 1.21 or higher
- Docker and Docker Compose
- Git
- Make (optional but recommended)

### Development Setup

1. **Fork the repository** on GitHub
2. **Clone your fork** locally:
   ```bash
   git clone https://github.com/YOUR_USERNAME/GreenLedger.git
   cd GreenLedger
   ```

3. **Add the upstream remote**:
   ```bash
   git remote add upstream https://github.com/sloweyyy/GreenLedger.git
   ```

4. **Set up the development environment**:
   ```bash
   make dev-setup
   ```

5. **Start the development environment**:
   ```bash
   make docker-up
   ```

6. **Verify the setup**:
   ```bash
   make test
   ```

## Development Workflow

### 1. Create a Feature Branch

Always create a new branch for your work:

```bash
git checkout main
git pull upstream main
git checkout -b feature/your-feature-name
```

### 2. Branch Naming Conventions

Use descriptive branch names with prefixes:

- `feature/` - New features
- `bugfix/` - Bug fixes
- `hotfix/` - Critical fixes
- `docs/` - Documentation updates
- `refactor/` - Code refactoring
- `test/` - Test improvements

Examples:
- `feature/carbon-offset-certificates`
- `bugfix/wallet-balance-calculation`
- `docs/api-documentation-update`

### 3. Make Your Changes

- Write clean, readable code following our style guidelines
- Add tests for new functionality
- Update documentation as needed
- Ensure all tests pass locally

### 4. Commit Your Changes

Follow our [commit message conventions](#commit-message-conventions):

```bash
git add .
git commit -m "feat(wallet): add credit transfer functionality"
```

### 5. Push and Create Pull Request

```bash
git push origin feature/your-feature-name
```

Then create a pull request on GitHub using our [PR template](.github/pull_request_template.md).

## Code Style Guidelines

### Go Code Style

We follow the standard Go conventions with additional rules:

#### 1. **Formatting**
- Use `gofmt` for code formatting
- Use `goimports` for import organization
- Line length should not exceed 120 characters

#### 2. **Naming Conventions**
- Use camelCase for variables and functions
- Use PascalCase for exported functions and types
- Use descriptive names (avoid abbreviations)
- Constants should be in ALL_CAPS with underscores

#### 3. **Package Organization**
- Each service should be in its own package under `services/`
- Shared code goes in the `shared/` package
- Use meaningful package names that reflect functionality

#### 4. **Error Handling**
- Always handle errors explicitly
- Use wrapped errors with context: `fmt.Errorf("operation failed: %w", err)`
- Log errors at appropriate levels

#### 5. **Documentation**
- All exported functions and types must have comments
- Use godoc format for documentation
- Include examples for complex functions

### Linting Rules

We use the following linters (configured in `.golangci.yml`):

- `golint` - Go style checker
- `govet` - Go static analysis
- `errcheck` - Check for unchecked errors
- `staticcheck` - Advanced static analysis
- `gosec` - Security checker
- `misspell` - Spell checker

Run linting locally:
```bash
golangci-lint run ./...
```

### Database Guidelines

- Use GORM for database operations
- Always use transactions for multi-step operations
- Include proper indexes in migration files
- Use meaningful table and column names
- Include foreign key constraints where appropriate

### API Guidelines

- Follow RESTful conventions
- Use proper HTTP status codes
- Include comprehensive Swagger documentation
- Validate all input parameters
- Use consistent error response format

## Testing Requirements

### Minimum Coverage

- **Unit tests**: Minimum 80% code coverage
- **Integration tests**: Cover all API endpoints
- **Load tests**: Performance benchmarks for critical paths

### Testing Structure

```
services/
├── calculator/
│   ├── internal/
│   │   ├── handler/
│   │   │   ├── calculator_handler.go
│   │   │   └── calculator_handler_test.go
│   │   ├── service/
│   │   │   ├── calculator_service.go
│   │   │   └── calculator_service_test.go
│   │   └── repository/
│   │       ├── calculator_repository.go
│   │       └── calculator_repository_test.go
│   └── integration_test.go
```

### Test Guidelines

1. **Unit Tests**
   - Test individual functions and methods
   - Use table-driven tests for multiple scenarios
   - Mock external dependencies
   - Test both success and error cases

2. **Integration Tests**
   - Test complete API workflows
   - Use test databases
   - Test service interactions

3. **Load Tests**
   - Test performance under load
   - Verify system behavior at scale
   - Include in CI/CD pipeline

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run specific service tests
cd services/calculator && go test ./...

# Run load tests
make load-test
```

## Commit Message Conventions

We follow the [Conventional Commits](https://www.conventionalcommits.org/) specification:

### Format

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

### Types

- `feat` - New feature
- `fix` - Bug fix
- `docs` - Documentation changes
- `style` - Code style changes (formatting, etc.)
- `refactor` - Code refactoring
- `test` - Adding or updating tests
- `chore` - Maintenance tasks
- `perf` - Performance improvements
- `ci` - CI/CD changes

### Scopes

Use service names or component names:
- `calculator` - Calculator service
- `tracker` - Activity tracker service
- `wallet` - Wallet service
- `auth` - Authentication service
- `reporting` - Reporting service
- `api` - API changes
- `db` - Database changes
- `docker` - Docker/deployment changes

### Examples

```bash
feat(wallet): add credit transfer between users
fix(calculator): correct CO2 calculation for electric vehicles
docs(api): update swagger documentation for tracker endpoints
test(wallet): add integration tests for transaction processing
refactor(auth): improve JWT token validation logic
chore(deps): update Go dependencies to latest versions
```

### Breaking Changes

For breaking changes, add `BREAKING CHANGE:` in the footer:

```
feat(api): change response format for calculation endpoints

BREAKING CHANGE: The calculation response now includes additional metadata fields
```

## Pull Request Process

### Before Submitting

1. **Sync with upstream**:
   ```bash
   git checkout main
   git pull upstream main
   git checkout your-feature-branch
   git rebase main
   ```

2. **Run all checks**:
   ```bash
   make test
   make test-coverage
   golangci-lint run ./...
   ```

3. **Update documentation** if needed

4. **Squash commits** if necessary to maintain clean history

### PR Requirements

- [ ] All tests pass
- [ ] Code coverage meets minimum requirements (80%)
- [ ] Linting passes without errors
- [ ] Documentation is updated
- [ ] Breaking changes are documented
- [ ] Security implications are considered

### Review Process

1. **Automated checks** must pass (CI/CD pipeline)
2. **Code review** by at least one maintainer
3. **Security review** for security-related changes
4. **Performance review** for performance-critical changes

### Merging

- Use "Squash and merge" for feature branches
- Use "Rebase and merge" for hotfixes
- Delete feature branches after merging

## Issue Reporting

### Before Creating an Issue

1. **Search existing issues** to avoid duplicates
2. **Check the FAQ** and documentation
3. **Try the latest version** to see if the issue is already fixed

### Issue Types

Use our issue templates:

- [Bug Report](.github/ISSUE_TEMPLATE/bug_report.md)
- [Feature Request](.github/ISSUE_TEMPLATE/feature_request.md)
- [Documentation Improvement](.github/ISSUE_TEMPLATE/documentation.md)

### Issue Labels

We use labels to categorize issues:

- `bug` - Something isn't working
- `enhancement` - New feature or request
- `documentation` - Improvements or additions to documentation
- `good first issue` - Good for newcomers
- `help wanted` - Extra attention is needed
- `priority/high` - High priority
- `service/calculator` - Calculator service related
- `service/wallet` - Wallet service related

## Security Vulnerabilities

**Do not report security vulnerabilities through public GitHub issues.**

Please report security vulnerabilities to [truonglevinhphuc2006@gmail.com](mailto:truonglevinhphuc2006@gmail.com). See our [Security Policy](SECURITY.md) for more details.

## Getting Help

- **Documentation**: Check our [docs](docs/) directory
- **Discussions**: Use GitHub Discussions for questions
- **Chat**: Join our community chat (link coming soon)
- **Email**: Contact us at [truonglevinhphuc2006@gmail.com](mailto:truonglevinhphuc2006@gmail.com)

## Recognition

Contributors will be recognized in our [CONTRIBUTORS.md](CONTRIBUTORS.md) file and release notes.

Thank you for contributing to a more sustainable future! 🌱
