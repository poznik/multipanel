const cardsRoot = document.getElementById("cards");
const healthCardsRoot = document.getElementById("health-cards");
const warningsRoot = document.getElementById("warnings");
const serversTableBody = document.getElementById("servers-table-body");
const datacentersRoot = document.getElementById("datacenters");
const runtimeHealthRoot = document.getElementById("runtime-health");
const transportHealthRoot = document.getElementById("transport-health");
const writersHealthRoot = document.getElementById("writers-health");
const usersHealthRoot = document.getElementById("users-health");
const reliabilityHealthRoot = document.getElementById("reliability-health");
const lastUpdatedRoot = document.getElementById("last-updated");
const helpToggleButton = document.getElementById("help-toggle");
const refreshButton = document.getElementById("refresh-button");
const viewButtons = Array.from(document.querySelectorAll("[data-view-trigger]"));
const viewPanels = Array.from(document.querySelectorAll("[data-view]"));

let refreshTimer = null;
const HELP_STORAGE_KEY = "multipanel.helpEnabled";

const TOOLTIP_TEXT = {
  "Servers": "Сколько серверов из конфигурации сейчас доступны по API и участвуют в сводке.",
  "Total Connections": "Общее количество активных клиентских подключений на всех серверах.",
  "Bad Connections": "Количество неуспешных или плохих подключений. Рост значения обычно указывает на деградацию сети или транспорта.",
  "Active IPs": "Количество уникальных IP-адресов с активностью по всем серверам.",
  "Traffic": "Суммарный трафик пользователей по всем серверам в текущем snapshot.",
  "Users Conn": "Сумма текущих пользовательских подключений по всем аккаунтам.",
  "Users IPs": "Сумма активных пользовательских IP по всем аккаунтам.",
  "Admission Open": "Сколько серверов сейчас открыты для приема новых подключений.",
  "Upstream Clean": "Сколько серверов не имеют нездоровых upstream и ошибок соединения с Telegram.",
  "Weak DCs": "Количество DC, где покрытие writers или свежесть покрытия уже ниже нормы.",
  "Active Users": "Сколько пользователей сейчас реально активны: есть соединения или активные IP.",
  "Route Drops": "Количество потерянных маршрутизацией событий. Рост значения обычно означает внутреннюю деградацию ME.",
  "Quarantine": "Сколько endpoint сейчас находятся в карантине и временно исключены из работы.",
  "NAT/STUN Live": "Сколько серверов имеют рабочий NAT/STUN блок и видят живые STUN-серверы.",
  "Admission": "Открыт ли сервер для приема новых клиентских подключений прямо сейчас.",
  "Startup": "Стадия и прогресс инициализации рантайма. В норме сервер должен быть в состоянии ready.",
  "ME Runtime": "Готов ли ME-рантайм обслуживать middle proxy трафик.",
  "Reroute": "Активна ли аварийная переадресация трафика в direct/DC режим.",
  "ME->DC Fallback": "Разрешен ли аварийный fallback из ME в direct/DC при проблемах.",
  "Conditional Cast": "Внутренний флаг условного режима маршрутизации в рантайме.",
  "Upstreams": "Сколько upstream сейчас healthy по отношению к общему числу сконфигурированных.",
  "Connect Fail": "Количество неудачных попыток подключения к upstream Telegram.",
  "Success Rate": "Доля успешных подключений к upstream Telegram.",
  "Latency": "Лучшее и худшее наблюдаемое время отклика до upstream/DC.",
  "Check Age": "Возраст последней проверки состояния upstream. Большое значение означает устаревшее состояние.",
  "STUN Live": "Сколько STUN-серверов реально доступны из текущего окружения.",
  "Reflection": "Есть ли отраженный внешний адрес через STUN и насколько он свежий.",
  "NAT Probe": "Включена ли проверка NAT/STUN для определения внешней связности.",
  "Backoff": "Оставшееся время до следующей попытки STUN после ошибки.",
  "Fresh Coverage": "Насколько свежие и живые writers покрывают требуемый объем по DC.",
  "Endpoints": "Сколько endpoint для writers сейчас доступны по отношению к сконфигурированным.",
  "Floor Capped": "Количество DC, где целевое число writers ограничено floor-логикой.",
  "Busy Writers": "Количество writers, на которых сейчас висят клиентские привязки.",
  "Degraded Writers": "Количество writers в деградированном состоянии.",
  "Healthy / Total": "Сколько writers в пуле healthy по отношению к общему числу.",
  "Degraded / Draining": "Сколько writers деградировали или находятся в состоянии drain.",
  "Warm / Active": "Сколько writers находятся в warm-контуре и сколько активно обслуживают трафик.",
  "Hardswap": "Есть ли сейчас ожидающий hardswap поколений writers.",
  "Refill": "Есть ли фоновые операции восполнения writers или endpoint по DC.",
  "Connections": "Текущее количество пользовательских TCP-подключений на сервере.",
  "Top Conn User": "Пользователь, который сейчас держит наибольшую долю соединений.",
  "Top Traffic User": "Пользователь, который сейчас дает наибольшую долю суммарного трафика.",
  "Top IP User": "Пользователь с наибольшей долей активных IP.",
  "Handshake Timeouts": "Количество таймаутов на этапе рукопожатия клиента с сервисом.",
  "Reconnect": "Сколько переподключений ME было выполнено и сколько из них завершились успешно.",
  "Unexpected Close": "Сумма reader EOF и idle close by peer. Часто указывает на нестабильный канал.",
  "Drain Gate": "Состояние gate, который решает, можно ли безопасно переводить writers в drain.",
  "Mode": "Текущий режим маршрутизации сервера: ME, DC или аварийный переход.",
  "Top Active Users": "Пользователи с наибольшей текущей активностью по соединениям, IP и трафику.",
  "Writers": "Показывает покрытие writers по DC, количество проблемных DC и признаки деградации пула.",
  "Pool": "Показывает здоровье runtime-пула writers, hardswap и refill-активность.",
  "Upstream": "Показывает здоровье выхода к Telegram и качество внешнего канала.",
  "NAT / STUN": "Показывает состояние NAT/STUN-проверок и наличие внешнего отраженного адреса.",
  "Users & Load": "Показывает активных пользователей и концентрацию нагрузки на отдельных аккаунтах.",
  "Reliability & Routing": "Показывает handshake timeouts, reconnect, route drops, quarantine и состояние ME-диагностики.",
  "Runtime": "Показывает, принимает ли сервер новые подключения, готов ли рантайм и активен ли fallback/reroute.",
};

