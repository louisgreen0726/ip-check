const $ = (selector) => document.querySelector(selector);
const $$ = (selector) => Array.from(document.querySelectorAll(selector));

const defaultTypes = ["A", "AAAA", "CNAME", "MX", "NS", "TXT", "SOA", "CAA", "SRV", "PTR", "HTTPS", "SVCB"];
const storageKey = "ipcheck.gui.settings.v1";
const defaultLanguage = "zh-CN";

const messages = {
  "zh-CN": {
    "brand.eyebrow": "本地 DNS 工作台",
    "language.aria": "语言 / Language",
    "settings.eyebrow": "查询配置",
    "settings.title": "解析请求",
    "domains.title": "域名",
    "actions.import": "导入",
    "actions.appendRootDot": "补根点",
    "dns.title": "DNS",
    "dns.presetsAria": "DNS 预设",
    "actions.add": "添加",
    "records.title": "记录类型",
    "records.subtitle": "默认 A / AAAA",
    "records.customPlaceholder": "自定义 TYPE，如 HTTPS,SVCB,CAA",
    "options.title": "参数",
    "options.subtitle": "超时、重试、并发",
    "options.timeout": "超时 ms",
    "options.retries": "重试",
    "options.concurrency": "并发",
    "options.doh": "DoH",
    "advanced.title": "高级选项",
    "advanced.ipInfo": "IP 信息",
    "advanced.strict": "严格域名",
    "advanced.insecure": "跳过 TLS 校验",
    "actions.resolve": "解析",
    "actions.resolving": "解析中...",
    "actions.cancel": "取消",
    "actions.sample": "示例",
    "actions.clear": "清空",
    "actions.remove": "删除",
    "actions.removeEndpoint": "移除",
    "actions.copy": "复制",
    "results.aria": "解析结果",
    "summary.aria": "结果摘要",
    "summary.total": "总计",
    "summary.totalSmall": "条结果",
    "summary.okSmall": "成功",
    "summary.ipSmall": "含信息",
    "summary.error": "错误",
    "summary.errorSmall": "需关注",
    "summary.duration": "耗时",
    "results.eyebrow": "查询结果",
    "results.title": "结果",
    "view.aria": "结果视图",
    "view.cards": "卡片",
    "view.table": "表格",
    "filter.label": "过滤结果",
    "filter.placeholder": "过滤结果",
    "statusFilter.aria": "状态筛选",
    "statusFilter.all": "全部状态",
    "privacy.label": "截图隐私",
    "export.json": "导出 JSON",
    "export.csv": "导出 CSV",
    "table.status": "状态",
    "table.input": "输入",
    "table.type": "类型",
    "table.dns": "DNS",
    "table.protocol": "协议",
    "table.rcode": "RCODE",
    "table.result": "结果",
    "table.location": "位置",
    "table.operator": "运营商",
    "detail.aria": "结果详情",
    "detail.eyebrow": "详情",
    "detail.emptyTitle": "未选择结果",
    "detail.noSelection": "No selection",
    "detail.fallbackTitle": "结果详情",
    "detail.overview": "概览",
    "detail.identity": "域名信息",
    "detail.response": "响应信息",
    "detail.raw": "原始 JSON",
    "detail.answerSection": "答案记录",
    "detail.authoritySection": "权威记录",
    "detail.additionalSection": "附加记录",
    "detail.status": "状态",
    "detail.input": "输入",
    "detail.domain": "域名",
    "detail.protocol": "协议",
    "detail.result": "结果",
    "detail.location": "位置",
    "detail.operator": "运营商",
    "detail.error": "错误",
    "detail.recordName": "名称",
    "detail.recordType": "类型",
    "detail.recordTTL": "TTL",
    "detail.recordData": "数据",
    "detail.warnings": "警告",
    "detail.ipInfo": "IP 信息",
    "detail.source": "来源",
    "health.ready": "Ready",
    "health.offline": "Offline",
    "health.resolving": "Resolving",
    "health.done": "Done · {count}",
    "health.canceled": "Canceled",
    "health.copied": "Copied",
    "validation.domainRequired": "请输入域名",
    "validation.dnsRequired": "请添加 DNS",
    "result.loading": "解析中...",
    "result.complete": "解析完成",
    "result.canceled": "已取消",
    "result.copied": "已复制",
    "result.copyFailed": "复制失败",
    "empty.busyTitle": "解析中",
    "empty.noMatchTitle": "没有匹配结果",
    "empty.waitingTitle": "等待查询",
    "empty.busySubtitle": "Resolving",
    "empty.noMatchSubtitle": "No match",
    "empty.readySubtitle": "Ready",
    "answer.noAnswer": "无答案",
    "answer.noRecordData": "无记录数据",
    "answer.noData": "No data",
    "count.domains": "{count} 个域名",
    "count.endpoints": "{count} 个端点",
    "meta.protocol": "协议",
    "meta.location": "位置",
    "meta.operator": "运营商",
    "meta.warning": "警告"
  },
  en: {
    "brand.eyebrow": "Local DNS Workspace",
    "language.aria": "Language",
    "settings.eyebrow": "Query Setup",
    "settings.title": "Resolve Request",
    "domains.title": "Domains",
    "actions.import": "Import",
    "actions.appendRootDot": "Add Root Dot",
    "dns.title": "DNS",
    "dns.presetsAria": "DNS presets",
    "actions.add": "Add",
    "records.title": "Record Types",
    "records.subtitle": "Default A / AAAA",
    "records.customPlaceholder": "Custom TYPE, e.g. HTTPS,SVCB,CAA",
    "options.title": "Options",
    "options.subtitle": "Timeout, retries, concurrency",
    "options.timeout": "Timeout ms",
    "options.retries": "Retries",
    "options.concurrency": "Concurrency",
    "options.doh": "DoH",
    "advanced.title": "Advanced Options",
    "advanced.ipInfo": "IP info",
    "advanced.strict": "Strict domains",
    "advanced.insecure": "Skip TLS verification",
    "actions.resolve": "Resolve",
    "actions.resolving": "Resolving...",
    "actions.cancel": "Cancel",
    "actions.sample": "Sample",
    "actions.clear": "Clear",
    "actions.remove": "Remove",
    "actions.removeEndpoint": "Remove endpoint",
    "actions.copy": "Copy",
    "results.aria": "Resolve results",
    "summary.aria": "Result summary",
    "summary.total": "Total",
    "summary.totalSmall": "results",
    "summary.okSmall": "Success",
    "summary.ipSmall": "With info",
    "summary.error": "Errors",
    "summary.errorSmall": "Needs attention",
    "summary.duration": "Duration",
    "results.eyebrow": "Query Results",
    "results.title": "Results",
    "view.aria": "Result view",
    "view.cards": "Cards",
    "view.table": "Table",
    "filter.label": "Filter results",
    "filter.placeholder": "Filter results",
    "statusFilter.aria": "Status filter",
    "statusFilter.all": "All statuses",
    "privacy.label": "Screenshot privacy",
    "export.json": "Export JSON",
    "export.csv": "Export CSV",
    "table.status": "Status",
    "table.input": "Input",
    "table.type": "Type",
    "table.dns": "DNS",
    "table.protocol": "Protocol",
    "table.rcode": "RCODE",
    "table.result": "Result",
    "table.location": "Location",
    "table.operator": "Operator",
    "detail.aria": "Result details",
    "detail.eyebrow": "Details",
    "detail.emptyTitle": "No Result Selected",
    "detail.noSelection": "No selection",
    "detail.fallbackTitle": "Result details",
    "detail.overview": "Overview",
    "detail.identity": "Domain Info",
    "detail.response": "Response Details",
    "detail.raw": "Raw JSON",
    "detail.answerSection": "Answer Records",
    "detail.authoritySection": "Authority Records",
    "detail.additionalSection": "Additional Records",
    "detail.status": "Status",
    "detail.input": "Input",
    "detail.domain": "Domain",
    "detail.protocol": "Protocol",
    "detail.result": "Result",
    "detail.location": "Location",
    "detail.operator": "Operator",
    "detail.error": "Error",
    "detail.recordName": "Name",
    "detail.recordType": "Type",
    "detail.recordTTL": "TTL",
    "detail.recordData": "Data",
    "detail.warnings": "Warnings",
    "detail.ipInfo": "IP info",
    "detail.source": "Source",
    "health.ready": "Ready",
    "health.offline": "Offline",
    "health.resolving": "Resolving",
    "health.done": "Done · {count}",
    "health.canceled": "Canceled",
    "health.copied": "Copied",
    "validation.domainRequired": "Enter at least one domain",
    "validation.dnsRequired": "Add at least one DNS endpoint",
    "result.loading": "Resolving...",
    "result.complete": "Resolve complete",
    "result.canceled": "Canceled",
    "result.copied": "Copied",
    "result.copyFailed": "Copy failed",
    "empty.busyTitle": "Resolving",
    "empty.noMatchTitle": "No Matching Results",
    "empty.waitingTitle": "Waiting for Query",
    "empty.busySubtitle": "Resolving",
    "empty.noMatchSubtitle": "No match",
    "empty.readySubtitle": "Ready",
    "answer.noAnswer": "No answer",
    "answer.noRecordData": "No record data",
    "answer.noData": "No data",
    "count.domains": "{count} domain{suffix}",
    "count.endpoints": "{count} endpoint{suffix}",
    "meta.protocol": "Protocol",
    "meta.location": "Location",
    "meta.operator": "Operator",
    "meta.warning": "Warning"
  }
};

