const $ = (selector) => document.querySelector(selector);
const $$ = (selector) => Array.from(document.querySelectorAll(selector));

const defaultTypes = ["A", "AAAA", "CNAME", "MX", "NS", "TXT", "SOA", "CAA", "SRV", "PTR", "HTTPS", "SVCB"];
const storageKey = "ipcheck.gui.settings.v1";

const state = {
  results: [],
  summary: null,
  selected: -1,
  abort: null,
  busy: false,
  viewMode: "cards",
  privacyMode: false,
  toastTimer: 0
};

function init() {
  renderTypeGrid();
  restoreSettings();
  if (!$("#endpoint-list").children.length) {
    addEndpoint("udp://1.1.1.1:53", false);
  }
  bindEvents();
  updateDomainMeta();
  updateEndpointMeta();
  setViewMode(state.viewMode, false);
  setPrivacyMode(state.privacyMode, false);
  ping();
  renderSummary();
  renderResults();
  renderDetail();
}

function bindEvents() {
  $("#add-endpoint").addEventListener("click", () => addEndpoint(""));
  $("#run").addEventListener("click", runResolve);
  $("#cancel").addEventListener("click", cancelResolve);
  $("#clear-all").addEventListener("click", clearAll);
  $("#load-sample").addEventListener("click", loadSample);
  $("#append-root-dot").addEventListener("click", appendRootDot);
  $("#filter").addEventListener("input", renderResults);
  $("#status-filter").addEventListener("change", renderResults);
  $("#export-json").addEventListener("click", exportJSON);
  $("#export-csv").addEventListener("click", exportCSV);
  $("#copy-detail").addEventListener("click", copySelected);
  $("#privacy-mode").addEventListener("change", (event) => setPrivacyMode(event.target.checked));
  $("#domain-file").addEventListener("change", importFile);
  $("#domains").addEventListener("input", updateDomainMeta);

  $$(".preset-row button").forEach((button) => {
    button.addEventListener("click", () => addEndpoint(button.dataset.preset));
  });

  $$(".view-button").forEach((button) => {
    button.addEventListener("click", () => setViewMode(button.dataset.view));
  });

  document.body.addEventListener("change", saveSettings);
  document.body.addEventListener("input", debounce(saveSettings, 250));
}

async function ping() {
  try {
    const response = await fetch("/api/health");
    const data = await response.json();
    $("#health").textContent = `${data.status} · ${data.version}`;
  } catch {
    $("#health").textContent = "Offline";
  }
}

function renderTypeGrid() {
  const grid = $("#type-grid");
  grid.innerHTML = defaultTypes.map((type) => `
    <label>
      <input type="checkbox" name="rtype" value="${type}" ${type === "A" || type === "AAAA" ? "checked" : ""}>
      <span>${type}</span>
    </label>
  `).join("");
}

function addEndpoint(value, persist = true) {
  const row = document.createElement("div");
  row.className = "endpoint-row";
  row.innerHTML = `
    <span class="endpoint-protocol">${escapeHTML(endpointProtocol(value))}</span>
    <input class="endpoint-input" type="text" spellcheck="false" value="${escapeAttr(value)}" placeholder="udp://8.8.8.8:5353">
    <button type="button" title="移除">删除</button>
  `;

  const input = row.querySelector(".endpoint-input");
  const protocol = row.querySelector(".endpoint-protocol");
  input.addEventListener("input", () => {
    protocol.textContent = endpointProtocol(input.value);
  });

  row.querySelector("button").addEventListener("click", () => {
    row.remove();
    updateEndpointMeta();
    saveSettings();
  });

  $("#endpoint-list").appendChild(row);
  updateEndpointMeta();
  if (persist) saveSettings();
}

function gatherPayload() {
  const typeSet = new Set($$("input[name='rtype']:checked").map((input) => input.value));
  splitList($("#custom-types").value).forEach((type) => typeSet.add(type.toUpperCase()));
  return {
    domainText: $("#domains").value,
    dns: $$(".endpoint-input").map((input) => input.value.trim()).filter(Boolean),
    types: Array.from(typeSet),
    timeoutMs: numberValue("#timeout-ms", 3000),
    retries: numberValue("#retries", 1),
    concurrency: numberValue("#concurrency", 16),
    strict: $("#strict").checked,
    ipInfo: $("#ip-info").checked,
    edns: $("#edns").checked,
    dnssec: $("#dnssec").checked,
    dohMethod: $("#doh-method").value,
    insecureSkipVerify: $("#insecure").checked
  };
}