initViewSwitcher();
initHelpToggle();
requestSnapshot();

async function requestSnapshot(force = false) {
  const endpoint = force ? "/api/v1/refresh" : "/api/v1/snapshot";
  const options = force ? { method: "POST" } : {};

  try {
    refreshButton.disabled = true;
    const response = await fetch(endpoint, options);
    if (!response.ok) {
      throw new Error(`HTTP ${response.status}`);
    }

    const snapshot = await response.json();
    render(snapshot);
    scheduleRefresh(snapshot.refresh_interval_ms || 10000);
  } catch (error) {
    console.error(error);
    warningsRoot.innerHTML = warningMarkup([`snapshot request failed: ${String(error)}`]);
  } finally {
    refreshButton.disabled = false;
  }
}

function initViewSwitcher() {
  const desired = normalizeView(window.location.hash.replace("#", "")) || "overview";
  setView(desired);

  viewButtons.forEach((button) => {
    button.addEventListener("click", () => {
      setView(button.dataset.viewTrigger);
    });
  });

  window.addEventListener("hashchange", () => {
    setView(normalizeView(window.location.hash.replace("#", "")) || "overview");
  });
}

function normalizeView(view) {
  return view === "health" ? "health" : view === "overview" ? "overview" : "";
}

function setView(view) {
  const normalized = normalizeView(view) || "overview";
  viewButtons.forEach((button) => {
    button.classList.toggle("is-active", button.dataset.viewTrigger === normalized);
  });
  viewPanels.forEach((panel) => {
    panel.classList.toggle("view--active", panel.dataset.view === normalized);
  });
  if (window.location.hash !== `#${normalized}`) {
    history.replaceState(null, "", `#${normalized}`);
  }
}

function initHelpToggle() {
  applyHelpPreference(loadHelpPreference());
  helpToggleButton.addEventListener("click", () => {
    const next = !document.body.classList.contains("help-enabled");
    applyHelpPreference(next);
    persistHelpPreference(next);
  });
}

function loadHelpPreference() {
  try {
    const raw = window.localStorage.getItem(HELP_STORAGE_KEY);
    if (raw == null) {
      return true;
    }
    return raw === "true";
  } catch {
    return true;
  }
}

function persistHelpPreference(enabled) {
  try {
    window.localStorage.setItem(HELP_STORAGE_KEY, String(enabled));
  } catch {
    // Ignore storage failures; the toggle still applies for the current page.
  }
}

function applyHelpPreference(enabled) {
  document.body.classList.toggle("help-enabled", enabled);
  helpToggleButton.classList.toggle("is-active", enabled);
  helpToggleButton.setAttribute("aria-pressed", String(enabled));
  helpToggleButton.textContent = enabled ? "Help On" : "Help Off";
}

function scheduleRefresh(intervalMs) {
  if (refreshTimer) {
    clearTimeout(refreshTimer);
  }
  refreshTimer = setTimeout(() => requestSnapshot(false), intervalMs);
}

function render(snapshot) {
  const servers = snapshot.servers || [];

  lastUpdatedRoot.textContent = `updated ${formatDateTime(snapshot.generated_at)}`;
  warningsRoot.innerHTML = warningMarkup(snapshot.warnings || []);

  cardsRoot.innerHTML = renderCards(snapshot.totals || {});
  serversTableBody.innerHTML = renderServers(servers);
  datacentersRoot.innerHTML = renderDatacenters(servers);

  healthCardsRoot.innerHTML = renderHealthCards(servers);
  runtimeHealthRoot.innerHTML = renderRuntimeHealth(servers);
  transportHealthRoot.innerHTML = renderTransportHealth(servers);
  writersHealthRoot.innerHTML = renderWritersHealth(servers);
  usersHealthRoot.innerHTML = renderUsersHealth(servers);
  reliabilityHealthRoot.innerHTML = renderReliabilityHealth(servers);
}

