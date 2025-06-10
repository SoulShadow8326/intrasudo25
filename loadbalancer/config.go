package main

import (
	"encoding/json"
	"os"
	"time"
)

type Config struct {
	Backends      []BackendConfig `json:"backends"`
	Security      SecurityConf    `json:"security"`
	Cache         CacheConf       `json:"cache"`
	LoadBalancing LBConfig        `json:"load_balancing"`
	Server        ServerConfig    `json:"server"`
	Monitoring    MonitoringConf  `json:"monitoring"`
	Features      FeaturesConf    `json:"features"`
}

type BackendConfig struct {
	Address string        `json:"address"`
	Weight  int           `json:"weight"`
	Timeout time.Duration `json:"timeout"`
}

type SecurityConf struct {
	DDoSThreshold     int                   `json:"ddos_threshold"`
	DoSThreshold      int                   `json:"dos_threshold"`
	BanDuration       time.Duration         `json:"ban_duration"`
	WindowSize        time.Duration         `json:"window_size"`
	JSChallengeSecret string                `json:"js_challenge_secret"`
	GeoBlocking       GeoBlockingConf       `json:"geo_blocking"`
	UserAgentBlocking UserAgentBlockingConf `json:"user_agent_blocking"`
}

type GeoBlockingConf struct {
	Enabled          bool     `json:"enabled"`
	BlockedCountries []string `json:"blocked_countries"`
}

type UserAgentBlockingConf struct {
	Enabled         bool     `json:"enabled"`
	BlockedPatterns []string `json:"blocked_patterns"`
}

type CacheConf struct {
	RedisAddr         string        `json:"redis_addr"`
	DefaultTTL        time.Duration `json:"default_ttl"`
	MaxSize           string        `json:"max_size"`
	EnableCompression bool          `json:"enable_compression"`
}

type LBConfig struct {
	Strategy            string             `json:"strategy"`
	HealthCheckInterval time.Duration      `json:"health_check_interval"`
	CircuitBreaker      CircuitBreakerConf `json:"circuit_breaker"`
}

type CircuitBreakerConf struct {
	FailureThreshold int           `json:"failure_threshold"`
	ResetTimeout     time.Duration `json:"reset_timeout"`
}

type ServerConfig struct {
	Port         string        `json:"port"`
	ReadTimeout  time.Duration `json:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout"`
	IdleTimeout  time.Duration `json:"idle_timeout"`
	EnableSSL    bool          `json:"enable_ssl"`
	CertFile     string        `json:"cert_file"`
	KeyFile      string        `json:"key_file"`
}

type MonitoringConf struct {
	EnableTracing      bool          `json:"enable_tracing"`
	EnableMetrics      bool          `json:"enable_metrics"`
	PrometheusEndpoint string        `json:"prometheus_endpoint"`
	TraceRetention     time.Duration `json:"trace_retention"`
}

type FeaturesConf struct {
	WebSocketSupport bool `json:"websocket_support"`
	Compression      bool `json:"compression"`
	RequestLogging   bool `json:"request_logging"`
	AdminAPI         bool `json:"admin_api"`
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

func (cb *CircuitBreakerConf) UnmarshalJSON(data []byte) error {
	type Alias CircuitBreakerConf
	aux := &struct {
		ResetTimeout string `json:"reset_timeout"`
		*Alias
	}{
		Alias: (*Alias)(cb),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if aux.ResetTimeout != "" {
		duration, err := time.ParseDuration(aux.ResetTimeout)
		if err != nil {
			return err
		}
		cb.ResetTimeout = duration
	}

	return nil
}

func (mc *MonitoringConf) UnmarshalJSON(data []byte) error {
	type Alias MonitoringConf
	aux := &struct {
		TraceRetention string `json:"trace_retention"`
		*Alias
	}{
		Alias: (*Alias)(mc),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if aux.TraceRetention != "" {
		duration, err := time.ParseDuration(aux.TraceRetention)
		if err != nil {
			return err
		}
		mc.TraceRetention = duration
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
