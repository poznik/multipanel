package telemt

import (
	"time"

	"multipanel/internal/config"
)

type envelope[T any] struct {
	OK    bool      `json:"ok"`
	Data  T         `json:"data"`
	Error ErrorBody `json:"error"`
}

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type HealthData struct {
	Status   string `json:"status"`
	ReadOnly bool   `json:"read_only"`
}

type SummaryData struct {
	UptimeSeconds          float64 `json:"uptime_seconds"`
	ConnectionsTotal       uint64  `json:"connections_total"`
	ConnectionsBadTotal    uint64  `json:"connections_bad_total"`
	HandshakeTimeoutsTotal uint64  `json:"handshake_timeouts_total"`
	ConfiguredUsers        int     `json:"configured_users"`
}

type UserInfo struct {
	Username           string `json:"username"`
	InRuntime          bool   `json:"in_runtime"`
	CurrentConnections uint64 `json:"current_connections"`
	ActiveUniqueIPs    int    `json:"active_unique_ips"`
	RecentUniqueIPs    int    `json:"recent_unique_ips"`
	TotalOctets        uint64 `json:"total_octets"`
}

type UserActiveIPs struct {
	Username  string   `json:"username"`
	ActiveIPs []string `json:"active_ips"`
}

type RuntimeGatesData struct {
	UseMiddleProxy          bool    `json:"use_middle_proxy"`
	RouteMode               string  `json:"route_mode"`
	MERuntimeReady          bool    `json:"me_runtime_ready"`
	ME2DCFallbackEnabled    bool    `json:"me2dc_fallback_enabled"`
	ME2DCFastEnabled        bool    `json:"me2dc_fast_enabled"`
	ConditionalCastEnabled  bool    `json:"conditional_cast_enabled"`
	RerouteActive           bool    `json:"reroute_active"`
	RerouteReason           string  `json:"reroute_reason"`
	AcceptingNewConnections bool    `json:"accepting_new_connections"`
	StartupStatus           string  `json:"startup_status"`
	StartupStage            string  `json:"startup_stage"`
	StartupProgressPct      float64 `json:"startup_progress_pct"`
}

type RuntimeMeQualityData struct {
	Enabled bool                     `json:"enabled"`
	Reason  string                   `json:"reason"`
	Data    *RuntimeMeQualityPayload `json:"data"`
}

type RuntimeMeQualityPayload struct {
	Counters     RuntimeMeQualityCounters   `json:"counters"`
	RouteDrops   RuntimeMeQualityRouteDrops `json:"route_drops"`
	FamilyStates []RuntimeMeQualityFamily   `json:"family_states"`
	DrainGate    RuntimeMeQualityDrainGate  `json:"drain_gate"`
	DCRTT        []RuntimeMeQualityDC       `json:"dc_rtt"`
}

type RuntimeMeQualityCounters struct {
	IdleCloseByPeerTotal  uint64 `json:"idle_close_by_peer_total"`
	ReaderEOFTotal        uint64 `json:"reader_eof_total"`
	KDFDriftTotal         uint64 `json:"kdf_drift_total"`
	KDFPortOnlyDriftTotal uint64 `json:"kdf_port_only_drift_total"`
	ReconnectAttemptTotal uint64 `json:"reconnect_attempt_total"`
	ReconnectSuccessTotal uint64 `json:"reconnect_success_total"`
}

type RuntimeMeQualityRouteDrops struct {
	NoConnTotal        uint64 `json:"no_conn_total"`
	ChannelClosedTotal uint64 `json:"channel_closed_total"`
	QueueFullTotal     uint64 `json:"queue_full_total"`
	QueueFullBaseTotal uint64 `json:"queue_full_base_total"`
	QueueFullHighTotal uint64 `json:"queue_full_high_total"`
}

type RuntimeMeQualityFamily struct {
	Family               string `json:"family"`
	State                string `json:"state"`
	StateSinceEpochSecs  uint64 `json:"state_since_epoch_secs"`
	FailStreak           uint32 `json:"fail_streak"`
	RecoverSuccessStreak uint32 `json:"recover_success_streak"`
}

type RuntimeMeQualityDrainGate struct {
	RouteQuorumOK      bool   `json:"route_quorum_ok"`
	RedundancyOK       bool   `json:"redundancy_ok"`
	BlockReason        string `json:"block_reason"`
	UpdatedAtEpochSecs uint64 `json:"updated_at_epoch_secs"`
}

type RuntimeMeQualityDC struct {
	DC              int      `json:"dc"`
	RTTEMAMs        *float64 `json:"rtt_ema_ms"`
	AliveWriters    int      `json:"alive_writers"`
	RequiredWriters int      `json:"required_writers"`
	CoveragePct     float64  `json:"coverage_pct"`
}

