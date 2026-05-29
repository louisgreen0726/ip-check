const $ = (selector) => document.querySelector(selector);
const $$ = (selector) => Array.from(document.querySelectorAll(selector));

const defaultTypes = ["A", "AAAA", "CNAME", "MX", "NS", "TXT", "SOA", "CAA", "SRV", "PTR", "HTTPS", "SVCB"];
const storageKey = "ipcheck.gui.settings.v1";

const state = {
  results: [],
  summary: null,
  selected: -1,
  abort: null
};

function init() {
  renderTypeGrid();
  restoreSettings();
  if (!$("#endpoint-list").children.length) {
    addEndpoint("udp://1.1.1.1:53");
  }
  bindEvents();
  ping();
  renderResults();
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
  $("#domain-file").addEventListener("change", importFile);

  $$(".preset-row button").forEach((button) => {
    button.addEventListener("click", () => addEndpoint(button.dataset.preset));
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
      ${type}
    </label>
  `).join("");
}

function addEndpoint(value) {
  const row = document.createElement("div");
  row.className = "endpoint-row";
  row.innerHTML = `
    <input class="endpoint-input" type="text" spellcheck="false" value="${escapeAttr(value)}" placeholder="udp://8.8.8.8:5353">
    <button type="button" title="移除">x</button>
  `;
  row.querySelector("button").addEventListener("click", () => {
    row.remove();
    saveSettings();
  });
  $("#endpoint-list").appendChild(row);
  saveSettings();
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
  if (!payload.domainText.trim()) {
    setHealth("请输入域名");
    return;
  }
  if (!payload.dns.length) {
    setHealth("请添加 DNS");
    return;
  }

  state.abort = new AbortController();
  setBusy(true);
  setHealth("Resolving");

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
    renderSummary();
    renderResults();
    renderDetail();
    setHealth("Done");
  } catch (error) {
    if (error.name === "AbortError") {
      setHealth("Canceled");
    } else {
      setHealth(error.message);
    }
  } finally {
    setBusy(false);
    state.abort = null;
  }
}

function cancelResolve() {
  if (state.abort) {
    state.abort.abort();
  }
}

function renderSummary() {
  const summary = state.summary || {};
  $("#sum-total").textContent = summary.total || 0;
  $("#sum-ok").textContent = summary.ok || 0;
  $("#sum-ip").textContent = summary.withIps || 0;
  $("#sum-error").textContent = (summary.error || 0) + (summary.dnsError || 0) + (summary.invalid || 0);
  $("#sum-ms").textContent = summary.durationMs || 0;
}

function renderResults() {
  const tbody = $("#result-body");
  const query = $("#filter").value.trim().toLowerCase();
  const status = $("#status-filter").value;
  const rows = state.results
    .map((item, index) => ({ item, index }))
    .filter(({ item }) => {
      if (status && item.status !== status) return false;
      if (!query) return true;
      return [
        item.input, item.domain, item.ascii, item.type, item.resolver,
        item.transportProtocol, item.status, item.rcode, answerText(item),
        locationText(item), operatorText(item), item.error
      ].filter(Boolean).join(" ").toLowerCase().includes(query);
    });

  if (!rows.length) {
    tbody.innerHTML = `<tr class="empty-row"><td colspan="10">${state.results.length ? "No match" : "Ready"}</td></tr>`;
    return;
  }

  tbody.innerHTML = rows.map(({ item, index }) => `
    <tr data-index="${index}" class="${index === state.selected ? "selected" : ""}">
      <td><span class="status ${item.status}">${item.status}</span></td>
      <td>${escapeHTML(item.input || "")}</td>
      <td>${escapeHTML(item.type || "")}</td>
      <td>${escapeHTML(item.resolver || "")}</td>
      <td>${escapeHTML(item.transportProtocol || "")}</td>
      <td>${escapeHTML(item.rcode || "")}</td>
      <td>${escapeHTML(answerText(item))}</td>
      <td>${escapeHTML(locationText(item))}</td>
      <td>${escapeHTML(operatorText(item))}</td>
      <td>${item.durationMs || 0}</td>
    </tr>
  `).join("");

  tbody.querySelectorAll("tr[data-index]").forEach((row) => {
    row.addEventListener("click", () => {
      state.selected = Number(row.dataset.index);
      renderResults();
      renderDetail();
    });
  });
}

function renderDetail() {
  const root = $("#detail-content");
  const item = state.results[state.selected];
  if (!item) {
    root.textContent = "No selection";
    return;
  }
  root.innerHTML = `
    <dl class="detail-list">
      <dt>输入</dt><dd>${escapeHTML(item.input || "")}</dd>
      <dt>ASCII</dt><dd>${escapeHTML(item.ascii || "")}</dd>
      <dt>FQDN</dt><dd>${escapeHTML(item.fqdn || "")}</dd>
      <dt>DNS</dt><dd>${escapeHTML(item.resolver || "")}</dd>
      <dt>状态</dt><dd>${escapeHTML(item.status || "")}</dd>
      <dt>结果</dt><dd>${escapeHTML(answerText(item))}</dd>
      <dt>位置</dt><dd>${escapeHTML(locationText(item))}</dd>
      <dt>运营商</dt><dd>${escapeHTML(operatorText(item))}</dd>
      <dt>错误</dt><dd>${escapeHTML(item.error || "")}</dd>
    </dl>
    ${ipInfoSection(item.ipInfo)}
    ${recordSection("Answer", item.answer)}
    ${recordSection("Authority", item.authority)}
    ${recordSection("Additional", item.additional)}
    <h3>JSON</h3>
    <pre>${escapeHTML(JSON.stringify(item, null, 2))}</pre>
  `;
}

function recordSection(title, records) {
  if (!records || !records.length) return "";
  return `
    <h3>${title}</h3>
    <table>
      <thead><tr><th>Name</th><th>Type</th><th>TTL</th><th>Data</th></tr></thead>
      <tbody>
        ${records.map((record) => `
          <tr>
            <td>${escapeHTML(record.name || "")}</td>
            <td>${escapeHTML(record.type || "")}</td>
            <td>${record.ttl || 0}</td>
            <td>${escapeHTML(record.data || "")}</td>
          </tr>
        `).join("")}
      </tbody>
    </table>
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

function ipInfoSection(infos) {
  if (!infos || !infos.length) return "";
  return `
    <h3>IP 信息</h3>
    <table>
      <thead><tr><th>IP</th><th>位置</th><th>ASN</th><th>运营商</th><th>来源</th></tr></thead>
      <tbody>
        ${infos.map((info) => `
          <tr>
            <td>${escapeHTML(info.ip || "")}</td>
            <td>${escapeHTML(locationFromInfo(info))}</td>
            <td>${escapeHTML(info.asn || "")}</td>
            <td>${escapeHTML(info.isp || info.org || "")}</td>
            <td>${escapeHTML(info.error || info.provider || "")}</td>
          </tr>
        `).join("")}
      </tbody>
    </table>
  `;
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

function exportJSON() {
  download("ipcheck-results.json", JSON.stringify({ summary: state.summary, results: state.results }, null, 2), "application/json");
}

function exportCSV() {
  const header = ["input", "domain", "ascii", "type", "resolver", "protocol", "status", "rcode", "ips", "answers", "location", "operator", "duration_ms", "error", "warnings"];
  const rows = state.results.map((item) => [
    item.input, item.domain, item.ascii, item.type, item.resolver, item.transportProtocol,
    item.status, item.rcode, (item.ips || []).join(";"), answerText(item),
    locationText(item), operatorText(item), item.durationMs || 0, item.error || "",
    (item.warnings || []).join(";")
  ]);
  const csv = [header, ...rows].map((row) => row.map(csvCell).join(",")).join("\n");
  download("ipcheck-results.csv", csv, "text/csv");
}

function copySelected() {
  const item = state.results[state.selected];
  if (!item) return;
  navigator.clipboard.writeText(JSON.stringify(item, null, 2)).then(() => setHealth("Copied"));
}

function download(filename, content, type) {
  const blob = new Blob([content], { type });
  const url = URL.createObjectURL(blob);
  const link = document.createElement("a");
  link.href = url;
  link.download = filename;
  link.click();
  URL.revokeObjectURL(url);
}

function importFile(event) {
  const file = event.target.files[0];
  if (!file) return;
  const reader = new FileReader();
  reader.onload = () => {
    $("#domains").value = String(reader.result || "");
    saveSettings();
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
  ].forEach(addEndpoint);
  saveSettings();
}

function clearAll() {
  state.results = [];
  state.summary = null;
  state.selected = -1;
  $("#domains").value = "";
  $("#endpoint-list").innerHTML = "";
  addEndpoint("udp://1.1.1.1:53");
  renderSummary();
  renderResults();
  renderDetail();
  saveSettings();
}

function setBusy(busy) {
  $("#run").disabled = busy;
  $("#cancel").disabled = !busy;
}

function setHealth(text) {
  $("#health").textContent = text;
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
    insecure: $("#insecure").checked
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
    (data.dns || []).forEach(addEndpoint);
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
