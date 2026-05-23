package metrics

import (
	"fmt"
	"log"
	"os"
	"sort"
	"time"

	"github.com/Giuse-04/load-tester/internal/worker"
	"github.com/schollz/progressbar/v3"
)

func Consume(totalReqs int, results <-chan worker.RequestResult) <-chan struct{} {
	done := make(chan struct{})
	go func() {
		defer close(done)

		file, err := os.OpenFile("salida.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to open log file: %v. Results will not be saved.\n", err)
			return
		}
		defer file.Close()

		log.SetOutput(file)
		log.SetFlags(log.LstdFlags)

		var latencies []time.Duration
		successCount := 0
		errorCount := 0
		statusCodeCount := make(map[int]int)

		bar := progressbar.Default(int64(totalReqs), "Processing results")

		for res := range results {
			bar.Add(1)
			if res.Error == nil && res.StatusCode >= 200 && res.StatusCode < 300 {
				successCount++
				latencies = append(latencies, res.Latency)
			} else {
				errorCount++
			}
			statusCodeCount[res.StatusCode]++
		}
		totalRequests := successCount + errorCount

		log.Printf("Total Requests: %d, Success: %d, Errors: %d\n", totalRequests, successCount, errorCount)
		log.Printf("Status Code Distribution: %v\n", statusCodeCount)

		if len(latencies) == 0 {
			log.Printf("No successful requests\n")
			fmt.Fprintf(os.Stderr, "\n=== Load Test Results ===\nTotal Requests: %d, Success: %d, Errors: %d\nStatus Codes: %v\n", totalRequests, successCount, errorCount, statusCodeCount)
			return
		}

		var totalLatency time.Duration
		for _, l := range latencies {
			totalLatency += l
		}
		avgLatency := totalLatency / time.Duration(len(latencies))

		sort.Slice(latencies, func(i, j int) bool {
			return latencies[i] < latencies[j]
		})
		p50 := latencies[(len(latencies)-1)/2]
		p90Idx := int(float64(len(latencies)-1) * 0.9)
		p99Idx := int(float64(len(latencies)-1) * 0.99)
		p90 := latencies[p90Idx]
		p99 := latencies[p99Idx]

		log.Printf("Average Latency: %v\n", avgLatency)
		log.Printf("P50 Latency: %v\n", p50)
		log.Printf("P90 Latency: %v\n", p90)
		log.Printf("P99 Latency: %v\n", p99)

		// Display results to user via stderr (doesn't interfere with progress bar)
		fmt.Fprintf(os.Stderr, "\n=== Load Test Results ===\nTotal Requests: %d, Success: %d, Errors: %d\nStatus Codes: %v\nAverage Latency: %v\nP50: %v | P90: %v | P99: %v\n", totalRequests, successCount, errorCount, statusCodeCount, avgLatency, p50, p90, p99)

	}()
	return done
}
