import { initPixelSoftUtilityApp } from "./design-system/app-shell.js";
import { bindRegisterWindow } from "./design-system/setup-flow.js";
import { PSU_THEMES, PSU_MODES, setPixelSoftUtilityMode, setPixelSoftUtilityTheme } from "./design-system/theme-controller.js";

const appConfig = JSON.parse(document.querySelector("[data-psu-app-config]").textContent);
initPixelSoftUtilityApp(appConfig);

const providerChoices = [
  "aliyun", "allinkl", "changeip", "cloudflare", "custom", "dd24", "ddnss", "desec",
  "digitalocean", "dnsomatic", "dnspod", "domeneshop", "dondominio", "dreamhost",
  "duckdns", "dyn", "dynu", "dynv6", "easydns", "freedns", "gandi", "gcp",
  "godaddy", "goip", "he", "hetzner", "hetznercloud", "infomaniak", "inwx",
  "ionos", "ipv64", "linode", "loopia", "luadns", "myaddr", "namecheap",
  "name.com", "namesilo", "netcup", "njalla", "noip", "nowdns", "opendns",
  "ovh", "porkbun", "route53", "scaleway", "selfhost.de", "servercow",
  "spaceship", "spdyn", "strato", "variomedia", "vercel", "vultr", "zoneedit"
];

