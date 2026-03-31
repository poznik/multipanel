package config

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	toml "github.com/pelletier/go-toml/v2"
)

const (
	defaultListen          = "0.0.0.0:8080"
	defaultRefreshInterval = 10 * time.Second
	defaultRequestTimeout  = 5 * time.Second
	defaultScheme          = "http"
)

type FileConfig struct {
	Server ServerConfig `toml:"server"`
	Telemt TelemtConfig `toml:"telemt"`
}

type ServerConfig struct {
	Listen          string `toml:"listen"`
	RefreshInterval string `toml:"refresh_interval"`
	RequestTimeout  string `toml:"request_timeout"`
}

type TelemtConfig struct {
	AllowInsecureTLS bool                 `toml:"allow_insecure_tls"`
	Endpoints        []EndpointFileConfig `toml:"endpoints"`
}

type EndpointFileConfig struct {
	Name       string `toml:"name"`
	Scheme     string `toml:"scheme"`
	Address    string `toml:"address"`
	Port       int    `toml:"port"`
	AuthHeader string `toml:"auth_header"`
	Enabled    *bool  `toml:"enabled"`
}

type RuntimeConfig struct {
	ConfigPath        string
	Listen            string
	RefreshInterval   time.Duration
	RequestTimeout    time.Duration
	AllowInsecureTLS  bool
	EnabledEndpoints  []Endpoint
	DisabledEndpoints []Endpoint
}

type Endpoint struct {
	Name       string
	Scheme     string
	Address    string
	Port       int
	AuthHeader string
	Enabled    bool
}

func (e Endpoint) AddressWithPort() string {
	return net.JoinHostPort(e.Address, strconv.Itoa(e.Port))
}

func (e Endpoint) BaseURL() string {
	return fmt.Sprintf("%s://%s", e.Scheme, e.AddressWithPort())
}

func Load(path string) (RuntimeConfig, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return RuntimeConfig{}, fmt.Errorf("read config: %w", err)
	}

	var fileCfg FileConfig
	if err := toml.Unmarshal(raw, &fileCfg); err != nil {
		return RuntimeConfig{}, fmt.Errorf("parse config: %w", err)
	}

	refreshInterval, err := parseDuration(fileCfg.Server.RefreshInterval, defaultRefreshInterval)
	if err != nil {
		return RuntimeConfig{}, fmt.Errorf("parse server.refresh_interval: %w", err)
	}

	requestTimeout, err := parseDuration(fileCfg.Server.RequestTimeout, defaultRequestTimeout)
	if err != nil {
		return RuntimeConfig{}, fmt.Errorf("parse server.request_timeout: %w", err)
	}

	listen := strings.TrimSpace(fileCfg.Server.Listen)
	if listen == "" {
		listen = defaultListen
	}
	if _, _, err := net.SplitHostPort(listen); err != nil {
		return RuntimeConfig{}, fmt.Errorf("server.listen must be in host:port format: %w", err)
	}

	enabled := make([]Endpoint, 0, len(fileCfg.Telemt.Endpoints))
	disabled := make([]Endpoint, 0, len(fileCfg.Telemt.Endpoints))
	for i, endpointCfg := range fileCfg.Telemt.Endpoints {
		endpoint, err := normalizeEndpoint(endpointCfg, i)
		if err != nil {
			return RuntimeConfig{}, err
		}
		if endpoint.Enabled {
			enabled = append(enabled, endpoint)
		} else {
			disabled = append(disabled, endpoint)
		}
	}

	return RuntimeConfig{
		ConfigPath:        path,
		Listen:            listen,
		RefreshInterval:   refreshInterval,
		RequestTimeout:    requestTimeout,
		AllowInsecureTLS:  fileCfg.Telemt.AllowInsecureTLS,
		EnabledEndpoints:  enabled,
		DisabledEndpoints: disabled,
	}, nil
}

func normalizeEndpoint(cfg EndpointFileConfig, idx int) (Endpoint, error) {
	address := strings.TrimSpace(cfg.Address)
	if address == "" {
		return Endpoint{}, fmt.Errorf("telemt.endpoints[%d].address is required", idx)
	}
	if cfg.Port < 1 || cfg.Port > 65535 {
		return Endpoint{}, fmt.Errorf("telemt.endpoints[%d].port must be within [1, 65535]", idx)
	}

	scheme := strings.ToLower(strings.TrimSpace(cfg.Scheme))
	if scheme == "" {
		scheme = defaultScheme
	}
	if scheme != "http" && scheme != "https" {
		return Endpoint{}, fmt.Errorf("telemt.endpoints[%d].scheme must be http or https", idx)
	}

	name := strings.TrimSpace(cfg.Name)
	if name == "" {
		name = net.JoinHostPort(address, strconv.Itoa(cfg.Port))
	}

	enabled := true
	if cfg.Enabled != nil {
		enabled = *cfg.Enabled
	}

	return Endpoint{
		Name:       name,
		Scheme:     scheme,
		Address:    address,
		Port:       cfg.Port,
		AuthHeader: strings.TrimSpace(cfg.AuthHeader),
		Enabled:    enabled,
	}, nil
}

func parseDuration(raw string, fallback time.Duration) (time.Duration, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return fallback, nil
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return 0, err
	}
	if parsed <= 0 {
		return 0, errors.New("must be > 0")
	}
	return parsed, nil
}