const diagnosticTranslations = [
  { pattern: /^JSON 请求解析失败: (.+)$/u, en: ([, reason]) => `Failed to parse JSON request: ${reason}` },
  { pattern: /^请至少输入一个域名$/u, en: () => "Enter at least one domain" },
  { pattern: /^查询类型为空$/u, en: () => "Query type is empty" },
  { pattern: /^不支持的查询类型: (.+)$/u, en: ([, type]) => `Unsupported query type: ${type}` },
  { pattern: /^不支持的输出格式: (.+)$/u, en: ([, format]) => `Unsupported output format: ${format}` },
  { pattern: /^不支持的 DNS 协议: (.+)$/u, en: ([, protocol]) => `Unsupported DNS protocol: ${protocol}` },
  { pattern: /^DNS endpoint 为空$/u, en: () => "DNS endpoint is empty" },
  { pattern: /^DNS endpoint URL 解析失败: (.+)$/u, en: ([, reason]) => `Failed to parse DNS endpoint URL: ${reason}` },
  { pattern: /^DNS endpoint 缺少 host: (.+)$/u, en: ([, endpoint]) => `DNS endpoint is missing a host: ${endpoint}` },
  { pattern: /^域名为空$/u, en: () => "Domain is empty" },
  { pattern: /^输入为空$/u, en: () => "Input is empty" },
  { pattern: /^输入是 IP 地址，不是可查询的域名: (.+)$/u, en: ([, input]) => `Input is an IP address, not a queryable domain: ${input}` },
  { pattern: /^域名包含连续的点，DNS 不允许空 label: (.+)$/u, en: ([, domain]) => `Domain contains consecutive dots; DNS does not allow empty labels: ${domain}` },
  { pattern: /^域名包含空 label: (.+)$/u, en: ([, domain]) => `Domain contains an empty label: ${domain}` },
  { pattern: /^域名 label 包含空白或控制字符: (.+)$/u, en: ([, label]) => `Domain label contains whitespace or control characters: ${label}` },
  { pattern: /^域名 label 超过 63 字节: (.+)$/u, en: ([, label]) => `Domain label exceeds 63 bytes: ${label}` },
  { pattern: /^完整域名超过 253 字符: (.+)$/u, en: ([, length]) => `Full domain exceeds 253 characters: ${length}` },
  { pattern: /^URL 解析失败: (.+)$/u, en: ([, reason]) => `URL parse failed: ${reason}` },
  { pattern: /^label (.+) 包含下划线；DNS 可查询，但它不是标准主机名 label$/u, en: ([, label]) => `Label ${label} contains an underscore; DNS can query it, but it is not a standard hostname label` },
  { pattern: /^检测到通配符 label \*；将按字面 DNS 名称查询$/u, en: () => "Wildcard label * detected; it will be queried as a literal DNS name" },
  { pattern: /^label (.+) 以连字符开头或结尾；DNS 可查询，但它不是标准主机名 label$/u, en: ([, label]) => `Label ${label} starts or ends with a hyphen; DNS can query it, but it is not a standard hostname label` },
  { pattern: /^域名 label 包含不支持的字符 (.+): (.+)$/u, en: ([, char, label]) => `Domain label contains unsupported character ${char}: ${label}` },
  { pattern: /^DoH 响应不是合法 DNS message: (.+)$/u, en: ([, reason]) => `DoH response is not a valid DNS message: ${reason}` },
  { pattern: /^DNS message 超过 DoQ 2 字节长度限制: (.+)$/u, en: ([, length]) => `DNS message exceeds the DoQ 2-byte length limit: ${length}` },
  { pattern: /^DoQ 响应不是合法 DNS message: (.+)$/u, en: ([, reason]) => `DoQ response is not a valid DNS message: ${reason}` }
];

