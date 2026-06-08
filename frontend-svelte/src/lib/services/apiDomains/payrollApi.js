import { API_URL, authHeaders, apiFetch, parseError } from '../apiCore.js';

function unwrapData(json) {
    if (json && typeof json === 'object' && json.success && json.data !== undefined) {
        return json.data;
    }
    return json;
}

export function createPayrollApi({ cacheInvalidate } = /** @type {{cacheInvalidate: Function}} */ ({})) {
    const BASE = `${API_URL}/payroll`;

    return {
        async getSummary() {
            const res = await apiFetch(`${BASE}/summary`, { headers: authHeaders() });
            if (!res.ok) throw new Error(await parseError(res));
            return unwrapData(await res.json());
        },

        async listEmployees() {
            const res = await apiFetch(`${BASE}/employees`, { headers: authHeaders() });
            if (!res.ok) throw new Error(await parseError(res));
            return unwrapData(await res.json());
        },

        async createEmployee(data) {
            const res = await apiFetch(`${BASE}/employees`, {
                method: 'POST',
                headers: authHeaders({ 'Content-Type': 'application/json' }),
                body: JSON.stringify(data),
            });
            if (!res.ok) throw new Error(await parseError(res));
            return unwrapData(await res.json());
        },

        async updateEmployee(id, data) {
            const res = await apiFetch(`${BASE}/employees/${id}`, {
                method: 'PUT',
                headers: authHeaders({ 'Content-Type': 'application/json' }),
                body: JSON.stringify(data),
            });
            if (!res.ok) throw new Error(await parseError(res));
            return unwrapData(await res.json());
        },

        async listSalarySlips(period) {
            const params = new URLSearchParams();
            if (period) params.set('period', period);
            const res = await apiFetch(`${BASE}/slips?${params}`, { headers: authHeaders() });
            if (!res.ok) throw new Error(await parseError(res));
            return unwrapData(await res.json());
        },

        async createSalarySlip(data) {
            const res = await apiFetch(`${BASE}/slips`, {
                method: 'POST',
                headers: authHeaders({ 'Content-Type': 'application/json' }),
                body: JSON.stringify(data),
            });
            if (!res.ok) throw new Error(await parseError(res));
            return unwrapData(await res.json());
        },

        async finalizeSlip(id) {
            const res = await apiFetch(`${BASE}/slips/${id}/finalize`, {
                method: 'PUT',
                headers: authHeaders(),
            });
            if (!res.ok) throw new Error(await parseError(res));
            return unwrapData(await res.json());
        },

        async listAdvances() {
            const res = await apiFetch(`${BASE}/advances`, { headers: authHeaders() });
            if (!res.ok) throw new Error(await parseError(res));
            return unwrapData(await res.json());
        },

        async createAdvance(data) {
            const res = await apiFetch(`${BASE}/advances`, {
                method: 'POST',
                headers: authHeaders({ 'Content-Type': 'application/json' }),
                body: JSON.stringify(data),
            });
            if (!res.ok) throw new Error(await parseError(res));
            return unwrapData(await res.json());
        },

        async repayAdvance(id, data) {
            const res = await apiFetch(`${BASE}/advances/${id}/repay`, {
                method: 'PUT',
                headers: authHeaders({ 'Content-Type': 'application/json' }),
                body: JSON.stringify(data),
            });
            if (!res.ok) throw new Error(await parseError(res));
            return unwrapData(await res.json());
        },
    };
}
