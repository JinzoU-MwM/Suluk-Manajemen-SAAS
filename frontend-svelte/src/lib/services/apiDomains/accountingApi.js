import { API_URL, authHeaders, apiFetch, parseError } from '../apiCore.js';

function unwrapData(json) {
    if (json && typeof json === 'object' && json.success && json.data !== undefined) {
        return json.data;
    }
    return json;
}

export function createAccountingApi({ cacheInvalidate, cacheGet, cacheSet } = /** @type {any} */ ({})) {
    const COA = `${API_URL}/coa`;
    const JOURNALS = `${API_URL}/journals`;
    const REPORTS = `${API_URL}/reports`;

    async function get(url) {
        const res = await apiFetch(url, { headers: authHeaders() });
        if (!res.ok) throw new Error(await parseError(res));
        return unwrapData(await res.json());
    }

    return {
        // Chart of Accounts
        async listCOA() {
            const cached = cacheGet?.('coa:list');
            if (cached) return cached;
            const data = await get(COA);
            cacheSet?.('coa:list', data, 60000);
            return data;
        },
        async createCOA(data) {
            const res = await apiFetch(COA, {
                method: 'POST',
                headers: authHeaders({ 'Content-Type': 'application/json' }),
                body: JSON.stringify(data),
            });
            if (!res.ok) throw new Error(await parseError(res));
            cacheInvalidate?.('coa:');
            return unwrapData(await res.json());
        },

        // Journals
        async listJournals({ page = 1, limit = 20 } = {}) {
            const params = new URLSearchParams({ page: String(page), limit: String(limit) });
            const res = await apiFetch(`${JOURNALS}?${params}`, { headers: authHeaders() });
            if (!res.ok) throw new Error(await parseError(res));
            const json = await res.json();
            // journals endpoint is paginated: {success, data, meta}
            return { items: unwrapData(json), meta: json.meta || null };
        },
        async getJournal(id) {
            return get(`${JOURNALS}/${id}`);
        },

        // Reports
        getTrialBalance(asOf) {
            return get(`${REPORTS}/trial-balance${asOf ? `?as_of=${asOf}` : ''}`);
        },
        getNeraca(asOf) {
            return get(`${REPORTS}/neraca${asOf ? `?as_of=${asOf}` : ''}`);
        },
        getLabaRugi(from, to) {
            const p = new URLSearchParams();
            if (from) p.set('from', from);
            if (to) p.set('to', to);
            const qs = p.toString();
            return get(`${REPORTS}/laba-rugi${qs ? `?${qs}` : ''}`);
        },
        getLedger(accountId, from, to) {
            const p = new URLSearchParams();
            if (from) p.set('from', from);
            if (to) p.set('to', to);
            const qs = p.toString();
            return get(`${REPORTS}/ledger/${accountId}${qs ? `?${qs}` : ''}`);
        },
    };
}
