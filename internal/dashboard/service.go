package dashboard

import (
	"context"
	"log/slog"
	"slices"
	"strings"
	"sync"
	"time"

	"multipanel/internal/config"
	"multipanel/internal/telemt"
)

type Service struct {
	logger   *slog.Logger
	cfg      config.RuntimeConfig
	client   *telemt.Client
	mu       sync.RWMutex
	snapshot Snapshot
}

type Snapshot struct {
	GeneratedAt             time.Time        `json:"generated_at"`
	RefreshIntervalMs       int64            `json:"refresh_interval_ms"`
	ConfigPath              string           `json:"config_path"`
	TrafficDerivedFromUsers bool             `json:"traffic_derived_from_users"`
	Warnings                []string         `json:"warnings,omitempty"`
	Totals                  TotalsSnapshot   `json:"totals"`
	Servers                 []ServerSnapshot `json:"servers"`
}

type TotalsSnapshot struct {
	ServersConfigured     int    `json:"servers_configured"`
	ServersDisabled       int    `json:"servers_disabled"`
	ServersReachable      int    `json:"servers_reachable"`
	ServersHealthy        int    `json:"servers_healthy"`
	ConnectionsTotal      uint64 `json:"connections_total"`
	BadConnectionsTotal   uint64 `json:"bad_connections_total"`
	ActiveIPsTotal        int    `json:"active_ips_total"`
	TotalTrafficBytes     uint64 `json:"total_traffic_bytes"`
	UsersTrafficBytes     uint64 `json:"users_traffic_bytes"`
	UsersConnectionsTotal uint64 `json:"users_connections_total"`
	UsersActiveIPsTotal   int    `json:"users_active_ips_total"`
}

type ServerSnapshot struct {
	Name                   string                    `json:"name"`
	BaseURL                string                    `json:"base_url"`
	Address                string                    `json:"address"`
	Reachable              bool                      `json:"reachable"`
	Partial                bool                      `json:"partial"`
	ReadOnly               bool                      `json:"read_only"`
	Status                 string                    `json:"status"`
	LastError              string                    `json:"last_error,omitempty"`
	Errors                 []string                  `json:"errors,omitempty"`
	CollectedAt            time.Time                 `json:"collected_at"`
	UptimeSeconds          float64                   `json:"uptime_seconds"`
	ConnectionsTotal       uint64                    `json:"connections_total"`
	BadConnectionsTotal    uint64                    `json:"bad_connections_total"`
	HandshakeTimeoutsTotal uint64                    `json:"handshake_timeouts_total"`
	ConfiguredUsers        int                       `json:"configured_users"`
	ActiveIPsTotal         int                       `json:"active_ips_total"`
	TotalTrafficBytes      uint64                    `json:"total_traffic_bytes"`
	UsersTrafficBytes      uint64                    `json:"users_traffic_bytes"`
	UsersConnectionsTotal  uint64                    `json:"users_connections_total"`
	UsersActiveIPsTotal    int                       `json:"users_active_ips_total"`
	Mode                   string                    `json:"mode"`
	RouteMode              string                    `json:"route_mode"`
	MERuntimeReady         bool                      `json:"me_runtime_ready"`
	RerouteActive          bool                      `json:"reroute_active"`
	QualityKind            string                    `json:"quality_kind,omitempty"`
	Datacenters            []DatacenterSnapshot      `json:"datacenters,omitempty"`
	DirectDatacenters      []DirectHealthSnapshot    `json:"direct_datacenters,omitempty"`
	DatacenterReason       string                    `json:"datacenter_reason,omitempty"`
	Runtime                RuntimeHealthSnapshot     `json:"runtime"`
	Upstream               UpstreamHealthSnapshot    `json:"upstream"`
	Writers                WritersHealthSnapshot     `json:"writers"`
	Pool                   PoolHealthSnapshot        `json:"pool"`
	Reliability            ReliabilityHealthSnapshot `json:"reliability"`
	UserActivity           UserActivitySnapshot      `json:"user_activity"`
	Network                NetworkHealthSnapshot     `json:"network"`
}

type DatacenterSnapshot struct {
	DC              int      `json:"dc"`
	RTTEMAMs        *float64 `json:"rtt_ema_ms,omitempty"`
	AliveWriters    int      `json:"alive_writers"`
	RequiredWriters int      `json:"required_writers"`
	CoveragePct     float64  `json:"coverage_pct"`
}

type DirectHealthSnapshot struct {
	DC               int      `json:"dc"`
	RTTEMAMs         *float64 `json:"rtt_ema_ms,omitempty"`
	IPPreference     string   `json:"ip_preference"`
	Healthy          bool     `json:"healthy"`
	HealthyUpstreams int      `json:"healthy_upstreams"`
	TotalUpstreams   int      `json:"total_upstreams"`
}

type RuntimeHealthSnapshot struct {
	Available               bool    `json:"available"`
	AcceptingNewConnections bool    `json:"accepting_new_connections"`
	ConditionalCastEnabled  bool    `json:"conditional_cast_enabled"`
	StartupStatus           string  `json:"startup_status,omitempty"`
	StartupStage            string  `json:"startup_stage,omitempty"`
	StartupProgressPct      float64 `json:"startup_progress_pct"`
	MERuntimeReady          bool    `json:"me_runtime_ready"`
	ME2DCFallbackEnabled    bool    `json:"me2dc_fallback_enabled"`
	ME2DCFastEnabled        bool    `json:"me2dc_fast_enabled"`
	RerouteActive           bool    `json:"reroute_active"`
	RerouteReason           string  `json:"reroute_reason,omitempty"`
}

