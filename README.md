# Load Tester

Simple and configurable HTTP load testing tool written in Go.

## Features

- Concurrent HTTP requests with configurable concurrency level
- Detailed metrics including latency percentiles (P50, P90, P99)
- Status code distribution tracking
- Real-time progress bar
- Configurable HTTP client settings (timeouts, connection pooling)
- Input validation and error handling
- Results logged to file with stderr display

## Installation

```bash
git clone https://github.com/Giuse-04/load-tester.git
cd load-tester
go build ./cmd/load-tester
```

## Usage

### Basic Usage

```bash
./load-tester -u https://httpbin.org/get -n 100 -c 10
```

### Options

```
-u, --url string                     Target URL to test (default: http://localhost:8080)
-n, --requests int                   Total number of requests (default: 1000)
-c, --concurrency int                Number of concurrent goroutines (default: 100)
--max-idle-conns int                 Max idle connections across all hosts (default: 100)
--max-idle-conns-per-host int        Max idle connections per host (default: 2)
--idle-conn-timeout duration         Idle connection timeout (default: 30s)
--request-timeout duration           Request timeout (default: 10s)
```

### Examples

```bash
# Test against local server with 50 concurrent workers
./load-tester -u http://localhost:8080 -n 1000 -c 50

# Test with custom timeouts
./load-tester -u https://example.com -n 500 -c 25 --request-timeout 20s

# Test with custom connection pooling
./load-tester -u https://api.example.com -n 10000 -c 100 \
  --max-idle-conns 200 --max-idle-conns-per-host 5
```

## Output

Results are displayed in real-time with a progress bar and saved to `salida.log`:

```
Processing results |████████████████████| 100%

=== Load Test Results ===
Total Requests: 100, Success: 100, Errors: 0
Status Codes: map[200:100]
Average Latency: 125.5ms
P50: 120ms | P90: 145ms | P99: 150ms
```

Status codes are tracked to help diagnose issues:
- `2xx`: Success
- `4xx`: Client errors (bad request, not found, etc)
- `5xx`: Server errors
- `0`: Network/connection errors (timeout, connection refused, etc)

## Project Structure

```
load-tester/
├── cmd/load-tester/
│   └── main.go           # CLI entry point with argument parsing
├── internal/
│   ├── config/           # Configuration structures
│   ├── metrics/          # Results collection and reporting
│   └── worker/           # HTTP request execution
├── go.mod
└── go.sum
```

## Building

```bash
go build ./cmd/load-tester
```

## Requirements

- Go 1.26.1 or later

## Dependencies

- `github.com/spf13/cobra` - CLI framework
- `github.com/schollz/progressbar/v3` - Progress bar display
