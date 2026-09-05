package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Config struct {
	BaseURL     string
	TotalReqs   int
	Concurrency int
	Scenario    string // "plans", "auth_login", "health", "mixed", "subscribe"
	Timeout     time.Duration
}

type FailureDetail struct {
	Timestamp    time.Time `json:"timestamp"`
	Endpoint     string    `json:"endpoint"`
	Method       string    `json:"method"`
	StatusCode   int       `json:"status_code"`
	ErrorMessage string    `json:"error_message"`
	ResponseBody string    `json:"response_body,omitempty"`
	DurationMs   float64   `json:"duration_ms"`
}

type FailureCategory struct {
	Endpoint     string `json:"endpoint"`
	Method       string `json:"method"`
	StatusCode   int    `json:"status_code"`
	Reason       string `json:"reason"`
	Count        int64  `json:"count"`
	SampleBody   string `json:"sample_body"`
}

type Result struct {
	TotalRequests    int                         `json:"total_requests"`
	SuccessCount     int64                       `json:"success_count"`
	ErrorCount       int64                       `json:"error_count"`
	StatusCodes      map[int]int64               `json:"status_codes"`
	TotalDurationSec float64                     `json:"total_duration_sec"`
	RPS              float64                     `json:"rps"`
	AvgLatencyMs     float64                     `json:"avg_latency_ms"`
	MinLatencyMs     float64                     `json:"min_latency_ms"`
	P50LatencyMs     float64                     `json:"p50_latency_ms"`
	P90LatencyMs     float64                     `json:"p90_latency_ms"`
	P95LatencyMs     float64                     `json:"p95_latency_ms"`
	P99LatencyMs     float64                     `json:"p99_latency_ms"`
	MaxLatencyMs     float64                     `json:"max_latency_ms"`
	FailureBreakdown []FailureCategory           `json:"failure_breakdown"`
	DetailedFailures []FailureDetail             `json:"detailed_failures_sample"`
}

func createHTTPClient(concurrency int, timeout time.Duration) *http.Client {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:        concurrency * 4,
		MaxIdleConnsPerHost: concurrency * 4,
		MaxConnsPerHost:     concurrency * 4,
		IdleConnTimeout:     90 * time.Second,
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
		DisableCompression: false,
	}

	return &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}
}

type RequestResult struct {
	Elapsed      time.Duration
	StatusCode   int
	Endpoint     string
	Method       string
	Err          error
	ResponseBody string
}

func runWorker(
	id int,
	cfg Config,
	client *http.Client,
	jobs <-chan int,
	results chan<- RequestResult,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	for jobID := range jobs {
		res := executeRequest(client, cfg, jobID)
		results <- res
	}
}