type UpstreamHealthSnapshot struct {
	Available                     bool     `json:"available"`
	Reason                        string   `json:"reason,omitempty"`
	ConfiguredTotal               int      `json:"configured_total"`
	HealthyTotal                  int      `json:"healthy_total"`
	UnhealthyTotal                int      `json:"unhealthy_total"`
	ConnectAttemptTotal           uint64   `json:"connect_attempt_total"`
	ConnectSuccessTotal           uint64   `json:"connect_success_total"`
	ConnectFailTotal              uint64   `json:"connect_fail_total"`
	ConnectFailfastHardErrorTotal uint64   `json:"connect_failfast_hard_error_total"`
	SuccessRatePct                *float64 `json:"success_rate_pct,omitempty"`
	BestLatencyMs                 *float64 `json:"best_latency_ms,omitempty"`
	WorstLatencyMs                *float64 `json:"worst_latency_ms,omitempty"`
	MaxLastCheckAgeSecs           uint64   `json:"max_last_check_age_secs"`
}

type WritersHealthSnapshot struct {
	Available           bool     `json:"available"`
	Reason              string   `json:"reason,omitempty"`
	ConfiguredDCGroups  int      `json:"configured_dc_groups"`
	ConfiguredEndpoints int      `json:"configured_endpoints"`
	AvailableEndpoints  int      `json:"available_endpoints"`
	AvailablePct        float64  `json:"available_pct"`
	RequiredWriters     int      `json:"required_writers"`
	AliveWriters        int      `json:"alive_writers"`
	CoveragePct         float64  `json:"coverage_pct"`
	FreshCoveragePct    float64  `json:"fresh_coverage_pct"`
	BusyWriters         int      `json:"busy_writers"`
	DegradedWriters     int      `json:"degraded_writers"`
	DrainingWriters     int      `json:"draining_writers"`
	WeakDCTotal         int      `json:"weak_dc_total"`
	FloorCappedDCTotal  int      `json:"floor_capped_dc_total"`
	WorstCoveragePct    *float64 `json:"worst_coverage_pct,omitempty"`
	MaxLoad             int      `json:"max_load"`
}

type PoolHealthSnapshot struct {
	Available                 bool    `json:"available"`
	Reason                    string  `json:"reason,omitempty"`
	TotalWriters              int     `json:"total_writers"`
	AliveNonDrainingWriters   int     `json:"alive_non_draining_writers"`
	HealthyWriters            int     `json:"healthy_writers"`
	DegradedWriters           int     `json:"degraded_writers"`
	DrainingWriters           int     `json:"draining_writers"`
	WarmWriters               int     `json:"warm_writers"`
	ActiveWriters             int     `json:"active_writers"`
	ContourDrainingWriters    int     `json:"contour_draining_writers"`
	HardswapEnabled           bool    `json:"hardswap_enabled"`
	HardswapPending           bool    `json:"hardswap_pending"`
	PendingHardswapGeneration uint64  `json:"pending_hardswap_generation"`
	PendingHardswapAgeSecs    *uint64 `json:"pending_hardswap_age_secs,omitempty"`
	InflightEndpointsTotal    int     `json:"inflight_endpoints_total"`
	InflightDCTotal           int     `json:"inflight_dc_total"`
}

type ReliabilityHealthSnapshot struct {
	Available                   bool                  `json:"available"`
	Reason                      string                `json:"reason,omitempty"`
	MEAvailable                 bool                  `json:"me_available"`
	HandshakeTimeoutsTotal      uint64                `json:"handshake_timeouts_total"`
	ReconnectAttemptTotal       uint64                `json:"reconnect_attempt_total"`
	ReconnectSuccessTotal       uint64                `json:"reconnect_success_total"`
	ReconnectSuccessRatePct     *float64              `json:"reconnect_success_rate_pct,omitempty"`
	ReaderEOFTotal              uint64                `json:"reader_eof_total"`
	IdleCloseByPeerTotal        uint64                `json:"idle_close_by_peer_total"`
	UnexpectedCloseTotal        uint64                `json:"unexpected_close_total"`
	RouteDropNoConnTotal        uint64                `json:"route_drop_no_conn_total"`
	RouteDropChannelClosedTotal uint64                `json:"route_drop_channel_closed_total"`
	RouteDropQueueFullTotal     uint64                `json:"route_drop_queue_full_total"`
	RouteDropQueueFullBaseTotal uint64                `json:"route_drop_queue_full_base_total"`
	RouteDropQueueFullHighTotal uint64                `json:"route_drop_queue_full_high_total"`
	RouteDropsTotal             uint64                `json:"route_drops_total"`
	DrainGateRouteQuorumOK      bool                  `json:"drain_gate_route_quorum_ok"`
	DrainGateRedundancyOK       bool                  `json:"drain_gate_redundancy_ok"`
	DrainGateBlockReason        string                `json:"drain_gate_block_reason,omitempty"`
	FamilyStates                []FamilyStateSnapshot `json:"family_states,omitempty"`
	QuarantinedEndpointsTotal   int                   `json:"quarantined_endpoints_total"`
	QuarantinedEndpoints        []QuarantineSnapshot  `json:"quarantined_endpoints,omitempty"`
}

