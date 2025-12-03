(function () {
  const apiBaseInput = document.getElementById("apiBase");
  const tenantHeaderNameInput = document.getElementById("tenantHeaderName");
  const tenantIdInput = document.getElementById("tenantId");
  const memberHeaderNameInput = document.getElementById("memberHeaderName");
  const memberIdInput = document.getElementById("memberId");

  const methodSelect = document.getElementById("method");
  const pathInput = document.getElementById("path");
  const extraHeadersInput = document.getElementById("extraHeaders");
  const bodyInput = document.getElementById("body");
  const sendBtn = document.getElementById("sendBtn");

  const responseStatus = document.getElementById("responseStatus");
  const responseHeaders = document.getElementById("responseHeaders");
  const responseBody = document.getElementById("responseBody");

  const presetHealth = document.getElementById("presetHealth");
  const presetListEvents = document.getElementById("presetListEvents");
  const presetCreateEvent = document.getElementById("presetCreateEvent");
  const presetCreateBusinessDay = document.getElementById("presetCreateBusinessDay");
  const presetListBusinessDays = document.getElementById("presetListBusinessDays");

  // デフォルト値
  apiBaseInput.value = "http://localhost:8082";
  tenantHeaderNameInput.value = "X-Tenant-ID";
  memberHeaderNameInput.value = "X-Member-ID";
  tenantIdInput.value = "01KBHMYWYKRV8PK8EVYGF1SHV0"; // 実際のテナントID
  methodSelect.value = "GET";
  pathInput.value = "/health";

  function parseExtraHeaders(text) {
    const headers = {};
    const lines = text
      .split("\n")
      .map((l) => l.trim())
      .filter((l) => l.length > 0);

    for (const line of lines) {
      const idx = line.indexOf(":");
      if (idx === -1) continue;
      const key = line.slice(0, idx).trim();
      const value = line.slice(idx + 1).trim();
      if (key) {
        headers[key] = value;
      }
    }
    return headers;
  }

  async function sendRequest() {
    const base = apiBaseInput.value.trim().replace(/\/$/, "");
    const path = pathInput.value.trim();
    if (!base || !path) {
      alert("API Base URL と Path を入力してください。");
      return;
    }

    const url = base + path;

    const headers = {};

    const tenantHeaderName = tenantHeaderNameInput.value.trim();
    const tenantId = tenantIdInput.value.trim();
    if (tenantHeaderName && tenantId) {
      headers[tenantHeaderName] = tenantId;
    }

    const memberHeaderName = memberHeaderNameInput.value.trim();
    const memberId = memberIdInput.value.trim();
    if (memberHeaderName && memberId) {
      headers[memberHeaderName] = memberId;
    }

    const extra = parseExtraHeaders(extraHeadersInput.value);
    for (const [k, v] of Object.entries(extra)) {
      headers[k] = v;
    }

    const method = methodSelect.value;
    const init = {
      method,
      headers,
    };

    const bodyText = bodyInput.value.trim();
    if (method !== "GET" && method !== "HEAD" && bodyText.length > 0) {
      init.body = bodyText;
      // Content-Type が未指定なら JSON を想定
      const hasContentType =
        Object.keys(headers).some(
          (k) => k.toLowerCase() === "content-type"
        );
      if (!hasContentType) {
        headers["Content-Type"] = "application/json";
      }
    }

    responseStatus.textContent = "送信中…";
    responseStatus.className = "";
    responseHeaders.textContent = "";
    responseBody.textContent = "";

    try {
      const res = await fetch(url, init);
      const statusText = `${res.status} ${res.statusText}`;
      responseStatus.textContent = statusText;
      responseStatus.className =
        res.ok || (res.status >= 200 && res.status < 300)
          ? "status-ok"
          : "status-error";

      // レスポンスヘッダー
      const headerLines = [];
      res.headers.forEach((value, key) => {
        headerLines.push(`${key}: ${value}`);
      });
      responseHeaders.textContent = headerLines.join("\n");

      // ボディ
      const text = await res.text();
      if (!text) {
        responseBody.textContent = "(empty)";
        return;
      }

      try {
        const json = JSON.parse(text);
        responseBody.textContent = JSON.stringify(json, null, 2);
      } catch (e) {
        // JSON でなければそのまま
        responseBody.textContent = text;
      }
    } catch (err) {
      responseStatus.textContent = "Request error";
      responseStatus.className = "status-error";
      responseBody.textContent = String(err);
    }
  }

  sendBtn.addEventListener("click", (e) => {
    e.preventDefault();
    sendRequest();
  });

  // プリセット: GET /health
  presetHealth.addEventListener("click", () => {
    methodSelect.value = "GET";
    pathInput.value = "/health";
    bodyInput.value = "";
  });

  // プリセット: GET /api/v1/events
  presetListEvents.addEventListener("click", () => {
    methodSelect.value = "GET";
    pathInput.value = "/api/v1/events";
    bodyInput.value = "";
  });

  // プリセット: POST /api/v1/events
  presetCreateEvent.addEventListener("click", () => {
    methodSelect.value = "POST";
    pathInput.value = "/api/v1/events";
    bodyInput.value = JSON.stringify(
      {
        event_name: "VRC定期イベント",
        event_type: "normal",
        description: "毎週開催されるVRCイベント"
      },
      null,
      2
    );
  });

  // プリセット: POST /api/v1/events/:event_id/business-days
  presetCreateBusinessDay.addEventListener("click", () => {
    methodSelect.value = "POST";
    const eventId = prompt("Event ID を入力してください（例: 01KBHN7M8QT9HS5XSJ6QR30WRQ）");
    if (!eventId) return;
    pathInput.value = `/api/v1/events/${eventId}/business-days`;
    
    // 今日から7日後の日付を生成
    const targetDate = new Date();
    targetDate.setDate(targetDate.getDate() + 7);
    const dateStr = targetDate.toISOString().split('T')[0];
    
    bodyInput.value = JSON.stringify(
      {
        target_date: dateStr,
        start_time: "20:00",
        end_time: "23:00",
        occurrence_type: "special"
      },
      null,
      2
    );
  });

  // プリセット: GET /api/v1/events/:event_id/business-days
  presetListBusinessDays.addEventListener("click", () => {
    methodSelect.value = "GET";
    const eventId = prompt("Event ID を入力してください（例: 01KBHN7M8QT9HS5XSJ6QR30WRQ）");
    if (!eventId) return;
    pathInput.value = `/api/v1/events/${eventId}/business-days`;
    bodyInput.value = "";
  });
})();