const state = {
  results: [],
  summary: null,
  selected: -1,
  abort: null,
  busy: false,
  viewMode: "cards",
  privacyMode: false,
  language: defaultLanguage,
  toastTimer: 0
};

function init() {
  state.language = savedLanguage();
  applyLanguage(state.language, false);
  renderTypeGrid();
  restoreSettings();
  if (!$("#endpoint-list").children.length) {
    addEndpoint("udp://1.1.1.1:53", false);
  }
  bindEvents();
  applyLanguage(state.language, false);
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

  $$(".lang-button").forEach((button) => {
    button.addEventListener("click", () => setLanguage(button.dataset.lang));
  });

  $$(".preset-row button").forEach((button) => {
    button.addEventListener("click", () => addEndpoint(button.dataset.preset));
  });

  $$(".view-button").forEach((button) => {
    button.addEventListener("click", () => setViewMode(button.dataset.view));
  });

  document.body.addEventListener("change", saveSettings);
  document.body.addEventListener("input", debounce(saveSettings, 250));
}

function savedLanguage() {
  const raw = localStorage.getItem(storageKey);
  if (!raw) return defaultLanguage;
  try {
    const data = JSON.parse(raw);
    return normalizeLanguage(data.language);
  } catch {
    return defaultLanguage;
  }
}