type FamilyStateSnapshot struct {
	Family               string `json:"family"`
	State                string `json:"state"`
	FailStreak           uint32 `json:"fail_streak"`
	RecoverSuccessStreak uint32 `json:"recover_success_streak"`
}

type QuarantineSnapshot struct {
	Endpoint    string `json:"endpoint"`
	RemainingMs uint64 `json:"remaining_ms"`
}

type UserActivitySnapshot struct {
	ConfiguredUsers    int                 `json:"configured_users"`
	LoadedUsers        int                 `json:"loaded_users"`
	RuntimeUsers       int                 `json:"runtime_users"`
	ActiveUsers        int                 `json:"active_users"`
	TopConnectionsUser *UserLeaderSnapshot `json:"top_connections_user,omitempty"`
	TopTrafficUser     *UserLeaderSnapshot `json:"top_traffic_user,omitempty"`
	TopIPsUser         *UserLeaderSnapshot `json:"top_ips_user,omitempty"`
	TopUsers           []UserLoadSnapshot  `json:"top_users,omitempty"`
}

type UserLoadSnapshot struct {
	Username           string `json:"username"`
	CurrentConnections uint64 `json:"current_connections"`
	ActiveUniqueIPs    int    `json:"active_unique_ips"`
	TotalTrafficBytes  uint64 `json:"total_traffic_bytes"`
}

type UserLeaderSnapshot struct {
	Username           string  `json:"username"`
	CurrentConnections uint64  `json:"current_connections"`
	ActiveUniqueIPs    int     `json:"active_unique_ips"`
	TotalTrafficBytes  uint64  `json:"total_traffic_bytes"`
	SharePct           float64 `json:"share_pct"`
}

type NetworkHealthSnapshot struct {
	Available               bool     `json:"available"`
	Reason                  string   `json:"reason,omitempty"`
	NATProbeEnabled         bool     `json:"nat_probe_enabled"`
	NATProbeDisabledRuntime bool     `json:"nat_probe_disabled_runtime"`
	NATProbeAttempts        uint8    `json:"nat_probe_attempts"`
	ConfiguredServers       int      `json:"configured_servers"`
	LiveServers             int      `json:"live_servers"`
	LivePct                 *float64 `json:"live_pct,omitempty"`
	HasReflectionV4         bool     `json:"has_reflection_v4"`
	HasReflectionV6         bool     `json:"has_reflection_v6"`
	ReflectionV4AgeSecs     *uint64  `json:"reflection_v4_age_secs,omitempty"`
	ReflectionV6AgeSecs     *uint64  `json:"reflection_v6_age_secs,omitempty"`
	BackoffRemainingMs      *uint64  `json:"backoff_remaining_ms,omitempty"`
}

func NewService(logger *slog.Logger, cfg config.RuntimeConfig, client *telemt.Client) *Service {
	return &Service{
		logger: logger,
		cfg:    cfg,
		client: client,
		snapshot: Snapshot{
			GeneratedAt:             time.Now().UTC(),
			RefreshIntervalMs:       cfg.RefreshInterval.Milliseconds(),
			ConfigPath:              cfg.ConfigPath,
			TrafficDerivedFromUsers: true,
		},
	}
}

func (s *Service) Start(ctx context.Context) {
	s.refresh(context.Background())

	go func() {
		ticker := time.NewTicker(s.cfg.RefreshInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.refresh(context.Background())
			}
		}
	}()
}

func (s *Service) Snapshot() Snapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.snapshot
}

func (s *Service) RefreshNow(ctx context.Context) Snapshot {
	s.refresh(ctx)
	return s.Snapshot()
}

func (s *Service) refresh(ctx context.Context) {
	snapshot := Snapshot{
		GeneratedAt:             time.Now().UTC(),
		RefreshIntervalMs:       s.cfg.RefreshInterval.Milliseconds(),
		ConfigPath:              s.cfg.ConfigPath,
		TrafficDerivedFromUsers: true,
		Totals: TotalsSnapshot{
			ServersConfigured: len(s.cfg.EnabledEndpoints),
			ServersDisabled:   len(s.cfg.DisabledEndpoints),
		},
	}

	if len(s.cfg.EnabledEndpoints) == 0 {
		snapshot.Warnings = append(snapshot.Warnings, "No telemt endpoints are enabled in config.toml")
		s.storeSnapshot(snapshot)
		return
	}

	type result struct {
		index  int
		source telemt.SourceData
	}

	results := make([]result, len(s.cfg.EnabledEndpoints))
	var wg sync.WaitGroup
	for i, endpoint := range s.cfg.EnabledEndpoints {
		wg.Add(1)
		go func(index int, endpoint config.Endpoint) {
			defer wg.Done()
			refreshCtx, cancel := context.WithTimeout(ctx, s.cfg.RequestTimeout+time.Second)
			defer cancel()
			results[index] = result{
				index:  index,
				source: s.client.Collect(refreshCtx, endpoint),
			}
		}(i, endpoint)
	}
	wg.Wait()

	clusterActiveIPs := make(map[string]struct{})
	snapshot.Servers = make([]ServerSnapshot, 0, len(results))
	for _, res := range results {
		server, uniqueIPs := buildServerSnapshot(res.source)
		snapshot.Servers = append(snapshot.Servers, server)

		if server.Reachable {
			snapshot.Totals.ServersReachable++
		}
		if server.Status == "ok" {
			snapshot.Totals.ServersHealthy++
		}
		snapshot.Totals.ConnectionsTotal += server.ConnectionsTotal
		snapshot.Totals.BadConnectionsTotal += server.BadConnectionsTotal
		snapshot.Totals.TotalTrafficBytes += server.TotalTrafficBytes
		snapshot.Totals.UsersTrafficBytes += server.UsersTrafficBytes
		snapshot.Totals.UsersConnectionsTotal += server.UsersConnectionsTotal
		snapshot.Totals.UsersActiveIPsTotal += server.UsersActiveIPsTotal
		for ip := range uniqueIPs {
			clusterActiveIPs[ip] = struct{}{}
		}
	}

	snapshot.Totals.ActiveIPsTotal = len(clusterActiveIPs)
	s.storeSnapshot(snapshot)
}

