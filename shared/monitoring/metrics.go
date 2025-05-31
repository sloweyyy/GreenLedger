package monitoring

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics holds all Prometheus metrics
type Metrics struct {
	// HTTP metrics
	HTTPRequestsTotal     *prometheus.CounterVec
	HTTPRequestDuration   *prometheus.HistogramVec
	HTTPRequestsInFlight  prometheus.Gauge
	HTTPResponseSize      *prometheus.HistogramVec

	// Database metrics
	DBConnectionsOpen     prometheus.Gauge
	DBConnectionsIdle     prometheus.Gauge
	DBQueryDuration       *prometheus.HistogramVec
	DBQueriesTotal        *prometheus.CounterVec

	// Business metrics
	CalculationsTotal     *prometheus.CounterVec
	ActivitiesTotal       *prometheus.CounterVec
	CreditsEarned         *prometheus.CounterVec
	CreditsSpent          *prometheus.CounterVec
	TransactionsTotal     *prometheus.CounterVec
	CertificatesIssued    *prometheus.CounterVec
	ReportsGenerated      *prometheus.CounterVec

	// System metrics
	GoroutinesActive      prometheus.Gauge
	MemoryUsage           prometheus.Gauge
	CPUUsage              prometheus.Gauge
}

// NewMetrics creates a new metrics instance
func NewMetrics(serviceName string) *Metrics {
	return &Metrics{
		// HTTP metrics
		HTTPRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"service", "method", "endpoint", "status_code"},
		),
		HTTPRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "Duration of HTTP requests in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"service", "method", "endpoint", "status_code"},
		),
		HTTPRequestsInFlight: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "http_requests_in_flight",
				Help: "Number of HTTP requests currently being processed",
			},
		),
		HTTPResponseSize: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_response_size_bytes",
				Help:    "Size of HTTP responses in bytes",
				Buckets: []float64{100, 1000, 10000, 100000, 1000000},
			},
			[]string{"service", "method", "endpoint"},
		),

		// Database metrics
		DBConnectionsOpen: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "db_connections_open",
				Help: "Number of open database connections",
			},
		),
		DBConnectionsIdle: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "db_connections_idle",
				Help: "Number of idle database connections",
			},
		),
		DBQueryDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "db_query_duration_seconds",
				Help:    "Duration of database queries in seconds",
				Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1.0, 5.0},
			},
			[]string{"service", "operation", "table"},
		),
		DBQueriesTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "db_queries_total",
				Help: "Total number of database queries",
			},
			[]string{"service", "operation", "table", "status"},
		),

		// Business metrics
		CalculationsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "calculations_total",
				Help: "Total number of carbon footprint calculations",
			},
			[]string{"service", "activity_type", "status"},
		),
		ActivitiesTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "activities_total",
				Help: "Total number of eco activities logged",
			},
			[]string{"service", "activity_type", "source", "verified"},
		),
		CreditsEarned: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "credits_earned_total",
				Help: "Total carbon credits earned",
			},
			[]string{"service", "source"},
		),
		CreditsSpent: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "credits_spent_total",
				Help: "Total carbon credits spent",
			},
			[]string{"service", "purpose"},
		),
		TransactionsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "transactions_total",
				Help: "Total number of wallet transactions",
			},
			[]string{"service", "type", "status"},
		),
		CertificatesIssued: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "certificates_issued_total",
				Help: "Total number of certificates issued",
			},
			[]string{"service", "type", "project_type"},
		),
		ReportsGenerated: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "reports_generated_total",
				Help: "Total number of reports generated",
			},
			[]string{"service", "type", "format", "status"},
		),

		// System metrics
		GoroutinesActive: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "goroutines_active",
				Help: "Number of active goroutines",
			},
		),
		MemoryUsage: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "memory_usage_bytes",
				Help: "Memory usage in bytes",
			},
		),
		CPUUsage: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "cpu_usage_percent",
				Help: "CPU usage percentage",
			},
		),
	}
}

