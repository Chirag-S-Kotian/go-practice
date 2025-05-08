package main

import (
	"context"
	"crypto/tls"
	"html/template"
	"io"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/time/rate"
)

type Backend struct {
	mu sync.Mutex
	URL string
	Alive bool
	Requests int
	Limiter *rate.Limiter
	FailureCount int
	LastFailure time.Time
}

type LoadBalancer struct {
	mu sync.Mutex
	Backends []*Backend
	Strategy string
	Current int
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
	lb.mu.Lock()
	defer lb.mu.Unlock()

	switch lb.Strategy {
	case "round-robin":
		start := lb.Current
		for {
			b := lb.Backends[lb.Current]
			lb.Current = (lb.Current + 1) % len(lb.Backends)

			b.mu.Lock()
			isAlive := b.Alive
			b.mu.Unlock()

			if isAlive {
				return b
			}
			if lb.Current == start {
				return nil
			}
		}
	case "least-connections":
		var selected *Backend
		minRequests := -1

		for _, b := range lb.Backends {
			b.mu.Lock()
			isAlive := b.Alive
			reqs := b.Requests
			b.mu.Unlock()

			if isAlive {
				if selected == nil || reqs < minRequests {
					selected = b
					minRequests = reqs
				}
			}
		}
		return selected
	case "weighted":
		var totalWeight int
		aliveBackends := []*Backend{}

		for _, b := range lb.Backends {
			b.mu.Lock()
			isAlive := b.Alive
			failures := b.FailureCount
			b.mu.Unlock()

			if isAlive {
				aliveBackends = append(aliveBackends, b)
				weight := 10 - failures
				if weight < 0 {
					weight = 0
				}
				totalWeight += weight
			}
		}

		if totalWeight == 0 {
			return nil
		}

		randVal := rand.Intn(totalWeight)

		cumulative := 0
		for _, b := range aliveBackends {
            b.mu.Lock()
            failures := b.FailureCount
            b.mu.Unlock()

            weight := 10 - failures
            if weight < 0 {
                weight = 0
            }
			cumulative += weight
			if cumulative > randVal {
				return b
			}
		}
        if len(aliveBackends) > 0 {
             return aliveBackends[len(aliveBackends)-1]
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
				resp, err := http.Get(b.URL + "/health")
				isAlive := err == nil && resp.StatusCode == http.StatusOK

				b.mu.Lock()
				b.Alive = isAlive
				if !b.Alive {
					b.FailureCount++
					b.LastFailure = time.Now()
				} else {
					b.FailureCount = 0
				}
				b.mu.Unlock()

				if resp != nil {
					resp.Body.Close()
				}

				if !isAlive {
					log.Printf("Backend %s failed health check: %v", b.URL, err)
				}
			}(backend)
		}
		time.Sleep(interval)
	}
}

