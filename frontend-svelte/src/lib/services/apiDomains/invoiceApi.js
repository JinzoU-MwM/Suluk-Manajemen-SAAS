import { API_URL, authHeaders, parseError, apiFetch } from '../apiCore.js';

function unwrapData(json) {
    if (json && typeof json === 'object' && json.success === true && json.data !== undefined) {
        return json.data;
    }
    return json;
}

export const invoiceApi = {
    async listInvoices({ status = '', page = 1, pageSize = 20 } = {}) {
        const params = new URLSearchParams({
            page: String(page),
            page_size: String(pageSize),
        });
        if (status) params.set('status', status);
        const response = await apiFetch(`${API_URL}/invoices?${params}`, {
            headers: authHeaders(),
        });
        if (!response.ok) throw new Error(await parseError(response));
        return unwrapData(await response.json());
    },

    async getInvoice(invoiceId) {
        const response = await apiFetch(`${API_URL}/invoices/${invoiceId}`, {
            headers: authHeaders(),
        });
        if (!response.ok) throw new Error(await parseError(response));
        return unwrapData(await response.json());
    },

    async createInvoice(data) {
        const response = await apiFetch(`${API_URL}/invoices`, {
            method: 'POST',
            headers: authHeaders({ 'Content-Type': 'application/json' }),
            body: JSON.stringify(data),
        });
        if (!response.ok) throw new Error(await parseError(response));
        return unwrapData(await response.json());
    },

    async updateInvoice(invoiceId, data) {
        const response = await apiFetch(`${API_URL}/invoices/${invoiceId}`, {
            method: 'PUT',
            headers: authHeaders({ 'Content-Type': 'application/json' }),
            body: JSON.stringify(data),
        });
        if (!response.ok) throw new Error(await parseError(response));
        return unwrapData(await response.json());
    },

    async cancelInvoice(invoiceId, reason) {
        const response = await apiFetch(`${API_URL}/invoices/${invoiceId}/cancel`, {
            method: 'PATCH',
            headers: authHeaders({ 'Content-Type': 'application/json' }),
            body: JSON.stringify({ reason }),
        });
        if (!response.ok) throw new Error(await parseError(response));
        return unwrapData(await response.json());
    },

    async recordPayment(invoiceId, data) {
        const response = await apiFetch(`${API_URL}/invoices/${invoiceId}/payments`, {
            method: 'POST',
            headers: authHeaders({ 'Content-Type': 'application/json' }),
            body: JSON.stringify(data),
        });
        if (!response.ok) throw new Error(await parseError(response));
        return unwrapData(await response.json());
    },

    async getPayments(invoiceId) {
        const response = await apiFetch(`${API_URL}/invoices/${invoiceId}/payments`, {
            headers: authHeaders(),
        });
        if (!response.ok) throw new Error(await parseError(response));
        return unwrapData(await response.json());
    },

    async getInvoiceSummary() {
        const response = await apiFetch(`${API_URL}/invoices/summary`, {
            headers: authHeaders(),
        });
        if (!response.ok) throw new Error(await parseError(response));
        return unwrapData(await response.json());
    },
};