func (s *Service) storeSnapshot(snapshot Snapshot) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.snapshot = snapshot
}

func buildServerSnapshot(source telemt.SourceData) (ServerSnapshot, map[string]struct{}) {
	server := ServerSnapshot{
		Name:        source.Endpoint.Name,
		BaseURL:     source.BaseURL,
		Address:     source.Endpoint.AddressWithPort(),
		Reachable:   source.Reachable,
		Partial:     source.Partial,
		ReadOnly:    source.ReadOnly,
		CollectedAt: source.CollectedAt,
		Errors:      slices.Clone(source.Errors),
	}

	if source.Summary != nil {
		server.UptimeSeconds = source.Summary.UptimeSeconds
		server.ConnectionsTotal = source.Summary.ConnectionsTotal
		server.BadConnectionsTotal = source.Summary.ConnectionsBadTotal
		server.HandshakeTimeoutsTotal = source.Summary.HandshakeTimeoutsTotal
		server.ConfiguredUsers = source.Summary.ConfiguredUsers
	}

	var usersActiveIPsTotal int
	var usersTrafficBytes uint64
	var usersConnectionsTotal uint64
	for _, user := range source.Users {
		usersActiveIPsTotal += user.ActiveUniqueIPs
		usersTrafficBytes += user.TotalOctets
		usersConnectionsTotal += user.CurrentConnections
	}
	server.UsersActiveIPsTotal = usersActiveIPsTotal
	server.UsersTrafficBytes = usersTrafficBytes
	server.UsersConnectionsTotal = usersConnectionsTotal
	server.TotalTrafficBytes = usersTrafficBytes

	activeIPSet := make(map[string]struct{})
	for _, entry := range source.ActiveIPs {
		for _, ip := range entry.ActiveIPs {
			activeIPSet[ip] = struct{}{}
		}
	}
	server.ActiveIPsTotal = len(activeIPSet)

	if source.Gates != nil {
		server.RouteMode = source.Gates.RouteMode
		server.MERuntimeReady = source.Gates.MERuntimeReady
		server.RerouteActive = source.Gates.RerouteActive
		server.Mode = deriveMode(*source.Gates)
	}

	if server.Mode == "ME" || server.Mode == "ME -> DC" || server.Mode == "" {
		server.QualityKind = "me"
	}
	if server.Mode == "DC" {
		server.QualityKind = "dc"
	}

	if server.QualityKind == "me" && source.MeQuality != nil {
		if source.MeQuality.Enabled && source.MeQuality.Data != nil {
			server.Datacenters = make([]DatacenterSnapshot, 0, len(source.MeQuality.Data.DCRTT))
			for _, dc := range source.MeQuality.Data.DCRTT {
				server.Datacenters = append(server.Datacenters, DatacenterSnapshot{
					DC:              dc.DC,
					RTTEMAMs:        dc.RTTEMAMs,
					AliveWriters:    dc.AliveWriters,
					RequiredWriters: dc.RequiredWriters,
					CoveragePct:     dc.CoveragePct,
				})
			}
			slices.SortFunc(server.Datacenters, func(a, b DatacenterSnapshot) int {
				return a.DC - b.DC
			})
		} else if source.MeQuality.Reason != "" {
			server.DatacenterReason = source.MeQuality.Reason
		}
	}

	if server.QualityKind == "dc" && source.UpstreamQuality != nil {
		if source.UpstreamQuality.Enabled && len(source.UpstreamQuality.Upstreams) > 0 {
			server.DirectDatacenters = buildDirectHealth(source.UpstreamQuality.Upstreams)
			if len(server.DirectDatacenters) == 0 {
				server.DatacenterReason = "No direct DC health rows returned by upstream_quality"
			}
		} else if source.UpstreamQuality.Reason != "" {
			server.DatacenterReason = source.UpstreamQuality.Reason
		}
	}

	server.Runtime = buildRuntimeHealth(source.Gates)
	server.Upstream = buildUpstreamHealth(source.UpstreamQuality)
	server.Writers = buildWritersHealth(source.MeWriters, source.DCStatus)
	server.Pool = buildPoolHealth(source.MePool)
	server.Reliability = buildReliabilityHealth(source.Summary, source.MeQuality, source.MinimalAll)
	server.UserActivity = buildUserActivity(source.Users, source.Summary, server.UsersConnectionsTotal, server.UsersTrafficBytes, server.UsersActiveIPsTotal)
	server.Network = buildNetworkHealth(source.NatStun)

	switch {
	case !server.Reachable:
		server.Status = "down"
	case server.Partial:
		server.Status = "partial"
	default:
		server.Status = "ok"
	}

	if len(server.Errors) > 0 {
		server.LastError = server.Errors[0]
	}

	return server, activeIPSet
}

