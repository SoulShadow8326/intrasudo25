package main

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	RoundRobin = iota
	WeightedRoundRobin
	FastestFirst
)

const (
	CircuitClosed = iota
	CircuitOpen
	CircuitHalfOpen
)

type Backend struct {
	ID               string
	URL              *url.URL
	Proxy            http.Handler
	mu               sync.RWMutex
	alive            bool
	weight           int
	currentWeight    int
	timeout          time.Duration
	isUnixSocket     bool
	socketPath       string
	totalRequests    int64
	failCount        int64
	responseTime     time.Duration
	retryWindow      time.Duration
	strictMode       bool
	circuitState     int32
	failureThreshold int
	resetTimeout     time.Duration
	lastStateChange  time.Time
	successCount     int64
	avgLatency       time.Duration
	lastAccess       time.Time
}

type RateLimitEntry struct {
	requests    []time.Time
	lastRequest time.Time
}

type SecurityManager struct {
	rateLimiters  map[string]*RateLimitEntry
	bannedIPs     map[string]time.Time
	ddosThreshold int
	dosThreshold  int
	banDuration   time.Duration
	windowSize    time.Duration
	maxTrackedIPs int
	hmacSecret    []byte
	jsSecret      string
	mu            sync.RWMutex
}

type CacheEntry struct {
	data      []byte
	headers   map[string]string
	expiresAt time.Time
}

type CacheManager struct {
	cache map[string]*CacheEntry
	mu    sync.RWMutex
}

type MetricsCollector struct {
	totalRequests    int64
	totalErrors      int64
	totalLatency     int64
	requestsByMinute map[int64]int64
	errorsByMinute   map[int64]int64
	mu               sync.RWMutex
}

type LoadBalancer struct {
	backends       []*Backend
	mu             sync.RWMutex
	strategy       int
	security       *SecurityManager
	cache          *CacheManager
	healthInterval time.Duration
	metrics        *MetricsCollector
	startTime      time.Time
}

type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	body       []byte
}

func (rr *responseRecorder) WriteHeader(statusCode int) {
	rr.statusCode = statusCode
	rr.ResponseWriter.WriteHeader(statusCode)
}

func (rr *responseRecorder) Write(data []byte) (int, error) {
	rr.body = append(rr.body, data...)
	return rr.ResponseWriter.Write(data)
}

func (b *Backend) IsAlive() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.alive
}

func (b *Backend) SetAlive(up bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if !up {
		atomic.AddInt64(&b.failCount, 1)
	}
	b.alive = up
}

func (b *Backend) IncrementRequests() {
	atomic.AddInt64(&b.totalRequests, 1)
}

func (b *Backend) UpdateResponseTime(duration time.Duration) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.responseTime = duration
	b.lastAccess = time.Now()

	if b.avgLatency == 0 {
		b.avgLatency = duration
	} else {
		b.avgLatency = time.Duration(float64(b.avgLatency)*0.9 + float64(duration)*0.1)
	}
}

func (b *Backend) GetResponseTime() time.Duration {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.responseTime
}

func (b *Backend) CanAcceptRequest() bool {
	state := atomic.LoadInt32(&b.circuitState)
	switch state {
	case CircuitClosed:
		return true
	case CircuitOpen:
		b.mu.RLock()
		canTry := time.Since(b.lastStateChange) > b.resetTimeout
		b.mu.RUnlock()
		if canTry {
			atomic.CompareAndSwapInt32(&b.circuitState, CircuitOpen, CircuitHalfOpen)
			b.mu.Lock()
			b.lastStateChange = time.Now()
			b.mu.Unlock()
			return true
		}
		return false
	case CircuitHalfOpen:
		return true
	default:
		return false
	}
}

func (b *Backend) RecordSuccess() {
	atomic.AddInt64(&b.successCount, 1)
	state := atomic.LoadInt32(&b.circuitState)
	if state == CircuitHalfOpen {
		atomic.CompareAndSwapInt32(&b.circuitState, CircuitHalfOpen, CircuitClosed)
		b.mu.Lock()
		b.lastStateChange = time.Now()
		b.failCount = 0
		b.mu.Unlock()
	}
}