const providerSchemas = {
  aliyun: { required: ["provider", "domain", "access_key_id", "access_secret"], optional: ["ip_version", "ipv6_suffix", "region"] },
  allinkl: { required: ["provider", "domain", "username", "password"], optional: ["ip_version", "ipv6_suffix"] },
  changeip: { required: ["provider", "domain", "username", "password"], optional: ["ip_version", "ipv6_suffix"] },
  cloudflare: { required: ["provider", "domain", "zone_identifier", "ttl", "token"], optional: ["email", "key", "user_service_key", "proxied", "ip_version", "ipv6_suffix"] },
  custom: { required: ["provider", "domain", "url", "ipv4key", "ipv6key", "success_regex"], optional: ["ip_version", "ipv6_suffix"] },
  dd24: { required: ["provider", "domain", "password"], optional: ["ip_version", "ipv6_suffix"] },
  ddnss: { required: ["provider", "domain", "username", "password"], optional: ["dual_stack", "ip_version", "ipv6_suffix"] },
  desec: { required: ["provider", "domain", "token"], optional: ["ip_version", "ipv6_suffix"] },
  digitalocean: { required: ["provider", "domain", "token"], optional: ["ip_version", "ipv6_suffix"] },
  dnsomatic: { required: ["provider", "domain", "username", "password"], optional: ["ip_version", "ipv6_suffix"] },
  dnspod: { required: ["provider", "domain", "token"], optional: ["ip_version", "ipv6_suffix"] },
  domeneshop: { required: ["provider", "domain", "token", "secret"], optional: ["ip_version", "ipv6_suffix"] },
  dondominio: { required: ["provider", "domain", "username", "password"], optional: ["ip_version", "ipv6_suffix"] },
  dreamhost: { required: ["provider", "domain", "key"], optional: ["ip_version", "ipv6_suffix"] },
  duckdns: { required: ["provider", "domain", "token"], optional: ["ip_version", "ipv6_suffix"] },
  dyn: { required: ["provider", "domain", "username", "client_key"], optional: ["password", "ip_version", "ipv6_suffix"] },
  dynu: { required: ["provider", "domain", "username", "password"], optional: ["group", "ip_version", "ipv6_suffix"] },
  dynv6: { required: ["provider", "domain", "token"], optional: ["ip_version", "ipv6_suffix"] },
  easydns: { required: ["provider", "domain", "username", "token"], optional: ["ip_version", "ipv6_suffix"] },
  example: { required: ["provider", "domain", "username", "password"], optional: ["ip_version", "ipv6_suffix"] },
  freedns: { required: ["provider", "domain", "token"], optional: ["ip_version", "ipv6_suffix"] },
  gandi: { required: ["provider", "domain", "personal_access_token"], optional: ["ttl", "ip_version", "ipv6_suffix"] },
  gcp: { required: ["provider", "domain", "project", "zone", "credentials"], optional: ["ip_version", "ipv6_suffix"] },
  godaddy: { required: ["provider", "domain", "key", "secret"], optional: ["ip_version", "ipv6_suffix"] },
  goip: { required: ["provider", "domain", "username", "password"], optional: ["ip_version", "ipv6_suffix"] },
  he: { required: ["provider", "domain", "password"], optional: ["ip_version", "ipv6_suffix"] },
  hetzner: { required: ["provider", "domain", "zone_identifier", "token"], optional: ["ttl", "ip_version", "ipv6_suffix"] },
  hetznercloud: { required: ["provider", "domain", "token"], optional: ["ttl", "ip_version", "ipv6_suffix"] },
  infomaniak: { required: ["provider", "domain", "username", "password"], optional: ["ip_version", "ipv6_suffix"] },
  inwx: { required: ["provider", "domain", "username", "password"], optional: ["ip_version", "ipv6_suffix"] },
  ionos: { required: ["provider", "domain", "api_key"], optional: ["ip_version", "ipv6_suffix"] },
  ipv64: { required: ["provider", "domain", "key"], optional: ["ip_version", "ipv6_suffix"] },
  linode: { required: ["provider", "domain", "token"], optional: ["ip_version", "ipv6_suffix"] },
  loopia: { required: ["provider", "domain", "username", "password"], optional: ["ip_version", "ipv6_suffix"] },
  luadns: { required: ["provider", "domain", "email", "token"], optional: ["ip_version", "ipv6_suffix"] },
  myaddr: { required: ["provider", "domain", "key"], optional: ["ip_version", "ipv6_suffix"] },
  "name.com": { required: ["provider", "domain", "username", "token"], optional: ["ttl", "ip_version", "ipv6_suffix"] },
  namecheap: { required: ["provider", "domain", "password"], optional: [] },
  namesilo: { required: ["provider", "domain", "key"], optional: ["ttl", "ip_version", "ipv6_suffix"] },
  netcup: { required: ["provider", "domain", "api_key", "password", "customer_number"], optional: ["ip_version", "ipv6_suffix"] },
  njalla: { required: ["provider", "domain", "key"], optional: ["ip_version", "ipv6_suffix"] },
  noip: { required: ["provider", "domain", "username", "password"], optional: ["ip_version", "ipv6_suffix"] },
  nowdns: { required: ["provider", "domain", "username", "password"], optional: ["ip_version", "ipv6_suffix"] },
  opendns: { required: ["provider", "domain", "username", "password"], optional: ["ip_version", "ipv6_suffix"] },
  ovh: { required: ["provider", "domain", "username", "password"], optional: ["mode", "api_endpoint", "app_key", "app_secret", "consumer_key", "ip_version", "ipv6_suffix"] },
  porkbun: { required: ["provider", "domain", "api_key", "secret_api_key"], optional: ["ttl", "ip_version", "ipv6_suffix"] },
  route53: { required: ["provider", "domain", "access_key", "secret_key", "zone_id"], optional: ["ttl", "ip_version", "ipv6_suffix"] },
  scaleway: { required: ["provider", "domain", "secret_key"], optional: ["ttl", "ip_version", "ipv6_suffix"] },
  "selfhost.de": { required: ["provider", "domain", "username", "password"], optional: ["ip_version", "ipv6_suffix"] },
  servercow: { required: ["provider", "domain", "username", "password"], optional: ["ttl", "ip_version", "ipv6_suffix"] },
  spaceship: { required: ["provider", "domain", "api_key", "api_secret"], optional: ["ttl", "ip_version", "ipv6_suffix"] },
  spdyn: { required: ["provider", "domain", "token"], optional: ["user", "password", "ip_version", "ipv6_suffix"] },
  strato: { required: ["provider", "domain", "password"], optional: ["ip_version", "ipv6_suffix"] },
  variomedia: { required: ["provider", "domain", "email", "password"], optional: ["ip_version", "ipv6_suffix"] },
  vercel: { required: ["provider", "domain", "token"], optional: ["team_id", "ttl", "ip_version", "ipv6_suffix"] },
  vultr: { required: ["provider", "domain", "apikey"], optional: ["ttl", "ip_version", "ipv6_suffix"] },
  zoneedit: { required: ["provider", "domain", "username", "token"], optional: ["ip_version", "ipv6_suffix"] }
};

const secretKeys = ["api_key", "apikey", "access_key", "access_secret", "key", "password", "secret", "token", "credentials", "consumer_key"];
const themeLabels = {
  lavender: "Lavender",
  mint: "Mint",
  sky: "Sky",
  amber: "Amber",
  rose: "Rose",
  graphite: "Graphite"
};

let loadedConfig = { settings: [] };
let savingDisabled = false;
let editingSettingIndex = null;
let lastStatus = { records: [], history_log: [] };
let currentUser = null;
let adminInfo = null;
let refreshTimer = null;