func deriveMode(gates telemt.RuntimeGatesData) string {
	if !gates.UseMiddleProxy {
		return "DC"
	}
	routeMode := strings.ToLower(strings.TrimSpace(gates.RouteMode))
	switch routeMode {
	case "middle":
		return "ME"
	case "direct":
		if gates.RerouteActive {
			return "ME -> DC"
		}
		return "DC"
	case "":
		return "ME"
	default:
		return strings.ToUpper(routeMode)
	}
}

func buildDirectHealth(upstreams []telemt.RuntimeUpstreamData) []DirectHealthSnapshot {
	type aggregate struct {
		dc               int
		bestRTT          *float64
		ipPreference     string
		healthyUpstreams int
		totalUpstreams   int
	}

	aggregates := map[int]*aggregate{}
	for _, upstream := range upstreams {
		for _, dc := range upstream.DC {
			entry, exists := aggregates[dc.DC]
			if !exists {
				entry = &aggregate{dc: dc.DC}
				aggregates[dc.DC] = entry
			}

			entry.totalUpstreams++
			if upstream.Healthy {
				entry.healthyUpstreams++
			}

			if entry.ipPreference == "" || (upstream.Healthy && dc.IPPreference != "") {
				entry.ipPreference = dc.IPPreference
			}

			if dc.LatencyEMAMs == nil {
				continue
			}
			if entry.bestRTT == nil || *dc.LatencyEMAMs < *entry.bestRTT {
				value := *dc.LatencyEMAMs
				entry.bestRTT = &value
			}
		}
	}

	rows := make([]DirectHealthSnapshot, 0, len(aggregates))
	for _, entry := range aggregates {
		rows = append(rows, DirectHealthSnapshot{
			DC:               entry.dc,
			RTTEMAMs:         entry.bestRTT,
			IPPreference:     entry.ipPreference,
			Healthy:          entry.healthyUpstreams > 0,
			HealthyUpstreams: entry.healthyUpstreams,
			TotalUpstreams:   entry.totalUpstreams,
		})
	}

	slices.SortFunc(rows, func(a, b DirectHealthSnapshot) int {
		return a.DC - b.DC
	})

	return rows
}

func buildRuntimeHealth(gates *telemt.RuntimeGatesData) RuntimeHealthSnapshot {
	if gates == nil {
		return RuntimeHealthSnapshot{}
	}
	return RuntimeHealthSnapshot{
		Available:               true,
		AcceptingNewConnections: gates.AcceptingNewConnections,
		ConditionalCastEnabled:  gates.ConditionalCastEnabled,
		StartupStatus:           gates.StartupStatus,
		StartupStage:            gates.StartupStage,
		StartupProgressPct:      gates.StartupProgressPct,
		MERuntimeReady:          gates.MERuntimeReady,
		ME2DCFallbackEnabled:    gates.ME2DCFallbackEnabled,
		ME2DCFastEnabled:        gates.ME2DCFastEnabled,
		RerouteActive:           gates.RerouteActive,
		RerouteReason:           gates.RerouteReason,
	}
}

func buildUpstreamHealth(data *telemt.RuntimeUpstreamQualityData) UpstreamHealthSnapshot {
	if data == nil {
		return UpstreamHealthSnapshot{}
	}

	snapshot := UpstreamHealthSnapshot{
		Available:                     data.Enabled,
		Reason:                        data.Reason,
		ConnectAttemptTotal:           data.Counters.ConnectAttemptTotal,
		ConnectSuccessTotal:           data.Counters.ConnectSuccessTotal,
		ConnectFailTotal:              data.Counters.ConnectFailTotal,
		ConnectFailfastHardErrorTotal: data.Counters.ConnectFailfastHardErrorTotal,
		SuccessRatePct:                percentPtr(data.Counters.ConnectSuccessTotal, data.Counters.ConnectAttemptTotal),
	}

	if data.Summary != nil {
		snapshot.ConfiguredTotal = data.Summary.ConfiguredTotal
		snapshot.HealthyTotal = data.Summary.HealthyTotal
		snapshot.UnhealthyTotal = data.Summary.UnhealthyTotal
	}

	for _, upstream := range data.Upstreams {
		if upstream.LastCheckAgeSecs > snapshot.MaxLastCheckAgeSecs {
			snapshot.MaxLastCheckAgeSecs = upstream.LastCheckAgeSecs
		}
		if upstream.EffectiveLatencyMs != nil {
			snapshot.BestLatencyMs = minLatency(snapshot.BestLatencyMs, *upstream.EffectiveLatencyMs)
			snapshot.WorstLatencyMs = maxLatency(snapshot.WorstLatencyMs, *upstream.EffectiveLatencyMs)
		}
		if upstream.EffectiveLatencyMs != nil {
			continue
		}
		for _, dc := range upstream.DC {
			if dc.LatencyEMAMs == nil {
				continue
			}
			snapshot.BestLatencyMs = minLatency(snapshot.BestLatencyMs, *dc.LatencyEMAMs)
			snapshot.WorstLatencyMs = maxLatency(snapshot.WorstLatencyMs, *dc.LatencyEMAMs)
		}
	}

	return snapshot
}