function renderCards(totals) {
  const items = [
    { label: "Servers", value: `${totals.servers_reachable ?? 0} / ${totals.servers_configured ?? 0}`, tone: "neutral" },
    { label: "Total Connections", value: formatNumber(totals.connections_total), tone: "accent" },
    { label: "Bad Connections", value: formatNumber(totals.bad_connections_total), tone: (totals.bad_connections_total ?? 0) > 0 ? "warning" : "neutral" },
    { label: "Active IPs", value: formatNumber(totals.active_ips_total), tone: "neutral" },
    { label: "Traffic", value: formatBytes(totals.total_traffic_bytes), tone: "accent" },
    { label: "Users Conn", value: formatNumber(totals.users_connections_total), tone: "neutral" },
    { label: "Users IPs", value: formatNumber(totals.users_active_ips_total), tone: "neutral" },
  ];

  return renderSummaryCards(items);
}

function renderHealthCards(servers) {
  const reachableServers = servers.filter((server) => server.reachable);
  const runtimeAvailable = servers.filter((server) => server.runtime?.available);
  const upstreamAvailable = servers.filter((server) => server.upstream?.available);
  const networkAvailable = servers.filter((server) => server.network?.available);

  const admissionOpen = runtimeAvailable.filter((server) => server.runtime.accepting_new_connections).length;
  const upstreamClean = upstreamAvailable.filter((server) => (server.upstream.unhealthy_total ?? 0) === 0 && (server.upstream.connect_fail_total ?? 0) === 0).length;
  const weakDCs = sumBy(servers, (server) => server.writers?.weak_dc_total ?? 0);
  const activeUsers = sumBy(servers, (server) => server.user_activity?.active_users ?? 0);
  const routeDrops = sumBy(servers, (server) => server.reliability?.route_drops_total ?? 0);
  const quarantined = sumBy(servers, (server) => server.reliability?.quarantined_endpoints_total ?? 0);
  const natLive = networkAvailable.filter((server) => (server.network.live_servers ?? 0) > 0).length;

  const items = [
    {
      label: "Admission Open",
      value: `${admissionOpen} / ${runtimeAvailable.length || reachableServers.length || servers.length}`,
      tone: admissionOpen === (runtimeAvailable.length || reachableServers.length || servers.length) ? "success" : "warning",
    },
    {
      label: "Upstream Clean",
      value: `${upstreamClean} / ${upstreamAvailable.length || servers.length}`,
      tone: upstreamAvailable.length > 0 && upstreamClean === upstreamAvailable.length ? "success" : "warning",
    },
    {
      label: "Weak DCs",
      value: formatNumber(weakDCs),
      tone: weakDCs === 0 ? "success" : "danger",
    },
    {
      label: "Active Users",
      value: formatNumber(activeUsers),
      tone: activeUsers > 0 ? "accent" : "neutral",
    },
    {
      label: "Route Drops",
      value: formatNumber(routeDrops),
      tone: routeDrops === 0 ? "success" : "danger",
    },
    {
      label: "Quarantine",
      value: formatNumber(quarantined),
      tone: quarantined === 0 ? "success" : "danger",
    },
    {
      label: "NAT/STUN Live",
      value: `${natLive} / ${networkAvailable.length || servers.length}`,
      tone: networkAvailable.length > 0 && natLive === networkAvailable.length ? "success" : "warning",
    },
  ];

  return renderSummaryCards(items);
}

function renderSummaryCards(items) {
  return items
    .map(
      (item) => `
        <article class="card"${tooltipAttrs(item.description || tooltipText(item.label))}>
          <p class="card-label">${escapeHtml(item.label)}</p>
          <p class="card-value card-value--${item.tone}">${escapeHtml(item.value)}</p>
        </article>
      `,
    )
    .join("");
}

function renderServers(servers) {
  if (!servers.length) {
    return `<tr><td colspan="12" class="empty">нет активных точек в конфиге</td></tr>`;
  }

  return servers
    .map((server) => {
      const debugParts = [];
      if (server.read_only) {
        debugParts.push(`<span class="pill pill--neutral">read-only</span>`);
      }
      if (server.me_runtime_ready) {
        debugParts.push(`<span class="pill pill--success">me ready</span>`);
      }
      if (server.reroute_active) {
        debugParts.push(`<span class="pill pill--warning">reroute</span>`);
      }
      if (server.last_error) {
        debugParts.push(`<span class="muted small">${escapeHtml(server.last_error)}</span>`);
      }

      return `
        <tr>
          <td>
            <div class="server-name">${escapeHtml(server.name)}</div>
            <div class="muted small">${formatDateTime(server.collected_at)}</div>
          </td>
          <td class="mono small">${escapeHtml(server.base_url)}</td>
          <td>${escapeHtml(formatUptime(server.uptime_seconds))}</td>
          <td>${escapeHtml(formatNumber(server.connections_total))}</td>
          <td>${escapeHtml(formatNumber(server.bad_connections_total))}</td>
          <td>${escapeHtml(formatNumber(server.active_ips_total))}</td>
          <td>${escapeHtml(formatBytes(server.total_traffic_bytes))}</td>
          <td>${escapeHtml(formatNumber(server.users_connections_total))}</td>
          <td>${escapeHtml(formatNumber(server.users_active_ips_total))}</td>
          <td>${server.mode ? `<span class="pill pill--neutral">${escapeHtml(server.mode)}</span>` : `<span class="muted">n/a</span>`}</td>
          <td>${statusBadge(server)}</td>
          <td>${debugParts.join("") || `<span class="muted">-</span>`}</td>
        </tr>
      `;
    })
    .join("");
}