async function runResolve() {
  const payload = gatherPayload();
  clearFormNotice();
  clearResultMessage();

  if (!payload.domainText.trim()) {
    setFormNotice("请输入域名", "error");
    setHealth("请输入域名");
    return;
  }
  if (!payload.dns.length) {
    setFormNotice("请添加 DNS", "error");
    setHealth("请添加 DNS");
    return;
  }

  state.abort = new AbortController();
  state.busy = true;
  setBusy(true);
  setHealth("Resolving");
  setResultMessage("解析中...", "loading");
  renderResults();

  try {
    const response = await fetch("/api/resolve", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
      signal: state.abort.signal
    });
    const data = await response.json();
    if (!response.ok) {
      throw new Error(data.error || `HTTP ${response.status}`);
    }
    state.results = data.results || [];
    state.summary = data.summary || null;
    state.selected = state.results.length ? 0 : -1;
    clearResultMessage();
    renderSummary();
    renderResults();
    renderDetail();
    setHealth(`Done · ${state.results.length}`);
    showToast("解析完成");
  } catch (error) {
    if (error.name === "AbortError") {
      setHealth("Canceled");
      setResultMessage("已取消", "");
    } else {
      setHealth(error.message);
      setResultMessage(error.message, "error");
    }
  } finally {
    state.busy = false;
    setBusy(false);
    state.abort = null;
    renderResults();
  }
}

function cancelResolve() {
  if (state.abort) {
    state.abort.abort();
  }
}

function renderSummary() {
  const summary = state.summary || {};
  const errorCount = (summary.error || 0) + (summary.dnsError || 0) + (summary.invalid || 0);
  $("#sum-total").textContent = summary.total || 0;
  $("#sum-ok").textContent = summary.ok || 0;
  $("#sum-ip").textContent = summary.withIps || 0;
  $("#sum-error").textContent = errorCount;
  $("#sum-ms").textContent = summary.durationMs || 0;
  $("#metric-error").classList.toggle("has-errors", errorCount > 0);
}

function renderResults() {
  const rows = filteredRows();
  syncSelection(rows);
  renderResultCards(rows);
  renderResultTable(rows);
  renderDetail();
}

function filteredRows() {
  const query = $("#filter").value.trim().toLowerCase();
  const status = $("#status-filter").value;
  return state.results
    .map((item, index) => ({ item, index }))
    .filter(({ item }) => {
      if (status && item.status !== status) return false;
      if (!query) return true;
      return [
        item.input, item.domain, item.ascii, item.fqdn, item.type, item.resolver,
        item.resolverProtocol, item.transportProtocol, item.status, item.rcode,
        answerText(item), locationText(item), operatorText(item), item.error,
        (item.warnings || []).join(" ")
      ].filter(Boolean).join(" ").toLowerCase().includes(query);
    });
}

function syncSelection(rows) {
  if (!rows.length) {
    state.selected = -1;
    return;
  }
  if (!rows.some(({ index }) => index === state.selected)) {
    state.selected = rows[0].index;
  }
}

function renderResultCards(rows) {
  const list = $("#result-list");
  if (!rows.length) {
    list.innerHTML = emptyStateHTML();
    return;
  }

  list.innerHTML = rows.map(({ item, index }) => resultCardHTML(item, index)).join("");
  list.querySelectorAll(".result-card").forEach((card) => {
    card.addEventListener("click", () => selectResult(card.dataset.index));
  });
}