const $ = (selector) => document.querySelector(selector);

document.addEventListener("DOMContentLoaded", () => {
  bindStaticActions();
  bindAuthActions();
  buildAppearanceControls();
  bootstrapAuth();
});

function bindStaticActions() {
  $(".view-tabs").addEventListener("click", (event) => {
    const button = event.target.closest("[data-view]");
    if (!button) return;
    showView(button.dataset.view);
  });

  $("#refresh-button").addEventListener("click", loadAll);
  $("#force-button").addEventListener("click", forceUpdate);
  $("#add-entry-button").addEventListener("click", () => addSetting(newSettingForProvider("netcup")));
  $("#save-config-button").addEventListener("click", saveConfig);
  $("#record-search").addEventListener("input", () => renderRecords(lastStatus.records || []));
  $("#copy-debug-button").addEventListener("click", copyDebugDetails);
}

function bindAuthActions() {
  bindRegisterWindow($("#register-form"), {
    appId: appConfig.app_id,
    appName: appConfig.app_name,
    async onSubmit(formData, form) {
      await submitSetup(formData, form);
    }
  });

  $("#login-form").addEventListener("submit", async (event) => {
    event.preventDefault();
    const data = new FormData(event.currentTarget);
    await submitLogin({
      username: data.get("username"),
      password: data.get("password")
    });
  });

  $("#logout-button").addEventListener("click", logout);
}

async function bootstrapAuth() {
  try {
    const state = await fetchJSON("api/auth/state");
    applyAuthState(state);
  } catch (error) {
    showAuthGate("error");
    $("#setup-error-message").textContent = `Authentication status could not be loaded: ${error.message}`;
  }
}

function applyAuthState(state) {
  currentUser = state.user || null;
  adminInfo = state.admin_info || null;
  renderUser();
  renderAdminInfo();

  if (state.setup_required) {
    if (state.setup_configured) {
      showAuthGate("setup");
    } else {
      showAuthGate("error");
      $("#setup-error-message").textContent = state.message || "Setup secret is missing.";
    }
    return;
  }

  if (!state.authenticated) {
    showAuthGate("login");
    return;
  }

  showApp();
}

function showAuthGate(mode) {
  stopRefreshTimer();
  $("[data-app-content]").hidden = true;
  $("#auth-gate").hidden = false;
  $("#setup-error-window").hidden = mode !== "error";
  $("#register-form").hidden = mode !== "setup";
  $("#login-form").hidden = mode !== "login";

  const focusTarget = {
    setup: "#register-form [name='setup_secret']",
    login: "#login-form [name='username']"
  }[mode];
  if (focusTarget) requestAnimationFrame(() => $(focusTarget)?.focus());
}

function showApp() {
  $("#auth-gate").hidden = true;
  $("[data-app-content]").hidden = false;
  loadAll();
  startRefreshTimer();
}

function startRefreshTimer() {
  if (refreshTimer) return;
  refreshTimer = window.setInterval(loadStatus, 30000);
}

function stopRefreshTimer() {
  if (!refreshTimer) return;
  window.clearInterval(refreshTimer);
  refreshTimer = null;
}

async function submitSetup(formData, form) {
  const button = form.querySelector("button[type='submit']");
  button.disabled = true;
  try {
    const result = await fetchJSON("api/setup", {
      method: "POST",
      headers: { "Content-Type": "application/json", Accept: "application/json" },
      body: JSON.stringify({
        setup_secret: formData.get("setup_secret"),
        admin_display_name: formData.get("admin_display_name"),
        admin_username: formData.get("admin_username"),
        admin_email: formData.get("admin_email"),
        admin_password: formData.get("admin_password"),
        admin_password_confirm: formData.get("admin_password_confirm")
      })
    });
    currentUser = result.user || null;
    showToast("Administrator account created");
    await bootstrapAuth();
  } catch (error) {
    showToast(error.message);
  } finally {
    button.disabled = false;
  }
}

async function submitLogin(credentials) {
  $("#login-error").hidden = true;
  try {
    const result = await fetchJSON("api/login", {
      method: "POST",
      headers: { "Content-Type": "application/json", Accept: "application/json" },
      body: JSON.stringify(credentials)
    });
    currentUser = result.user || null;
    showToast("Angemeldet");
    await bootstrapAuth();
  } catch (error) {
    $("#login-error").hidden = false;
    $("#login-error").textContent = error.message;
  }
}