function renderDatacenters(servers) {
  if (!servers.length) {
    return `<div class="empty-panel">datacenter tables появятся после первой успешной загрузки</div>`;
  }

  return servers
    .map((server) => {
      if (server.quality_kind === "dc" && server.direct_datacenters && server.direct_datacenters.length) {
        const rows = server.direct_datacenters
          .map(
            (dc) => `
              <tr>
                <td>DC ${escapeHtml(String(dc.dc))}</td>
                <td>${escapeHtml(dc.rtt_ema_ms != null ? `${dc.rtt_ema_ms.toFixed(1)} ms` : "-")}</td>
                <td>${escapeHtml(dc.ip_preference || "-")}</td>
                <td>${escapeHtml(`${dc.healthy_upstreams} / ${dc.total_upstreams}`)}</td>
                <td>${healthBadge(dc.healthy)}</td>
              </tr>
            `,
          )
          .join("");

        return `
          <article class="dc-panel"${tooltipAttrs("Показывает здоровье direct/DC пути: RTT по DC, предпочтение IP и наличие healthy upstream.")}>
            <div class="dc-header">
              <div>
                <p class="eyebrow">${escapeHtml(server.name)}</p>
                <h3>DC Health</h3>
              </div>
              <span class="pill pill--neutral">${escapeHtml(server.mode || "n/a")}</span>
            </div>
            <div class="table-wrap">
              <table class="grid-table grid-table--compact">
                <thead>
                  <tr>
                    <th>DC</th>
                    <th>RTT</th>
                    <th>IP Pref</th>
                    <th>Upstreams</th>
                    <th>Health</th>
                  </tr>
                </thead>
                <tbody>${rows}</tbody>
              </table>
            </div>
          </article>
        `;
      }

      if (!server.datacenters || !server.datacenters.length) {
        return `
          <article class="dc-panel"${tooltipAttrs(server.quality_kind === "dc" ? "Показывает здоровье direct/DC пути: RTT по DC, предпочтение IP и наличие healthy upstream." : "Показывает качество ME по DC: RTT, writers и покрытие.")}>
            <div class="dc-header">
              <div>
                <p class="eyebrow">${escapeHtml(server.name)}</p>
                <h3>${escapeHtml(server.quality_kind === "dc" ? "DC Health" : "ME Quality")}</h3>
              </div>
              <span class="muted">${escapeHtml(server.datacenter_reason || (server.quality_kind === "dc" ? "DC health unavailable" : "ME quality unavailable"))}</span>
            </div>
          </article>
        `;
      }

      const rows = server.datacenters
        .map(
          (dc) => `
            <tr>
              <td>DC ${escapeHtml(String(dc.dc))}</td>
              <td>${escapeHtml(dc.rtt_ema_ms != null ? `${dc.rtt_ema_ms.toFixed(1)} ms` : "-")}</td>
              <td>${escapeHtml(`${dc.alive_writers} / ${dc.required_writers}`)}</td>
              <td>${coverageBadge(dc.coverage_pct)}</td>
            </tr>
          `,
        )
        .join("");

      return `
        <article class="dc-panel"${tooltipAttrs("Показывает качество ME по DC: RTT, writers и покрытие.")}>
          <div class="dc-header">
            <div>
              <p class="eyebrow">${escapeHtml(server.name)}</p>
              <h3>ME Quality</h3>
            </div>
            <span class="pill pill--neutral">${escapeHtml(server.mode || "n/a")}</span>
          </div>
          <div class="table-wrap">
            <table class="grid-table grid-table--compact">
              <thead>
                <tr>
                  <th>DC</th>
                  <th>RTT</th>
                  <th>Writers</th>
                  <th>Coverage</th>
                </tr>
              </thead>
              <tbody>${rows}</tbody>
            </table>
          </div>
        </article>
      `;
    })
    .join("");
}

function renderRuntimeHealth(servers) {
  if (!servers.length) {
    return renderEmptyState("runtime cards появятся после первой успешной загрузки");
  }

  return servers
    .map((server) => {
      const runtime = server.runtime || {};
      if (!runtime.available) {
        return renderHealthCard(server, "Runtime", renderUnavailable("runtime gates unavailable"), "", tooltipText("Runtime"), "health-card--runtime");
      }

      const startupTone = toneFromStartup(runtime.startup_status, runtime.startup_progress_pct);
      const meRuntimeValue = server.mode === "DC"
        ? badge("n/a", "neutral")
        : badge(runtime.me_runtime_ready ? "ready" : "not ready", runtime.me_runtime_ready ? "success" : "danger");

      const rows = [
        metricRow("Mode", badge(server.mode || "n/a", "neutral")),
        metricRow("Admission", badge(runtime.accepting_new_connections ? "open" : "closed", runtime.accepting_new_connections ? "success" : "danger")),
        metricRow("Startup", valueSpan(`${runtime.startup_stage || runtime.startup_status || "n/a"} · ${formatPercent(runtime.startup_progress_pct)}`, startupTone)),
        metricRow("ME Runtime", meRuntimeValue),
        metricRow("Reroute", badge(runtime.reroute_active ? "active" : "off", runtime.reroute_active ? "danger" : "success")),
        metricRow("ME->DC Fallback", badge(runtime.me2dc_fallback_enabled ? "enabled" : "off", runtime.me2dc_fallback_enabled ? "success" : "neutral")),
        metricRow("Conditional Cast", badge(runtime.conditional_cast_enabled ? "on" : "off", runtime.conditional_cast_enabled ? "success" : "neutral")),
      ];

      const footer = runtime.reroute_reason
        ? renderNote(`reroute reason: ${runtime.reroute_reason}`, "danger")
        : "";

      return renderHealthCard(server, "Runtime", rows.join(""), footer, tooltipText("Runtime"), "health-card--runtime");
    })
    .join("");
}

