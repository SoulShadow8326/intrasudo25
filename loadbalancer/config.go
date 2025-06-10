package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

type Config struct {
	Backends      []BackendConfig `json:"backends"`
	Security      SecurityConf    `json:"security"`
	Cache         CacheConf       `json:"cache"`
	LoadBalancing LBConfig        `json:"load_balancing"`
	Server        ServerConfig    `json:"server"`
}

type BackendConfig struct {
	Address string        `json:"address"`
	Weight  int           `json:"weight"`
	Timeout time.Duration `json:"timeout"`
}

type SecurityConf struct {
	DDoSThreshold     int           `json:"ddos_threshold"`
	DoSThreshold      int           `json:"dos_threshold"`
	BanDuration       time.Duration `json:"ban_duration"`
	WindowSize        time.Duration `json:"window_size"`
	JSChallengeSecret string        `json:"js_challenge_secret"`
}

type CacheConf struct {
	RedisAddr  string        `json:"redis_addr"`
	DefaultTTL time.Duration `json:"default_ttl"`
}

type LBConfig struct {
	Strategy            string        `json:"strategy"`
	HealthCheckInterval time.Duration `json:"health_check_interval"`
}

type ServerConfig struct {
	Port         string        `json:"port"`
	ReadTimeout  time.Duration `json:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout"`
	IdleTimeout  time.Duration `json:"idle_timeout"`
}

func (bc *BackendConfig) UnmarshalJSON(data []byte) error {
	type Alias BackendConfig
	aux := &struct {
		Timeout string `json:"timeout"`
		*Alias
	}{
		Alias: (*Alias)(bc),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if aux.Timeout != "" {
		duration, err := time.ParseDuration(aux.Timeout)
		if err != nil {
			return err
		}
		bc.Timeout = duration
	}

	return nil
}

func (sc *SecurityConf) UnmarshalJSON(data []byte) error {
	type Alias SecurityConf
	aux := &struct {
		BanDuration string `json:"ban_duration"`
		WindowSize  string `json:"window_size"`
		*Alias
	}{
		Alias: (*Alias)(sc),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if aux.BanDuration != "" {
		duration, err := time.ParseDuration(aux.BanDuration)
		if err != nil {
			return err
		}
		sc.BanDuration = duration
	}

	if aux.WindowSize != "" {
		duration, err := time.ParseDuration(aux.WindowSize)
		if err != nil {
			return err
		}
		sc.WindowSize = duration
	}

	return nil
}

func (cc *CacheConf) UnmarshalJSON(data []byte) error {
	type Alias CacheConf
	aux := &struct {
		DefaultTTL string `json:"default_ttl"`
		*Alias
	}{
		Alias: (*Alias)(cc),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if aux.DefaultTTL != "" {
		duration, err := time.ParseDuration(aux.DefaultTTL)
		if err != nil {
			return err
		}
		cc.DefaultTTL = duration
	}

	return nil
}

func (lbc *LBConfig) UnmarshalJSON(data []byte) error {
	type Alias LBConfig
	aux := &struct {
		HealthCheckInterval string `json:"health_check_interval"`
		*Alias
	}{
		Alias: (*Alias)(lbc),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if aux.HealthCheckInterval != "" {
		duration, err := time.ParseDuration(aux.HealthCheckInterval)
		if err != nil {
			return err
		}
		lbc.HealthCheckInterval = duration
	}

	return nil
}

func (svc *ServerConfig) UnmarshalJSON(data []byte) error {
	type Alias ServerConfig
	aux := &struct {
		ReadTimeout  string `json:"read_timeout"`
		WriteTimeout string `json:"write_timeout"`
		IdleTimeout  string `json:"idle_timeout"`
		*Alias
	}{
		Alias: (*Alias)(svc),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if aux.ReadTimeout != "" {
		duration, err := time.ParseDuration(aux.ReadTimeout)
		if err != nil {
			return err
		}
		svc.ReadTimeout = duration
	}

	if aux.WriteTimeout != "" {
		duration, err := time.ParseDuration(aux.WriteTimeout)
		if err != nil {
			return err
		}
		svc.WriteTimeout = duration
	}

	if aux.IdleTimeout != "" {
		duration, err := time.ParseDuration(aux.IdleTimeout)
		if err != nil {
			return err
		}
		svc.IdleTimeout = duration
	}

	return nil
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

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := map[string]string{
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"service":   "SoulLoad",
	}
	json.NewEncoder(w).Encode(response)
}

func metricsHandler(lb *LoadBalancer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")

		lb.mu.RLock()
		defer lb.mu.RUnlock()

		for _, backend := range lb.backends {
			backend.mu.RLock()
			fmt.Fprintf(w, "# HELP backend_requests_total Total requests to backend\n")
			fmt.Fprintf(w, "# TYPE backend_requests_total counter\n")
			fmt.Fprintf(w, "backend_requests_total{backend=\"%s\"} %d\n", backend.ID, backend.totalRequests)

			fmt.Fprintf(w, "# HELP backend_failures_total Total failures for backend\n")
			fmt.Fprintf(w, "# TYPE backend_failures_total counter\n")
			fmt.Fprintf(w, "backend_failures_total{backend=\"%s\"} %d\n", backend.ID, backend.failCount)

			fmt.Fprintf(w, "# HELP backend_response_time_ms Response time in milliseconds\n")
			fmt.Fprintf(w, "# TYPE backend_response_time_ms gauge\n")
			fmt.Fprintf(w, "backend_response_time_ms{backend=\"%s\"} %f\n", backend.ID, float64(backend.responseTime.Nanoseconds())/1e6)

			alive := 0
			if backend.alive {
				alive = 1
			}
			fmt.Fprintf(w, "# HELP backend_alive Backend alive status\n")
			fmt.Fprintf(w, "# TYPE backend_alive gauge\n")
			fmt.Fprintf(w, "backend_alive{backend=\"%s\"} %d\n", backend.ID, alive)
			backend.mu.RUnlock()
		}
	}
}
