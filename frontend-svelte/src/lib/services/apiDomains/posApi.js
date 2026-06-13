import { API_URL, authHeaders, apiFetch, parseError } from '../apiCore.js';

function unwrapData(json) {
    if (json && typeof json === 'object' && json.success && json.data !== undefined) return json.data;
    return json;
}

// Kasir POS — cash-drawer sessions (buka/tutup kas). Backed by invoice-service
// (/api/v1/cash-sessions). Closing a session emits pos.cash.session.closed which
// the accounting-service posts as a cash over/short journal.
export function createPosApi({ cacheInvalidate } = /** @type {any} */ ({})) {
    const BASE = `${API_URL}/cash-sessions`;
    async function send(url, method, body) {
        const res = await apiFetch(url, {
            method,
            headers: authHeaders(body ? { 'Content-Type': 'application/json' } : {}),
            body: body ? JSON.stringify(body) : undefined,
        });
        if (!res.ok) throw new Error(await parseError(res));
        return unwrapData(await res.json());
    }
    return {
        async listCashSessions({ limit = 30 } = {}) {
            const res = await apiFetch(`${BASE}?limit=${limit}`, { headers: authHeaders() });
            if (!res.ok) throw new Error(await parseError(res));
            return unwrapData(await res.json()) || [];
        },
        getActiveCashSession() { return send(`${BASE}/active`, 'GET'); },
        openCashSession(data) {
            const r = send(`${BASE}`, 'POST', data);
            cacheInvalidate?.('cash-sessions:');
            return r;
        },
        closeCashSession(id, data) {
            const r = send(`${BASE}/${id}/close`, 'POST', data);
            cacheInvalidate?.('cash-sessions:');
            return r;
        },
    };
}
