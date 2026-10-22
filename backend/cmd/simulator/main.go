package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"streamforge/internal/event"
)

type Config struct {
	BaseURL     string
	Rate        int
	Duration    time.Duration
	Concurrency int
	Scenario    string
}

type Metrics struct {
	TotalSent uint64
	Success   uint64
	Failed    uint64
	Latencies []time.Duration
	mu        sync.Mutex
}

func main() {
	cfg := parseFlags()

	fmt.Println("=================================================================")
	fmt.Println("             STREAMFORGE LOAD SIMULATOR (PHASE 18)               ")
	fmt.Println("=================================================================")
	fmt.Printf(" Target URL:   %s\n", cfg.BaseURL)
	fmt.Printf(" Target Rate:  %d events/sec\n", cfg.Rate)
	fmt.Printf(" Duration:     %v\n", cfg.Duration)
	fmt.Printf(" Concurrency:  %d workers\n", cfg.Concurrency)
	if cfg.Scenario != "" {
		fmt.Printf(" Scenario:     %s\n", cfg.Scenario)
	}
	fmt.Println("=================================================================")
	fmt.Println(" Starting traffic generation...")

	client := &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        2000,
			MaxIdleConnsPerHost: 1000,
			IdleConnTimeout:     90 * time.Second,
			DisableKeepAlives:   false,
		},
		Timeout: 5 * time.Second,
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Duration)
	defer cancel()

	// Handle graceful interrupt
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("\nReceived interrupt signal, stopping generation early...")
		cancel()
	}()

	metrics := &Metrics{
		Latencies: make([]time.Duration, 0, cfg.Rate*int(cfg.Duration.Seconds())),
	}

	workCh := make(chan struct{}, cfg.Rate)
	var wg sync.WaitGroup

	// Start worker pool for sending HTTP requests
	for i := 0; i < cfg.Concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for range workCh {
				sendEvent(client, cfg.BaseURL, metrics)
			}
		}(i)
	}

	startTime := time.Now()
	tickerInterval := time.Second / time.Duration(cfg.Rate)
	ticker := time.NewTicker(tickerInterval)
	defer ticker.Stop()

	// Rate-limited dispatch loop
dispatchLoop:
	for {
		select {
		case <-ctx.Done():
			break dispatchLoop
		case <-ticker.C:
			select {
			case workCh <- struct{}{}:
			default:
				// Worker queue saturated
			}
		}
	}

	close(workCh)
	wg.Wait()
	elapsed := time.Since(startTime)

	printReport(metrics, elapsed, cfg.Rate)
}

func parseFlags() Config {
	defaultURL := getEnv("API_URL", "http://localhost:8080")
	defaultRate := getIntEnv("RATE", 100)
	defaultDuration := getDurationEnv("DURATION", 10*time.Second)

	apiURL := flag.String("url", defaultURL, "Target StreamForge API base URL")
	rate := flag.Int("rate", defaultRate, "Target event generation rate (events/second)")
	duration := flag.Duration("duration", defaultDuration, "Load test duration (e.g. 10s, 30s, 1m)")
	concurrency := flag.Int("concurrency", 0, "Number of concurrent HTTP dispatchers (default auto-scaled)")
	scenario := flag.String("scenario", "", "Scenario identifier (scenario1=100/s, scenario2=1k/s, scenario3=5k/s, scenario4=10k/s)")

	flag.Parse()

	if *scenario != "" {
		switch *scenario {
		case "1", "100", "scenario1":
			*rate = 100
		case "2", "1k", "1000", "scenario2":
			*rate = 1000
		case "3", "5k", "5000", "scenario3":
			*rate = 5000
		case "4", "10k", "10000", "scenario4":
			*rate = 10000
		}
	}

	workers := *concurrency
	if workers <= 0 {
		workers = *rate / 10
		if workers < 10 {
			workers = 10
		}
		if workers > 200 {
			workers = 200
		}
	}

	return Config{
		BaseURL:     *apiURL,
		Rate:        *rate,
		Duration:    *duration,
		Concurrency: workers,
		Scenario:    *scenario,
	}
}

type EventRequest struct {
	Type     string         `json:"type"`
	Priority int            `json:"priority"`
	Payload  map[string]any `json:"payload"`
}

var eventTypes = []event.EventType{
	event.TypeOrderCreated,
	event.TypePaymentSuccess,
	event.TypePaymentFailed,
	event.TypeInventoryUpdated,
	event.TypeOrderCancelled,
}

