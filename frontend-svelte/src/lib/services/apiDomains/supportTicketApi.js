import { API_URL, authHeaders, parseError, apiFetch } from '../apiCore.js';

function unwrapData(json) {
    if (json && typeof json === 'object' && json.success === true && json.data !== undefined) {
        return json.data;
    }
    return json;
}

export const supportTicketApi = {
    async createTicket(subject, message, priority = 'medium') {
        const response = await apiFetch(`${API_URL}/tickets`, {
            method: 'POST',
            headers: authHeaders({ 'Content-Type': 'application/json' }),
            body: JSON.stringify({ subject, message, priority }),
        });
        if (!response.ok) throw new Error(await parseError(response));
        return unwrapData(await response.json());
    },

    async listMyTickets(filters = {}) {
        const { page = 1, pageSize = 20, status } = filters;
        const params = new URLSearchParams({
            page: String(page),
            page_size: String(pageSize),
        });
        if (status) params.append('status', status);

        const response = await apiFetch(`${API_URL}/tickets?${params}`, {
            headers: authHeaders(),
        });
        if (!response.ok) throw new Error(await parseError(response));
        return unwrapData(await response.json());
    },

    async getMyTicketDetail(ticketId) {
        const response = await apiFetch(`${API_URL}/tickets/${ticketId}/messages`, {
            headers: authHeaders(),
        });
        if (!response.ok) throw new Error(await parseError(response));
        return unwrapData(await response.json());
    },

    async replyToTicket(ticketId, content) {
        const response = await apiFetch(`${API_URL}/tickets/${ticketId}/messages`, {
            method: 'POST',
            headers: authHeaders({ 'Content-Type': 'application/json' }),
            body: JSON.stringify({ content }),
        });
        if (!response.ok) throw new Error(await parseError(response));
        return unwrapData(await response.json());
    },
};