function resultCardHTML(item, index) {
  const rawAnswer = answerText(item);
  const selected = index === state.selected;
  const protocol = protocolForItem(item);
  const warningText = (item.warnings || []).join("; ");
  const answerValue = rawAnswer ? displayAnswerText(item) : emptyAnswerText(item);
  return `
    <button type="button" class="result-card ${selected ? "selected" : ""}" data-index="${index}" aria-pressed="${selected}">
      <span class="result-card-head">
        <span class="result-title">
          ${statusBadge(item.status)}
          <span class="record-chip">${escapeHTML(item.type || "TYPE")}</span>
          <strong>${escapeHTML(displayDomain(item.input || item.domain || ""))}</strong>
        </span>
        <span class="duration-pill">${Number(item.durationMs || 0)} ms</span>
      </span>
      <span class="answer-block ${rawAnswer ? "" : "is-empty"}">${escapeHTML(answerValue)}</span>
      <span class="result-meta">
        ${metaPill("DNS", displayEndpoint(item.resolver))}
        ${metaPill("协议", protocol)}
        ${metaPill("RCODE", item.rcode)}
        ${metaPill("位置", locationText(item))}
        ${metaPill("运营商", operatorText(item))}
        ${metaPill("警告", warningText)}
      </span>
    </button>
  `;
}

function renderResultTable(rows) {
  const tbody = $("#result-body");
  if (!rows.length) {
    tbody.innerHTML = `<tr class="empty-row"><td colspan="10">${escapeHTML(emptyStateTitle())}</td></tr>`;
    return;
  }

  tbody.innerHTML = rows.map(({ item, index }) => `
    <tr data-index="${index}" class="${index === state.selected ? "selected" : ""}" tabindex="0">
      <td>${statusBadge(item.status)}</td>
      <td>${escapeHTML(displayDomain(item.input || ""))}</td>
      <td>${escapeHTML(item.type || "")}</td>
      <td>${escapeHTML(displayEndpoint(item.resolver || ""))}</td>
      <td>${escapeHTML(protocolForItem(item))}</td>
      <td>${escapeHTML(item.rcode || "")}</td>
      <td>${escapeHTML(displayAnswerText(item))}</td>
      <td>${escapeHTML(locationText(item))}</td>
      <td>${escapeHTML(operatorText(item))}</td>
      <td>${Number(item.durationMs || 0)}</td>
    </tr>
  `).join("");

  tbody.querySelectorAll("tr[data-index]").forEach((row) => {
    row.addEventListener("click", () => selectResult(row.dataset.index));
    row.addEventListener("keydown", (event) => {
      if (event.key === "Enter" || event.key === " ") {
        event.preventDefault();
        selectResult(row.dataset.index);
      }
    });
  });
}

function selectResult(index) {
  state.selected = Number(index);
  renderResults();
}

function renderDetail() {
  const root = $("#detail-content");
  const item = state.results[state.selected];
  const copyButton = $("#copy-detail");

  if (!item) {
    $("#detail-title").textContent = "未选择结果";
    $("#detail-subtitle").textContent = "No selection";
    copyButton.disabled = true;
    root.innerHTML = `<div class="empty-state"><div><h3>未选择结果</h3><p>No selection</p></div></div>`;
    return;
  }

  $("#detail-title").textContent = displayDomain(item.input || item.domain || "结果详情");
  $("#detail-subtitle").textContent = [item.status, item.type, protocolForItem(item)].filter(Boolean).join(" · ");
  copyButton.disabled = false;

  root.innerHTML = `
    <dl class="detail-list">
      ${detailItem("状态", statusBadge(item.status))}
      ${detailItem("输入", displayDomain(item.input))}
      ${detailItem("域名", displayDomain(item.domain))}
      ${detailItem("ASCII", displayDomain(item.ascii))}
      ${detailItem("FQDN", displayDomain(item.fqdn))}
      ${detailItem("DNS", displayEndpoint(item.resolver))}
      ${detailItem("协议", protocolForItem(item))}
      ${detailItem("RCODE", item.rcode)}
      ${detailItem("结果", displayAnswerText(item) || emptyAnswerText(item))}
      ${detailItem("位置", locationText(item))}
      ${detailItem("运营商", operatorText(item))}
      ${detailItem("错误", displaySensitive(item.error))}
    </dl>
    ${warningsSection(item.warnings)}
    ${ipInfoSection(item.ipInfo)}
    ${recordSection("Answer", item.answer, true)}
    ${recordSection("Authority", item.authority, false)}
    ${recordSection("Additional", item.additional, false)}
    <details class="detail-section">
      <summary>JSON</summary>
      <pre>${escapeHTML(JSON.stringify(displayResult(item), null, 2))}</pre>
    </details>
  `;
}

