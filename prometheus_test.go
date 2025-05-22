package prometheus

import (
	"context"
	"strings"
	"testing"
	"time"

	configjson "github.com/akademic/go-config-json"
)

func TestPrometheus_Populate(t *testing.T) {
	tests := []struct {
		name                string
		setupCounters       func(*Prometheus)
		expectedContains    []string
		expectedNotContains []string
	}{
		{
			name: "empty prometheus instance",
			setupCounters: func(prom *Prometheus) {
				// No counters added
			},
			expectedContains:    []string{},
			expectedNotContains: []string{"# HELP", "# TYPE"},
		},
		{
			name: "single counter without metadata",
			setupCounters: func(prom *Prometheus) {
				counter := prom.NewCounter("test_counter", "Test counter description", "counter", nil)
				counter.Set(42)
			},
			expectedContains: []string{
				"# HELP test_counter Test counter description",
				"# TYPE test_counter counter",
				"test_counter{project=\"test_project\"} 42",
			},
		},
		{
			name: "single counter with custom metadata",
			setupCounters: func(prom *Prometheus) {
				metadata := map[string]string{
					"instance": "server1",
					"job":      "webapp",
				}
				counter := prom.NewCounter("http_requests", "HTTP requests total", "counter", metadata)
				counter.Set(100)
			},
			expectedContains: []string{
				"# HELP http_requests HTTP requests total",
				"# TYPE http_requests counter",
				"http_requests{",
				"project=\"test_project\"",
				"instance=\"server1\"",
				"job=\"webapp\"",
				"} 100",
			},
		},
		{
			name: "multiple counters with same name",
			setupCounters: func(prom *Prometheus) {
				metadata1 := map[string]string{"instance": "server1"}
				metadata2 := map[string]string{"instance": "server2"}

				counter1 := prom.NewCounter("requests_total", "Total requests", "counter", metadata1)
				counter1.Set(50)

				counter2 := prom.NewCounter("requests_total", "Total requests", "counter", metadata2)
				counter2.Set(75)
			},
			expectedContains: []string{
				"# HELP requests_total Total requests",
				"# TYPE requests_total counter",
				"requests_total{instance=\"server1\", project=\"test_project\"} 50",
				"requests_total{instance=\"server2\", project=\"test_project\"} 75",
			},
		},
		{
			name: "multiple counters with different names",
			setupCounters: func(prom *Prometheus) {
				counter1 := prom.NewCounter("http_requests", "HTTP requests", "counter", nil)
				counter1.Set(10)

				counter2 := prom.NewCounter("database_queries", "Database queries", "counter", nil)
				counter2.Set(20)
			},
			expectedContains: []string{
				"# HELP http_requests HTTP requests",
				"# TYPE http_requests counter",
				"http_requests{project=\"test_project\"} 10",
				"# HELP database_queries Database queries",
				"# TYPE database_queries counter",
				"database_queries{project=\"test_project\"} 20",
			},
		},
		{
			name: "counter with zero value",
			setupCounters: func(prom *Prometheus) {
				prom.NewCounter("zero_counter", "Counter with zero value", "counter", nil)
				// Default value should be 0
			},
			expectedContains: []string{
				"# HELP zero_counter Counter with zero value",
				"# TYPE zero_counter counter",
				"zero_counter{project=\"test_project\"} 0",
			},
		},
		{
			name: "counter after increment operations",
			setupCounters: func(prom *Prometheus) {
				counter := prom.NewCounter("incremented_counter", "Incremented counter", "counter", nil)
				counter.Inc()
				counter.Inc()
				counter.IncC(5)
			},
			expectedContains: []string{
				"# HELP incremented_counter Incremented counter",
				"# TYPE incremented_counter counter",
				"incremented_counter{project=\"test_project\"} 7",
			},
		},
		{
			name: "counter with float value",
			setupCounters: func(prom *Prometheus) {
				counter := prom.NewCounter("float_counter", "Float counter", "gauge", nil)
				counter.SetF(42.5)
			},
			expectedContains: []string{
				"# HELP float_counter Float counter",
				"# TYPE float_counter gauge",
				"float_counter{project=\"test_project\"} 42.500000",
			},
		},
		{
			name: "counter with both int and float values - float takes precedence",
			setupCounters: func(prom *Prometheus) {
				counter := prom.NewCounter("mixed_counter", "Mixed counter", "gauge", nil)
				counter.Set(100)
				counter.SetF(3.14)
			},
			expectedContains: []string{
				"# HELP mixed_counter Mixed counter",
				"# TYPE mixed_counter gauge",
				"mixed_counter{project=\"test_project\"} 3.140000",
			},
		},
		{
			name: "counter with zero float value shows int value",
			setupCounters: func(prom *Prometheus) {
				counter := prom.NewCounter("zero_float_counter", "Zero float counter", "counter", nil)
				counter.Set(42)
				counter.SetF(0.0)
			},
			expectedContains: []string{
				"# HELP zero_float_counter Zero float counter",
				"# TYPE zero_float_counter counter",
				"zero_float_counter{project=\"test_project\"} 42",
			},
		},
		{
			name: "multiple counters with mixed value types",
			setupCounters: func(prom *Prometheus) {
				counter1 := prom.NewCounter("int_metric", "Integer metric", "counter", nil)
				counter1.Set(100)

				counter2 := prom.NewCounter("float_metric", "Float metric", "gauge", nil)
				counter2.SetF(99.99)
			},
			expectedContains: []string{
				"# HELP int_metric Integer metric",
				"# TYPE int_metric counter",
				"int_metric{project=\"test_project\"} 100",
				"# HELP float_metric Float metric",
				"# TYPE float_metric gauge",
				"float_metric{project=\"test_project\"} 99.990000",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{
				ProjectName:  "test_project",
				DumpPath:     "",
				DumpInterval: configjson.Duration{Duration: time.Second},
			}

			prom := NewPrometheus(context.Background(), cfg, &MockLogger{})

			tt.setupCounters(prom)

			// CALL TESTED FUNCTION
			result := prom.Populate()

			for _, expected := range tt.expectedContains {
				if !strings.Contains(result, expected) {
					t.Errorf("Expected result to contain %q, but it didn't. Result:\n%s", expected, result)
				}
			}

			for _, notExpected := range tt.expectedNotContains {
				if strings.Contains(result, notExpected) {
					t.Errorf("Expected result to NOT contain %q, but it did. Result:\n%s", notExpected, result)
				}
			}
		})
	}
}