type RuntimeUpstreamQualityData struct {
	Enabled   bool                           `json:"enabled"`
	Reason    string                         `json:"reason"`
	Policy    RuntimeUpstreamQualityPolicy   `json:"policy"`
	Counters  RuntimeUpstreamQualityCounters `json:"counters"`
	Summary   *RuntimeUpstreamQualitySummary `json:"summary"`
	Upstreams []RuntimeUpstreamData          `json:"upstreams"`
}

type RuntimeUpstreamQualityPolicy struct {
	ConnectRetryAttempts      uint32 `json:"connect_retry_attempts"`
	ConnectRetryBackoffMs     uint64 `json:"connect_retry_backoff_ms"`
	ConnectBudgetMs           uint64 `json:"connect_budget_ms"`
	UnhealthyFailThreshold    uint32 `json:"unhealthy_fail_threshold"`
	ConnectFailfastHardErrors bool   `json:"connect_failfast_hard_errors"`
}

type RuntimeUpstreamQualityCounters struct {
	ConnectAttemptTotal           uint64 `json:"connect_attempt_total"`
	ConnectSuccessTotal           uint64 `json:"connect_success_total"`
	ConnectFailTotal              uint64 `json:"connect_fail_total"`
	ConnectFailfastHardErrorTotal uint64 `json:"connect_failfast_hard_error_total"`
}

type RuntimeUpstreamQualitySummary struct {
	ConfiguredTotal  int `json:"configured_total"`
	HealthyTotal     int `json:"healthy_total"`
	UnhealthyTotal   int `json:"unhealthy_total"`
	DirectTotal      int `json:"direct_total"`
	Socks4Total      int `json:"socks4_total"`
	Socks5Total      int `json:"socks5_total"`
	ShadowsocksTotal int `json:"shadowsocks_total"`
}

type RuntimeUpstreamData struct {
	UpstreamID         int                       `json:"upstream_id"`
	RouteKind          string                    `json:"route_kind"`
	Address            string                    `json:"address"`
	Weight             int                       `json:"weight"`
	Scopes             string                    `json:"scopes"`
	Healthy            bool                      `json:"healthy"`
	Fails              uint64                    `json:"fails"`
	LastCheckAgeSecs   uint64                    `json:"last_check_age_secs"`
	EffectiveLatencyMs *float64                  `json:"effective_latency_ms"`
	DC                 []RuntimeUpstreamDCStatus `json:"dc"`
}

type RuntimeUpstreamDCStatus struct {
	DC           int      `json:"dc"`
	LatencyEMAMs *float64 `json:"latency_ema_ms"`
	IPPreference string   `json:"ip_preference"`
}

type MeWritersData struct {
	MiddleProxyEnabled bool             `json:"middle_proxy_enabled"`
	Reason             string           `json:"reason"`
	Summary            MeWritersSummary `json:"summary"`
	Writers            []MeWriterStatus `json:"writers"`
}

type MeWritersSummary struct {
	ConfiguredDCGroups  int     `json:"configured_dc_groups"`
	ConfiguredEndpoints int     `json:"configured_endpoints"`
	AvailableEndpoints  int     `json:"available_endpoints"`
	AvailablePct        float64 `json:"available_pct"`
	RequiredWriters     int     `json:"required_writers"`
	AliveWriters        int     `json:"alive_writers"`
	CoveragePct         float64 `json:"coverage_pct"`
	FreshAliveWriters   int     `json:"fresh_alive_writers"`
	FreshCoveragePct    float64 `json:"fresh_coverage_pct"`
}

type MeWriterStatus struct {
	WriterID     uint64   `json:"writer_id"`
	DC           *int     `json:"dc"`
	Endpoint     string   `json:"endpoint"`
	Generation   uint64   `json:"generation"`
	State        string   `json:"state"`
	Draining     bool     `json:"draining"`
	Degraded     bool     `json:"degraded"`
	BoundClients int      `json:"bound_clients"`
	IdleForSecs  *uint64  `json:"idle_for_secs"`
	RTTEMAMs     *float64 `json:"rtt_ema_ms"`
	DrainOverTTL bool     `json:"drain_over_ttl"`
}

type DCSummaryData struct {
	MiddleProxyEnabled bool       `json:"middle_proxy_enabled"`
	Reason             string     `json:"reason"`
	DCS                []DCStatus `json:"dcs"`
}

type DCStatus struct {
	DC                 int      `json:"dc"`
	AvailableEndpoints int      `json:"available_endpoints"`
	AvailablePct       float64  `json:"available_pct"`
	RequiredWriters    int      `json:"required_writers"`
	FloorMin           int      `json:"floor_min"`
	FloorTarget        int      `json:"floor_target"`
	FloorMax           int      `json:"floor_max"`
	FloorCapped        bool     `json:"floor_capped"`
	AliveWriters       int      `json:"alive_writers"`
	CoveragePct        float64  `json:"coverage_pct"`
	FreshAliveWriters  int      `json:"fresh_alive_writers"`
	FreshCoveragePct   float64  `json:"fresh_coverage_pct"`
	RTTMs              *float64 `json:"rtt_ms"`
	Load               int      `json:"load"`
}