function detailItem(label, value) {
  if (value === undefined || value === null || value === "") return "";
  const content = String(value).startsWith("<span class=\"status") ? value : escapeHTML(value);
  return `<dt>${escapeHTML(label)}</dt><dd>${content}</dd>`;
}

function recordSection(title, records, open) {
  if (!records || !records.length) return "";
  return `
    <details class="detail-section" ${open ? "open" : ""}>
      <summary>${escapeHTML(title)} · ${records.length}</summary>
      <table>
        <thead><tr><th>Name</th><th>Type</th><th>TTL</th><th>Data</th></tr></thead>
        <tbody>
          ${records.map((record) => `
            <tr>
              <td>${escapeHTML(displayDomain(record.name || ""))}</td>
              <td>${escapeHTML(record.type || "")}</td>
              <td>${Number(record.ttl || 0)}</td>
              <td>${escapeHTML(displaySensitive(record.data || ""))}</td>
            </tr>
          `).join("")}
        </tbody>
      </table>
    </details>
  `;
}

function warningsSection(warnings) {
  if (!warnings || !warnings.length) return "";
  return `
    <details class="detail-section" open>
      <summary>Warnings · ${warnings.length}</summary>
      <div class="answer-block">${escapeHTML(displaySensitive(warnings.join("; ")))}</div>
    </details>
  `;
}

function ipInfoSection(infos) {
  if (!infos || !infos.length) return "";
  return `
    <details class="detail-section" open>
      <summary>IP 信息 · ${infos.length}</summary>
      <table>
        <thead><tr><th>IP</th><th>位置</th><th>ASN</th><th>运营商</th><th>来源</th></tr></thead>
        <tbody>
          ${infos.map((info) => `
            <tr>
              <td>${escapeHTML(displayIP(info.ip || ""))}</td>
              <td>${escapeHTML(locationFromInfo(info))}</td>
              <td>${escapeHTML(info.asn || "")}</td>
              <td>${escapeHTML(info.isp || info.org || "")}</td>
              <td>${escapeHTML(info.error || info.provider || "")}</td>
            </tr>
          `).join("")}
        </tbody>
      </table>
    </details>
  `;
}

function answerText(item) {
  if (item.ips && item.ips.length) return item.ips.join(", ");
  if (item.answer && item.answer.length) {
    return item.answer.map((record) => `${record.type}=${record.data}`).join("; ");
  }
  if (item.cnameChain && item.cnameChain.length) return item.cnameChain.join(" -> ");
  return item.error || "";
}

function emptyAnswerText(item) {
  if (item.status === "NO_ANSWER") return "无答案";
  if (item.status === "OK") return "无记录数据";
  return item.error || "No data";
}

function locationText(item) {
  return (item.ipInfo || [])
    .map(locationFromInfo)
    .filter(Boolean)
    .join("; ");
}

function operatorText(item) {
  return (item.ipInfo || [])
    .map((info) => [info.asn, info.isp || info.org].filter(Boolean).join(" "))
    .filter(Boolean)
    .join("; ");
}

function locationFromInfo(info) {
  if (!info || info.error) return info && info.error ? info.error : "";
  return [info.country, info.region, info.city].filter(Boolean).join(" / ");
}

function statusBadge(status) {
  const label = status || "UNKNOWN";
  return `<span class="status ${statusClass(label)}">${escapeHTML(label)}</span>`;
}

function statusClass(status) {
  const key = String(status || "unknown").toLowerCase().replaceAll("_", "-").replace(/[^a-z0-9-]/g, "");
  return `status-${key || "unknown"}`;
}

function metaPill(label, value) {
  if (!value) return "";
  return `<span class="meta-pill"><b>${escapeHTML(label)}</b><span>${escapeHTML(value)}</span></span>`;
}

function displayDomain(value) {
  return displaySensitive(value);
}

function displayEndpoint(value) {
  if (!value) return "";
  return state.privacyMode ? maskEndpoint(value) : value;
}

function displayIP(value) {
  if (!value) return "";
  return state.privacyMode ? maskIP(String(value)) : value;
}

function displayAnswerText(item) {
  return displaySensitive(answerText(item));
}

function displaySensitive(value) {
  if (value === undefined || value === null || value === "") return "";
  return state.privacyMode ? maskSensitiveText(value) : value;
}

function displayResult(value) {
  return state.privacyMode ? maskForDisplay(value) : value;
}