func (b *Backend) RecordFailure() {
	atomic.AddInt64(&b.failCount, 1)
	failures := atomic.LoadInt64(&b.failCount)

	if failures >= int64(b.failureThreshold) {
		atomic.CompareAndSwapInt32(&b.circuitState, CircuitClosed, CircuitOpen)
		b.mu.Lock()
		b.lastStateChange = time.Now()
		b.mu.Unlock()
	}
}

func (b *Backend) GetCircuitState() string {
	state := atomic.LoadInt32(&b.circuitState)
	switch state {
	case CircuitClosed:
		return "closed"
	case CircuitOpen:
		return "open"
	case CircuitHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
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

	client := &http.Client{
		Timeout: timeout,
	}

	if isUnixSocket {
		client.Transport = &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return net.Dial("unix", socketPath)
			},
		}
	}

	proxy := &customProxy{
		target: u,
		client: client,
	}

	backend := &Backend{
		ID:               id,
		URL:              u,
		Proxy:            proxy,
		alive:            true,
		weight:           weight,
		currentWeight:    weight,
		timeout:          timeout,
		isUnixSocket:     isUnixSocket,
		socketPath:       socketPath,
		retryWindow:      30 * time.Second,
		strictMode:       false,
		circuitState:     CircuitClosed,
		failureThreshold: 5,
		resetTimeout:     30 * time.Second,
		lastStateChange:  time.Now(),
		lastAccess:       time.Now(),
	}

	proxy.backend = backend
	return backend
}

func NewSecurityManager(ddosThreshold, dosThreshold int, banDuration, windowSize time.Duration, hmacSecret []byte, jsSecret string) *SecurityManager {
	return &SecurityManager{
		rateLimiters:  make(map[string]*RateLimitEntry),
		bannedIPs:     make(map[string]time.Time),
		ddosThreshold: ddosThreshold,
		dosThreshold:  dosThreshold,
		banDuration:   banDuration,
		windowSize:    windowSize,
		maxTrackedIPs: 1000,
		hmacSecret:    hmacSecret,
		jsSecret:      jsSecret,
	}
}

func (sm *SecurityManager) IsBlocked(ip string) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	if banTime, exists := sm.bannedIPs[ip]; exists {
		if time.Since(banTime) < sm.banDuration {
			return true
		}
		delete(sm.bannedIPs, ip)
	}
	return false
}

func (sm *SecurityManager) CheckDDoS(ip string) bool {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if len(sm.rateLimiters) >= sm.maxTrackedIPs {
		oldest := ""
		oldestTime := time.Now()
		for k, v := range sm.rateLimiters {
			if v.lastRequest.Before(oldestTime) {
				oldestTime = v.lastRequest
				oldest = k
			}
		}
		if oldest != "" {
			delete(sm.rateLimiters, oldest)
		}
	}

	if _, exists := sm.rateLimiters[ip]; !exists {
		sm.rateLimiters[ip] = &RateLimitEntry{}
	}

	entry := sm.rateLimiters[ip]
	now := time.Now()
	cutoff := now.Add(-sm.windowSize)

	var validRequests []time.Time
	for _, req := range entry.requests {
		if req.After(cutoff) {
			validRequests = append(validRequests, req)
		}
	}
	entry.requests = validRequests
	entry.requests = append(entry.requests, now)

	if len(entry.requests) > sm.ddosThreshold {
		sm.bannedIPs[ip] = now
		return true
	}
	return false
}

func (sm *SecurityManager) CheckDoS(ip string) bool {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if _, exists := sm.rateLimiters[ip]; !exists {
		sm.rateLimiters[ip] = &RateLimitEntry{}
	}

	entry := sm.rateLimiters[ip]
	now := time.Now()

	if time.Since(entry.lastRequest) < time.Second {
		return len(entry.requests) > sm.dosThreshold
	}

	entry.lastRequest = now
	return false
}

func (sm *SecurityManager) VerifyHMAC(data, signature string) bool {
	mac := hmac.New(sha256.New, sm.hmacSecret)
	mac.Write([]byte(data))
	expectedMAC := mac.Sum(nil)
	receivedMAC, err := hex.DecodeString(signature)
	return err == nil && hmac.Equal(expectedMAC, receivedMAC)
}

