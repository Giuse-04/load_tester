package worker

import (
	"bytes"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/Giuse-04/load-tester/internal/config"
)

type RequestResult struct {
	StatusCode int
	Latency    time.Duration
	Error      error
}

func Run(id int, client *http.Client, tasks <-chan config.Request, results chan<- RequestResult, wg *sync.WaitGroup) {
	defer wg.Done()
	for req := range tasks {
		start := time.Now()
		httpReq, err := http.NewRequest(req.Method, req.URL, bytes.NewReader(req.Body))
		if err != nil {
			results <- RequestResult{StatusCode: 0, Latency: time.Since(start), Error: err}
			continue
		}
		for k, v := range req.Headers {
			httpReq.Header.Set(k, v)
		}
		resp, err := client.Do(httpReq)
		latency := time.Since(start)
		if err != nil {
			results <- RequestResult{StatusCode: 0, Latency: latency, Error: err}
			continue
		}
		_, err = io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
		if err != nil {
			results <- RequestResult{StatusCode: 0, Latency: latency, Error: err}
			continue
		}
		results <- RequestResult{StatusCode: resp.StatusCode, Latency: latency, Error: nil}
	}
}
