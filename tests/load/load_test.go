package load

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// LoadTestConfig represents configuration for load testing
type LoadTestConfig struct {
	BaseURL         string
	Duration        time.Duration
	ConcurrentUsers int
	RampUpTime      time.Duration
	TestScenarios   []TestScenario
}

// TestScenario represents a test scenario
type TestScenario struct {
	Name       string
	Weight     int // Percentage of requests
	Endpoint   string
	Method     string
	Headers    map[string]string
	Body       interface{}
	Validation func(*http.Response) error
}

// LoadTestResult represents the results of a load test
type LoadTestResult struct {
	TotalRequests  int64
	SuccessfulReqs int64
	FailedReqs     int64
	AverageLatency time.Duration
	MinLatency     time.Duration
	MaxLatency     time.Duration
	P95Latency     time.Duration
	P99Latency     time.Duration
	RequestsPerSec float64
	ErrorRate      float64
	Errors         map[string]int64
}

// LoadTester handles load testing
type LoadTester struct {
	config     LoadTestConfig
	httpClient *http.Client
	results    *LoadTestResult
	latencies  []time.Duration
	errors     map[string]int64
	mutex      sync.RWMutex
}

// NewLoadTester creates a new load tester
func NewLoadTester(config LoadTestConfig) *LoadTester {
	return &LoadTester{
		config: config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		results: &LoadTestResult{
			Errors: make(map[string]int64),
		},
		latencies: make([]time.Duration, 0),
		errors:    make(map[string]int64),
	}
}

// Run executes the load test
func (lt *LoadTester) Run(ctx context.Context) (*LoadTestResult, error) {
	fmt.Printf("Starting load test with %d concurrent users for %v\n",
		lt.config.ConcurrentUsers, lt.config.Duration)

	startTime := time.Now()
	endTime := startTime.Add(lt.config.Duration)

	// Create worker pool
	var wg sync.WaitGroup
	userChan := make(chan int, lt.config.ConcurrentUsers)

	// Start workers
	for i := 0; i < lt.config.ConcurrentUsers; i++ {
		wg.Add(1)
		go lt.worker(ctx, &wg, userChan, endTime)

		// Ramp up gradually
		if lt.config.RampUpTime > 0 {
			time.Sleep(lt.config.RampUpTime / time.Duration(lt.config.ConcurrentUsers))
		}
	}

	// Send user IDs
	for i := 0; i < lt.config.ConcurrentUsers; i++ {
		userChan <- i
	}
	close(userChan)

	// Wait for all workers to complete
	wg.Wait()

	// Calculate results
	lt.calculateResults(time.Since(startTime))

	fmt.Printf("Load test completed. Results:\n")
	fmt.Printf("Total Requests: %d\n", lt.results.TotalRequests)
	fmt.Printf("Successful: %d\n", lt.results.SuccessfulReqs)
	fmt.Printf("Failed: %d\n", lt.results.FailedReqs)
	fmt.Printf("Error Rate: %.2f%%\n", lt.results.ErrorRate)
	fmt.Printf("Average Latency: %v\n", lt.results.AverageLatency)
	fmt.Printf("P95 Latency: %v\n", lt.results.P95Latency)
	fmt.Printf("P99 Latency: %v\n", lt.results.P99Latency)
	fmt.Printf("Requests/sec: %.2f\n", lt.results.RequestsPerSec)

	return lt.results, nil
}

// worker represents a single user making requests
func (lt *LoadTester) worker(ctx context.Context, wg *sync.WaitGroup, userChan <-chan int, endTime time.Time) {
	defer wg.Done()

	userID := <-userChan

	for time.Now().Before(endTime) {
		select {
		case <-ctx.Done():
			return
		default:
			scenario := lt.selectScenario()
			lt.executeRequest(userID, scenario)

			// Small delay between requests
			time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
		}
	}
}

// selectScenario selects a test scenario based on weights
func (lt *LoadTester) selectScenario() TestScenario {
	if len(lt.config.TestScenarios) == 0 {
		return TestScenario{}
	}

	totalWeight := 0
	for _, scenario := range lt.config.TestScenarios {
		totalWeight += scenario.Weight
	}

	random := rand.Intn(totalWeight)
	currentWeight := 0

	for _, scenario := range lt.config.TestScenarios {
		currentWeight += scenario.Weight
		if random < currentWeight {
			return scenario
		}
	}

	return lt.config.TestScenarios[0]
}

// executeRequest executes a single HTTP request
func (lt *LoadTester) executeRequest(userID int, scenario TestScenario) {
	startTime := time.Now()

	// Prepare request
	var body []byte
	if scenario.Body != nil {
		var err error
		body, err = json.Marshal(scenario.Body)
		if err != nil {
			lt.recordError("marshal_error")
			return
		}
	}

	url := lt.config.BaseURL + scenario.Endpoint
	req, err := http.NewRequest(scenario.Method, url, bytes.NewBuffer(body))
	if err != nil {
		lt.recordError("request_creation_error")
		return
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", fmt.Sprintf("LoadTester-User-%d", userID))

	for key, value := range scenario.Headers {
		req.Header.Set(key, value)
	}

	// Execute request
	resp, err := lt.httpClient.Do(req)
	if err != nil {
		lt.recordError("network_error")
		return
	}
	defer resp.Body.Close()

	latency := time.Since(startTime)

	// Record metrics
	lt.mutex.Lock()
	lt.results.TotalRequests++
	lt.latencies = append(lt.latencies, latency)
	lt.mutex.Unlock()

	// Validate response
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		if scenario.Validation != nil {
			if err := scenario.Validation(resp); err != nil {
				lt.recordError("validation_error")
				return
			}
		}
		lt.mutex.Lock()
		lt.results.SuccessfulReqs++
		lt.mutex.Unlock()
	} else {
		lt.recordError(fmt.Sprintf("http_%d", resp.StatusCode))
	}
}