function renderTransportHealth(servers) {
  if (!servers.length) {
    return renderEmptyState("transport cards появятся после первой успешной загрузки");
  }

  return servers
    .map((server) => {
      const upstream = server.upstream || {};
      const network = server.network || {};

      const upstreamRows = upstream.available
        ? [
            metricRow("Upstreams", valueSpan(`${upstream.healthy_total ?? 0} / ${upstream.configured_total ?? 0}`, toneFromRatio(upstream.healthy_total ?? 0, upstream.configured_total ?? 0))),
            metricRow("Connect Fail", valueSpan(formatNumber(upstream.connect_fail_total), toneFromCount(upstream.connect_fail_total))),
            metricRow("Success Rate", valueSpan(upstream.success_rate_pct != null ? `${upstream.success_rate_pct.toFixed(1)}%` : "n/a", toneFromPercent(upstream.success_rate_pct))),
            metricRow("Latency", valueSpan(formatLatencyRange(upstream.best_latency_ms, upstream.worst_latency_ms), toneFromLatency(upstream.worst_latency_ms))),
            metricRow("Check Age", valueSpan(formatAge(upstream.max_last_check_age_secs), toneFromAge(upstream.max_last_check_age_secs))),
          ]
        : [metricRow("Upstream", valueSpan(upstream.reason || "unavailable", isDirectMode(server) ? "warning" : "neutral"))];

      const networkRows = network.available
        ? [
            metricRow("STUN Live", valueSpan(`${network.live_servers ?? 0} / ${network.configured_servers ?? 0}`, toneFromRatio(network.live_servers ?? 0, network.configured_servers ?? 0))),
            metricRow("Reflection", valueSpan(formatReflection(network), toneFromReflection(network))),
            metricRow("NAT Probe", badge(network.nat_probe_enabled ? "enabled" : "off", network.nat_probe_enabled ? "success" : "neutral")),
            metricRow("Backoff", valueSpan(network.backoff_remaining_ms != null ? `${formatNumber(network.backoff_remaining_ms)} ms` : "none", network.backoff_remaining_ms ? "warning" : "success")),
          ]
        : [metricRow("NAT/STUN", valueSpan(network.reason || "n/a", isDirectMode(server) ? "neutral" : "warning"))];

      return renderHealthCard(
        server,
        "Upstream & Network",
        `
          <div class="metric-section">
            <div class="metric-section-title">Upstream</div>
            <div class="metric-list">${upstreamRows.join("")}</div>
          </div>
          <div class="health-divider"></div>
          <div class="metric-section">
            <div class="metric-section-title">NAT / STUN</div>
            <div class="metric-list">${networkRows.join("")}</div>
          </div>
        `,
        "",
        "Показывает здоровье upstream Telegram, качество внешнего канала и состояние NAT/STUN.",
      );
    })
    .join("");
}

