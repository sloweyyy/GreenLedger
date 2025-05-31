---
name: ⚡ Performance Issue
about: Report a performance problem or suggest performance improvements
title: '[PERF] '
labels: ['performance', 'bug']
assignees: ''
---

# ⚡ Performance Issue

## 📋 Performance Problem Type

<!-- Mark the relevant option with an "x" -->

- [ ] 🐌 Slow response times
- [ ] 💾 High memory usage
- [ ] 🔥 High CPU usage
- [ ] 💽 Disk I/O issues
- [ ] 🌐 Network latency
- [ ] 🗄️ Database performance
- [ ] 📊 Inefficient algorithms
- [ ] 🔄 Memory leaks
- [ ] 📈 Scaling issues
- [ ] 🎯 Other: ___________

## 🎯 Affected Service(s)

<!-- Mark all affected services -->

- [ ] 🧮 Calculator Service
- [ ] 📊 Tracker Service
- [ ] 💰 Wallet Service
- [ ] 🔐 User Auth Service
- [ ] 📈 Reporting Service
- [ ] 🏆 Certificate Service
- [ ] 🌐 API Gateway
- [ ] 🗄️ Database
- [ ] 🔄 Message Queue (Kafka)
- [ ] 💾 Cache (Redis)

## 📊 Performance Metrics

### Current Performance

**Response Time:** 
<!-- e.g., 2.5 seconds average -->

**Throughput:** 
<!-- e.g., 100 requests/second -->

**Resource Usage:**
- CPU: ___% average, ___% peak
- Memory: ___MB average, ___MB peak
- Disk I/O: ___MB/s read, ___MB/s write
- Network: ___MB/s

### Expected Performance

**Target Response Time:** 
<!-- e.g., < 500ms -->

**Target Throughput:** 
<!-- e.g., 1000 requests/second -->

**Target Resource Usage:**
- CPU: < ___%
- Memory: < ___MB
- Disk I/O: < ___MB/s
- Network: < ___MB/s

## 🔍 Steps to Reproduce

<!-- Provide detailed steps to reproduce the performance issue -->

1. 
2. 
3. 
4. 

### Test Environment

**Environment:** 
<!-- e.g., local development, staging, production -->

**Load Conditions:**
- Concurrent users: ___
- Request rate: ___ requests/second
- Data volume: ___
- Duration: ___

## 📈 Performance Data

### Monitoring Screenshots

<!-- Include screenshots from monitoring tools (Grafana, Prometheus, etc.) -->

### Profiling Data

<!-- Include profiling data if available -->

```
# CPU profiling
go tool pprof cpu.prof

# Memory profiling
go tool pprof mem.prof

# Trace data
go tool trace trace.out
```

### Database Query Performance

<!-- If database-related, include slow query logs -->

```sql
-- Slow queries
EXPLAIN ANALYZE SELECT ...;
```

### Load Test Results

<!-- Include load test results if available -->

```bash
# Load test command
make load-test

# Results
Average response time: ___ms
95th percentile: ___ms
99th percentile: ___ms
Error rate: ___%
```

## 🔧 Environment Details

**Operating System:** 
<!-- e.g., Ubuntu 22.04, macOS 13.0, Windows 11 -->

**Go Version:** 
<!-- e.g., 1.21.0 -->

**Docker Version:** 
<!-- e.g., 24.0.0 -->

**Hardware:**
- CPU: ___
- RAM: ___GB
- Storage: ___
- Network: ___

**Container Resources:**
- CPU limit: ___
- Memory limit: ___MB
- Storage: ___

## 💡 Potential Solutions

<!-- If you have ideas for fixing the performance issue -->

### Suspected Root Cause

<!-- What do you think is causing the performance issue? -->

### Proposed Solutions

1. 
2. 
3. 

### Alternative Approaches

<!-- Any alternative solutions to consider -->

## 📊 Impact Assessment

### Business Impact

- [ ] 🔥 Critical (service unusable)
- [ ] ⚡ High (significant user impact)
- [ ] 📋 Medium (noticeable degradation)
- [ ] 🔍 Low (minor impact)

### User Experience Impact

<!-- How does this affect users? -->

### Cost Impact

<!-- Any cost implications (infrastructure, etc.) -->

## 🧪 Testing Strategy

### Performance Testing Plan

<!-- How should we test the fix? -->

1. 
2. 
3. 

### Success Criteria

<!-- How will we know the issue is resolved? -->

- [ ] Response time < ___ms
- [ ] Throughput > ___ requests/second
- [ ] CPU usage < ___%
- [ ] Memory usage < ___MB
- [ ] Error rate < ___%

## 🔗 Related Issues

<!-- Link any related performance issues -->

- Related to #
- Blocks #
- Blocked by #

## 📝 Additional Context

<!-- Any other relevant information -->

### Recent Changes

<!-- Any recent changes that might have caused this issue -->

### Monitoring Alerts

<!-- Any alerts that have been triggered -->

### Similar Issues

<!-- Any similar issues you've encountered -->

---

**Thank you for helping improve GreenLedger's performance! ⚡**