func (sm *SecurityManager) GenerateJSChallenge() (string, string) {
	challenge := make([]byte, 16)
	rand.Read(challenge)
	challengeStr := hex.EncodeToString(challenge)

	answer := fmt.Sprintf("%x", sha256.Sum256([]byte(challengeStr+sm.jsSecret)))

	jsCode := fmt.Sprintf(`
		const challenge = '%s';
		const secret = '%s';
		async function solve() {
			const encoder = new TextEncoder();
			const data = encoder.encode(challenge + secret);
			const hashBuffer = await crypto.subtle.digest('SHA-256', data);
			const hashArray = Array.from(new Uint8Array(hashBuffer));
			return hashArray.map(b => b.toString(16).padStart(2, '0')).join('');
		}
	`, challengeStr, sm.jsSecret)

	return jsCode, answer
}

func NewCacheManager() *CacheManager {
	return &CacheManager{
		cache: make(map[string]*CacheEntry),
	}
}

func (cm *CacheManager) Get(key string) ([]byte, map[string]string, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	if entry, exists := cm.cache[key]; exists && time.Now().Before(entry.expiresAt) {
		return entry.data, entry.headers, true
	}
	return nil, nil, false
}

func (cm *CacheManager) Set(key string, data []byte, headers map[string]string, ttl time.Duration) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.cache[key] = &CacheEntry{
		data:      data,
		headers:   headers,
		expiresAt: time.Now().Add(ttl),
	}
}

func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		requestsByMinute: make(map[int64]int64),
		errorsByMinute:   make(map[int64]int64),
	}
}

func (mc *MetricsCollector) RecordRequest(latency time.Duration, isError bool) {
	atomic.AddInt64(&mc.totalRequests, 1)
	atomic.AddInt64(&mc.totalLatency, int64(latency))

	if isError {
		atomic.AddInt64(&mc.totalErrors, 1)
	}

	minute := time.Now().Unix() / 60
	mc.mu.Lock()
	mc.requestsByMinute[minute]++
	if isError {
		mc.errorsByMinute[minute]++
	}
	mc.mu.Unlock()
}

func (mc *MetricsCollector) GetStats() map[string]interface{} {
	totalReqs := atomic.LoadInt64(&mc.totalRequests)
	totalErrs := atomic.LoadInt64(&mc.totalErrors)
	totalLat := atomic.LoadInt64(&mc.totalLatency)

	var avgLatency float64
	if totalReqs > 0 {
		avgLatency = float64(totalLat) / float64(totalReqs) / float64(time.Millisecond)
	}

	errorRate := float64(0)
	if totalReqs > 0 {
		errorRate = float64(totalErrs) / float64(totalReqs) * 100
	}

	return map[string]interface{}{
		"total_requests":      totalReqs,
		"total_errors":        totalErrs,
		"error_rate_percent":  errorRate,
		"avg_latency_ms":      avgLatency,
		"requests_per_minute": mc.getRecentRequests(),
	}
}

func (mc *MetricsCollector) getRecentRequests() int64 {
	minute := time.Now().Unix() / 60
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	return mc.requestsByMinute[minute]
}

func NewLoadBalancer(targets []string, weights []int, config *Config) *LoadBalancer {
	var backends []*Backend

	for i, target := range targets {
		weight := 1
		if i < len(weights) {
			weight = weights[i]
		}

		timeout := 10 * time.Second
		if i < len(config.Backends) {
			timeout = config.Backends[i].Timeout
		}

		backend := NewBackend(fmt.Sprintf("backend-%d", i), target, weight, timeout)
		backends = append(backends, backend)
	}

	hmacSecret := []byte("default-hmac-secret")
	if config.Security.JSChallengeSecret != "" {
		hmacSecret = []byte(config.Security.JSChallengeSecret)
	}

	lb := &LoadBalancer{
		backends:       backends,
		strategy:       RoundRobin,
		security:       NewSecurityManager(config.Security.DDoSThreshold, config.Security.DoSThreshold, config.Security.BanDuration, config.Security.WindowSize, hmacSecret, config.Security.JSChallengeSecret),
		cache:          NewCacheManager(),
		healthInterval: config.LoadBalancing.HealthCheckInterval,
		metrics:        NewMetricsCollector(),
		startTime:      time.Now(),
	}

	if config.LoadBalancing.Strategy == "weighted_round_robin" {
		lb.strategy = WeightedRoundRobin
	} else if config.LoadBalancing.Strategy == "fastest_first" {
		lb.strategy = FastestFirst
	}

	go lb.cleanupLoop()
	go lb.healthLoop()

	return lb
}