// recordError records an error
func (lt *LoadTester) recordError(errorType string) {
	lt.mutex.Lock()
	defer lt.mutex.Unlock()

	lt.results.FailedReqs++
	lt.errors[errorType]++
	lt.results.Errors[errorType]++
}

// calculateResults calculates final test results
func (lt *LoadTester) calculateResults(totalDuration time.Duration) {
	lt.mutex.Lock()
	defer lt.mutex.Unlock()

	if len(lt.latencies) == 0 {
		return
	}

	// Sort latencies for percentile calculations
	for i := 0; i < len(lt.latencies)-1; i++ {
		for j := i + 1; j < len(lt.latencies); j++ {
			if lt.latencies[i] > lt.latencies[j] {
				lt.latencies[i], lt.latencies[j] = lt.latencies[j], lt.latencies[i]
			}
		}
	}

	// Calculate statistics
	var totalLatency time.Duration
	for _, latency := range lt.latencies {
		totalLatency += latency
	}

	lt.results.AverageLatency = totalLatency / time.Duration(len(lt.latencies))
	lt.results.MinLatency = lt.latencies[0]
	lt.results.MaxLatency = lt.latencies[len(lt.latencies)-1]

	// Calculate percentiles
	p95Index := int(float64(len(lt.latencies)) * 0.95)
	p99Index := int(float64(len(lt.latencies)) * 0.99)

	if p95Index < len(lt.latencies) {
		lt.results.P95Latency = lt.latencies[p95Index]
	}
	if p99Index < len(lt.latencies) {
		lt.results.P99Latency = lt.latencies[p99Index]
	}

	// Calculate rates
	lt.results.RequestsPerSec = float64(lt.results.TotalRequests) / totalDuration.Seconds()
	if lt.results.TotalRequests > 0 {
		lt.results.ErrorRate = float64(lt.results.FailedReqs) / float64(lt.results.TotalRequests) * 100
	}
}

// TestGreenLedgerLoadTest runs a comprehensive load test for GreenLedger
func TestGreenLedgerLoadTest(t *testing.T) {
	config := LoadTestConfig{
		BaseURL:         "http://localhost:8080/api/v1",
		Duration:        2 * time.Minute,
		ConcurrentUsers: 50,
		RampUpTime:      30 * time.Second,
		TestScenarios: []TestScenario{
			{
				Name:     "Calculate Carbon Footprint",
				Weight:   30,
				Endpoint: "/calculator/calculate",
				Method:   "POST",
				Headers: map[string]string{
					"Authorization": "Bearer test-token",
				},
				Body: map[string]interface{}{
					"activity_type": "vehicle",
					"distance":      100.0,
					"fuel_type":     "gasoline",
				},
				Validation: func(resp *http.Response) error {
					if resp.StatusCode != 200 {
						return fmt.Errorf("expected 200, got %d", resp.StatusCode)
					}
					return nil
				},
			},
			{
				Name:     "Log Eco Activity",
				Weight:   25,
				Endpoint: "/tracker/activities",
				Method:   "POST",
				Headers: map[string]string{
					"Authorization": "Bearer test-token",
				},
				Body: map[string]interface{}{
					"activity_type": "biking",
					"distance":      10.0,
					"description":   "Biked to work",
				},
			},
			{
				Name:     "Get Wallet Balance",
				Weight:   20,
				Endpoint: "/wallet/balance",
				Method:   "GET",
				Headers: map[string]string{
					"Authorization": "Bearer test-token",
				},
			},
			{
				Name:     "Get User Activities",
				Weight:   15,
				Endpoint: "/tracker/activities",
				Method:   "GET",
				Headers: map[string]string{
					"Authorization": "Bearer test-token",
				},
			},
			{
				Name:     "Generate Report",
				Weight:   10,
				Endpoint: "/reporting/reports",
				Method:   "POST",
				Headers: map[string]string{
					"Authorization": "Bearer test-token",
				},
				Body: map[string]interface{}{
					"type":       "summary",
					"format":     "json",
					"start_date": "2024-01-01T00:00:00Z",
					"end_date":   "2024-12-31T23:59:59Z",
				},
			},
		},
	}

	loadTester := NewLoadTester(config)
	ctx := context.Background()

	result, err := loadTester.Run(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Performance assertions
	assert.Less(t, result.ErrorRate, 5.0, "Error rate should be less than 5%")
	assert.Less(t, result.P95Latency, 2*time.Second, "P95 latency should be less than 2 seconds")
	assert.Greater(t, result.RequestsPerSec, 10.0, "Should handle at least 10 requests per second")
}
