package main

import (
	"context"
	"crypto/tls"
	"html/template"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/time/rate"
)

type Backend struct {
	URL          string
	Alive        bool
	Requests     int
	Limiter      *rate.Limiter
	FailureCount int
	LastFailure  time.Time
}

type LoadBalancer struct {
	Backends []*Backend
	Strategy string
	Current  int
}

var (
	requestCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "loadbalancer_requests_total",
			Help: "Total number of requests forwarded",
		},
		[]string{"backend"},
	)
	requestLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "loadbalancer_request_latency_seconds",
			Help:    "Request latency per backend",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"backend"},
	)
)

func init() {
	prometheus.MustRegister(requestCount, requestLatency)
}

func (lb *LoadBalancer) GetNextBackend() *Backend {
	switch lb.Strategy {
	case "round-robin":
		start := lb.Current
		for {
			b := lb.Backends[lb.Current]
			lb.Current = (lb.Current + 1) % len(lb.Backends)
			if b.Alive {
				return b
			}
			if lb.Current == start {
				return nil
			}
		}
	case "least-connections":
		var selected *Backend
		for _, b := range lb.Backends {
			if b.Alive && (selected == nil || b.Requests < selected.Requests) {
				selected = b
			}
		}
		return selected
	case "weighted":
		var totalWeight int
		for _, b := range lb.Backends {
			if b.Alive {
				totalWeight += (10 - b.FailureCount)
			}
		}
		randVal := rand.Intn(totalWeight + 1)
		cumulative := 0
		for _, b := range lb.Backends {
			if b.Alive {
				cumulative += (10 - b.FailureCount)
				if cumulative >= randVal {
					return b
				}
			}
		}
		return nil
	default:
		return nil
	}
}

func (lb *LoadBalancer) HealthCheck(interval time.Duration) {
	for {
		for _, backend := range lb.Backends {
			go func(b *Backend) {
				_, err := http.Get(b.URL + "/health")
				b.Alive = err == nil
				if !b.Alive {
					b.FailureCount++
					b.LastFailure = time.Now()
				} else {
					b.FailureCount = 0
				}
			}(backend)
		}
		time.Sleep(interval)
	}
}

func (lb *LoadBalancer) ServeProxy(w http.ResponseWriter, r *http.Request) {
	backend := lb.GetNextBackend()
	if backend == nil {
		http.Error(w, "No backend available", http.StatusServiceUnavailable)
		return
	}

	if !backend.Limiter.Allow() {
		http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
		return
	}

	start := time.Now()
	proxy := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	newReq, err := http.NewRequest(r.Method, backend.URL+r.RequestURI, r.Body)
	if err != nil {
		http.Error(w, "Request creation failed", http.StatusInternalServerError)
		return
	}
	newReq.Header = r.Header
	resp, err := proxy.RoundTrip(newReq)
	if err != nil {
		backend.FailureCount++
		backend.LastFailure = time.Now()
		http.Error(w, "Upstream error", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	for k, v := range resp.Header {
		for _, h := range v {
			w.Header().Add(k, h)
		}
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)

	duration := time.Since(start).Seconds()
	requestCount.WithLabelValues(backend.URL).Inc()
	requestLatency.WithLabelValues(backend.URL).Observe(duration)
	backend.Requests++
}

func serveAdminUI(lb *LoadBalancer) http.HandlerFunc {
	tmpl := template.Must(template.New("dashboard").Parse(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Load Balancer Admin</title>
    <link href="https://cdn.jsdelivr.net/npm/tailwindcss@2.2.19/dist/tailwind.min.css" rel="stylesheet">
</head>
<body class="bg-gray-100">
    <div class="max-w-4xl mx-auto p-8">
        <h2 class="text-3xl font-semibold text-center text-gray-800 mb-6">Load Balancer Status</h2>
        <div class="bg-white p-6 rounded-lg shadow-md">
            <p class="text-xl text-gray-700">Current Strategy: <span class="font-bold">{{.Strategy}}</span></p>
            <table class="min-w-full bg-white border border-gray-300 mt-6">
                <thead>
                    <tr class="border-b">
                        <th class="px-4 py-2 text-left">Backend</th>
                        <th class="px-4 py-2 text-left">Alive</th>
                        <th class="px-4 py-2 text-left">Requests</th>
                        <th class="px-4 py-2 text-left">Failures</th>
                        <th class="px-4 py-2 text-left">Last Failure</th>
                    </tr>
                </thead>
                <tbody>
                    {{range .Backends}}
                    <tr>
                        <td class="px-4 py-2">{{.URL}}</td>
                        <td class="px-4 py-2">{{.Alive}}</td>
                        <td class="px-4 py-2">{{.Requests}}</td>
                        <td class="px-4 py-2">{{.FailureCount}}</td>
                        <td class="px-4 py-2">{{.LastFailure}}</td>
                    </tr>
                    {{end}}
                </tbody>
            </table>
            <form action="/admin/strategy" method="post" class="mt-6">
                <label for="strategy" class="block text-sm text-gray-700">Select Load Balancer Strategy:</label>
                <select name="s" id="strategy" class="w-full px-4 py-2 border rounded-md mt-2">
                    <option value="round-robin" {{if eq .Strategy "round-robin"}}selected{{end}}>Round Robin</option>
                    <option value="least-connections" {{if eq .Strategy "least-connections"}}selected{{end}}>Least Connections</option>
                    <option value="weighted" {{if eq .Strategy "weighted"}}selected{{end}}>Weighted</option>
                </select>
                <button type="submit" class="mt-4 px-6 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 focus:outline-none">Set Strategy</button>
            </form>
        </div>
    </div>
</body>
</html>`))

	return func(w http.ResponseWriter, r *http.Request) {
		tmpl.Execute(w, lb)
	}
}

func main() {
	backends := []*Backend{
		{URL: "http://localhost:8081", Alive: true, Limiter: rate.NewLimiter(5, 10)},
		{URL: "http://localhost:8082", Alive: true, Limiter: rate.NewLimiter(5, 10)},
		{URL: "http://localhost:8083", Alive: true, Limiter: rate.NewLimiter(5, 10)},
	}

	lb := &LoadBalancer{
		Backends: backends,
		Strategy: "round-robin",
		Current:  0,
	}

	go lb.HealthCheck(10 * time.Second)

	mux := http.NewServeMux()
	mux.HandleFunc("/", lb.ServeProxy)
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/admin", serveAdminUI(lb))
	mux.HandleFunc("/admin/strategy", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		strategy := r.FormValue("s")
		if strategy != "round-robin" && strategy != "least-connections" && strategy != "weighted" {
			http.Error(w, "Unsupported strategy", http.StatusBadRequest)
			return
		}
		lb.Strategy = strategy
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
	})

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		TLSConfig:    &tls.Config{MinVersion: tls.VersionTLS12},
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		log.Println("Starting HTTPS server on :8080")
		if err := srv.ListenAndServeTLS("server.crt", "server.key"); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTPS server error: %s", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server Shutdown Failed:%+v", err)
	}
	log.Println("Server exited gracefully")
}