function normalizeLanguage(language) {
  const value = String(language || "").toLowerCase();
  return value === "en" || value.startsWith("en-") ? "en" : defaultLanguage;
}

function setLanguage(language) {
  applyLanguage(language);
}

function applyLanguage(language, persist = true) {
  state.language = normalizeLanguage(language);
  document.documentElement.lang = state.language;

  $$("[data-i18n]").forEach((node) => {
    node.textContent = t(node.dataset.i18n);
  });
  $$("[data-i18n-placeholder]").forEach((node) => {
    node.setAttribute("placeholder", t(node.dataset.i18nPlaceholder));
  });
  $$("[data-i18n-aria-label]").forEach((node) => {
    node.setAttribute("aria-label", t(node.dataset.i18nAriaLabel));
  });
  $$("[data-i18n-title]").forEach((node) => {
    node.setAttribute("title", t(node.dataset.i18nTitle));
  });

  $$(".lang-button").forEach((button) => {
    const active = normalizeLanguage(button.dataset.lang) === state.language;
    button.classList.toggle("is-active", active);
    button.setAttribute("aria-pressed", String(active));
  });

  translateEndpointRows();
  updateDomainMeta();
  updateEndpointMeta();
  setBusy(state.busy);
  renderSummary();
  renderResults();
  renderDetail();

  if (persist) saveSettings();
}

function translateEndpointRows() {
  $$(".endpoint-row button").forEach((button) => {
    button.textContent = t("actions.remove");
    button.title = t("actions.removeEndpoint");
  });
}

function t(key, params = {}) {
  const template = (messages[state.language] && messages[state.language][key])
    || messages[defaultLanguage][key]
    || key;
  return template.replace(/\{(\w+)\}/g, (_match, name) => {
    if (name === "suffix" && params.suffix === undefined) return countSuffix(params.count);
    return params[name] ?? "";
  });
}

function countSuffix(count) {
  return state.language === "en" && Number(count) !== 1 ? "s" : "";
}