func buildWritersHealth(meWriters *telemt.MeWritersData, dcStatus *telemt.DCSummaryData) WritersHealthSnapshot {
	snapshot := WritersHealthSnapshot{}
	if meWriters == nil && dcStatus == nil {
		return snapshot
	}

	if meWriters != nil {
		snapshot.ConfiguredDCGroups = meWriters.Summary.ConfiguredDCGroups
		snapshot.ConfiguredEndpoints = meWriters.Summary.ConfiguredEndpoints
		snapshot.AvailableEndpoints = meWriters.Summary.AvailableEndpoints
		snapshot.AvailablePct = meWriters.Summary.AvailablePct
		snapshot.RequiredWriters = meWriters.Summary.RequiredWriters
		snapshot.AliveWriters = meWriters.Summary.AliveWriters
		snapshot.CoveragePct = meWriters.Summary.CoveragePct
		snapshot.FreshCoveragePct = meWriters.Summary.FreshCoveragePct
		for _, writer := range meWriters.Writers {
			if writer.BoundClients > 0 {
				snapshot.BusyWriters++
			}
			if writer.Degraded {
				snapshot.DegradedWriters++
			}
			if writer.Draining {
				snapshot.DrainingWriters++
			}
		}
	}

	if dcStatus != nil {
		for _, dc := range dcStatus.DCS {
			if dc.FreshCoveragePct < 100 || dc.CoveragePct < 100 {
				snapshot.WeakDCTotal++
			}
			if dc.FloorCapped {
				snapshot.FloorCappedDCTotal++
			}
			if dc.Load > snapshot.MaxLoad {
				snapshot.MaxLoad = dc.Load
			}
			snapshot.WorstCoveragePct = minLatency(snapshot.WorstCoveragePct, dc.FreshCoveragePct)
		}
	}

	snapshot.Reason = firstNonEmpty(reasonFromMeWriters(meWriters), reasonFromDCStatus(dcStatus))
	snapshot.Available = meWriters != nil && dcStatus != nil && meWriters.MiddleProxyEnabled && dcStatus.MiddleProxyEnabled
	return snapshot
}

func buildPoolHealth(data *telemt.RuntimeMePoolStateData) PoolHealthSnapshot {
	if data == nil {
		return PoolHealthSnapshot{}
	}
	snapshot := PoolHealthSnapshot{
		Available: data.Enabled && data.Data != nil,
		Reason:    data.Reason,
	}
	if data.Data == nil {
		return snapshot
	}

	snapshot.TotalWriters = data.Data.Writers.Total
	snapshot.AliveNonDrainingWriters = data.Data.Writers.AliveNonDraining
	snapshot.HealthyWriters = data.Data.Writers.Health.Healthy
	snapshot.DegradedWriters = data.Data.Writers.Health.Degraded
	snapshot.DrainingWriters = data.Data.Writers.Health.Draining
	snapshot.WarmWriters = data.Data.Writers.Contour.Warm
	snapshot.ActiveWriters = data.Data.Writers.Contour.Active
	snapshot.ContourDrainingWriters = data.Data.Writers.Contour.Draining
	snapshot.HardswapEnabled = data.Data.Hardswap.Enabled
	snapshot.HardswapPending = data.Data.Hardswap.Pending
	snapshot.PendingHardswapGeneration = data.Data.Generations.PendingHardswapGeneration
	snapshot.PendingHardswapAgeSecs = data.Data.Generations.PendingHardswapAgeSecs
	snapshot.InflightEndpointsTotal = data.Data.Refill.InflightEndpointsTotal
	snapshot.InflightDCTotal = data.Data.Refill.InflightDCTotal
	return snapshot
}

