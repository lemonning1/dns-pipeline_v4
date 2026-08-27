const DEFAULT_PAGE_SIZE = 20;

let currentPage = 1;
let totalPages = 1;

window.onload = function () {
    loadLatest();
};

function getQRFilter() {
    return document.getElementById("qrFilter").value;
}

function buildAPIUrl() {
    const params = new URLSearchParams();
    const domain = document.getElementById("domainInput").value.trim();
    const qr = getQRFilter();

    if (domain !== "") {
        params.set("domain", domain);
    }
    if (qr !== "all") {
        params.set("qr", qr);
    }
    params.set("page", currentPage);
    params.set("page_size", DEFAULT_PAGE_SIZE);

    return "/api/dns?" + params.toString();
}

function searchDomain() {
    currentPage = 1;
    fetchData(buildAPIUrl());
}

function loadLatest() {
    document.getElementById("domainInput").value = "";
    document.getElementById("qrFilter").value = "all";
    currentPage = 1;
    fetchData(buildAPIUrl());
}

function onFilterChange() {
    currentPage = 1;
    fetchData(buildAPIUrl());
}

function prevPage() {
    if (currentPage > 1) {
        currentPage--;
        fetchData(buildAPIUrl());
    }
}

function nextPage() {
    if (currentPage < totalPages) {
        currentPage++;
        fetchData(buildAPIUrl());
    }
}

function goToPage() {
    const raw = document.getElementById("pageJump").value;
    let page = parseInt(raw, 10);
    if (Number.isNaN(page) || page < 1) {
        page = 1;
    }
    if (page > totalPages) {
        page = totalPages;
    }
    currentPage = page;
    fetchData(buildAPIUrl());
}

window.addEventListener("DOMContentLoaded", function () {
    const input = document.getElementById("pageJump");
    if (!input) {
        return;
    }
    input.addEventListener("keydown", function (e) {
        if (e.key === "Enter") {
            e.preventDefault();
            goToPage();
        }
    });
});

function fetchData(url) {
    const tbody = document.getElementById("resultBody");
    tbody.innerHTML =
        '<tr><td colspan="8" class="empty-row">加载中...</td></tr>';

    fetch(url)
        .then((res) => {
            if (!res.ok) {
                return res.json().then((data) => {
                    throw new Error(data.error || "HTTP " + res.status);
                }).catch((e) => {
                    if (e.message && e.message.indexOf("HTTP") === 0) {
                        throw e;
                    }
                    throw new Error(e.message || "HTTP " + res.status);
                });
            }
            return res.json();
        })
        .then((data) => {
            if (!data || !Array.isArray(data.items)) {
                throw new Error("返回数据格式错误");
            }
            renderTable(data.items);
            updatePagination(data.page, data.total, data.page_size);
        })
        .catch((err) => {
            tbody.innerHTML =
                '<tr><td colspan="8" class="empty-row status-error">请求失败: ' +
                escapeHtml(err.message) +
                "</td></tr>";
            updatePagination(1, 0, DEFAULT_PAGE_SIZE);
        });
}

function updatePagination(page, total, pageSize) {
    totalPages = Math.max(1, Math.ceil(total / pageSize));
    currentPage = page;

    document.getElementById("pageInfo").textContent =
        "第 " + page + " / " + totalPages + " 页，共 " + total + " 条";

    document.getElementById("prevPage").disabled = page <= 1;
    document.getElementById("nextPage").disabled = page >= totalPages || total === 0;

    const jump = document.getElementById("pageJump");
    jump.max = totalPages;
    jump.value = page;
}

function renderTable(data) {
    const tbody = document.getElementById("resultBody");
    if (data.length === 0) {
        tbody.innerHTML =
            '<tr class="empty-row"><td colspan="8">暂无 DNS 记录</td></tr>';
        return;
    }

    let html = "";
    data.forEach((item) => {
        let statusText = "查询";
        let statusClass = "status-query";
        if (item.qr === 1) {
            if (item.rcode === 0) {
                statusText = "成功";
                statusClass = "status-success";
            } else {
                statusText = "错误(" + item.rcode + ")";
                statusClass = "status-error";
            }
        }

        const direction = item.qr === 0 ? "请求" : "响应";
        const cname = item.cnamechain || item.cname_chain || "-";
        const ips = item.responseips || item.response_ips || "-";
        const qtype = item.qtype ?? item.query_type;

        const typeMap = { 1: "A", 28: "AAAA", 5: "CNAME", 16: "TXT", 15: "MX", 2: "NS", 65: "HTTPS" };
        const typeLabel = typeMap[qtype] || qtype;

        html += `<tr>
            <td>${escapeHtml(item.domain || "")}</td>
            <td>${escapeHtml(String(typeLabel))}</td>
            <td>${direction}</td>
            <td class="muted">${escapeHtml(String(cname))}</td>
            <td>${escapeHtml(String(ips))}</td>
            <td>${item.ttl || "-"}</td>
            <td class="${statusClass}">${statusText}</td>
            <td class="muted">${escapeHtml(formatTime(item.created_at)+"\nUTC+8")}</td>
        </tr>`;
    });

    tbody.innerHTML = html;
}

function escapeHtml(text) {
    const div = document.createElement("div");
    div.textContent = text == null ? "" : String(text);
    return div.innerHTML;
}

function formatTime(ts) {
    if (!ts) return "-";
    try {
        const d = new Date(ts);
        if (Number.isNaN(d.getTime())) {
            return String(ts);
        }
        // API 多为 UTC（带 Z）；按北京时间（Asia/Shanghai）展示
        return new Intl.DateTimeFormat("zh-CN", {
            timeZone: "Asia/Shanghai",
            year: "numeric",
            month: "2-digit",
            day: "2-digit",
            hour: "2-digit",
            minute: "2-digit",
            second: "2-digit",
            hour12: false,
        })
            .format(d)
            .replace(/\//g, "-");
    } catch (_) {
        return String(ts);
    }
}