func TestPrometheus_Populate_HeadersOnlyOnce(t *testing.T) {
	cfg := Config{
		ProjectName:  "test_project",
		DumpPath:     "",
		DumpInterval: configjson.Duration{Duration: time.Second},
	}

	prom := NewPrometheus(context.Background(), cfg, &MockLogger{})

	// Add multiple counters with same name
	counter1 := prom.NewCounter("same_name", "Description", "counter", map[string]string{"instance": "1"})
	counter1.Set(10)

	counter2 := prom.NewCounter("same_name", "Description", "counter", map[string]string{"instance": "2"})
	counter2.Set(20)

	result := prom.Populate()

	// Count occurrences of HELP and TYPE headers
	helpCount := strings.Count(result, "# HELP same_name")
	typeCount := strings.Count(result, "# TYPE same_name")

	if helpCount != 1 {
		t.Errorf("Expected exactly 1 HELP header for same_name, got %d", helpCount)
	}

	if typeCount != 1 {
		t.Errorf("Expected exactly 1 TYPE header for same_name, got %d", typeCount)
	}

	// Verify both counter values are present
	if !strings.Contains(result, "same_name{instance=\"1\", project=\"test_project\"} 10") {
		t.Error("Expected first counter value to be present")
	}

	if !strings.Contains(result, "same_name{instance=\"2\", project=\"test_project\"} 20") {
		t.Error("Expected second counter value to be present")
	}
}

func TestPrometheus_Populate_MetadataFormatting(t *testing.T) {
	cfg := Config{
		ProjectName:  "test_project",
		DumpPath:     "",
		DumpInterval: configjson.Duration{Duration: time.Second},
	}

	prom := NewPrometheus(context.Background(), cfg, &MockLogger{})

	metadata := map[string]string{
		"method":  "GET",
		"status":  "200",
		"handler": "index",
	}

	counter := prom.NewCounter("http_requests", "HTTP requests", "counter", metadata)
	counter.Set(42)

	result := prom.Populate()

	// Check that metadata is properly formatted with commas and quotes
	expectedParts := []string{
		"method=\"GET\"",
		"status=\"200\"",
		"handler=\"index\"",
		"project=\"test_project\"",
	}

	for _, part := range expectedParts {
		if !strings.Contains(result, part) {
			t.Errorf("Expected result to contain metadata part %q, but it didn't. Result:\n%s", part, result)
		}
	}

	// Check that metadata is wrapped in curly braces
	if !strings.Contains(result, "http_requests{") {
		t.Error("Expected metadata to be wrapped in curly braces")
	}

	if !strings.Contains(result, "} 42") {
		t.Error("Expected metadata to end with closing brace followed by value")
	}
}