func buildReliabilityHealth(summary *telemt.SummaryData, meQuality *telemt.RuntimeMeQualityData, minimalAll *telemt.MinimalAllData) ReliabilityHealthSnapshot {
	snapshot := ReliabilityHealthSnapshot{}
	if summary != nil {
		snapshot.Available = true
		snapshot.HandshakeTimeoutsTotal = summary.HandshakeTimeoutsTotal
	}

	if meQuality != nil {
		snapshot.Reason = meQuality.Reason
	}
	if meQuality != nil && meQuality.Enabled && meQuality.Data != nil {
		snapshot.Available = true
		snapshot.MEAvailable = true
		snapshot.ReconnectAttemptTotal = meQuality.Data.Counters.ReconnectAttemptTotal
		snapshot.ReconnectSuccessTotal = meQuality.Data.Counters.ReconnectSuccessTotal
		snapshot.ReconnectSuccessRatePct = percentPtr(meQuality.Data.Counters.ReconnectSuccessTotal, meQuality.Data.Counters.ReconnectAttemptTotal)
		snapshot.ReaderEOFTotal = meQuality.Data.Counters.ReaderEOFTotal
		snapshot.IdleCloseByPeerTotal = meQuality.Data.Counters.IdleCloseByPeerTotal
		snapshot.UnexpectedCloseTotal = meQuality.Data.Counters.ReaderEOFTotal + meQuality.Data.Counters.IdleCloseByPeerTotal
		snapshot.RouteDropNoConnTotal = meQuality.Data.RouteDrops.NoConnTotal
		snapshot.RouteDropChannelClosedTotal = meQuality.Data.RouteDrops.ChannelClosedTotal
		snapshot.RouteDropQueueFullTotal = meQuality.Data.RouteDrops.QueueFullTotal
		snapshot.RouteDropQueueFullBaseTotal = meQuality.Data.RouteDrops.QueueFullBaseTotal
		snapshot.RouteDropQueueFullHighTotal = meQuality.Data.RouteDrops.QueueFullHighTotal
		snapshot.RouteDropsTotal = meQuality.Data.RouteDrops.NoConnTotal +
			meQuality.Data.RouteDrops.ChannelClosedTotal +
			meQuality.Data.RouteDrops.QueueFullTotal
		snapshot.DrainGateRouteQuorumOK = meQuality.Data.DrainGate.RouteQuorumOK
		snapshot.DrainGateRedundancyOK = meQuality.Data.DrainGate.RedundancyOK
		snapshot.DrainGateBlockReason = meQuality.Data.DrainGate.BlockReason
		snapshot.FamilyStates = make([]FamilyStateSnapshot, 0, len(meQuality.Data.FamilyStates))
		for _, state := range meQuality.Data.FamilyStates {
			snapshot.FamilyStates = append(snapshot.FamilyStates, FamilyStateSnapshot{
				Family:               state.Family,
				State:                state.State,
				FailStreak:           state.FailStreak,
				RecoverSuccessStreak: state.RecoverSuccessStreak,
			})
		}
	}

	if minimalAll != nil && minimalAll.Data != nil && minimalAll.Data.MeRuntime != nil {
		snapshot.Available = true
		snapshot.QuarantinedEndpointsTotal = minimalAll.Data.MeRuntime.QuarantinedEndpointsTotal
		snapshot.QuarantinedEndpoints = make([]QuarantineSnapshot, 0, len(minimalAll.Data.MeRuntime.QuarantinedEndpoints))
		for _, item := range minimalAll.Data.MeRuntime.QuarantinedEndpoints {
			snapshot.QuarantinedEndpoints = append(snapshot.QuarantinedEndpoints, QuarantineSnapshot{
				Endpoint:    item.Endpoint,
				RemainingMs: item.RemainingMs,
			})
		}
	}

	if snapshot.Reason == "" && minimalAll != nil {
		snapshot.Reason = minimalAll.Reason
	}

	return snapshot
}

func buildUserActivity(users []telemt.UserInfo, summary *telemt.SummaryData, totalConnections uint64, totalTraffic uint64, totalIPs int) UserActivitySnapshot {
	snapshot := UserActivitySnapshot{}
	if summary != nil {
		snapshot.ConfiguredUsers = summary.ConfiguredUsers
	}

	snapshot.LoadedUsers = len(users)
	if len(users) == 0 {
		return snapshot
	}

	topUsers := make([]UserLoadSnapshot, 0, len(users))
	var topConnections *UserLeaderSnapshot
	var topTraffic *UserLeaderSnapshot
	var topIPs *UserLeaderSnapshot

	for _, user := range users {
		if user.InRuntime {
			snapshot.RuntimeUsers++
		}
		if user.CurrentConnections > 0 || user.ActiveUniqueIPs > 0 {
			snapshot.ActiveUsers++
		}

		load := UserLoadSnapshot{
			Username:           user.Username,
			CurrentConnections: user.CurrentConnections,
			ActiveUniqueIPs:    user.ActiveUniqueIPs,
			TotalTrafficBytes:  user.TotalOctets,
		}

		if user.CurrentConnections > 0 || user.ActiveUniqueIPs > 0 || user.TotalOctets > 0 {
			topUsers = append(topUsers, load)
		}

		if topConnections == nil || compareUserConnections(user, *topConnections) < 0 {
			topConnections = &UserLeaderSnapshot{
				Username:           user.Username,
				CurrentConnections: user.CurrentConnections,
				ActiveUniqueIPs:    user.ActiveUniqueIPs,
				TotalTrafficBytes:  user.TotalOctets,
				SharePct:           percentValue(user.CurrentConnections, totalConnections),
			}
		}
		if topTraffic == nil || compareUserTraffic(user, *topTraffic) < 0 {
			topTraffic = &UserLeaderSnapshot{
				Username:           user.Username,
				CurrentConnections: user.CurrentConnections,
				ActiveUniqueIPs:    user.ActiveUniqueIPs,
				TotalTrafficBytes:  user.TotalOctets,
				SharePct:           percentValue(user.TotalOctets, totalTraffic),
			}
		}
		if topIPs == nil || compareUserIPs(user, *topIPs) < 0 {
			topIPs = &UserLeaderSnapshot{
				Username:           user.Username,
				CurrentConnections: user.CurrentConnections,
				ActiveUniqueIPs:    user.ActiveUniqueIPs,
				TotalTrafficBytes:  user.TotalOctets,
				SharePct:           percentValue(uint64(user.ActiveUniqueIPs), uint64(totalIPs)),
			}
		}
	}

	slices.SortFunc(topUsers, func(a, b UserLoadSnapshot) int {
		switch {
		case a.CurrentConnections != b.CurrentConnections:
			if a.CurrentConnections > b.CurrentConnections {
				return -1
			}
			return 1
		case a.ActiveUniqueIPs != b.ActiveUniqueIPs:
			if a.ActiveUniqueIPs > b.ActiveUniqueIPs {
				return -1
			}
			return 1
		case a.TotalTrafficBytes != b.TotalTrafficBytes:
			if a.TotalTrafficBytes > b.TotalTrafficBytes {
				return -1
			}
			return 1
		default:
			return strings.Compare(a.Username, b.Username)
		}
	})
	if len(topUsers) > 5 {
		topUsers = topUsers[:5]
	}

	if topConnections != nil && topConnections.CurrentConnections > 0 {
		snapshot.TopConnectionsUser = topConnections
	}
	if topTraffic != nil && topTraffic.TotalTrafficBytes > 0 {
		snapshot.TopTrafficUser = topTraffic
	}
	if topIPs != nil && topIPs.ActiveUniqueIPs > 0 {
		snapshot.TopIPsUser = topIPs
	}
	snapshot.TopUsers = topUsers
	return snapshot
}