function renderWritersHealth(servers) {
  if (!servers.length) {
    return renderEmptyState("writer cards появятся после первой успешной загрузки");
  }

  return servers
    .map((server) => {
      const writers = server.writers || {};
      const pool = server.pool || {};

      if (!writers.available && !pool.available && isDirectMode(server)) {
        return renderHealthCard(server, "Writers & Pool", renderUnavailable("ME-only diagnostics unavailable in DC mode"), "", "Показывает покрытие writers по DC, деградацию пула и фоновую перестройку поколений.");
      }

      const writerRows = writers.available
        ? [
            metricRow("Fresh Coverage", valueSpan(`${formatPercent(writers.fresh_coverage_pct)}`, toneFromCoverage(writers.fresh_coverage_pct))),
            metricRow("Endpoints", valueSpan(`${writers.available_endpoints ?? 0} / ${writers.configured_endpoints ?? 0}`, toneFromCoverage(writers.available_pct))),
            metricRow("Weak DCs", valueSpan(formatNumber(writers.weak_dc_total), toneFromCount(writers.weak_dc_total))),
            metricRow("Floor Capped", valueSpan(formatNumber(writers.floor_capped_dc_total), toneFromCount(writers.floor_capped_dc_total, "warning"))),
            metricRow("Busy Writers", valueSpan(formatNumber(writers.busy_writers), writers.busy_writers > 0 ? "accent" : "neutral")),
            metricRow("Degraded Writers", valueSpan(formatNumber(writers.degraded_writers), toneFromCount(writers.degraded_writers))),
          ]
        : [metricRow("Writers", valueSpan(writers.reason || "unavailable", "warning"))];

      const poolRows = pool.available
        ? [
            metricRow("Healthy / Total", valueSpan(`${pool.healthy_writers ?? 0} / ${pool.total_writers ?? 0}`, toneFromRatio(pool.healthy_writers ?? 0, pool.total_writers ?? 0))),
            metricRow("Degraded / Draining", valueSpan(`${pool.degraded_writers ?? 0} / ${pool.draining_writers ?? 0}`, (pool.degraded_writers ?? 0) > 0 || (pool.draining_writers ?? 0) > 0 ? "danger" : "success")),
            metricRow("Warm / Active", valueSpan(`${pool.warm_writers ?? 0} / ${pool.active_writers ?? 0}`, "neutral")),
            metricRow("Hardswap", badge(pool.hardswap_pending ? "pending" : "stable", pool.hardswap_pending ? "warning" : "success")),
            metricRow("Refill", valueSpan(`${pool.inflight_endpoints_total ?? 0} ep / ${pool.inflight_dc_total ?? 0} dc`, (pool.inflight_endpoints_total ?? 0) > 0 || (pool.inflight_dc_total ?? 0) > 0 ? "warning" : "success")),
          ]
        : [metricRow("Pool", valueSpan(pool.reason || "unavailable", "warning"))];

      return renderHealthCard(
        server,
        "Writers & Pool",
        `
          <div class="metric-section">
            <div class="metric-section-title">Coverage</div>
            <div class="metric-list">${writerRows.join("")}</div>
          </div>
          <div class="health-divider"></div>
          <div class="metric-section">
            <div class="metric-section-title">Pool</div>
            <div class="metric-list">${poolRows.join("")}</div>
          </div>
        `,
        "",
        "Показывает покрытие writers по DC, деградацию пула и фоновую перестройку поколений.",
      );
    })
    .join("");
}

function renderUsersHealth(servers) {
  if (!servers.length) {
    return renderEmptyState("user cards появятся после первой успешной загрузки");
  }

  return servers
    .map((server) => {
      const activity = server.user_activity || {};
      const rows = [
        metricRow("Active Users", valueSpan(`${activity.active_users ?? 0} / ${activity.runtime_users ?? 0}`, (activity.active_users ?? 0) > 0 ? "accent" : "neutral")),
        metricRow("Connections", valueSpan(formatNumber(server.users_connections_total), "neutral")),
        metricRow("Active IPs", valueSpan(formatNumber(server.users_active_ips_total), "neutral")),
        metricRow("Traffic", valueSpan(formatBytes(server.users_traffic_bytes), "accent")),
        metricRow("Top Conn User", renderLeader(activity.top_connections_user, "connections")),
        metricRow("Top Traffic User", renderLeader(activity.top_traffic_user, "traffic")),
        metricRow("Top IP User", renderLeader(activity.top_ips_user, "ips")),
      ];

      const topUsersMarkup = activity.top_users && activity.top_users.length
        ? renderTopUsersTable(activity.top_users)
        : renderUnavailable("no active user concentration yet");

      return renderHealthCard(
        server,
        "Users & Load",
        `
          <div class="metric-list">${rows.join("")}</div>
          <div class="health-divider"></div>
          <div class="metric-section">
            <div class="metric-section-title">Top Active Users</div>
            ${topUsersMarkup}
          </div>
        `,
        "",
        tooltipText("Users & Load"),
      );
    })
    .join("");
}

function renderReliabilityHealth(servers) {
  if (!servers.length) {
    return renderEmptyState("reliability cards появятся после первой успешной загрузки");
  }

  return servers
    .map((server) => {
      const reliability = server.reliability || {};
      const rows = [
        metricRow("Handshake Timeouts", valueSpan(formatNumber(reliability.handshake_timeouts_total), toneFromCount(reliability.handshake_timeouts_total))),
        metricRow("Reconnect", renderReconnect(reliability)),
        metricRow("Unexpected Close", valueSpan(formatNumber(reliability.unexpected_close_total), toneFromCount(reliability.unexpected_close_total))),
        metricRow("Route Drops", valueSpan(formatNumber(reliability.route_drops_total), toneFromCount(reliability.route_drops_total))),
        metricRow("Drain Gate", renderDrainGate(reliability, server)),
        metricRow("Quarantine", valueSpan(formatNumber(reliability.quarantined_endpoints_total), toneFromCount(reliability.quarantined_endpoints_total))),
      ];

      const familyMarkup = reliability.me_available && reliability.family_states && reliability.family_states.length
        ? `<div class="chip-list">${reliability.family_states.map(renderFamilyChip).join("")}</div>`
        : renderUnavailable(isDirectMode(server) ? "ME diagnostics unavailable in DC mode" : reliability.reason || "no family-state data");

      const quarantineMarkup = reliability.quarantined_endpoints && reliability.quarantined_endpoints.length
        ? `<div class="chip-list">${reliability.quarantined_endpoints.map(renderQuarantineChip).join("")}</div>`
        : renderUnavailable("no quarantined endpoints");

      return renderHealthCard(
        server,
        "Reliability & Routing",
        `
          <div class="metric-list">${rows.join("")}</div>
          <div class="health-divider"></div>
          <div class="metric-section">
            <div class="metric-section-title">Family States</div>
            ${familyMarkup}
          </div>
          <div class="health-divider"></div>
          <div class="metric-section">
            <div class="metric-section-title">Quarantined Endpoints</div>
            ${quarantineMarkup}
          </div>
        `,
        "",
        tooltipText("Reliability & Routing"),
      );
    })
    .join("");
}