func (lb *LoadBalancer) ServeProxy(w http.ResponseWriter, r *http.Request) {
	backend := lb.GetNextBackend()
	if backend == nil {
		log.Printf("No backend available for %s %s", r.Method, r.URL.Path)
		http.Error(w, "No backend available", http.StatusServiceUnavailable)
		return
	}

	if !backend.Limiter.Allow() {
		log.Printf("Rate limit exceeded for backend %s", backend.URL)
		http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
		return
	}

	start := time.Now()

	newReq, err := http.NewRequest(r.Method, backend.URL+r.RequestURI, r.Body)
	if err != nil {
		log.Printf("Failed to create request for backend %s: %v", backend.URL, err)
		http.Error(w, "Request creation failed", http.StatusInternalServerError)
		return
	}
	newReq.Header = make(http.Header)
	for k, v := range r.Header {
		newReq.Header[k] = v
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
        DialContext: (&net.Dialer{
            Timeout:   30 * time.Second,
            KeepAlive: 30 * time.Second,
        }).DialContext,
        TLSHandshakeTimeout: 10 * time.Second,
        ResponseHeaderTimeout: 10 * time.Second,
        ExpectContinueTimeout: 1 * time.Second,
	}
    client := &http.Client{Transport: transport}

	resp, err := client.Do(newReq)
	if err != nil {
		log.Printf("Upstream error from backend %s: %v", backend.URL, err)
		backend.mu.Lock()
		backend.FailureCount++
		backend.LastFailure = time.Now()
		backend.Alive = false
		backend.mu.Unlock()

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

	if _, copyErr := io.Copy(w, resp.Body); copyErr != nil {
         log.Printf("Error copying response body from backend %s: %v", backend.URL, copyErr)
    }


	duration := time.Since(start).Seconds()
	requestCount.WithLabelValues(backend.URL).Inc()
	requestLatency.WithLabelValues(backend.URL).Observe(duration)

	backend.mu.Lock()
	backend.Requests++
	backend.mu.Unlock()

	log.Printf("Request %s %s forwarded to %s, status %d, took %.2f ms",
		r.Method, r.URL.Path, backend.URL, resp.StatusCode, duration*1000)
}

func serveAdminUI(lb *LoadBalancer) http.HandlerFunc {
	tmpl := template.Must(template.New("dashboard").Parse(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Go Load Balancer Admin</title>
    <link href="https://cdn.jsdelivr.net/npm/tailwindcss@2.2.19/dist/tailwind.min.css" rel="stylesheet">
</head>
<body class="bg-gray-100 font-sans">
    <div class="max-w-4xl mx-auto p-8 bg-white shadow-md rounded-lg mt-10">
        <h2 class="text-3xl font-bold text-center text-gray-800 mb-8">Load Balancer Status</h2>
        <div class="mb-6">
            <p class="text-xl text-gray-700">Current Strategy: <span class="font-bold text-blue-600">{{.Strategy}}</span></p>
        </div>

        <div class="overflow-x-auto">
            <table class="min-w-full bg-white border border-gray-300 rounded-lg">
                <thead>
                    <tr class="bg-gray-200 text-gray-700 uppercase text-sm leading-normal border-b border-gray-300">
                        <th class="py-3 px-6 text-left">Backend URL</th>
                        <th class="py-3 px-6 text-left">Alive</th>
                        <th class="py-3 px-6 text-left">Requests Handled</th>
                        <th class="py-3 px-6 text-left">Failure Count</th>
                        <th class="py-3 px-6 text-left">Last Failure Time</th>
                    </tr>
                </thead>
                <tbody class="text-gray-600 text-sm font-light">
                    {{range .Backends}}
                    <tr class="border-b border-gray-200 hover:bg-gray-100">
                        <td class="py-3 px-6 text-left whitespace-nowrap">{{.URL}}</td>
                        <td class="py-3 px-6 text-left">
                            {{if .Alive}}
                                <span class="bg-green-200 text-green-600 py-1 px-3 rounded-full text-xs">Yes</span>
                            {{else}}
                                <span class="bg-red-200 text-red-600 py-1 px-3 rounded-full text-xs">No</span>
                            {{end}}
                        </td>
                        <td class="py-3 px-6 text-left">{{.Requests}}</td>
                        <td class="py-3 px-6 text-left">{{.FailureCount}}</td>
                        <td class="py-3 px-6 text-left">{{if not .LastFailure.IsZero}}{{.LastFailure.Format "2006-01-02 15:04:05"}}{{else}}N/A{{end}}</td>
                    </tr>
                    {{end}}
                </tbody>
            </table>
        </div>

        <form action="/admin/strategy" method="post" class="mt-8">
            <label for="strategy" class="block text-gray-700 text-sm font-bold mb-2">Select Load Balancer Strategy:</label>
            <div class="flex items-center">
                 <select name="s" id="strategy" class="block appearance-none w-full bg-gray-200 border border-gray-200 text-gray-700 py-3 px-4 pr-8 rounded leading-tight focus:outline-none focus:bg-white focus:border-gray-500" id="grid-state">
                    <option value="round-robin" {{if eq .Strategy "round-robin"}}selected{{end}}>Round Robin</option>
                    <option value="least-connections" {{if eq .Strategy "least-connections"}}selected{{end}}>Least Connections</option>
                    <option value="weighted" {{if eq .Strategy "weighted"}}selected{{end}}>Weighted</option>
                </select>
                <button type="submit" class="ml-4 bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded focus:outline-none focus:shadow-outline">Set Strategy</button>
            </div>
        </form>
         <div class="mt-8 text-center text-gray-500 text-sm">
            Metrics available at <a href="/metrics" class="text-blue-600 hover:underline">/metrics</a>
        </div>
    </div>
</body>
</html>`))

	return func(w http.ResponseWriter, r *http.Request) {
        lb.mu.Lock()
        currentStrategy := lb.Strategy
        displayBackends := make([]struct{
            URL string
            Alive bool
            Requests int
            FailureCount int
            LastFailure time.Time
        }, len(lb.Backends))

        for i, b := range lb.Backends {
            b.mu.Lock()
            displayBackends[i] = struct{
                URL string
                Alive bool
                Requests int
                FailureCount int
                LastFailure time.Time
            }{
                URL: b.URL,
                Alive: b.Alive,
                Requests: b.Requests,
                FailureCount: b.FailureCount,
                LastFailure: b.LastFailure,
            }
            b.mu.Unlock()
        }
        lb.mu.Unlock()

        data := struct {
            Strategy string
            Backends []struct {
                URL string
                Alive bool
                Requests int
                FailureCount int
                LastFailure time.Time
            }
        }{
            Strategy: currentStrategy,
            Backends: displayBackends,
        }

		err := tmpl.Execute(w, data)
        if err != nil {
            log.Printf("Error rendering admin UI: %v", err)
            http.Error(w, "Error rendering template", http.StatusInternalServerError)
        }
	}
}


func main() {
    rand.Seed(time.Now().UnixNano())

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

        lb.mu.Lock()
		lb.Strategy = strategy
        lb.mu.Unlock()

        log.Printf("Load balancer strategy changed to %s", strategy)

		http.Redirect(w, r, "/admin", http.StatusSeeOther)
	})

	srv := &http.Server{
		Addr: ":8080",
		Handler: mux,
		TLSConfig: &tls.Config{MinVersion: tls.VersionTLS12},
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

    if _, err := os.Stat("server.crt"); os.IsNotExist(err) {
        log.Fatal("server.crt not found. Cannot start HTTPS server.")
    }
     if _, err := os.Stat("server.key"); os.IsNotExist(err) {
        log.Fatal("server.key not found. Cannot start HTTPS server.")
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

	log.Println("Shutdown signal received, shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server Shutdown Failed: %+v", err)
	}

	log.Println("Server exited gracefully")
}