func executeRequest(client *http.Client, cfg Config, jobID int) RequestResult {
	var req *http.Request
	var endpoint string
	var method string

	switch cfg.Scenario {
	case "plans":
		endpoint = "/api/v1/plans"
		method = "GET"
		req, _ = http.NewRequest("GET", cfg.BaseURL+endpoint, nil)

	case "health":
		endpoint = "/health"
		method = "GET"
		req, _ = http.NewRequest("GET", cfg.BaseURL+endpoint, nil)

	case "auth_login":
		endpoint = "/api/v1/auth/login"
		method = "POST"
		body := bytes.NewBufferString(`{"email":"admin@example.com","password":"password"}`)
		req, _ = http.NewRequest("POST", cfg.BaseURL+endpoint, body)
		req.Header.Set("Content-Type", "application/json")

	case "mixed":
		r := rand.Intn(100)
		if r < 60 {
			endpoint = "/api/v1/plans"
			method = "GET"
			req, _ = http.NewRequest("GET", cfg.BaseURL+endpoint, nil)
		} else if r < 85 {
			endpoint = "/health"
			method = "GET"
			req, _ = http.NewRequest("GET", cfg.BaseURL+endpoint, nil)
		} else {
			endpoint = "/api/v1/auth/login"
			method = "POST"
			body := bytes.NewBufferString(`{"email":"admin@example.com","password":"password"}`)
			req, _ = http.NewRequest("POST", cfg.BaseURL+endpoint, body)
			req.Header.Set("Content-Type", "application/json")
		}

	default:
		endpoint = "/api/v1/plans"
		method = "GET"
		req, _ = http.NewRequest("GET", cfg.BaseURL+endpoint, nil)
	}

	start := time.Now()
	resp, err := client.Do(req)
	elapsed := time.Since(start)

	if err != nil {
		return RequestResult{
			Elapsed:      elapsed,
			StatusCode:   0,
			Endpoint:     endpoint,
			Method:       method,
			Err:          err,
			ResponseBody: "",
		}
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	respBody := strings.TrimSpace(string(bodyBytes))

	return RequestResult{
		Elapsed:      elapsed,
		StatusCode:   resp.StatusCode,
		Endpoint:     endpoint,
		Method:       method,
		Err:          nil,
		ResponseBody: respBody,
	}
}

func runLoadTest(cfg Config) Result {
	client := createHTTPClient(cfg.Concurrency, cfg.Timeout)

	jobs := make(chan int, cfg.TotalReqs)
	results := make(chan RequestResult, cfg.TotalReqs)

	fmt.Printf("\n🚀 Starting Load Test: %d Requests | %d Concurrency | Scenario: %s\n", cfg.TotalReqs, cfg.Concurrency, cfg.Scenario)
	fmt.Printf("🎯 Target: %s\n\n", cfg.BaseURL)

	var wg sync.WaitGroup

	startTime := time.Now()

	// Spawn workers
	for i := 0; i < cfg.Concurrency; i++ {
		wg.Add(1)
		go runWorker(i, cfg, client, jobs, results, &wg)
	}

	// Feed jobs in background
	go func() {
		for i := 0; i < cfg.TotalReqs; i++ {
			jobs <- i
		}
		close(jobs)
	}()

	// Collector in background
	var successCount int64
	var errorCount int64
	statusCounts := make(map[int]int64)
	latencies := make([]time.Duration, 0, cfg.TotalReqs)

	failureCategories := make(map[string]*FailureCategory)
	var sampleFailures []FailureDetail

	var collectorWg sync.WaitGroup
	collectorWg.Add(1)

	var processedCount int64

	// Progress reporter
	ticker := time.NewTicker(1 * time.Second)
	doneProgress := make(chan bool)
	go func() {
		for {
			select {
			case <-ticker.C:
				done := atomic.LoadInt64(&processedCount)
				pct := float64(done) / float64(cfg.TotalReqs) * 100.0
				elapsed := time.Since(startTime).Seconds()
				rps := float64(done) / elapsed
				fmt.Printf("\r  ⏳ Progress: %d / %d (%.1f%%) | Rate: %.0f req/sec | Errors: %d", done, cfg.TotalReqs, pct, rps, atomic.LoadInt64(&errorCount))
			case <-doneProgress:
				return
			}
		}
	}()

	go func() {
		defer collectorWg.Done()
		for r := range results {
			atomic.AddInt64(&processedCount, 1)
			latencies = append(latencies, r.Elapsed)
			statusCounts[r.StatusCode]++

			isFailed := r.Err != nil || r.StatusCode >= 400

			if isFailed {
				atomic.AddInt64(&errorCount, 1)

				errMsg := ""
				if r.Err != nil {
					errMsg = r.Err.Error()
				} else {
					errMsg = fmt.Sprintf("HTTP %d", r.StatusCode)
				}

				// Extract error reason
				reason := errMsg
				if r.ResponseBody != "" {
					var jsonMap map[string]interface{}
					if err := json.Unmarshal([]byte(r.ResponseBody), &jsonMap); err == nil {
						if errVal, ok := jsonMap["error"]; ok {
							reason = fmt.Sprintf("HTTP %d: %v", r.StatusCode, errVal)
						} else if msgVal, ok := jsonMap["message"]; ok {
							reason = fmt.Sprintf("HTTP %d: %v", r.StatusCode, msgVal)
						}
					} else {
						if len(r.ResponseBody) > 120 {
							reason = fmt.Sprintf("HTTP %d: %s...", r.StatusCode, r.ResponseBody[:120])
						} else {
							reason = fmt.Sprintf("HTTP %d: %s", r.StatusCode, r.ResponseBody)
						}
					}
				}

				catKey := fmt.Sprintf("%s %s [%d] %s", r.Method, r.Endpoint, r.StatusCode, reason)
				if cat, exists := failureCategories[catKey]; exists {
					cat.Count++
				} else {
					failureCategories[catKey] = &FailureCategory{
						Endpoint:   r.Endpoint,
						Method:     r.Method,
						StatusCode: r.StatusCode,
						Reason:     reason,
						Count:      1,
						SampleBody: r.ResponseBody,
					}
				}

				if len(sampleFailures) < 1000 {
					sampleFailures = append(sampleFailures, FailureDetail{
						Timestamp:    time.Now(),
						Endpoint:     r.Endpoint,
						Method:       r.Method,
						StatusCode:   r.StatusCode,
						ErrorMessage: errMsg,
						ResponseBody: r.ResponseBody,
						DurationMs:   float64(r.Elapsed.Microseconds()) / 1000.0,
					})
				}
			} else {
				atomic.AddInt64(&successCount, 1)
			}
		}
	}()

	// Wait for workers to finish
	wg.Wait()
	close(results)
	collectorWg.Wait()

	totalDuration := time.Since(startTime)
	ticker.Stop()
	doneProgress <- true

	// Latency percentiles
	sort.Slice(latencies, func(i, j int) bool {
		return latencies[i] < latencies[j]
	})

	var totalLatency time.Duration
	minLat := time.Duration(1<<63 - 1)
	maxLat := time.Duration(0)

	for _, lat := range latencies {
		totalLatency += lat
		if lat < minLat {
			minLat = lat
		}
		if lat > maxLat {
			maxLat = lat
		}
	}
	if len(latencies) == 0 {
		minLat = 0
	}

	p50 := time.Duration(0)
	p90 := time.Duration(0)
	p95 := time.Duration(0)
	p99 := time.Duration(0)

	if len(latencies) > 0 {
		p50 = latencies[int(float64(len(latencies))*0.50)]
		p90 = latencies[int(float64(len(latencies))*0.90)]
		p95 = latencies[int(float64(len(latencies))*0.95)]
		p99 = latencies[int(float64(len(latencies))*0.99)]
	}

	avgLat := time.Duration(0)
	if len(latencies) > 0 {
		avgLat = totalLatency / time.Duration(len(latencies))
	}

	// Format failure breakdown
	var breakdown []FailureCategory
	for _, v := range failureCategories {
		breakdown = append(breakdown, *v)
	}
	sort.Slice(breakdown, func(i, j int) bool {
		return breakdown[i].Count > breakdown[j].Count
	})

	rps := float64(cfg.TotalReqs) / totalDuration.Seconds()

	res := Result{
		TotalRequests:    cfg.TotalReqs,
		SuccessCount:     successCount,
		ErrorCount:       errorCount,
		StatusCodes:      statusCounts,
		TotalDurationSec: totalDuration.Seconds(),
		RPS:              rps,
		AvgLatencyMs:     float64(avgLat.Microseconds()) / 1000.0,
		MinLatencyMs:     float64(minLat.Microseconds()) / 1000.0,
		P50LatencyMs:     float64(p50.Microseconds()) / 1000.0,
		P90LatencyMs:     float64(p90.Microseconds()) / 1000.0,
		P95LatencyMs:     float64(p95.Microseconds()) / 1000.0,
		P99LatencyMs:     float64(p99.Microseconds()) / 1000.0,
		MaxLatencyMs:     float64(maxLat.Microseconds()) / 1000.0,
		FailureBreakdown: breakdown,
		DetailedFailures: sampleFailures,
	}

	printAndSaveSummary(res)
	return res
}

func printAndSaveSummary(r Result) {
	fmt.Printf("\n\n=======================================================================\n")
	fmt.Printf("                    🏁 LOAD TEST BENCHMARK RESULTS                    \n")
	fmt.Printf("=======================================================================\n")
	fmt.Printf(" Total Requests:     %d\n", r.TotalRequests)
	fmt.Printf(" Successful (2xx):   %d (%.2f%%)\n", r.SuccessCount, float64(r.SuccessCount)/float64(r.TotalRequests)*100.0)
	fmt.Printf(" Failed / Errored:   %d (%.2f%%)\n", r.ErrorCount, float64(r.ErrorCount)/float64(r.TotalRequests)*100.0)
	fmt.Printf(" Total Duration:     %.2fs\n", r.TotalDurationSec)
	fmt.Printf(" Throughput:         %.2f Requests / Second (RPS)\n", r.RPS)
	fmt.Printf("-----------------------------------------------------------------------\n")
	fmt.Printf(" ⚡ Latency Breakdown:\n")
	fmt.Printf("   - Average:        %.2f ms\n", r.AvgLatencyMs)
	fmt.Printf("   - Min:            %.2f ms\n", r.MinLatencyMs)
	fmt.Printf("   - P50 (Median):   %.2f ms\n", r.P50LatencyMs)
	fmt.Printf("   - P90:            %.2f ms\n", r.P90LatencyMs)
	fmt.Printf("   - P95:            %.2f ms\n", r.P95LatencyMs)
	fmt.Printf("   - P99:            %.2f ms\n", r.P99LatencyMs)
	fmt.Printf("   - Max:            %.2f ms\n", r.MaxLatencyMs)
	fmt.Printf("-----------------------------------------------------------------------\n")
	fmt.Printf(" 📊 HTTP Status Codes Distribution:\n")
	for code, count := range r.StatusCodes {
		fmt.Printf("   - HTTP %d:         %d (%.2f%%)\n", code, count, float64(count)/float64(r.TotalRequests)*100.0)
	}

	if len(r.FailureBreakdown) > 0 {
		fmt.Printf("-----------------------------------------------------------------------\n")
		fmt.Printf(" ❌ FAILED ENDPOINTS & ROOT CAUSE ANALYSIS:\n")
		for idx, cat := range r.FailureBreakdown {
			fmt.Printf("\n   [%d] %s %s (HTTP %d) -> %d Failures (%.2f%%)\n",
				idx+1, cat.Method, cat.Endpoint, cat.StatusCode, cat.Count, float64(cat.Count)/float64(r.TotalRequests)*100.0)
			fmt.Printf("       Why it failed: %s\n", cat.Reason)
			if cat.SampleBody != "" {
				fmt.Printf("       Sample Response: %s\n", cat.SampleBody)
			}
		}
	} else {
		fmt.Printf("-----------------------------------------------------------------------\n")
		fmt.Printf(" ✨ 0 Failures! All requests completed with 100%% success!\n")
	}
	fmt.Printf("=======================================================================\n\n")

	// 1. Write comprehensive JSON report
	jsonData, _ := json.MarshalIndent(r, "", "  ")
	_ = os.WriteFile("loadtest_report.json", jsonData, 0644)
	fmt.Printf("📁 Full JSON report saved to: loadtest_report.json\n")

	// 2. Write dedicated failure analysis report
	var failureTxt strings.Builder
	failureTxt.WriteString("=======================================================================\n")
	failureTxt.WriteString("                 FAILED ENDPOINTS & ROOT CAUSE REPORT                  \n")
	failureTxt.WriteString(fmt.Sprintf(" Generated: %s\n", time.Now().Format(time.RFC3339)))
	failureTxt.WriteString(fmt.Sprintf(" Total Requests: %d | Total Failures: %d (%.2f%%)\n",
		r.TotalRequests, r.ErrorCount, float64(r.ErrorCount)/float64(r.TotalRequests)*100.0))
	failureTxt.WriteString("=======================================================================\n\n")

	if len(r.FailureBreakdown) == 0 {
		failureTxt.WriteString("No failed endpoints recorded. All requests succeeded (100% 2xx responses).\n")
	} else {
		for idx, cat := range r.FailureBreakdown {
			failureTxt.WriteString(fmt.Sprintf("-----------------------------------------------------------------------\n"))
			failureTxt.WriteString(fmt.Sprintf("[%d] Endpoint:   %s %s\n", idx+1, cat.Method, cat.Endpoint))
			failureTxt.WriteString(fmt.Sprintf("    Status:     HTTP %d\n", cat.StatusCode))
			failureTxt.WriteString(fmt.Sprintf("    Count:      %d occurrences (%.2f%% of all traffic)\n",
				cat.Count, float64(cat.Count)/float64(r.TotalRequests)*100.0))
			failureTxt.WriteString(fmt.Sprintf("    Failure:    %s\n", cat.Reason))
			if cat.SampleBody != "" {
				failureTxt.WriteString(fmt.Sprintf("    Raw Body:   %s\n", cat.SampleBody))
			}
			failureTxt.WriteString("\n")
		}

		failureTxt.WriteString("\n=======================================================================\n")
		failureTxt.WriteString("SAMPLE ERROR TIMELINE (FIRST UP TO 50 DETAILED EVENTS):\n")
		failureTxt.WriteString("=======================================================================\n")
		maxSample := len(r.DetailedFailures)
		if maxSample > 50 {
			maxSample = 50
		}
		for i := 0; i < maxSample; i++ {
			f := r.DetailedFailures[i]
			failureTxt.WriteString(fmt.Sprintf("[%s] %s %s -> HTTP %d (%.1fms): %s | %s\n",
				f.Timestamp.Format("15:04:05.000"), f.Method, f.Endpoint, f.StatusCode, f.DurationMs, f.ErrorMessage, f.ResponseBody))
		}
	}

	_ = os.WriteFile("loadtest_failures.txt", []byte(failureTxt.String()), 0644)
	fmt.Printf("📄 Human-readable Failure Report saved to: loadtest_failures.txt\n\n")
}

func main() {
	baseURL := flag.String("url", "http://localhost:8080", "Base URL of API Gateway")
	totalReqs := flag.Int("n", 100000, "Total number of requests")
	concurrency := flag.Int("c", 1000, "Concurrent worker routines")
	scenario := flag.String("scenario", "mixed", "Scenario: plans | health | auth_login | mixed")
	timeoutSec := flag.Int("t", 10, "Request timeout in seconds")
	flag.Parse()

	cfg := Config{
		BaseURL:     *baseURL,
		TotalReqs:   *totalReqs,
		Concurrency: *concurrency,
		Scenario:    *scenario,
		Timeout:     time.Duration(*timeoutSec) * time.Second,
	}

	_ = runLoadTest(cfg)
}
