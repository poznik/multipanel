package dashboard

import (
	"context"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
	"time"

	"multipanel/internal/config"
	"multipanel/internal/telemt"
)

func TestRefreshNowAggregatesTelemtEndpoints(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/v1/health":
			io.WriteString(w, `{"ok":true,"data":{"status":"ok","read_only":false}}`)
		case "/v1/stats/summary":
			io.WriteString(w, `{"ok":true,"data":{"uptime_seconds":3661,"connections_total":42,"connections_bad_total":2,"handshake_timeouts_total":5,"configured_users":3}}`)
		case "/v1/users":
			io.WriteString(w, `{"ok":true,"data":[{"username":"alice","in_runtime":true,"current_connections":3,"active_unique_ips":2,"recent_unique_ips":2,"total_octets":1500},{"username":"bob","in_runtime":true,"current_connections":1,"active_unique_ips":1,"recent_unique_ips":1,"total_octets":500},{"username":"carol","in_runtime":false,"current_connections":0,"active_unique_ips":0,"recent_unique_ips":0,"total_octets":0}]}`)
		case "/v1/stats/users/active-ips":
			io.WriteString(w, `{"ok":true,"data":[{"username":"alice","active_ips":["1.1.1.1","2.2.2.2"]},{"username":"bob","active_ips":["2.2.2.2"]}]}`)
		case "/v1/runtime/gates":
			io.WriteString(w, `{"ok":true,"data":{"use_middle_proxy":true,"route_mode":"middle","me_runtime_ready":true,"me2dc_fallback_enabled":true,"me2dc_fast_enabled":false,"conditional_cast_enabled":true,"reroute_active":false,"accepting_new_connections":true,"startup_status":"ready","startup_stage":"ready","startup_progress_pct":100}}`)
		case "/v1/runtime/me_quality":
			io.WriteString(w, `{"ok":true,"data":{"enabled":true,"data":{"counters":{"idle_close_by_peer_total":1,"reader_eof_total":2,"kdf_drift_total":0,"kdf_port_only_drift_total":0,"reconnect_attempt_total":4,"reconnect_success_total":3},"route_drops":{"no_conn_total":5,"channel_closed_total":0,"queue_full_total":0,"queue_full_base_total":0,"queue_full_high_total":0},"family_states":[{"family":"v4","state":"healthy","state_since_epoch_secs":1,"fail_streak":0,"recover_success_streak":0}],"drain_gate":{"route_quorum_ok":true,"redundancy_ok":false,"block_reason":"redundancy_low","updated_at_epoch_secs":2},"dc_rtt":[{"dc":1,"rtt_ema_ms":12.5,"alive_writers":2,"required_writers":2,"coverage_pct":100}]}}}`)
		case "/v1/runtime/upstream_quality":
			io.WriteString(w, `{"ok":true,"data":{"enabled":true,"policy":{"connect_retry_attempts":2,"connect_retry_backoff_ms":100,"connect_budget_ms":3000,"unhealthy_fail_threshold":5,"connect_failfast_hard_errors":false},"counters":{"connect_attempt_total":10,"connect_success_total":9,"connect_fail_total":1,"connect_failfast_hard_error_total":0},"summary":{"configured_total":1,"healthy_total":1,"unhealthy_total":0,"direct_total":1,"socks4_total":0,"socks5_total":0,"shadowsocks_total":0},"upstreams":[{"upstream_id":0,"route_kind":"direct","address":"direct","weight":1,"scopes":"","healthy":true,"fails":0,"last_check_age_secs":1,"effective_latency_ms":11.2,"dc":[{"dc":1,"latency_ema_ms":12.5,"ip_preference":"prefer_v4"}]}]}}`)
		case "/v1/stats/me-writers":
			io.WriteString(w, `{"ok":true,"data":{"middle_proxy_enabled":true,"summary":{"configured_dc_groups":1,"configured_endpoints":2,"available_endpoints":2,"available_pct":100,"required_writers":2,"alive_writers":2,"coverage_pct":100,"fresh_alive_writers":2,"fresh_coverage_pct":100},"writers":[{"writer_id":1,"dc":1,"endpoint":"dc1-a","generation":2,"state":"active","draining":false,"degraded":false,"bound_clients":1,"idle_for_secs":null,"rtt_ema_ms":12.5,"drain_over_ttl":false},{"writer_id":2,"dc":1,"endpoint":"dc1-b","generation":2,"state":"active","draining":false,"degraded":true,"bound_clients":0,"idle_for_secs":30,"rtt_ema_ms":14.5,"drain_over_ttl":false}]}}`)
		case "/v1/stats/dcs":
			io.WriteString(w, `{"ok":true,"data":{"middle_proxy_enabled":true,"dcs":[{"dc":1,"available_endpoints":2,"available_pct":100,"required_writers":2,"floor_min":1,"floor_target":2,"floor_max":4,"floor_capped":false,"alive_writers":2,"coverage_pct":100,"fresh_alive_writers":2,"fresh_coverage_pct":100,"rtt_ms":12.5,"load":1}]}}`)
		case "/v1/runtime/me_pool_state":
			io.WriteString(w, `{"ok":true,"data":{"enabled":true,"data":{"generations":{"active_generation":2,"warm_generation":0,"pending_hardswap_generation":0,"pending_hardswap_age_secs":null,"draining_generations":[]},"hardswap":{"enabled":true,"pending":false},"writers":{"total":2,"alive_non_draining":2,"draining":0,"degraded":1,"contour":{"warm":0,"active":2,"draining":0},"health":{"healthy":1,"degraded":1,"draining":0}},"refill":{"inflight_endpoints_total":0,"inflight_dc_total":0,"by_dc":[]}}}}`)
		case "/v1/stats/minimal/all":
			io.WriteString(w, `{"ok":true,"data":{"enabled":true,"data":{"me_runtime":{"quarantined_endpoints_total":1,"quarantined_endpoints":[{"endpoint":"dc1-b","remaining_ms":120000}]}}}}`)
		case "/v1/runtime/nat_stun":
			io.WriteString(w, `{"ok":true,"data":{"enabled":true,"data":{"flags":{"nat_probe_enabled":true,"nat_probe_disabled_runtime":false,"nat_probe_attempts":0},"servers":{"configured":["a","b"],"live":["a"],"live_total":1},"reflection":{"v4":{"addr":"1.2.3.4:1234","age_secs":60}},"stun_backoff_remaining_ms":null}}}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	endpoint := endpointFromURL(t, server.URL)
	cfg := config.RuntimeConfig{
		ConfigPath:       "config.test.toml",
		Listen:           "127.0.0.1:8080",
		RefreshInterval:  10 * time.Second,
		RequestTimeout:   2 * time.Second,
		EnabledEndpoints: []config.Endpoint{endpoint},
	}
	service := NewService(
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		cfg,
		telemt.NewClient(cfg.RequestTimeout, false),
	)

	snapshot := service.RefreshNow(context.Background())

	if got, want := len(snapshot.Servers), 1; got != want {
		t.Fatalf("servers len = %d, want %d", got, want)
	}
	if got, want := snapshot.Totals.ConnectionsTotal, uint64(42); got != want {
		t.Fatalf("connections_total = %d, want %d", got, want)
	}
	if got, want := snapshot.Totals.ActiveIPsTotal, 2; got != want {
		t.Fatalf("active_ips_total = %d, want %d", got, want)
	}
	if got, want := snapshot.Servers[0].Mode, "ME"; got != want {
		t.Fatalf("mode = %q, want %q", got, want)
	}
	if got, want := snapshot.Servers[0].UsersConnectionsTotal, uint64(4); got != want {
		t.Fatalf("users_connections_total = %d, want %d", got, want)
	}
	if got, want := snapshot.Servers[0].QualityKind, "me"; got != want {
		t.Fatalf("quality_kind = %q, want %q", got, want)
	}
	if got, want := snapshot.Servers[0].Runtime.Available, true; got != want {
		t.Fatalf("runtime.available = %v, want %v", got, want)
	}
	if got, want := snapshot.Servers[0].Upstream.Available, true; got != want {
		t.Fatalf("upstream.available = %v, want %v", got, want)
	}
	if got, want := snapshot.Servers[0].Writers.Available, true; got != want {
		t.Fatalf("writers.available = %v, want %v", got, want)
	}
	if got, want := snapshot.Servers[0].Pool.Available, true; got != want {
		t.Fatalf("pool.available = %v, want %v", got, want)
	}
	if got, want := snapshot.Servers[0].Reliability.MEAvailable, true; got != want {
		t.Fatalf("reliability.me_available = %v, want %v", got, want)
	}
	if got, want := snapshot.Servers[0].Reliability.QuarantinedEndpointsTotal, 1; got != want {
		t.Fatalf("quarantined_endpoints_total = %d, want %d", got, want)
	}
	if got, want := snapshot.Servers[0].UserActivity.ActiveUsers, 2; got != want {
		t.Fatalf("active_users = %d, want %d", got, want)
	}
	if snapshot.Servers[0].UserActivity.TopConnectionsUser == nil || snapshot.Servers[0].UserActivity.TopConnectionsUser.Username != "alice" {
		t.Fatalf("top_connections_user = %#v, want alice", snapshot.Servers[0].UserActivity.TopConnectionsUser)
	}
	if got, want := snapshot.Servers[0].Network.Available, true; got != want {
		t.Fatalf("network.available = %v, want %v", got, want)
	}
}

func TestRefreshNowBuildsDirectDatacenterHealthForDCMode(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/v1/health":
			io.WriteString(w, `{"ok":true,"data":{"status":"ok","read_only":true}}`)
		case "/v1/stats/summary":
			io.WriteString(w, `{"ok":true,"data":{"uptime_seconds":900,"connections_total":7,"connections_bad_total":0,"handshake_timeouts_total":0,"configured_users":2}}`)
		case "/v1/users":
			io.WriteString(w, `{"ok":true,"data":[{"username":"work","in_runtime":true,"current_connections":2,"active_unique_ips":1,"recent_unique_ips":1,"total_octets":900},{"username":"idle","in_runtime":true,"current_connections":0,"active_unique_ips":0,"recent_unique_ips":0,"total_octets":0}]}`)
		case "/v1/stats/users/active-ips":
			io.WriteString(w, `{"ok":true,"data":[{"username":"work","active_ips":["1.1.1.1"]}]}`)
		case "/v1/runtime/gates":
			io.WriteString(w, `{"ok":true,"data":{"use_middle_proxy":false,"route_mode":"direct","me_runtime_ready":true,"me2dc_fallback_enabled":true,"me2dc_fast_enabled":false,"conditional_cast_enabled":false,"reroute_active":false,"accepting_new_connections":true,"startup_status":"ready","startup_stage":"ready","startup_progress_pct":100}}`)
		case "/v1/runtime/me_quality":
			io.WriteString(w, `{"ok":true,"data":{"enabled":false,"reason":"source_unavailable","data":null}}`)
		case "/v1/runtime/upstream_quality":
			io.WriteString(w, `{"ok":true,"data":{"enabled":true,"policy":{"connect_retry_attempts":2,"connect_retry_backoff_ms":100,"connect_budget_ms":3000,"unhealthy_fail_threshold":5,"connect_failfast_hard_errors":false},"counters":{"connect_attempt_total":4,"connect_success_total":3,"connect_fail_total":1,"connect_failfast_hard_error_total":0},"summary":{"configured_total":2,"healthy_total":1,"unhealthy_total":1,"direct_total":2,"socks4_total":0,"socks5_total":0,"shadowsocks_total":0},"upstreams":[{"upstream_id":0,"route_kind":"direct","address":"direct","weight":1,"scopes":"","healthy":true,"fails":0,"last_check_age_secs":1,"effective_latency_ms":21.1,"dc":[{"dc":1,"latency_ema_ms":30.5,"ip_preference":"prefer_v4"},{"dc":2,"latency_ema_ms":10.5,"ip_preference":"prefer_v6"}]},{"upstream_id":1,"route_kind":"direct","address":"backup","weight":1,"scopes":"","healthy":false,"fails":4,"last_check_age_secs":2,"effective_latency_ms":31.1,"dc":[{"dc":1,"latency_ema_ms":40.5,"ip_preference":"prefer_v4"}]}]}}`)
		case "/v1/stats/me-writers":
			io.WriteString(w, `{"ok":true,"data":{"middle_proxy_enabled":false,"reason":"source_unavailable","summary":{"configured_dc_groups":0,"configured_endpoints":0,"available_endpoints":0,"available_pct":0,"required_writers":0,"alive_writers":0,"coverage_pct":0,"fresh_alive_writers":0,"fresh_coverage_pct":0},"writers":[]}}`)
		case "/v1/stats/dcs":
			io.WriteString(w, `{"ok":true,"data":{"middle_proxy_enabled":false,"reason":"source_unavailable","dcs":[]}}`)
		case "/v1/runtime/me_pool_state":
			io.WriteString(w, `{"ok":true,"data":{"enabled":false,"reason":"source_unavailable","data":null}}`)
		case "/v1/stats/minimal/all":
			io.WriteString(w, `{"ok":true,"data":{"enabled":true,"reason":"source_unavailable","data":null}}`)
		case "/v1/runtime/nat_stun":
			io.WriteString(w, `{"ok":true,"data":{"enabled":false,"reason":"source_unavailable","data":null}}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	endpoint := endpointFromURL(t, server.URL)
	cfg := config.RuntimeConfig{
		ConfigPath:       "config.test.toml",
		Listen:           "127.0.0.1:8080",
		RefreshInterval:  10 * time.Second,
		RequestTimeout:   2 * time.Second,
		EnabledEndpoints: []config.Endpoint{endpoint},
	}
	service := NewService(
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		cfg,
		telemt.NewClient(cfg.RequestTimeout, false),
	)

	snapshot := service.RefreshNow(context.Background())
	serverSnapshot := snapshot.Servers[0]

	if got, want := serverSnapshot.Mode, "DC"; got != want {
		t.Fatalf("mode = %q, want %q", got, want)
	}
	if got, want := serverSnapshot.QualityKind, "dc"; got != want {
		t.Fatalf("quality_kind = %q, want %q", got, want)
	}
	if got, want := len(serverSnapshot.DirectDatacenters), 2; got != want {
		t.Fatalf("direct_datacenters len = %d, want %d", got, want)
	}
	if got, want := serverSnapshot.DirectDatacenters[0].HealthyUpstreams, 1; got != want {
		t.Fatalf("healthy_upstreams = %d, want %d", got, want)
	}
	if got, want := serverSnapshot.DirectDatacenters[0].TotalUpstreams, 2; got != want {
		t.Fatalf("total_upstreams = %d, want %d", got, want)
	}
	if got, want := serverSnapshot.Writers.Available, false; got != want {
		t.Fatalf("writers.available = %v, want %v", got, want)
	}
	if got, want := serverSnapshot.Pool.Available, false; got != want {
		t.Fatalf("pool.available = %v, want %v", got, want)
	}
	if got, want := serverSnapshot.Reliability.MEAvailable, false; got != want {
		t.Fatalf("reliability.me_available = %v, want %v", got, want)
	}
	if got, want := serverSnapshot.Network.Available, false; got != want {
		t.Fatalf("network.available = %v, want %v", got, want)
	}
	if serverSnapshot.UserActivity.TopConnectionsUser == nil || serverSnapshot.UserActivity.TopConnectionsUser.Username != "work" {
		t.Fatalf("top_connections_user = %#v, want work", serverSnapshot.UserActivity.TopConnectionsUser)
	}
}

func endpointFromURL(t *testing.T, rawURL string) config.Endpoint {
	t.Helper()

	parsed, err := url.Parse(rawURL)
	if err != nil {
		t.Fatalf("parse url: %v", err)
	}

	host, portRaw, err := net.SplitHostPort(parsed.Host)
	if err != nil {
		t.Fatalf("split host port: %v", err)
	}

	port, err := strconv.Atoi(portRaw)
	if err != nil {
		t.Fatalf("atoi port: %v", err)
	}

	return config.Endpoint{
		Name:    "test",
		Scheme:  parsed.Scheme,
		Address: host,
		Port:    port,
		Enabled: true,
	}
}