function maskForDisplay(value) {
  if (Array.isArray(value)) {
    return value.map(maskForDisplay);
  }
  if (value && typeof value === "object") {
    return Object.fromEntries(Object.entries(value).map(([key, nested]) => [key, maskForDisplay(nested)]));
  }
  if (typeof value === "string") {
    return maskSensitiveText(value);
  }
  return value;
}

function maskEndpoint(value) {
  const text = String(value || "");
  const match = text.match(/^([a-z][a-z0-9+.-]*:\/\/)(.+)$/i);
  if (!match) return maskSensitiveText(text);
  return `${match[1]}${maskSensitiveText(match[2])}`;
}

function maskSensitiveText(value) {
  let text = String(value ?? "");
  if (!text) return "";

  const protectedValues = [];
  text = text.replace(/\b(?:[a-f0-9]{1,4}:){2,}[a-f0-9:.]{1,}\b/gi, (ip) => protectMasked(protectedValues, maskIPv6(ip)));
  text = text.replace(/\b(?:\d{1,3}\.){3}\d{1,3}\b/g, (ip) => protectMasked(protectedValues, maskIPv4(ip)));
  text = text.replace(/(^|[^\p{L}\p{N}_-])((?:[\p{L}\p{N}_-]+\.)+[\p{L}\p{N}_-]{2,}\.?)/gu, (_match, prefix, domain) => {
    return `${prefix}${maskDomainToken(domain)}`;
  });

  protectedValues.forEach((masked, index) => {
    text = text.replaceAll(`__MASKED_${index}__`, masked);
  });
  return text;
}

function protectMasked(list, masked) {
  const token = `__MASKED_${list.length}__`;
  list.push(masked);
  return token;
}

function maskIP(value) {
  const text = String(value || "");
  if (text.includes(":")) return maskIPv6(text);
  if (text.includes(".")) return maskIPv4(text);
  return maskToken(text);
}

function maskIPv4(ip) {
  const parts = String(ip).split(".");
  if (parts.length !== 4) return maskToken(ip);
  return `${parts[0]}.${parts[1]}.xxx.xxx`;
}

function maskIPv6(ip) {
  const parts = String(ip).split(":");
  if (parts.length <= 3) return maskToken(ip);
  const head = parts.slice(0, 2).join(":");
  const tail = parts[parts.length - 1] || "****";
  return `${head}:****:${tail}`;
}

function maskDomainToken(value) {
  const text = String(value || "");
  const trailingDot = text.endsWith(".");
  const body = trailingDot ? text.slice(0, -1) : text;
  const labels = body.split(".");
  if (labels.length < 2) return maskToken(text);
  return labels.map((label, index) => {
    if (index === labels.length - 1) return label;
    return maskToken(label);
  }).join(".") + (trailingDot ? "." : "");
}

function maskToken(value) {
  const chars = Array.from(String(value || ""));
  if (!chars.length) return "";
  if (chars.length === 1) return "*";
  if (chars.length === 2) return `${chars[0]}*`;
  const visible = Math.ceil(chars.length / 2);
  const start = Math.max(1, Math.floor(visible / 2));
  const end = Math.max(1, visible - start);
  const middle = Math.max(3, chars.length - start - end);
  return `${chars.slice(0, start).join("")}${"*".repeat(middle)}${chars.slice(-end).join("")}`;
}

function protocolForItem(item) {
  return item.transportProtocol || item.resolverProtocol || endpointProtocol(item.resolver || "");
}

function endpointProtocol(value) {
  const scheme = String(value || "").trim().split("://")[0].toLowerCase();
  switch (scheme) {
    case "udp":
    case "dns":
      return "UDP";
    case "tcp":
      return "TCP";
    case "tls":
    case "dot":
      return "DoT";
    case "https":
    case "doh":
      return "DoH";
    case "http":
      return "HTTP";
    case "quic":
    case "doq":
      return "DoQ";
    default:
      return "DNS";
  }
}

function emptyStateHTML() {
  return `
    <div class="empty-state">
      <div>
        <h3>${escapeHTML(emptyStateTitle())}</h3>
        <p>${state.busy ? "Resolving" : state.results.length ? "No match" : "Ready"}</p>
      </div>
    </div>
  `;
}

