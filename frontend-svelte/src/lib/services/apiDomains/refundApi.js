import { API_URL, authHeaders, apiFetch, parseError } from '../apiCore.js';

function unwrapData(json) {
    if (json && typeof json === 'object' && json.success && json.data !== undefined) {
        return json.data;
    }
    return json;
}

export function createRefundApi({ cacheInvalidate } = /** @type {{cacheInvalidate: Function}} */ ({})) {
    const BASE = `${API_URL}/refunds`;

    return {
        async listPolicies() {
            const res = await apiFetch(`${BASE}/policies`, { headers: authHeaders() });
            if (!res.ok) throw new Error(await parseError(res));
            return unwrapData(await res.json());
        },

        async createPolicy(data) {
            const res = await apiFetch(`${BASE}/policies`, {
                method: 'POST',
                headers: authHeaders({ 'Content-Type': 'application/json' }),
                body: JSON.stringify(data),
            });
            if (!res.ok) throw new Error(await parseError(res));
            return unwrapData(await res.json());
        },

        async updatePolicy(id, data) {
            const res = await apiFetch(`${BASE}/policies/${id}`, {
                method: 'PUT',
                headers: authHeaders({ 'Content-Type': 'application/json' }),
                body: JSON.stringify(data),
            });
            if (!res.ok) throw new Error(await parseError(res));
            return unwrapData(await res.json());
        },

        async deletePolicy(id) {
            const res = await apiFetch(`${BASE}/policies/${id}`, {
                method: 'DELETE',
                headers: authHeaders(),
            });
            if (!res.ok) throw new Error(await parseError(res));
            return unwrapData(await res.json());
        },

        async listRefunds(/** @type {{status?: string, page?: number, limit?: number}} */ { status, page, limit } = {}) {
            const params = new URLSearchParams();
            if (status) params.set('status', status);
            if (page) params.set('page', String(page));
            if (limit) params.set('limit', String(limit));
            const res = await apiFetch(`${BASE}/?${params}`, { headers: authHeaders() });
            if (!res.ok) throw new Error(await parseError(res));
            return unwrapData(await res.json());
        },

        async getRefund(id) {
            const res = await apiFetch(`${BASE}/${id}`, { headers: authHeaders() });
            if (!res.ok) throw new Error(await parseError(res));
            return unwrapData(await res.json());
        },

        async initiateRefund(invoiceId, data) {
            const res = await apiFetch(`${API_URL}/invoices/${invoiceId}/refund`, {
                method: 'POST',
                headers: authHeaders({ 'Content-Type': 'application/json' }),
                body: JSON.stringify(data),
            });
            if (!res.ok) throw new Error(await parseError(res));
            return unwrapData(await res.json());
        },

        async approveRefund(id) {
            const res = await apiFetch(`${BASE}/${id}/approve`, {
                method: 'PUT',
                headers: authHeaders(),
            });
            if (!res.ok) throw new Error(await parseError(res));
            return unwrapData(await res.json());
        },

        async processRefund(id) {
            const res = await apiFetch(`${BASE}/${id}/process`, {
                method: 'PUT',
                headers: authHeaders(),
            });
            if (!res.ok) throw new Error(await parseError(res));
            return unwrapData(await res.json());
        },

        async completeRefund(id) {
            const res = await apiFetch(`${BASE}/${id}/complete`, {
                method: 'PUT',
                headers: authHeaders(),
            });
            if (!res.ok) throw new Error(await parseError(res));
            return unwrapData(await res.json());
        },

        async rejectRefund(id) {
            const res = await apiFetch(`${BASE}/${id}/reject`, {
                method: 'PUT',
                headers: authHeaders(),
            });
            if (!res.ok) throw new Error(await parseError(res));
            return unwrapData(await res.json());
        },

        async getRefundsByInvoice(invoiceId) {
            const res = await apiFetch(`${BASE}/by-invoice/${invoiceId}`, { headers: authHeaders() });
            if (!res.ok) throw new Error(await parseError(res));
            return unwrapData(await res.json());
        },
    };
}