func sendEvent(client *http.Client, baseURL string, m *Metrics) {
	atomic.AddUint64(&m.TotalSent, 1)

	// Pick realistic random event type and priority
	evtType := string(eventTypes[rand.Intn(len(eventTypes))])
	priorityRoll := rand.Intn(100)
	priority := event.PriorityNormal
	if priorityRoll < 15 {
		priority = event.PriorityLow
	} else if priorityRoll > 80 {
		priority = event.PriorityHigh
	}

	reqBody := EventRequest{
		Type:     evtType,
		Priority: priority,
		Payload: map[string]any{
			"order_id":  fmt.Sprintf("ord-%d-%d", time.Now().UnixNano(), rand.Intn(1000)),
			"amount":    10.0 + rand.Float64()*490.0,
			"timestamp": time.Now().UTC().Format(time.RFC3339Nano),
			"source":    "simulator",
		},
	}

	bodyBytes, _ := json.Marshal(reqBody)

	start := time.Now()
	resp, err := client.Post(baseURL+"/api/v1/events", "application/json", bytes.NewReader(bodyBytes))
	latency := time.Since(start)

	if err != nil {
		atomic.AddUint64(&m.Failed, 1)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusAccepted {
		atomic.AddUint64(&m.Success, 1)
		m.mu.Lock()
		m.Latencies = append(m.Latencies, latency)
		m.mu.Unlock()
	} else {
		atomic.AddUint64(&m.Failed, 1)
	}
}

func printReport(m *Metrics, elapsed time.Duration, targetRate int) {
	total := atomic.LoadUint64(&m.TotalSent)
	success := atomic.LoadUint64(&m.Success)
	failed := atomic.LoadUint64(&m.Failed)

	m.mu.Lock()
	lats := make([]time.Duration, len(m.Latencies))
	copy(lats, m.Latencies)
	m.mu.Unlock()

	sort.Slice(lats, func(i, j int) bool {
		return lats[i] < lats[j]
	})

	var minLat, maxLat, avgLat, p50, p90, p95, p99 time.Duration
	if len(lats) > 0 {
		minLat = lats[0]
		maxLat = lats[len(lats)-1]
		var totalLat time.Duration
		for _, l := range lats {
			totalLat += l
		}
		avgLat = totalLat / time.Duration(len(lats))
		p50 = lats[int(float64(len(lats))*0.50)]
		p90 = lats[int(float64(len(lats))*0.90)]
		p95 = lats[int(float64(len(lats))*0.95)]
		p99 = lats[int(float64(len(lats))*0.99)]
	}

	actualRate := float64(total) / elapsed.Seconds()
	successRate := 0.0
	if total > 0 {
		successRate = float64(success) / float64(total) * 100.0
	}

	fmt.Println("\n=================================================================")
	fmt.Println("                     BENCHMARK RESULTS REPORT                    ")
	fmt.Println("=================================================================")
	fmt.Printf(" Execution Time:       %v\n", elapsed.Round(time.Millisecond))
	fmt.Printf(" Total Events Sent:    %d\n", total)
	fmt.Printf(" Successful Events:    %d (%.1f%%)\n", success, successRate)
	fmt.Printf(" Failed/Dropped:       %d\n", failed)
	fmt.Printf(" Target Throughput:    %d events/sec\n", targetRate)
	fmt.Printf(" Actual Throughput:    %.1f events/sec\n", actualRate)
	fmt.Println("-----------------------------------------------------------------")
	fmt.Println(" LATENCY DISTRIBUTION (API HTTP Response Time):")
	fmt.Printf("   Min Latency:        %v\n", minLat.Round(time.Microsecond))
	fmt.Printf("   Average Latency:    %v\n", avgLat.Round(time.Microsecond))
	fmt.Printf("   p50 (Median):       %v\n", p50.Round(time.Microsecond))
	fmt.Printf("   p90:                %v\n", p90.Round(time.Microsecond))
	fmt.Printf("   p95:                %v\n", p95.Round(time.Microsecond))
	fmt.Printf("   p99:                %v\n", p99.Round(time.Microsecond))
	fmt.Printf("   Max Latency:        %v\n", maxLat.Round(time.Microsecond))
	fmt.Println("=================================================================")
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getIntEnv(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func getDurationEnv(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
		if n, err := strconv.Atoi(v); err == nil {
			return time.Duration(n) * time.Second
		}
	}
	return fallback
}