function emptyStateTitle() {
  if (state.busy) return "解析中";
  return state.results.length ? "没有匹配结果" : "等待查询";
}

function setViewMode(mode, persist = true) {
  state.viewMode = mode === "table" ? "table" : "cards";
  $("#view-cards").classList.toggle("is-active", state.viewMode === "cards");
  $("#view-table").classList.toggle("is-active", state.viewMode === "table");
  $("#result-list").hidden = state.viewMode !== "cards";
  $("#table-view").hidden = state.viewMode !== "table";
  if (persist) saveSettings();
}

function setPrivacyMode(enabled, persist = true) {
  state.privacyMode = Boolean(enabled);
  $("#privacy-mode").checked = state.privacyMode;
  document.body.classList.toggle("privacy-mode", state.privacyMode);
  renderResults();
  if (persist) saveSettings();
}

function updateDomainMeta() {
  const count = $("#domains").value.split(/\r?\n/).map((line) => line.trim()).filter(Boolean).length;
  $("#domain-count").textContent = `${count} 个域名`;
}

function updateEndpointMeta() {
  const count = $$(".endpoint-input").filter((input) => input.value.trim()).length;
  $("#endpoint-count").textContent = `${count} 个端点`;
}

function setBusy(busy) {
  state.busy = busy;
  $("#run").disabled = busy;
  $("#run").textContent = busy ? "解析中..." : "解析";
  $("#cancel").disabled = !busy;
}

function setHealth(text) {
  $("#health").textContent = text;
}

function setFormNotice(text, tone) {
  setNotice("#form-notice", text, tone);
}

function clearFormNotice() {
  clearNotice("#form-notice");
}

function setResultMessage(text, tone) {
  setNotice("#result-message", text, tone);
}

function clearResultMessage() {
  clearNotice("#result-message");
}

function setNotice(selector, text, tone) {
  const node = $(selector);
  node.textContent = text;
  node.hidden = false;
  node.classList.toggle("is-error", tone === "error");
  node.classList.toggle("is-loading", tone === "loading");
}

function clearNotice(selector) {
  const node = $(selector);
  node.textContent = "";
  node.hidden = true;
  node.classList.remove("is-error", "is-loading");
}

function showToast(text) {
  const toast = $("#toast");
  window.clearTimeout(state.toastTimer);
  toast.textContent = text;
  toast.hidden = false;
  state.toastTimer = window.setTimeout(() => {
    toast.hidden = true;
  }, 1800);
}

function exportJSON() {
  download("ipcheck-results.json", JSON.stringify({ summary: state.summary, results: state.results }, null, 2), "application/json");
}

function exportCSV() {
  const header = ["input", "domain", "ascii", "type", "resolver", "protocol", "status", "rcode", "ips", "answers", "location", "operator", "duration_ms", "error", "warnings"];
  const rows = state.results.map((item) => [
    item.input, item.domain, item.ascii, item.type, item.resolver, protocolForItem(item),
    item.status, item.rcode, (item.ips || []).join(";"), answerText(item),
    locationText(item), operatorText(item), item.durationMs || 0, item.error || "",
    (item.warnings || []).join(";")
  ]);
  const csv = [header, ...rows].map((row) => row.map(csvCell).join(",")).join("\n");
  download("ipcheck-results.csv", csv, "text/csv");
}

function copySelected() {
  const item = state.results[state.selected];
  if (!item || !navigator.clipboard) return;
  navigator.clipboard.writeText(JSON.stringify(displayResult(item), null, 2))
    .then(() => {
      setHealth("Copied");
      showToast("已复制");
    })
    .catch((error) => {
      setHealth(error.message);
      showToast("复制失败");
    });
}

function download(filename, content, type) {
  const blob = new Blob([content], { type });
  const url = URL.createObjectURL(blob);
  const link = document.createElement("a");
  link.href = url;
  link.download = filename;
  document.body.appendChild(link);
  link.click();
  link.remove();
  URL.revokeObjectURL(url);
}

function importFile(event) {
  const file = event.target.files[0];
  if (!file) return;
  const reader = new FileReader();
  reader.onload = () => {
    $("#domains").value = String(reader.result || "");
    updateDomainMeta();
    saveSettings();
    event.target.value = "";
  };
  reader.readAsText(file);
}