async function logout() {
  try {
    await fetchJSON("api/logout", { method: "POST" });
  } finally {
    currentUser = null;
    adminInfo = null;
    document.querySelectorAll(".psu-backdrop:not([hidden])").forEach((element) => element.setAttribute("hidden", ""));
    showAuthGate("login");
  }
}

function buildAppearanceControls() {
  const themePicker = $("#theme-picker");
  themePicker.innerHTML = "";
  for (const theme of PSU_THEMES) {
    const button = document.createElement("button");
    button.type = "button";
    button.className = "theme-button";
    button.dataset.theme = theme;
    button.dataset.themeChoice = theme;
    button.textContent = themeLabels[theme] || theme;
    button.addEventListener("click", () => {
      setPixelSoftUtilityTheme(theme);
      syncAppearanceControls();
    });
    themePicker.append(button);
  }

  $("#mode-picker").addEventListener("click", (event) => {
    const button = event.target.closest("[data-mode-choice]");
    if (!button || !PSU_MODES.includes(button.dataset.mode)) return;
    setPixelSoftUtilityMode(button.dataset.mode);
    syncAppearanceControls();
  });
  window.addEventListener("psu:themechange", syncAppearanceControls);
  syncAppearanceControls();
}

function syncAppearanceControls() {
  const root = document.documentElement;
  document.querySelectorAll("[data-theme-choice]").forEach((button) => {
    button.setAttribute("aria-pressed", String(button.dataset.themeChoice === root.dataset.theme));
  });
  document.querySelectorAll("[data-mode-choice]").forEach((button) => {
    button.setAttribute("aria-selected", String(button.dataset.modeChoice === root.dataset.mode));
  });
}

async function loadAll() {
  await Promise.allSettled([loadStatus(), loadConfig()]);
}

async function loadStatus() {
  try {
    lastStatus = await fetchJSON("api/status", { headers: { Accept: "application/json" } });
    renderStatus(lastStatus);
    $("#runtime-info").textContent = "API: verbunden";
  } catch (error) {
    if (await handleAuthError(error)) return;
    showToast("Could not load status");
    $("#runtime-info").textContent = `API: ${error.message}`;
  }
}

async function loadConfig() {
  try {
    const result = await fetchJSON("api/config", { headers: { Accept: "application/json" } });
    loadedConfig = result.config || { settings: [] };
    if (!Array.isArray(loadedConfig.settings)) loadedConfig.settings = [];
    for (const setting of loadedConfig.settings) {
      if (!setting.provider) setting.provider = "netcup";
    }
    editingSettingIndex = null;
    savingDisabled = Boolean(result.env_config);
    $("#config-path").textContent = result.path || "";
    $("#save-config-button").disabled = savingDisabled;
    renderConfigNotice(result);
    renderSettings();
  } catch (error) {
    if (await handleAuthError(error)) return;
    showConfigWarning(`Could not load config: ${error.message}`);
  }
}

function renderStatus(status) {
  $("#current-ip").textContent = status.current_ip || "N/A";
  $("#current-since").textContent = status.current_since
    ? `Since ${formatDate(status.current_since)} (${durationSince(status.current_since)})`
    : "Waiting for the first successful update";

  const records = status.records || [];
  const healthy = records.filter((record) => ["success", "up to date"].includes(String(record.status).toLowerCase())).length;
  $("#record-count").textContent = String(records.length);
  $("#healthy-count").textContent = String(healthy);
  $("#attention-count").textContent = String(Math.max(records.length - healthy, 0));

  renderRecords(records);
  renderHistory(status.history_log || []);
}

function renderRecords(records) {
  const query = $("#record-search").value.trim().toLowerCase();
  const filtered = query
    ? records.filter((record) => [record.owner, record.domain, record.provider, record.ip_version, record.status, record.current_ip]
      .some((value) => String(value || "").toLowerCase().includes(query)))
    : records;

  $("#records-empty").hidden = filtered.length > 0;
  $("#records-table-wrap").hidden = filtered.length === 0;

  const body = $("#records-body");
  body.innerHTML = "";
  filtered.forEach((record, index) => {
    const row = document.createElement("tr");
    const statusName = record.status || "unset";
    row.innerHTML = `
      <td data-label="Owner">${escapeHTML(record.owner || "@")}</td>
      <td data-label="Domain" class="domain-cell">${escapeHTML(record.domain || "")}</td>
      <td data-label="Provider">${escapeHTML(record.provider || "")}</td>
      <td data-label="IP Version">${escapeHTML(record.ip_version || "")}</td>
      <td data-label="Status"><span class="status-pill ${statusClass(statusName)}">${escapeHTML(statusName)}</span></td>
      <td data-label="Current IP" class="mono-cell">${escapeHTML(record.current_ip || "N/A")}</td>
      <td data-label="Since">${record.since ? escapeHTML(durationSince(record.since)) : "N/A"}</td>
      <td data-label="Log"><button class="psu-button psu-button--outlined log-toggle" type="button" aria-expanded="false" aria-controls="record-log-${index}">Log</button></td>
    `;
    body.append(row);

    const logRow = document.createElement("tr");
    logRow.className = "record-log-row hidden";
    logRow.id = `record-log-${index}`;
    logRow.innerHTML = `<td colspan="8">${renderRecordLog(record.history || [])}</td>`;
    body.append(logRow);

    row.querySelector(".log-toggle").addEventListener("click", (event) => {
      const expanded = event.currentTarget.getAttribute("aria-expanded") === "true";
      event.currentTarget.setAttribute("aria-expanded", String(!expanded));
      logRow.classList.toggle("hidden", expanded);
    });
  });
}

