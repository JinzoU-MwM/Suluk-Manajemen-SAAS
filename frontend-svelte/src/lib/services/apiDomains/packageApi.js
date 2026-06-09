import { API_URL, authHeaders, parseError, apiFetch } from '../apiCore.js';

function unwrapData(json) {
    if (json && typeof json === 'object' && json.success === true && json.data !== undefined) {
        return json.data;
    }
    return json;
}

export function createPackageApi({ cacheInvalidate, cacheGet, cacheSet }) {
    return {
        async listPackages({ status = '', page = 1, pageSize = 100 } = {}) {
            const params = new URLSearchParams({
                page: String(page),
                page_size: String(pageSize),
            });
            if (status) {
                params.set('status', status);
            }
            // Short-TTL cache: packages change rarely but the list is fetched as a
            // dropdown source across many pages. Mutations call cacheInvalidate('packages:').
            const cacheKey = `packages:list:${params.toString()}`;
            const cached = cacheGet?.(cacheKey);
            if (cached) return cached;
            const response = await apiFetch(`${API_URL}/packages/?${params.toString()}`, {
                headers: authHeaders(),
            });
            if (!response.ok) throw new Error(await parseError(response));
            const data = unwrapData(await response.json());
            cacheSet?.(cacheKey, data, 60_000);
            return data;
        },

        async getPackage(packageId) {
            const response = await apiFetch(`${API_URL}/packages/${packageId}`, {
                headers: authHeaders(),
            });
            if (!response.ok) throw new Error(await parseError(response));
            return unwrapData(await response.json());
        },

        async createPackage(data) {
            const response = await apiFetch(`${API_URL}/packages/`, {
                method: 'POST',
                headers: authHeaders({ 'Content-Type': 'application/json' }),
                body: JSON.stringify(data),
            });
            if (!response.ok) throw new Error(await parseError(response));
            cacheInvalidate?.('packages:');
            return unwrapData(await response.json());
        },

        async updatePackage(packageId, data) {
            const response = await apiFetch(`${API_URL}/packages/${packageId}`, {
                method: 'PUT',
                headers: authHeaders({ 'Content-Type': 'application/json' }),
                body: JSON.stringify(data),
            });
            if (!response.ok) throw new Error(await parseError(response));
            cacheInvalidate?.('packages:');
            return unwrapData(await response.json());
        },

        async deletePackage(packageId) {
            const response = await apiFetch(`${API_URL}/packages/${packageId}`, {
                method: 'DELETE',
                headers: authHeaders(),
            });
            if (!response.ok) throw new Error(await parseError(response));
            cacheInvalidate?.('packages:');
            return unwrapData(await response.json());
        },

        async updatePackageStatus(packageId, status) {
            const response = await apiFetch(`${API_URL}/packages/${packageId}/status`, {
                method: 'PATCH',
                headers: authHeaders({ 'Content-Type': 'application/json' }),
                body: JSON.stringify({ status }),
            });
            if (!response.ok) throw new Error(await parseError(response));
            cacheInvalidate?.('packages:');
            return unwrapData(await response.json());
        },

        async createPricingTier(packageId, data) {
            const response = await apiFetch(`${API_URL}/packages/${packageId}/tiers`, {
                method: 'POST',
                headers: authHeaders({ 'Content-Type': 'application/json' }),
                body: JSON.stringify(data),
            });
            if (!response.ok) throw new Error(await parseError(response));
            cacheInvalidate?.('packages:');
            return unwrapData(await response.json());
        },

        async updatePricingTier(packageId, tierId, data) {
            const response = await apiFetch(`${API_URL}/packages/${packageId}/tiers/${tierId}`, {
                method: 'PUT',
                headers: authHeaders({ 'Content-Type': 'application/json' }),
                body: JSON.stringify(data),
            });
            if (!response.ok) throw new Error(await parseError(response));
            cacheInvalidate?.('packages:');
            return unwrapData(await response.json());
        },

        async deletePricingTier(packageId, tierId) {
            const response = await apiFetch(`${API_URL}/packages/${packageId}/tiers/${tierId}`, {
                method: 'DELETE',
                headers: authHeaders(),
            });
            if (!response.ok) throw new Error(await parseError(response));
            cacheInvalidate?.('packages:');
            return unwrapData(await response.json());
        },
    };
}