func (lb *LoadBalancer) cleanupLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()

		lb.security.mu.Lock()
		for ip, entry := range lb.security.rateLimiters {
			if now.Sub(entry.lastRequest) > 10*time.Minute {
				delete(lb.security.rateLimiters, ip)
			}
		}
		lb.security.mu.Unlock()

		lb.cache.mu.Lock()
		for key, entry := range lb.cache.cache {
			if now.After(entry.expiresAt) {
				delete(lb.cache.cache, key)
			}
		}
		lb.cache.mu.Unlock()
	}
}

func (lb *LoadBalancer) healthLoop() {
	ticker := time.NewTicker(lb.healthInterval)
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
				resp, err := client.Get(backend.URL.String() + "/health")
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
		if b.IsAlive() && b.CanAcceptRequest() {
			aliveBackends = append(aliveBackends, b)
		}
	}

	if len(aliveBackends) == 0 {
		return nil
	}

	switch lb.strategy {
	case FastestFirst:
		sort.Slice(aliveBackends, func(i, j int) bool {
			return aliveBackends[i].GetResponseTime() < aliveBackends[j].GetResponseTime()
		})
		return aliveBackends[0]
	default:
		return lb.getWeightedBackend(aliveBackends)
	}
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
	start := time.Now()
	clientIP := getClientIP(r)

	if lb.security.IsBlocked(clientIP) {
		http.Error(w, "IP banned", http.StatusForbidden)
		lb.metrics.RecordRequest(time.Since(start), true)
		return
	}

	if lb.security.CheckDDoS(clientIP) {
		http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
		lb.metrics.RecordRequest(time.Since(start), true)
		return
	}

	if lb.security.CheckDoS(clientIP) {
		http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
		lb.metrics.RecordRequest(time.Since(start), true)
		return
	}

	if hmacSig := r.Header.Get("X-HMAC-Signature"); hmacSig != "" {
		body, _ := io.ReadAll(r.Body)
		if !lb.security.VerifyHMAC(string(body), hmacSig) {
			http.Error(w, "Invalid signature", http.StatusUnauthorized)
			lb.metrics.RecordRequest(time.Since(start), true)
			return
		}
	}

	if r.Header.Get("X-JS-Challenge") == "request" {
		jsCode, _ := lb.security.GenerateJSChallenge()
		w.Header().Set("Content-Type", "application/javascript")
		w.Write([]byte(jsCode))
		lb.metrics.RecordRequest(time.Since(start), false)
		return
	}

	cacheKey := r.Method + ":" + r.URL.String()
	if r.Method == "GET" && !strings.Contains(r.URL.Path, "/api/") {
		if data, headers, found := lb.cache.Get(cacheKey); found {
			for k, v := range headers {
				w.Header().Set(k, v)
			}
			w.Write(data)
			lb.metrics.RecordRequest(time.Since(start), false)
			return
		}
	}

	backend := lb.getNextBackend()
	if backend == nil {
		http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
		lb.metrics.RecordRequest(time.Since(start), true)
		return
	}

	backend.IncrementRequests()

	ctx, cancel := context.WithTimeout(r.Context(), backend.timeout)
	defer cancel()
	r = r.WithContext(ctx)

	recorder := &responseRecorder{ResponseWriter: w, statusCode: http.StatusOK}
	backend.Proxy.ServeHTTP(recorder, r)

	duration := time.Since(start)
	backend.UpdateResponseTime(duration)

	isError := recorder.statusCode >= 500
	lb.metrics.RecordRequest(duration, isError)

	if isError {
		backend.SetAlive(false)
		backend.RecordFailure()
	} else {
		backend.RecordSuccess()
	}

	if r.Method == "GET" && recorder.statusCode == http.StatusOK && recorder.body != nil && !strings.Contains(r.URL.Path, "/api/") && !strings.Contains(r.URL.Path, "/admin") {
		headers := make(map[string]string)
		for k, v := range recorder.Header() {
			if len(v) > 0 {
				headers[k] = v[0]
			}
		}
		lb.cache.Set(cacheKey, recorder.body, headers, 5*time.Minute)
	}
}