function renderHistory(lines) {
  $("#history-count").textContent = String(lines.length);
  const list = $("#history-list");
  list.innerHTML = "";
  if (lines.length === 0) {
    list.innerHTML = `<div class="empty-state"><img src="static/dyniku-logo.png" alt=""><strong>No IP changes yet</strong><span>Dyniku will list public IP changes here after they are logged.</span></div>`;
    return;
  }

  const newestFirst = [...lines].reverse();
  newestFirst.forEach((line, index) => {
    const current = parseLogLine(line);
    const previous = parseLogLine(newestFirst[index + 1]);
    const duration = previous && current ? formatDuration(current.date - previous.date) : "current";
    const item = document.createElement("div");
    item.className = "history-item";
    item.innerHTML = `
      <span>${escapeHTML(current?.stamp || line)}</span>
      <strong>${escapeHTML(current?.ip || "")}</strong>
      <span>${escapeHTML(duration)}</span>
    `;
    list.append(item);
  });
}

function renderConfigNotice(result) {
  if (savingDisabled) {
    showConfigWarning("CONFIG environment variable is active. File editing is disabled because it would be overwritten on restart.");
    return;
  }
  const warningText = (result.warnings || []).join(" ");
  const notice = [result.restart_hint, warningText].filter(Boolean).join(" ");
  if (notice) showConfigWarning(notice);
  else $("#config-warning").hidden = true;
}

function renderSettings() {
  const list = $("#settings-list");
  list.innerHTML = "";

  if (loadedConfig.settings.length === 0) {
    list.innerHTML = `<div class="empty-state"><img src="static/dyniku-logo.png" alt=""><strong>No provider entries</strong><span>Add an entry and save the config to begin updating DNS records.</span></div>`;
    return;
  }

  loadedConfig.settings.forEach((setting, index) => {
    if (editingSettingIndex === index) {
      list.append(makeSettingEditor(setting, index));
      return;
    }
    list.append(makeSettingSummary(setting, index));
  });
  renderDuplicateWarnings();
}

function makeSettingSummary(setting, index) {
  const node = document.importNode($("#setting-summary-template").content, true);
  const article = node.querySelector(".config-card");
  node.querySelector(".provider-mark").textContent = providerInitials(setting.provider);
  node.querySelector(".config-title").textContent = setting.domain || `Entry ${index + 1}`;
  node.querySelector(".config-meta").textContent = `${setting.provider || "unknown"} · ${setting.owner || setting.host || "@"} · ${setting.ip_version || "default IP"}`;
  node.querySelector(".duplicate-entry").addEventListener("click", () => duplicateSetting(index));
  node.querySelector(".edit-entry").addEventListener("click", () => {
    editingSettingIndex = index;
    renderSettings();
  });
  return article;
}

function makeSettingEditor(setting, index) {
  const node = document.importNode($("#setting-edit-template").content, true);
  const article = node.querySelector(".config-card");
  node.querySelector(".config-title").textContent = setting.domain || `Entry ${index + 1}`;
  node.querySelector(".config-meta").textContent = setting.provider || "unknown provider";
  node.querySelector(".close-entry").addEventListener("click", () => {
    editingSettingIndex = null;
    renderSettings();
  });
  node.querySelector(".add-field").addEventListener("click", () => {
    setting[""] = "";
    renderSettings();
  });
  node.querySelector(".duplicate-entry").addEventListener("click", () => duplicateSetting(index));
  node.querySelector(".delete-entry").addEventListener("click", () => {
    loadedConfig.settings.splice(index, 1);
    editingSettingIndex = null;
    renderSettings();
  });

  const fields = node.querySelector(".fields");
  for (const key of orderedConfigKeys(setting)) {
    fields.append(makeField(setting, key));
  }

  const missingOptional = schemaFields(setting.provider).optional.filter((key) => !(key in setting));
  if (missingOptional.length > 0) {
    const addOptional = document.createElement("button");
    addOptional.type = "button";
    addOptional.className = "psu-button psu-button--tonal";
    addOptional.textContent = "Add optional fields";
    addOptional.addEventListener("click", () => {
      for (const key of missingOptional) setting[key] = defaultValueForField(key);
      renderSettings();
    });
    fields.append(addOptional);
  }

  return article;
}