func buildNetworkHealth(data *telemt.RuntimeNatStunData) NetworkHealthSnapshot {
	if data == nil {
		return NetworkHealthSnapshot{}
	}

	snapshot := NetworkHealthSnapshot{
		Available: data.Enabled && data.Data != nil,
		Reason:    data.Reason,
	}
	if data.Data == nil {
		return snapshot
	}

	snapshot.NATProbeEnabled = data.Data.Flags.NATProbeEnabled
	snapshot.NATProbeDisabledRuntime = data.Data.Flags.NATProbeDisabledRuntime
	snapshot.NATProbeAttempts = data.Data.Flags.NATProbeAttempts
	snapshot.ConfiguredServers = len(data.Data.Servers.Configured)
	snapshot.LiveServers = data.Data.Servers.LiveTotal
	snapshot.LivePct = percentPtr(uint64(data.Data.Servers.LiveTotal), uint64(len(data.Data.Servers.Configured)))
	snapshot.HasReflectionV4 = data.Data.Reflection.V4 != nil
	snapshot.HasReflectionV6 = data.Data.Reflection.V6 != nil
	if data.Data.Reflection.V4 != nil {
		snapshot.ReflectionV4AgeSecs = &data.Data.Reflection.V4.AgeSecs
	}
	if data.Data.Reflection.V6 != nil {
		snapshot.ReflectionV6AgeSecs = &data.Data.Reflection.V6.AgeSecs
	}
	snapshot.BackoffRemainingMs = data.Data.StunBackoffRemainingMs
	return snapshot
}

func compareUserConnections(user telemt.UserInfo, leader UserLeaderSnapshot) int {
	switch {
	case user.CurrentConnections != leader.CurrentConnections:
		if user.CurrentConnections > leader.CurrentConnections {
			return -1
		}
		return 1
	case user.ActiveUniqueIPs != leader.ActiveUniqueIPs:
		if user.ActiveUniqueIPs > leader.ActiveUniqueIPs {
			return -1
		}
		return 1
	case user.TotalOctets != leader.TotalTrafficBytes:
		if user.TotalOctets > leader.TotalTrafficBytes {
			return -1
		}
		return 1
	default:
		return strings.Compare(user.Username, leader.Username)
	}
}

func compareUserTraffic(user telemt.UserInfo, leader UserLeaderSnapshot) int {
	switch {
	case user.TotalOctets != leader.TotalTrafficBytes:
		if user.TotalOctets > leader.TotalTrafficBytes {
			return -1
		}
		return 1
	case user.CurrentConnections != leader.CurrentConnections:
		if user.CurrentConnections > leader.CurrentConnections {
			return -1
		}
		return 1
	case user.ActiveUniqueIPs != leader.ActiveUniqueIPs:
		if user.ActiveUniqueIPs > leader.ActiveUniqueIPs {
			return -1
		}
		return 1
	default:
		return strings.Compare(user.Username, leader.Username)
	}
}

func compareUserIPs(user telemt.UserInfo, leader UserLeaderSnapshot) int {
	switch {
	case user.ActiveUniqueIPs != leader.ActiveUniqueIPs:
		if user.ActiveUniqueIPs > leader.ActiveUniqueIPs {
			return -1
		}
		return 1
	case user.CurrentConnections != leader.CurrentConnections:
		if user.CurrentConnections > leader.CurrentConnections {
			return -1
		}
		return 1
	case user.TotalOctets != leader.TotalTrafficBytes:
		if user.TotalOctets > leader.TotalTrafficBytes {
			return -1
		}
		return 1
	default:
		return strings.Compare(user.Username, leader.Username)
	}
}

func reasonFromMeWriters(data *telemt.MeWritersData) string {
	if data == nil {
		return ""
	}
	return data.Reason
}

func reasonFromDCStatus(data *telemt.DCSummaryData) string {
	if data == nil {
		return ""
	}
	return data.Reason
}

func percentPtr(numerator, denominator uint64) *float64 {
	if denominator == 0 {
		return nil
	}
	value := float64(numerator) * 100 / float64(denominator)
	return &value
}

func percentValue(numerator, denominator uint64) float64 {
	if denominator == 0 {
		return 0
	}
	return float64(numerator) * 100 / float64(denominator)
}

func minLatency(current *float64, next float64) *float64 {
	if current == nil || next < *current {
		value := next
		return &value
	}
	return current
}

func maxLatency(current *float64, next float64) *float64 {
	if current == nil || next > *current {
		value := next
		return &value
	}
	return current
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
