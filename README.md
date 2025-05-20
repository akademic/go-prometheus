# Go Prometheus Metrics Library

A lightweight Prometheus metrics collection and export library for Go applications. This library allows you to easily create and manage metrics, collect Go runtime statistics, and export metrics in Prometheus text format.

## Features

- Simple counter metrics with metadata support
- Automatic Go runtime metrics collection
- Configurable metrics export to file
- Project name labeling for all metrics

## Installation

```bash
go get github.com/akademic/go-prometheus
```

## Dependencies

This library requires:

- [github.com/akademic/go-config-json](https://github.com/akademic/go-config-json) - For configuration handling
- [github.com/akademic/go-logger2](https://github.com/akademic/go-logger2) - For logging

## Basic Usage

### Creating a Prometheus instance

```go
package main

import (
	"context"
	"os"
	"time"

	"github.com/akademic/go-prometheus"
	"github.com/akademic/go-logger2"
	configjson "github.com/akademic/go-config-json"
)

func main() {
	// Create context
	ctx := context.Background()

	// Set up logger
	baseLogger := log.New(os.Stdout, "", log.LstdFlags)
	logConfig := &logger.Config{
		Level: logger.LogDebug,
	}
	log := logger.New(baseLogger, "prometheus", logConfig)

	// Configure prometheus
	cfg := prometheus.Config{
		ProjectName:  "my-project",
		DumpInterval: configjson.Duration{Duration: 30 * time.Second},
		DumpPath:     "/tmp/metrics.prom",
	}

	// Create prometheus instance
	prom := prometheus.NewPrometheus(ctx, cfg, log)

	// Start prometheus metrics collection in a separate goroutine
	go prom.Start()

	// Create and use counters...

	// Application shutdown
	defer prom.Stop()
}
```

### Creating and using counters

```go
// Create a simple counter
requestCounter := prom.NewCounter(
	"http_requests_total",
	"Total number of HTTP requests",
	"counter",
	map[string]string{"service": "api"}
)

// Increment counter
requestCounter.Inc()

// Increment counter by specific value
requestCounter.IncC(10)

// Set counter to specific value
requestCounter.Set(100)
```

### Collecting Go runtime metrics

```go
// Create app metrics collector
appMetrics := prometheus.NewAppMetrics(prom, ctx)

// Start collecting metrics every 5 seconds in a separate goroutine
go appMetrics.Start(5 * time.Second)

// Application shutdown
defer appMetrics.Stop()
```

### Manual metrics export

```go
// Get metrics in Prometheus text format
metricsText := prom.Populate()
fmt.Println(metricsText)

// Output example:
// # HELP http_requests_total Total number of HTTP requests
// # TYPE http_requests_total counter
// http_requests_total{project="my-project", service="api"} 100
// # HELP go_runtime_goroutines_count Count of goroutines
// # TYPE go_runtime_goroutines_count gauge
// go_runtime_goroutines_count{project="my-project"} 8
// ...
```