function renderHealthCard(server, title, content, footer = "", description = "", extraClass = "") {
  const className = ["health-card", extraClass].filter(Boolean).join(" ");
  return `
    <article class="${escapeHtml(className)}"${tooltipAttrs(description || tooltipText(title))}>
      <div class="health-card-header">
        <div>
          <p class="eyebrow">${escapeHtml(server.name)}</p>
          <h3>${escapeHtml(title)}</h3>
        </div>
        ${statusBadge(server)}
      </div>
      ${content}
      ${footer}
    </article>
  `;
}

function renderEmptyState(message) {
  return `<div class="empty-panel">${escapeHtml(message)}</div>`;
}

function renderUnavailable(message) {
  return `<div class="health-empty">${escapeHtml(message)}</div>`;
}

function renderNote(message, tone = "neutral") {
  return `<p class="health-note health-note--${tone}">${escapeHtml(message)}</p>`;
}

function renderTopUsersTable(users) {
  const rows = users
    .map(
      (user) => `
        <tr>
          <td>${escapeHtml(user.username)}</td>
          <td>${escapeHtml(formatNumber(user.current_connections))}</td>
          <td>${escapeHtml(formatNumber(user.active_unique_ips))}</td>
          <td>${escapeHtml(formatBytes(user.total_traffic_bytes))}</td>
        </tr>
      `,
    )
    .join("");

  return `
    <div class="table-wrap">
      <table class="grid-table grid-table--compact grid-table--tight">
        <thead>
          <tr>
            <th>User</th>
            <th>Conn</th>
            <th>IPs</th>
            <th>Traffic</th>
          </tr>
        </thead>
        <tbody>${rows}</tbody>
      </table>
    </div>
  `;
}

function renderLeader(leader) {
  if (!leader) {
    return valueSpan("n/a", "neutral");
  }

  return `
    <span class="leader">
      <span class="leader-name">${escapeHtml(leader.username)}</span>
      <span class="metric-value metric-value--${toneFromShare(leader.share_pct)}">${escapeHtml(`${leader.share_pct.toFixed(1)}%`)}</span>
    </span>
  `;
}

function renderReconnect(reliability) {
  if (!reliability.me_available) {
    return valueSpan(isFiniteNumber(reliability.handshake_timeouts_total) ? "ME n/a" : "n/a", "neutral");
  }
  const attempts = reliability.reconnect_attempt_total ?? 0;
  const success = reliability.reconnect_success_total ?? 0;
  const rate = reliability.reconnect_success_rate_pct;
  return valueSpan(`${formatNumber(success)} / ${formatNumber(attempts)}${rate != null ? ` · ${rate.toFixed(1)}%` : ""}`, toneFromPercent(rate));
}

function renderDrainGate(reliability, server) {
  if (!reliability.me_available) {
    return valueSpan(isDirectMode(server) ? "DC mode" : reliability.reason || "n/a", "neutral");
  }
  const isHealthy = reliability.drain_gate_route_quorum_ok && reliability.drain_gate_redundancy_ok && (!reliability.drain_gate_block_reason || reliability.drain_gate_block_reason === "open");
  return valueSpan(
    `${reliability.drain_gate_route_quorum_ok ? "quorum" : "no quorum"} / ${reliability.drain_gate_redundancy_ok ? "redundant" : "thin"}${reliability.drain_gate_block_reason ? ` · ${reliability.drain_gate_block_reason}` : ""}`,
    isHealthy ? "success" : "danger",
  );
}

function renderFamilyChip(state) {
  return `<span class="chip chip--${toneFromFamilyState(state.state)}">${escapeHtml(`${state.family}: ${state.state}`)}</span>`;
}

function renderQuarantineChip(item) {
  return `<span class="chip chip--danger">${escapeHtml(`${item.endpoint} · ${formatDurationMs(item.remaining_ms)}`)}</span>`;
}

function metricRow(label, valueMarkup) {
  return `
    <div class="metric-row">
      <span class="metric-label"${tooltipAttrs(tooltipText(label))}>${escapeHtml(label)}</span>
      <span class="metric-value-wrap">${valueMarkup}</span>
    </div>
  `;
}

function valueSpan(value, tone = "neutral") {
  return `<span class="metric-value metric-value--${tone}">${escapeHtml(value)}</span>`;
}

function badge(label, tone = "neutral") {
  return `<span class="metric-badge metric-badge--${tone}">${escapeHtml(label)}</span>`;
}

function warningMarkup(warnings) {
  if (!warnings.length) {
    return "";
  }

  return warnings
    .map((warning) => `<div class="warning">${escapeHtml(warning)}</div>`)
    .join("");
}

function statusBadge(server) {
  if (!server.reachable) {
    return `<span class="pill pill--danger">down</span>`;
  }
  if (server.partial) {
    return `<span class="pill pill--warning">partial</span>`;
  }
  return `<span class="pill pill--success">ok</span>`;
}