function duplicateSetting(index) {
  const original = loadedConfig.settings[index];
  if (!original) return;
  const clone = JSON.parse(JSON.stringify(original));
  loadedConfig.settings.splice(index + 1, 0, clone);
  editingSettingIndex = index + 1;
  renderSettings();
  showView("config");
  showToast("Entry duplicated");
}

function makeField(setting, key) {
  const node = document.importNode($("#field-template").content, true);
  const row = node.querySelector(".field-row");
  const keyInput = node.querySelector(".field-key");
  const valueInput = node.querySelector(".field-value");
  let currentKey = key;
  const schema = schemaFields(setting.provider);
  const badge = node.querySelector(".field-badge");
  badge.textContent = schema.required.includes(key) ? "required" : schema.optional.includes(key) ? "optional" : "custom";
  keyInput.value = key;
  valueInput.value = setting[key] ?? "";
  valueInput.type = isSecretKey(key) ? "password" : "text";

  if (key === "provider") {
    const select = document.createElement("select");
    select.className = "psu-input field-value";
    for (const provider of providerChoices) {
      const option = document.createElement("option");
      option.value = provider;
      option.textContent = provider;
      select.append(option);
    }
    select.value = setting[key] || "netcup";
    valueInput.replaceWith(select);
    select.addEventListener("change", () => {
      setting[currentKey] = select.value;
      pruneSettingForProvider(setting, select.value);
      addProviderDefaults(setting, select.value);
      renderSettings();
    });
  } else if (key === "ip_version") {
    replaceValueWithSelect(valueInput, ["ipv4", "ipv6", "ipv4 or ipv6"], setting[key] || "ipv4", (value) => {
      setting[currentKey] = value;
    });
  } else if (typeof setting[key] === "boolean") {
    replaceValueWithSelect(valueInput, ["false", "true"], String(setting[key]), (value) => {
      setting[currentKey] = value === "true";
    });
  } else {
    valueInput.addEventListener("input", () => {
      setting[currentKey] = coerceFieldValue(currentKey, valueInput.value);
      renderDuplicateWarnings();
    });
  }

  keyInput.addEventListener("input", () => {
    const newKey = keyInput.value;
    if (currentKey === newKey) return;
    setting[newKey] = setting[currentKey];
    delete setting[currentKey];
    currentKey = newKey;
    badge.textContent = "custom";
    renderDuplicateWarnings();
  });

  row.querySelector(".delete-field").addEventListener("click", () => {
    delete setting[currentKey];
    renderSettings();
  });

  return row;
}

function renderDuplicateWarnings() {
  document.querySelectorAll(".duplicate-domain-warning").forEach((warning) => {
    const card = warning.closest(".config-card");
    const index = Array.from(document.querySelectorAll(".config-card")).indexOf(card);
    const setting = loadedConfig.settings[index];
    const duplicates = duplicateDomainTargets(setting, index);
    if (duplicates.length === 0) {
      warning.hidden = true;
      warning.textContent = "";
      return;
    }
    warning.hidden = false;
    warning.textContent = `Warning: This domain/subdomain entry already exists in entry ${duplicates.join(", ")}. You can still save it.`;
  });
}

function duplicateDomainTargets(setting, currentIndex) {
  const current = domainTarget(setting);
  if (!current) return [];
  return loadedConfig.settings
    .map((candidate, index) => ({ index, target: domainTarget(candidate) }))
    .filter((candidate) => candidate.index !== currentIndex && candidate.target === current)
    .map((candidate) => String(candidate.index + 1));
}

function domainTarget(setting) {
  if (!setting) return "";
  const domain = String(setting.domain || "").trim().toLowerCase();
  if (!domain) return "";
  const owner = String(setting.owner || setting.host || "").trim().toLowerCase();
  if (!owner || owner === "@") return domain;
  if (domain === owner || domain.startsWith(`${owner}.`)) return domain;
  return `${owner}.${domain}`;
}