type RuntimeMePoolStateData struct {
	Enabled bool                       `json:"enabled"`
	Reason  string                     `json:"reason"`
	Data    *RuntimeMePoolStatePayload `json:"data"`
}

type RuntimeMePoolStatePayload struct {
	Generations RuntimeMePoolGenerations `json:"generations"`
	Hardswap    RuntimeMePoolHardswap    `json:"hardswap"`
	Writers     RuntimeMePoolWriters     `json:"writers"`
	Refill      RuntimeMePoolRefill      `json:"refill"`
}

type RuntimeMePoolGenerations struct {
	ActiveGeneration          uint64   `json:"active_generation"`
	WarmGeneration            uint64   `json:"warm_generation"`
	PendingHardswapGeneration uint64   `json:"pending_hardswap_generation"`
	PendingHardswapAgeSecs    *uint64  `json:"pending_hardswap_age_secs"`
	DrainingGenerations       []uint64 `json:"draining_generations"`
}

type RuntimeMePoolHardswap struct {
	Enabled bool `json:"enabled"`
	Pending bool `json:"pending"`
}

type RuntimeMePoolWriters struct {
	Total            int                        `json:"total"`
	AliveNonDraining int                        `json:"alive_non_draining"`
	Draining         int                        `json:"draining"`
	Degraded         int                        `json:"degraded"`
	Contour          RuntimeMePoolWriterContour `json:"contour"`
	Health           RuntimeMePoolWriterHealth  `json:"health"`
}

type RuntimeMePoolWriterContour struct {
	Warm     int `json:"warm"`
	Active   int `json:"active"`
	Draining int `json:"draining"`
}

type RuntimeMePoolWriterHealth struct {
	Healthy  int `json:"healthy"`
	Degraded int `json:"degraded"`
	Draining int `json:"draining"`
}

type RuntimeMePoolRefill struct {
	InflightEndpointsTotal int                     `json:"inflight_endpoints_total"`
	InflightDCTotal        int                     `json:"inflight_dc_total"`
	ByDC                   []RuntimeMePoolRefillDC `json:"by_dc"`
}

type RuntimeMePoolRefillDC struct {
	DC       int    `json:"dc"`
	Family   string `json:"family"`
	Inflight int    `json:"inflight"`
}

type MinimalAllData struct {
	Enabled bool               `json:"enabled"`
	Reason  string             `json:"reason"`
	Data    *MinimalAllPayload `json:"data"`
}

type MinimalAllPayload struct {
	MeRuntime *MinimalMeRuntimeData `json:"me_runtime"`
}

type MinimalMeRuntimeData struct {
	QuarantinedEndpointsTotal int                 `json:"quarantined_endpoints_total"`
	QuarantinedEndpoints      []MinimalQuarantine `json:"quarantined_endpoints"`
}

type MinimalQuarantine struct {
	Endpoint    string `json:"endpoint"`
	RemainingMs uint64 `json:"remaining_ms"`
}

type RuntimeNatStunData struct {
	Enabled bool                   `json:"enabled"`
	Reason  string                 `json:"reason"`
	Data    *RuntimeNatStunPayload `json:"data"`
}

type RuntimeNatStunPayload struct {
	Flags                  RuntimeNatStunFlags      `json:"flags"`
	Servers                RuntimeNatStunServers    `json:"servers"`
	Reflection             RuntimeNatStunReflection `json:"reflection"`
	StunBackoffRemainingMs *uint64                  `json:"stun_backoff_remaining_ms"`
}

type RuntimeNatStunFlags struct {
	NATProbeEnabled         bool  `json:"nat_probe_enabled"`
	NATProbeDisabledRuntime bool  `json:"nat_probe_disabled_runtime"`
	NATProbeAttempts        uint8 `json:"nat_probe_attempts"`
}

type RuntimeNatStunServers struct {
	Configured []string `json:"configured"`
	Live       []string `json:"live"`
	LiveTotal  int      `json:"live_total"`
}

type RuntimeNatStunReflection struct {
	V4 *RuntimeNatStunReflectionEntry `json:"v4"`
	V6 *RuntimeNatStunReflectionEntry `json:"v6"`
}

type RuntimeNatStunReflectionEntry struct {
	Addr    string `json:"addr"`
	AgeSecs uint64 `json:"age_secs"`
}

type SourceData struct {
	Endpoint        config.Endpoint
	BaseURL         string
	CollectedAt     time.Time
	Reachable       bool
	Partial         bool
	ReadOnly        bool
	Errors          []string
	Summary         *SummaryData
	Users           []UserInfo
	ActiveIPs       []UserActiveIPs
	Gates           *RuntimeGatesData
	MeQuality       *RuntimeMeQualityData
	UpstreamQuality *RuntimeUpstreamQualityData
	MeWriters       *MeWritersData
	DCStatus        *DCSummaryData
	MePool          *RuntimeMePoolStateData
	MinimalAll      *MinimalAllData
	NatStun         *RuntimeNatStunData
}