async function ping() {
  try {
    const response = await fetch("/api/health");
    const data = await response.json();
    $("#health").textContent = `${data.status} · ${data.version}`;
  } catch {
    $("#health").textContent = t("health.offline");
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
    <button type="button" title="${escapeAttr(t("actions.removeEndpoint"))}">${escapeHTML(t("actions.remove"))}</button>
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
    setFormNotice(t("validation.domainRequired"), "error");
    setHealth(t("validation.domainRequired"));
    return;
  }
  if (!payload.dns.length) {
    setFormNotice(t("validation.dnsRequired"), "error");
    setHealth(t("validation.dnsRequired"));
    return;
  }

  state.abort = new AbortController();
  state.busy = true;
  setBusy(true);
  setHealth(t("health.resolving"));
  setResultMessage(t("result.loading"), "loading");
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
    setHealth(t("health.done", { count: state.results.length }));
    showToast(t("result.complete"));
  } catch (error) {
    if (error.name === "AbortError") {
      setHealth(t("health.canceled"));
      setResultMessage(t("result.canceled"), "");
    } else {
      const message = localizeDiagnostic(error.message);
      setHealth(message);
      setResultMessage(message, "error");
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
        answerText(item), displayAnswerText(item), locationText(item), operatorText(item),
        item.error, localizeDiagnostic(item.error), (item.warnings || []).join(" "),
        displayWarnings(item.warnings)
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
  const warningText = displayWarnings(item.warnings);
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
        ${metaPill(t("meta.protocol"), protocol)}
        ${metaPill("RCODE", item.rcode)}
        ${metaPill(t("meta.location"), locationText(item))}
        ${metaPill(t("meta.operator"), operatorText(item))}
        ${metaPill(t("meta.warning"), warningText)}
      </span>
      ${item.warnings && item.warnings.length ? `<span class="result-warning-badge">${escapeHTML(t("meta.warning"))} · ${item.warnings.length}</span>` : ""}
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
    $("#detail-title").textContent = t("detail.emptyTitle");
    $("#detail-subtitle").textContent = t("detail.noSelection");
    copyButton.disabled = true;
    root.innerHTML = `<div class="empty-state"><div><h3>${escapeHTML(t("detail.emptyTitle"))}</h3><p>${escapeHTML(t("detail.noSelection"))}</p></div></div>`;
    return;
  }

  const title = displayDomain(item.input || item.domain || t("detail.fallbackTitle"));
  const subtitle = [item.status, item.type, protocolForItem(item)].filter(Boolean).join(" · ");
  const answerValue = displayAnswerText(item) || emptyAnswerText(item);

  $("#detail-title").textContent = title;
  $("#detail-subtitle").textContent = subtitle;
  copyButton.disabled = false;

  root.innerHTML = `
    <div class="detail-stack">
      <section class="detail-panel detail-panel-hero">
        <div class="detail-panel-head">
          <div>
            <p class="eyebrow">${escapeHTML(t("detail.overview"))}</p>
            <h3>${escapeHTML(title)}</h3>
            <p class="detail-panel-subtitle">${escapeHTML(subtitle || t("detail.fallbackTitle"))}</p>
          </div>
        </div>
        <div class="detail-chip-row">
          ${statusBadge(item.status)}
          ${metaPill("TYPE", item.type)}
          ${metaPill(t("detail.protocol"), protocolForItem(item))}
          ${metaPill("RCODE", item.rcode)}
          ${metaPill(t("summary.duration"), `${Number(item.durationMs || 0)} ms`)}
        </div>
        <div class="detail-answer-block">
          <span class="detail-answer-label">${escapeHTML(t("detail.result"))}</span>
          <div class="answer-block ${answerText(item) ? "" : "is-empty"}">${escapeHTML(answerValue)}</div>
        </div>
        ${item.error ? `<div class="detail-alert">${escapeHTML(displayDiagnostic(item.error))}</div>` : ""}
      </section>

      <section class="detail-panel">
        <div class="detail-panel-head">
          <div>
            <p class="eyebrow">${escapeHTML(t("detail.identity"))}</p>
            <h3>${escapeHTML(t("detail.domain"))}</h3>
          </div>
        </div>
        <dl class="detail-list">
          ${detailItem(t("detail.input"), displayDomain(item.input))}
          ${detailItem(t("detail.domain"), displayDomain(item.domain))}
          ${detailItem("ASCII", displayDomain(item.ascii))}
          ${detailItem("FQDN", displayDomain(item.fqdn))}
          ${detailItem("DNS", displayEndpoint(item.resolver))}
          ${detailItem(t("detail.protocol"), protocolForItem(item))}
          ${detailItem("RCODE", item.rcode)}
          ${detailItem(t("summary.duration"), item.durationMs ? `${Number(item.durationMs)} ms` : "")}
        </dl>
      </section>

      ${warningsSection(item.warnings)}
      ${ipInfoSection(item.ipInfo)}
      ${recordSection(t("detail.answerSection"), item.answer, true)}
      ${recordSection(t("detail.authoritySection"), item.authority, false)}
      ${recordSection(t("detail.additionalSection"), item.additional, false)}
      ${rawSection(item)}
    </div>
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
        <thead><tr><th>${escapeHTML(t("detail.recordName"))}</th><th>${escapeHTML(t("detail.recordType"))}</th><th>${escapeHTML(t("detail.recordTTL"))}</th><th>${escapeHTML(t("detail.recordData"))}</th></tr></thead>
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
    <section class="detail-panel">
      <div class="detail-panel-head">
        <div>
          <p class="eyebrow">${escapeHTML(t("detail.warnings"))}</p>
          <h3>${escapeHTML(t("detail.warnings"))}</h3>
        </div>
        <span class="detail-count">${warnings.length}</span>
      </div>
      <ul class="detail-note-list">
        ${warnings.map((warning) => `<li>${escapeHTML(displayDiagnostic(warning))}</li>`).join("")}
      </ul>
    </section>
  `;
}