function replaceValueWithSelect(input, values, selectedValue, onChange) {
  const select = document.createElement("select");
  select.className = "psu-input field-value";
  for (const value of values) {
    const option = document.createElement("option");
    option.value = value;
    option.textContent = value;
    select.append(option);
  }
  select.value = selectedValue;
  input.replaceWith(select);
  select.addEventListener("change", () => onChange(select.value));
}

function addSetting(setting) {
  loadedConfig.settings.push(setting);
  editingSettingIndex = loadedConfig.settings.length - 1;
  renderSettings();
  showView("config");
}

function newSettingForProvider(provider) {
  const setting = { provider };
  addProviderDefaults(setting, provider);
  return setting;
}

function addProviderDefaults(setting, provider) {
  const schema = schemaFields(provider);
  for (const key of schema.required) {
    if (!(key in setting)) setting[key] = defaultValueForField(key, provider);
  }
  if (!("ip_version" in setting) && schema.optional.includes("ip_version")) {
    setting.ip_version = "ipv4";
  }
}

function pruneSettingForProvider(setting, provider) {
  const schema = schemaFields(provider);
  const allowed = new Set([...schema.required, ...schema.optional, "owner", "host"]);
  for (const key of Object.keys(setting)) {
    if (!allowed.has(key)) delete setting[key];
  }
  setting.provider = provider;
}

function defaultValueForField(key, provider) {
  switch (key) {
    case "provider":
      return provider || "netcup";
    case "ip_version":
      return "ipv4";
    case "ttl":
      return 600;
    case "proxied":
    case "dual_stack":
      return false;
    default:
      return "";
  }
}

async function saveConfig() {
  if (savingDisabled) return;
  $("#save-config-button").disabled = true;
  try {
    const result = await fetchJSON("api/config", {
      method: "PUT",
      headers: { "Content-Type": "application/json", Accept: "application/json" },
      body: JSON.stringify(loadedConfig)
    });
    loadedConfig = result.config || loadedConfig;
    editingSettingIndex = null;
    renderSettings();
    renderConfigNotice(result);
    showToast("Config saved");
  } catch (error) {
    if (await handleAuthError(error)) return;
    showConfigWarning(error.message);
  } finally {
    $("#save-config-button").disabled = savingDisabled;
  }
}

async function forceUpdate() {
  $("#force-button").disabled = true;
  try {
    const response = await fetch("update", { method: "POST" });
    if (!response.ok) throw await makeFetchError(response);
    await loadStatus();
    showToast("Update triggered");
  } catch (error) {
    if (await handleAuthError(error)) return;
    showToast(`Update failed: ${error.message}`);
  } finally {
    $("#force-button").disabled = false;
  }
}

async function fetchJSON(url, options = {}) {
  const response = await fetch(url, {
    ...options,
    headers: { Accept: "application/json", ...(options.headers || {}) }
  });
  if (!response.ok) throw await makeFetchError(response);
  if (response.status === 204) return null;
  return response.json();
}

async function makeFetchError(response) {
  let message = response.statusText;
  try {
    const data = await response.clone().json();
    message = data.error || (Array.isArray(data.errors) ? data.errors.join(" ") : message);
  } catch {
    message = await response.text();
  }
  const error = new Error(message || response.statusText);
  error.status = response.status;
  return error;
}

async function handleAuthError(error) {
  if (![401, 423].includes(error.status)) return false;
  await bootstrapAuth();
  return true;
}

function renderUser() {
  const initials = currentUser?.initials || "D";
  const displayName = currentUser?.display_name || "Dyniku Admin";
  const username = currentUser?.username || "admin";
  $("#avatar-initials").textContent = initials;
  $("#profile-avatar").textContent = initials;
  $("#profile-name").textContent = displayName;
  $("#profile-username").textContent = `@${username}`;
}

function renderAdminInfo() {
  const info = adminInfo || {};
  $("#admin-version").textContent = info.app_version || "unknown";
  $("#admin-build-date").textContent = info.build_date || "unknown";
  $("#admin-sha").textContent = info.github_sha || "unknown";
  $("#admin-data-dir").textContent = info.data_directory || "unknown";
  $("#admin-config-path").textContent = info.config_path || "unknown";
  $("#admin-database").textContent = info.database_status || "unknown";
  $("#admin-setup").textContent = info.setup_state || "unknown";
  $("#admin-health").textContent = info.health_status || "unknown";
  $("#admin-log-level").textContent = info.log_level || "unknown";
}