function appendRootDot() {
  const lines = $("#domains").value.split(/\r?\n/).map((line) => {
    const trimmed = line.trim();
    if (!trimmed || trimmed.endsWith(".")) return line;
    return `${trimmed}.`;
  });
  $("#domains").value = lines.join("\n");
  updateDomainMeta();
  saveSettings();
}

function loadSample() {
  $("#domains").value = "example.com\na.b.c.d.e.f.g.example.com\n例子.测试\na..b.com\n_sip._tcp.example.com";
  $("#endpoint-list").innerHTML = "";
  [
    "udp://1.1.1.1:53",
    "tcp://1.1.1.1:53",
    "https://dns.google:443/dns-query",
    "tls://1.1.1.1:853",
    "quic://dns.google:853"
  ].forEach((endpoint) => addEndpoint(endpoint, false));
  updateDomainMeta();
  updateEndpointMeta();
  saveSettings();
}

function clearAll() {
  state.results = [];
  state.summary = null;
  state.selected = -1;
  $("#domains").value = "";
  $("#endpoint-list").innerHTML = "";
  $("#filter").value = "";
  $("#status-filter").value = "";
  addEndpoint("udp://1.1.1.1:53", false);
  clearFormNotice();
  clearResultMessage();
  updateDomainMeta();
  updateEndpointMeta();
  renderSummary();
  renderResults();
  renderDetail();
  saveSettings();
}

function saveSettings() {
  const data = {
    domains: $("#domains").value,
    dns: $$(".endpoint-input").map((input) => input.value),
    types: $$("input[name='rtype']:checked").map((input) => input.value),
    customTypes: $("#custom-types").value,
    timeoutMs: $("#timeout-ms").value,
    retries: $("#retries").value,
    concurrency: $("#concurrency").value,
    strict: $("#strict").checked,
    ipInfo: $("#ip-info").checked,
    edns: $("#edns").checked,
    dnssec: $("#dnssec").checked,
    dohMethod: $("#doh-method").value,
    insecure: $("#insecure").checked,
    viewMode: state.viewMode,
    privacyMode: state.privacyMode
  };
  localStorage.setItem(storageKey, JSON.stringify(data));
}

function restoreSettings() {
  const raw = localStorage.getItem(storageKey);
  if (!raw) return;
  try {
    const data = JSON.parse(raw);
    $("#domains").value = data.domains ?? $("#domains").value;
    $("#endpoint-list").innerHTML = "";
    (data.dns || []).forEach((endpoint) => addEndpoint(endpoint, false));
    $$("input[name='rtype']").forEach((input) => {
      input.checked = (data.types || ["A", "AAAA"]).includes(input.value);
    });
    $("#custom-types").value = data.customTypes || "";
    $("#timeout-ms").value = data.timeoutMs || "3000";
    $("#retries").value = data.retries || "1";
    $("#concurrency").value = data.concurrency || "16";
    $("#strict").checked = Boolean(data.strict);
    $("#ip-info").checked = data.ipInfo !== false;
    $("#edns").checked = data.edns !== false;
    $("#dnssec").checked = Boolean(data.dnssec);
    $("#doh-method").value = data.dohMethod || "POST";
    $("#insecure").checked = Boolean(data.insecure);
    state.viewMode = data.viewMode === "table" ? "table" : "cards";
    state.privacyMode = Boolean(data.privacyMode);
  } catch {
    localStorage.removeItem(storageKey);
  }
}

function splitList(value) {
  return value.split(/[,\s]+/).map((part) => part.trim()).filter(Boolean);
}

function numberValue(selector, fallback) {
  const value = Number($(selector).value);
  return Number.isFinite(value) && value >= 0 ? value : fallback;
}

function csvCell(value) {
  const text = String(value ?? "");
  if (/[",\n\r]/.test(text)) return `"${text.replaceAll('"', '""')}"`;
  return text;
}

function escapeHTML(value) {
  return String(value ?? "")
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#039;");
}

function escapeAttr(value) {
  return escapeHTML(value).replaceAll("`", "&#096;");
}

function debounce(fn, delay) {
  let timer = 0;
  return (...args) => {
    window.clearTimeout(timer);
    timer = window.setTimeout(() => fn(...args), delay);
  };
}

document.addEventListener("DOMContentLoaded", init);