function ipInfoSection(infos) {
  if (!infos || !infos.length) return "";
  return `
    <section class="detail-panel">
      <div class="detail-panel-head">
        <div>
          <p class="eyebrow">${escapeHTML(t("detail.ipInfo"))}</p>
          <h3>${escapeHTML(t("detail.ipInfo"))}</h3>
        </div>
        <span class="detail-count">${infos.length}</span>
      </div>
      <div class="detail-ip-list">
        ${infos.map((info) => `
          <article class="detail-ip-row">
            <div class="detail-ip-main">
              <strong>${escapeHTML(displayIP(info.ip || ""))}</strong>
              <span>${escapeHTML(locationFromInfo(info) || t("answer.noData"))}</span>
            </div>
            <div class="detail-chip-row detail-chip-row-tight">
              ${metaPill("ASN", info.asn)}
              ${metaPill(t("detail.operator"), info.isp || info.org)}
              ${metaPill(t("detail.source"), info.provider || "")}
              ${info.error ? metaPill(t("meta.warning"), localizeDiagnostic(info.error)) : ""}
            </div>
          </article>
        `).join("")}
      </div>
    </section>
  `;
}

function rawSection(item) {
  return `
    <details class="detail-section">
      <summary>
        <span class="detail-summary-title">${escapeHTML(t("detail.raw"))}</span>
        <span class="detail-summary-meta">JSON</span>
      </summary>
      <pre>${escapeHTML(JSON.stringify(displayResult(item), null, 2))}</pre>
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
  if (item.status === "NO_ANSWER") return t("answer.noAnswer");
  if (item.status === "OK") return t("answer.noRecordData");
  return item.error ? displayDiagnostic(item.error) : t("answer.noData");
}

function locationText(item) {
  return (item.ipInfo || [])
    .map(locationFromInfo)
    .filter(Boolean)
    .join("; ");
}

function operatorText(item) {
  return (item.ipInfo || [])
    .filter((info) => !info.error)
    .map((info) => [info.asn, info.isp || info.org].filter(Boolean).join(" "))
    .filter(Boolean)
    .join("; ");
}

function locationFromInfo(info) {
  if (!info || info.error) return "";
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
  const text = answerText(item);
  return text === item.error ? displayDiagnostic(text) : displaySensitive(text);
}

function displaySensitive(value) {
  if (value === undefined || value === null || value === "") return "";
  return state.privacyMode ? maskSensitiveText(value) : value;
}

function displayDiagnostic(value) {
  if (value === undefined || value === null || value === "") return "";
  return displaySensitive(localizeDiagnostic(value));
}

function displayWarnings(warnings) {
  return (warnings || []).map(displayDiagnostic).filter(Boolean).join("; ");
}

function localizeDiagnostic(value) {
  const text = String(value ?? "");
  if (state.language !== "en" || !text) return text;

  for (const rule of diagnosticTranslations) {
    const match = text.match(rule.pattern);
    if (match) return rule.en(match);
  }
  return text;
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
        <p>${escapeHTML(emptyStateSubtitle())}</p>
      </div>
    </div>
  `;
}

function emptyStateTitle() {
  if (state.busy) return t("empty.busyTitle");
  return state.results.length ? t("empty.noMatchTitle") : t("empty.waitingTitle");
}

function emptyStateSubtitle() {
  if (state.busy) return t("empty.busySubtitle");
  return state.results.length ? t("empty.noMatchSubtitle") : t("empty.readySubtitle");
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
  $("#domain-count").textContent = t("count.domains", { count });
}

function updateEndpointMeta() {
  const count = $$(".endpoint-input").filter((input) => input.value.trim()).length;
  $("#endpoint-count").textContent = t("count.endpoints", { count });
}

function setBusy(busy) {
  state.busy = busy;
  $("#run").disabled = busy;
  $("#run").textContent = busy ? t("actions.resolving") : t("actions.resolve");
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
      setHealth(t("health.copied"));
      showToast(t("result.copied"));
    })
    .catch((error) => {
      setHealth(localizeDiagnostic(error.message));
      showToast(t("result.copyFailed"));
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
    privacyMode: state.privacyMode,
    language: state.language
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
    state.language = normalizeLanguage(data.language);
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
