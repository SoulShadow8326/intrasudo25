package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

type Backend struct {
	ID            string
	URL           *url.URL
	Proxy         *httputil.ReverseProxy
	mu            sync.RWMutex
	alive         bool
	weight        int
	currentWeight int
	timeout       time.Duration
	isUnixSocket  bool
	socketPath    string
}

func (b *Backend) IsAlive() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.alive
}

func (b *Backend) SetAlive(up bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.alive = up
}

func NewBackend(id, addr string, weight int, timeout time.Duration) *Backend {
	var u *url.URL
	var err error
	isUnixSocket := false
	socketPath := ""

	if strings.HasPrefix(addr, "unix://") {
		socketPath = strings.TrimPrefix(addr, "unix://")
		u, _ = url.Parse("http://localhost")
		isUnixSocket = true
	} else {
		u, err = url.Parse(addr)
		if err != nil {
			log.Fatalf("invalid backend URL %q: %v", addr, err)
		}
	}

	proxy := httputil.NewSingleHostReverseProxy(u)

	if isUnixSocket {
		proxy.Transport = &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return net.Dial("unix", socketPath)
			},
		}

		proxy.Director = func(req *http.Request) {
			req.URL.Scheme = "http"
			req.URL.Host = "localhost"
			req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
			req.Header.Set("X-Origin-Host", "unix-socket")
		}
	} else {
		originalDirector := proxy.Director
		proxy.Director = func(req *http.Request) {
			originalDirector(req)
			req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
			req.Header.Set("X-Origin-Host", u.Host)
		}
	}

	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("Backend %s error: %v", id, err)
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte("Bad Gateway"))
	}

	return &Backend{
		ID:            id,
		URL:           u,
		Proxy:         proxy,
		alive:         true,
		weight:        weight,
		currentWeight: weight,
		timeout:       timeout,
		isUnixSocket:  isUnixSocket,
		socketPath:    socketPath,
	}
}

type LoadBalancer struct {
	backends []*Backend
	counter  uint64
	mu       sync.RWMutex
}

func NewLoadBalancer(addrs []string, weights []int) *LoadBalancer {
	lb := &LoadBalancer{}

	for i, addr := range addrs {
		weight := 1
		if i < len(weights) {
			weight = weights[i]
		}
		backend := NewBackend(fmt.Sprintf("backend-%d", i), addr, weight, 10*time.Second)
		lb.backends = append(lb.backends, backend)
	}

	go lb.healthLoop()
	return lb
}

func (lb *LoadBalancer) healthLoop() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		lb.performHealthChecks()
	}
}

func (lb *LoadBalancer) performHealthChecks() {
	var wg sync.WaitGroup
	for _, b := range lb.backends {
		wg.Add(1)
		go func(backend *Backend) {
			defer wg.Done()

			if backend.isUnixSocket {
				conn, err := net.DialTimeout("unix", backend.socketPath, backend.timeout)
				if err != nil {
					backend.SetAlive(false)
					return
				}
				conn.Close()
				backend.SetAlive(true)
			} else {
				client := &http.Client{Timeout: backend.timeout}
				resp, err := client.Get(backend.URL.String() + "/landing")
				up := err == nil && resp != nil && resp.StatusCode < 500
				if resp != nil {
					resp.Body.Close()
				}
				backend.SetAlive(up)
			}
		}(b)
	}
	wg.Wait()
}

func (lb *LoadBalancer) getNextBackend() *Backend {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	var aliveBackends []*Backend
	for _, b := range lb.backends {
		if b.IsAlive() {
			aliveBackends = append(aliveBackends, b)
		}
	}

	if len(aliveBackends) == 0 {
		return nil
	}

	return lb.getWeightedBackend(aliveBackends)
}

func (lb *LoadBalancer) getWeightedBackend(backends []*Backend) *Backend {
	totalWeight := 0
	for _, b := range backends {
		b.mu.Lock()
		b.currentWeight += b.weight
		totalWeight += b.weight
		b.mu.Unlock()
	}

	var selected *Backend
	maxWeight := -1
	for _, b := range backends {
		b.mu.RLock()
		if b.currentWeight > maxWeight {
			maxWeight = b.currentWeight
			selected = b
		}
		b.mu.RUnlock()
	}

	if selected != nil {
		selected.mu.Lock()
		selected.currentWeight -= totalWeight
		selected.mu.Unlock()
	}

	return selected
}

func (lb *LoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	backend := lb.getNextBackend()
	if backend == nil {
		http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), backend.timeout)
	defer cancel()
	r = r.WithContext(ctx)

	backend.Proxy.ServeHTTP(w, r)
}

type Config struct {
	Backends      []BackendConfig `json:"backends"`
	LoadBalancing LBConfig        `json:"load_balancing"`
	Server        ServerConfig    `json:"server"`
}

type BackendConfig struct {
	Address string `json:"address"`
	Weight  int    `json:"weight"`
}

type LBConfig struct {
	Strategy string `json:"strategy"`
}

type ServerConfig struct {
	Port string `json:"port"`
}

func LoadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func main() {
	config, err := LoadConfig("config.json")
	if err != nil {
		log.Fatalf("Could not load config file: %v", err)
	}

	var targets []string
	var weights []int

	for _, backend := range config.Backends {
		targets = append(targets, backend.Address)
		weights = append(weights, backend.Weight)
	}

	if len(targets) == 0 {
		log.Fatal("No backends configured")
	}

	lb := NewLoadBalancer(targets, weights)

	mux := http.NewServeMux()
	mux.Handle("/", lb)

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	server := &http.Server{
		Addr:    config.Server.Port,
		Handler: mux,
	}

	log.Printf("Load balancer listening on %s", config.Server.Port)
	log.Printf("Backends: %d", len(targets))
	for i, target := range targets {
		log.Printf("  Backend %d: %s (weight: %d)", i, target, weights[i])
	}

	log.Fatal(server.ListenAndServe())
}