// PrometheusMiddleware creates a Gin middleware for Prometheus metrics
func (m *Metrics) PrometheusMiddleware(serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip metrics endpoint
		if c.Request.URL.Path == "/metrics" {
			c.Next()
			return
		}

		start := time.Now()
		m.HTTPRequestsInFlight.Inc()

		// Process request
		c.Next()

		// Record metrics
		duration := time.Since(start).Seconds()
		statusCode := strconv.Itoa(c.Writer.Status())
		
		labels := prometheus.Labels{
			"service":     serviceName,
			"method":      c.Request.Method,
			"endpoint":    c.FullPath(),
			"status_code": statusCode,
		}

		m.HTTPRequestsTotal.With(labels).Inc()
		m.HTTPRequestDuration.With(labels).Observe(duration)
		m.HTTPResponseSize.With(prometheus.Labels{
			"service":  serviceName,
			"method":   c.Request.Method,
			"endpoint": c.FullPath(),
		}).Observe(float64(c.Writer.Size()))

		m.HTTPRequestsInFlight.Dec()
	}
}

// MetricsHandler returns a Gin handler for the /metrics endpoint
func MetricsHandler() gin.HandlerFunc {
	h := promhttp.Handler()
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

// RecordCalculation records a carbon footprint calculation metric
func (m *Metrics) RecordCalculation(serviceName, activityType, status string) {
	m.CalculationsTotal.With(prometheus.Labels{
		"service":       serviceName,
		"activity_type": activityType,
		"status":        status,
	}).Inc()
}

// RecordActivity records an eco activity metric
func (m *Metrics) RecordActivity(serviceName, activityType, source string, verified bool) {
	m.ActivitiesTotal.With(prometheus.Labels{
		"service":       serviceName,
		"activity_type": activityType,
		"source":        source,
		"verified":      strconv.FormatBool(verified),
	}).Inc()
}

// RecordCreditsEarned records credits earned metric
func (m *Metrics) RecordCreditsEarned(serviceName, source string, amount float64) {
	m.CreditsEarned.With(prometheus.Labels{
		"service": serviceName,
		"source":  source,
	}).Add(amount)
}

// RecordCreditsSpent records credits spent metric
func (m *Metrics) RecordCreditsSpent(serviceName, purpose string, amount float64) {
	m.CreditsSpent.With(prometheus.Labels{
		"service": serviceName,
		"purpose": purpose,
	}).Add(amount)
}

// RecordTransaction records a wallet transaction metric
func (m *Metrics) RecordTransaction(serviceName, txType, status string) {
	m.TransactionsTotal.With(prometheus.Labels{
		"service": serviceName,
		"type":    txType,
		"status":  status,
	}).Inc()
}

// RecordCertificateIssued records a certificate issuance metric
func (m *Metrics) RecordCertificateIssued(serviceName, certType, projectType string) {
	m.CertificatesIssued.With(prometheus.Labels{
		"service":      serviceName,
		"type":         certType,
		"project_type": projectType,
	}).Inc()
}

// RecordReportGenerated records a report generation metric
func (m *Metrics) RecordReportGenerated(serviceName, reportType, format, status string) {
	m.ReportsGenerated.With(prometheus.Labels{
		"service": serviceName,
		"type":    reportType,
		"format":  format,
		"status":  status,
	}).Inc()
}

// RecordDBQuery records a database query metric
func (m *Metrics) RecordDBQuery(serviceName, operation, table, status string, duration time.Duration) {
	labels := prometheus.Labels{
		"service":   serviceName,
		"operation": operation,
		"table":     table,
		"status":    status,
	}

	m.DBQueriesTotal.With(labels).Inc()
	m.DBQueryDuration.With(prometheus.Labels{
		"service":   serviceName,
		"operation": operation,
		"table":     table,
	}).Observe(duration.Seconds())
}

// UpdateSystemMetrics updates system-level metrics
func (m *Metrics) UpdateSystemMetrics(goroutines int, memoryBytes uint64, cpuPercent float64) {
	m.GoroutinesActive.Set(float64(goroutines))
	m.MemoryUsage.Set(float64(memoryBytes))
	m.CPUUsage.Set(cpuPercent)
}

// UpdateDBConnectionMetrics updates database connection metrics
func (m *Metrics) UpdateDBConnectionMetrics(open, idle int) {
	m.DBConnectionsOpen.Set(float64(open))
	m.DBConnectionsIdle.Set(float64(idle))
}
