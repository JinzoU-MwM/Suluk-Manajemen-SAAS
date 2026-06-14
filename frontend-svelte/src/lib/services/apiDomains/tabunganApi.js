import { API_URL, authHeaders, apiFetch, parseError } from '../apiCore.js';

function unwrapData(json) {
    if (json && typeof json === 'object' && json.success && json.data !== undefined) return json.data;
    return json;
}

export function createTabunganApi({ cacheInvalidate } = /** @type {any} */ ({})) {
    const BASE = `${API_URL}/tabungan`;
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
        async listTabungan({ page = 1, limit = 50 } = {}) {
            const res = await apiFetch(`${BASE}?page=${page}&limit=${limit}`, { headers: authHeaders() });
            if (!res.ok) throw new Error(await parseError(res));
            const json = await res.json();
            return { items: unwrapData(json), meta: json.meta || null };
        },
        getTabungan(id) { return send(`${BASE}/${id}`, 'GET'); },
        createTabungan(data) { const r = send(`${BASE}`, 'POST', data); cacheInvalidate?.('tabungan:'); return r; },
        depositTabungan(id, data) { return send(`${BASE}/${id}/deposit`, 'POST', data); },
        convertTabungan(id, data) { return send(`${BASE}/${id}/convert`, 'POST', data); },
    };
}