function coverageBadge(coverage) {
  const tone = toneFromCoverage(coverage);
  return `<span class="pill pill--${tone}">${escapeHtml(`${coverage.toFixed(1)}%`)}</span>`;
}

function healthBadge(healthy) {
  return `<span class="pill pill--${healthy ? "success" : "danger"}">${healthy ? "healthy" : "down"}</span>`;
}

function toneFromCoverage(value) {
  if ((value ?? 0) >= 99) return "success";
  if ((value ?? 0) >= 90) return "warning";
  return "danger";
}

function toneFromPercent(value) {
  if (value == null) return "neutral";
  if (value >= 99) return "success";
  if (value >= 90) return "warning";
  return "danger";
}

function toneFromRatio(numerator, denominator) {
  if (!denominator) return "neutral";
  return toneFromPercent((numerator * 100) / denominator);
}

function toneFromCount(value, nonZeroTone = "danger") {
  return (value ?? 0) > 0 ? nonZeroTone : "success";
}

function toneFromLatency(value) {
  if (value == null) return "neutral";
  if (value < 80) return "success";
  if (value < 180) return "warning";
  return "danger";
}

function toneFromAge(value) {
  if (value == null) return "neutral";
  if (value <= 60) return "success";
  if (value <= 180) return "warning";
  return "danger";
}

function toneFromShare(value) {
  if (value == null || value === 0) return "neutral";
  if (value < 40) return "success";
  if (value < 70) return "warning";
  return "danger";
}

function toneFromStartup(status, progress) {
  if (status === "ready" && (progress ?? 0) >= 100) return "success";
  if ((progress ?? 0) >= 90) return "warning";
  return "danger";
}

function toneFromReflection(network) {
  if (network.has_reflection_v4 || network.has_reflection_v6) return "success";
  return "danger";
}

function toneFromFamilyState(state) {
  switch (state) {
    case "healthy":
      return "success";
    case "recovering":
    case "suppressed":
      return "warning";
    default:
      return "danger";
  }
}

function formatNumber(value) {
  return new Intl.NumberFormat("ru-RU").format(value ?? 0);
}

function formatBytes(value) {
  const bytes = Number(value ?? 0);
  if (!bytes) return "0 B";
  const units = ["B", "KB", "MB", "GB", "TB", "PB"];
  let unitIndex = 0;
  let current = bytes;
  while (current >= 1024 && unitIndex < units.length - 1) {
    current /= 1024;
    unitIndex += 1;
  }
  const digits = current >= 100 || unitIndex === 0 ? 0 : 1;
  return `${current.toFixed(digits)} ${units[unitIndex]}`;
}

function formatDateTime(value) {
  if (!value) return "n/a";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "n/a";
  return new Intl.DateTimeFormat("ru-RU", {
    day: "2-digit",
    month: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
  }).format(date);
}

function formatUptime(seconds) {
  const total = Math.max(0, Math.floor(seconds || 0));
  const days = Math.floor(total / 86400);
  const hours = Math.floor((total % 86400) / 3600);
  const minutes = Math.floor((total % 3600) / 60);

  if (days > 0) return `${days}d ${hours}h`;
  if (hours > 0) return `${hours}h ${minutes}m`;
  return `${minutes}m`;
}

function formatPercent(value) {
  if (value == null) return "n/a";
  return `${Number(value).toFixed(1)}%`;
}

function formatAge(seconds) {
  if (seconds == null) return "n/a";
  if (seconds < 60) return `${formatNumber(seconds)}s`;
  if (seconds < 3600) return `${formatNumber(Math.floor(seconds / 60))}m`;
  return `${formatNumber((seconds / 3600).toFixed(1))}h`;
}

function formatLatencyRange(best, worst) {
  if (best == null && worst == null) return "n/a";
  if (best != null && worst != null) return `${best.toFixed(1)}-${worst.toFixed(1)} ms`;
  const value = best ?? worst;
  return `${value.toFixed(1)} ms`;
}

function formatReflection(network) {
  if (network.has_reflection_v4 && network.has_reflection_v6) {
    return "v4 + v6";
  }
  if (network.has_reflection_v4) {
    return `v4 · ${formatAge(network.reflection_v4_age_secs)}`;
  }
  if (network.has_reflection_v6) {
    return `v6 · ${formatAge(network.reflection_v6_age_secs)}`;
  }
  return "none";
}

function formatDurationMs(value) {
  const ms = Number(value ?? 0);
  if (!ms) return "0 ms";
  if (ms < 1000) return `${formatNumber(ms)} ms`;
  if (ms < 60000) return `${(ms / 1000).toFixed(1)} s`;
  return `${(ms / 60000).toFixed(1)} m`;
}

function sumBy(items, pick) {
  return items.reduce((sum, item) => sum + Number(pick(item) || 0), 0);
}

function isDirectMode(server) {
  return server.mode === "DC";
}

function isFiniteNumber(value) {
  return Number.isFinite(Number(value));
}

function tooltipText(label) {
  return TOOLTIP_TEXT[label] || "";
}

function tooltipAttrs(text) {
  if (!text) {
    return "";
  }
  return ` data-tooltip="${escapeHtml(text)}"`;
}

function escapeHtml(value) {
  return String(value ?? "")
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#39;");
}

refreshButton.addEventListener("click", () => requestSnapshot(true));