func getClientIP(r *http.Request) string {
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		return strings.Split(forwarded, ",")[0]
	}

	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	return strings.Split(r.RemoteAddr, ":")[0]
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

	lb := NewLoadBalancer(targets, weights, config)

	mux := http.NewServeMux()
	mux.Handle("/", lb)

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	mux.HandleFunc("/admin/stats", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		stats := make(map[string]interface{})
		lb.mu.RLock()
		for _, b := range lb.backends {
			stats[b.ID] = map[string]interface{}{
				"alive":        b.IsAlive(),
				"responseTime": b.GetResponseTime().String(),
				"requests":     atomic.LoadInt64(&b.totalRequests),
				"failures":     atomic.LoadInt64(&b.failCount),
				"circuitState": b.GetCircuitState(),
				"avgLatency":   b.avgLatency.String(),
			}
		}
		lb.mu.RUnlock()

		stats["metrics"] = lb.metrics.GetStats()
		stats["uptime"] = time.Since(lb.startTime).String()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(stats)
	})

	mux.HandleFunc("/admin/backend", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			var data map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
				http.Error(w, "Invalid JSON", http.StatusBadRequest)
				return
			}

			var addr string
			var weight int = 1

			if url, ok := data["url"].(string); ok {
				addr = url
			} else if address, ok := data["address"].(string); ok {
				addr = address
			} else {
				http.Error(w, "Missing url or address field", http.StatusBadRequest)
				return
			}

			if w, ok := data["weight"].(float64); ok {
				weight = int(w)
			}

			backend := NewBackend(fmt.Sprintf("backend-%d", len(lb.backends)), addr, weight, 10*time.Second)
			lb.mu.Lock()
			lb.backends = append(lb.backends, backend)
			lb.mu.Unlock()

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"status": "Backend added"})

		case "DELETE":
			id := r.URL.Query().Get("id")
			if id == "" {
				http.Error(w, "Backend ID required", http.StatusBadRequest)
				return
			}
			lb.mu.Lock()
			for i, b := range lb.backends {
				if b.ID == id {
					lb.backends = append(lb.backends[:i], lb.backends[i+1:]...)
					break
				}
			}
			lb.mu.Unlock()

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"status": "Backend removed"})

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/admin/strategy", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			strategyStr := r.FormValue("strategy")
			var strategy int
			switch strategyStr {
			case "fastest_first":
				strategy = FastestFirst
			case "weighted_round_robin":
				strategy = WeightedRoundRobin
			default:
				strategy = RoundRobin
			}

			lb.mu.Lock()
			lb.strategy = strategy
			lb.mu.Unlock()

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"status": "Strategy updated"})
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	server := &http.Server{
		Addr:         config.Server.Port,
		Handler:      mux,
		ReadTimeout:  config.Server.ReadTimeout,
		WriteTimeout: config.Server.WriteTimeout,
		IdleTimeout:  config.Server.IdleTimeout,
	}

	log.Printf("Load balancer listening on %s", config.Server.Port)
	log.Printf("Backends: %d", len(targets))
	for i, target := range targets {
		log.Printf("  Backend %d: %s (weight: %d)", i, target, weights[i])
	}

	log.Fatal(server.ListenAndServe())
}

type customProxy struct {
	backend *Backend
	target  *url.URL
	client  *http.Client
}

func (cp *customProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	target := cp.target
	if cp.backend.isUnixSocket {
		target, _ = url.Parse("http://localhost")
	}

	req := &http.Request{
		Method: r.Method,
		URL: &url.URL{
			Scheme:   target.Scheme,
			Host:     target.Host,
			Path:     r.URL.Path,
			RawQuery: r.URL.RawQuery,
		},
		Header: make(http.Header),
		Body:   r.Body,
	}

	for k, vv := range r.Header {
		for _, v := range vv {
			req.Header.Add(k, v)
		}
	}

	resp, err := cp.client.Do(req)
	if err != nil {
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	for k, vv := range resp.Header {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}
