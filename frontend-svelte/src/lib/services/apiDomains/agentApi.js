import { API_URL, authHeaders, apiFetch, parseError } from '../apiCore.js';

function unwrapData(json) {
    if (json && typeof json === 'object' && json.success && json.data !== undefined) return json.data;
    return json;
}

export function createAgentApi({ cacheInvalidate } = /** @type {{cacheInvalidate: Function}} */ ({})) {
    const AGENTS = `${API_URL}/agents`;
    const COMM = `${API_URL}/commissions`;

    return {
        async listAgents(params) {
            const q = new URLSearchParams();
            if (params?.search) q.set('search', params.search);
            if (params?.page) q.set('page', String(params.page));
            if (params?.limit) q.set('limit', String(params.limit));
            const res = await apiFetch(`${AGENTS}/?${q}`, { headers: authHeaders() });
            if (!res.ok) throw new Error(await parseError(res));
            return unwrapData(await res.json());
        },

        async getAgent(id) {
            const res = await apiFetch(`${AGENTS}/${id}`, { headers: authHeaders() });
            if (!res.ok) throw new Error(await parseError(res));
            return unwrapData(await res.json());
        },

        async createAgent(data) {
            const res = await apiFetch(`${AGENTS}/`, { method: 'POST', headers: authHeaders({ 'Content-Type': 'application/json' }), body: JSON.stringify(data) });
            if (!res.ok) throw new Error(await parseError(res));
            return unwrapData(await res.json());
        },

        async updateAgent(id, data) {
            const res = await apiFetch(`${AGENTS}/${id}`, { method: 'PUT', headers: authHeaders({ 'Content-Type': 'application/json' }), body: JSON.stringify(data) });
            if (!res.ok) throw new Error(await parseError(res));
            return unwrapData(await res.json());
        },

        async listCommissions(params) {
            const q = new URLSearchParams();
            if (params?.agent_id) q.set('agent_id', params.agent_id);
            if (params?.status) q.set('status', params.status);
            if (params?.page) q.set('page', String(params.page));
            if (params?.limit) q.set('limit', String(params.limit));
            const res = await apiFetch(`${COMM}/?${q}`, { headers: authHeaders() });
            if (!res.ok) throw new Error(await parseError(res));
            return unwrapData(await res.json());
        },

        async createCommission(data) {
            const res = await apiFetch(`${COMM}/`, { method: 'POST', headers: authHeaders({ 'Content-Type': 'application/json' }), body: JSON.stringify(data) });
            if (!res.ok) throw new Error(await parseError(res));
            return unwrapData(await res.json());
        },

        async payCommission(id) {
            const res = await apiFetch(`${COMM}/${id}/pay`, { method: 'PUT', headers: authHeaders() });
            if (!res.ok) throw new Error(await parseError(res));
            return unwrapData(await res.json());
        },

        async getAgentCommissions(agentId) {
            const res = await apiFetch(`${COMM}/agent/${agentId}`, { headers: authHeaders() });
            if (!res.ok) throw new Error(await parseError(res));
            return unwrapData(await res.json());
        },
    };
}