async function copyDebugDetails() {
  const info = adminInfo || {};
  const lines = [
    `App Version: ${info.app_version || "unknown"}`,
    `Build Date: ${info.build_date || "unknown"}`,
    `GitHub SHA: ${info.github_sha || "unknown"}`,
    `Data Directory: ${info.data_directory || "unknown"}`,
    `Config Path: ${info.config_path || "unknown"}`,
    `Database: ${info.database_status || "unknown"}`,
    `Setup: ${info.setup_state || "unknown"}`,
    `Health: ${info.health_status || "unknown"}`,
    `Log Level: ${info.log_level || "unknown"}`
  ].join("\n");
  await navigator.clipboard.writeText(lines);
  showToast("Debug details copied");
}

function showView(name) {
  document.querySelectorAll(".view-tabs [data-view]").forEach((tab) => {
    tab.setAttribute("aria-selected", String(tab.dataset.view === name));
  });
  document.querySelectorAll(".view").forEach((view) => view.classList.remove("active"));
  $(`#${name}-view`).classList.add("active");
}

function orderedConfigKeys(setting) {
  const schema = schemaFields(setting.provider);
  const known = [...schema.required, ...schema.optional];
  const keys = new Set(Object.keys(setting));
  const ordered = known.filter((key) => keys.has(key));
  const extra = [...keys].filter((key) => !known.includes(key)).sort((a, b) => a.localeCompare(b));
  return [...ordered, ...extra];
}

function schemaFields(provider) {
  return providerSchemas[provider] || { required: ["provider", "domain"], optional: ["ip_version", "ipv6_suffix"] };
}

function providerInitials(provider = "") {
  const clean = provider.replace(/[^a-z0-9]+/gi, " ").trim();
  if (!clean) return "DD";
  return clean.split(/\s+/).map((part) => part[0]).join("").slice(0, 2);
}

function isSecretKey(key) {
  return secretKeys.some((secret) => key.toLowerCase().includes(secret));
}

function coerceFieldValue(key, value) {
  if (["ttl", "customer_number"].includes(key) && value !== "" && Number.isFinite(Number(value))) {
    return Number(value);
  }
  return value;
}

function statusClass(status) {
  return String(status).toLowerCase().replace(/[^a-z0-9]+/g, "-").replace(/^-|-$/g, "");
}

function renderRecordLog(events) {
  if (events.length === 0) {
    return `<div class="record-log"><div class="empty-state"><strong>No updates logged</strong><span>This record has no successful IP events yet.</span></div></div>`;
  }
  return `<div class="record-log">${events.map((event) => `
    <div class="record-log-item">
      <span>${escapeHTML(formatDate(event.time))}</span>
      <strong class="mono-cell">${escapeHTML(event.ip || "")}</strong>
    </div>
  `).join("")}</div>`;
}

function parseLogLine(line) {
  if (!line) return null;
  const [stamp, ip] = line.trim().split(/\s+/);
  const match = /^(\d{4})(\d{2})(\d{2})-(\d{2})(\d{2})$/.exec(stamp || "");
  if (!match) return null;
  const [, year, month, day, hour, minute] = match;
  return {
    stamp,
    ip,
    date: new Date(Number(year), Number(month) - 1, Number(day), Number(hour), Number(minute))
  };
}

function formatDate(value) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "N/A";
  return date.toLocaleString();
}

function durationSince(value) {
  return formatDuration(Date.now() - new Date(value).getTime());
}

function formatDuration(ms) {
  if (!Number.isFinite(ms) || ms < 0) return "N/A";
  const minutes = Math.floor(ms / 60000);
  if (minutes < 1) return "less than 1 min";
  if (minutes < 60) return `${minutes} min`;
  const hours = Math.floor(minutes / 60);
  if (hours < 48) return `${hours} h`;
  return `${Math.floor(hours / 24)} d`;
}

function showConfigWarning(message) {
  const warning = $("#config-warning");
  warning.hidden = false;
  warning.textContent = message;
}

function showToast(message) {
  const host = $("#psu-toast-host");
  host.innerHTML = "";
  const toast = document.createElement("div");
  toast.className = "psu-toast";
  toast.textContent = message;
  host.append(toast);
  window.setTimeout(() => {
    if (toast.isConnected) toast.remove();
  }, 3000);
}

function escapeHTML(value) {
  return String(value).replace(/[&<>"']/g, (char) => ({
    "&": "&amp;",
    "<": "&lt;",
    ">": "&gt;",
    "\"": "&quot;",
    "'": "&#39;"
  }[char]));
}
