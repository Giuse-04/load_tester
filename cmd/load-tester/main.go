package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/Giuse-04/load-tester/internal/config"
	"github.com/Giuse-04/load-tester/internal/metrics"
	"github.com/Giuse-04/load-tester/internal/worker"
	"github.com/spf13/cobra"
)

var cfg = config.AttackConfig{
	URL:           "http://localhost:8080",
	Concurrency:   100,
	TotalRequests: 1000,
}

var (
	maxIdleConns        = 100
	maxIdleConnsPerHost = 2
	idleConnTimeout     = 30 * time.Second
	requestTimeout      = 10 * time.Second
)

var rootCmd = &cobra.Command{
	Use:   "load-tester",
	Short: "A simple load testing tool in Go",
	Long:  `A simple load testing tool in Go that can be used to test the performance of web applications.`,
	Run: func(cmd *cobra.Command, args []string) {
		runLoadTest(cfg)
	},
}

func init() {
	rootCmd.Flags().StringVarP(&cfg.URL, "url", "u", cfg.URL, "Target URL to attack")
	rootCmd.Flags().IntVarP(&cfg.Concurrency, "concurrency", "c", cfg.Concurrency, "Number of goroutines to use for sending requests")
	rootCmd.Flags().IntVarP(&cfg.TotalRequests, "requests", "n", cfg.TotalRequests, "Total number of requests to send")
	rootCmd.Flags().IntVar(&maxIdleConns, "max-idle-conns", maxIdleConns, "Maximum number of idle connections across all hosts")
	rootCmd.Flags().IntVar(&maxIdleConnsPerHost, "max-idle-conns-per-host", maxIdleConnsPerHost, "Maximum number of idle connections per host")
	rootCmd.Flags().DurationVar(&idleConnTimeout, "idle-conn-timeout", idleConnTimeout, "Idle connection timeout (e.g., 30s, 5m)")
	rootCmd.Flags().DurationVar(&requestTimeout, "request-timeout", requestTimeout, "Request timeout (e.g., 10s, 1m)")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runLoadTest(cfg config.AttackConfig) {
	// Validate inputs
	if err := validateConfig(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Invalid configuration: %v\n", err)
		os.Exit(1)
	}

	transport := &http.Transport{
		MaxIdleConns:        maxIdleConns,
		MaxIdleConnsPerHost: maxIdleConnsPerHost,
		IdleConnTimeout:     idleConnTimeout,
	}

	client := &http.Client{Transport: transport, Timeout: requestTimeout}

	tasks := make(chan config.Request, cfg.TotalRequests)
	results := make(chan worker.RequestResult, cfg.TotalRequests)

	wg := &sync.WaitGroup{}

	metricsDone := metrics.Consume(cfg.TotalRequests, results)

	fmt.Println("Starting load test...")
	for i := 0; i < cfg.Concurrency; i++ {
		wg.Add(1)
		go worker.Run(i, client, tasks, results, wg)
	}

	for i := 0; i < cfg.TotalRequests; i++ {
		tasks <- config.Request{
			ID:     i,
			Method: "GET",
			URL:    cfg.URL,
			Headers: map[string]string{
				"User-Agent": "LoadTester/1.0",
			},
			Body: nil,
		}
	}
	close(tasks) // Signal that no more tasks will be sent
	// Wait for all workers to finish processing
	wg.Wait()
	close(results) // Signal that no more results will be sent
	// Wait for metrics to finish consuming results
	<-metricsDone
}

func validateConfig(cfg config.AttackConfig) error {
	// Validate URL
	parsedURL, err := url.Parse(cfg.URL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	// Only allow http and https schemes
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("unsupported URL scheme '%s': only http and https are allowed", parsedURL.Scheme)
	}

	// Validate concurrency
	if cfg.Concurrency <= 0 {
		return fmt.Errorf("concurrency must be greater than 0, got %d", cfg.Concurrency)
	}

	// Validate total requests
	if cfg.TotalRequests <= 0 {
		return fmt.Errorf("total requests must be greater than 0, got %d", cfg.TotalRequests)
	}

	// Validate transport parameters
	if maxIdleConns < 0 {
		return fmt.Errorf("max-idle-conns must be >= 0, got %d", maxIdleConns)
	}

	if maxIdleConnsPerHost < 0 {
		return fmt.Errorf("max-idle-conns-per-host must be >= 0, got %d", maxIdleConnsPerHost)
	}

	if idleConnTimeout < 0 {
		return fmt.Errorf("idle-conn-timeout must be >= 0, got %v", idleConnTimeout)
	}

	if requestTimeout <= 0 {
		return fmt.Errorf("request-timeout must be > 0, got %v", requestTimeout)
	}

	return nil
}
