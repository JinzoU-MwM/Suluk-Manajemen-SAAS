import { API_URL, authHeaders, parseError, apiFetch } from '../apiCore.js';

function unwrapData(json) {
    if (json && json.success === true && json.data !== undefined) {
        return json.data;
    }
    return json;
}

export function createContentApi({ cacheGet, cacheSet }) {
    return {
        async getDashboardStats() {
            const cached = cacheGet('analytics:dashboard');
            if (cached) return cached;
            const response = await apiFetch(`${API_URL}/analytics/dashboard`, {
                headers: authHeaders(),
            });
            if (!response.ok) throw new Error(await parseError(response));
            const result = unwrapData(await response.json());
            cacheSet('analytics:dashboard', result, 30000);
            return result;
        },

        async getOwnerDashboard() {
            const cached = cacheGet('owner:dashboard');
            if (cached) return cached;
            const response = await apiFetch(`${API_URL}/dashboard/owner`, {
                headers: authHeaders(),
            });
            if (!response.ok) throw new Error(await parseError(response));
            const result = unwrapData(await response.json());
            cacheSet('owner:dashboard', result, 15000);
            return result;
        },

        async getItinerary(groupId) {
            const response = await apiFetch(`${API_URL}/itineraries/${groupId}`, {
                headers: authHeaders(),
            });
            if (!response.ok) throw new Error(await parseError(response));
            return await response.json();
        },

        async createItinerary(groupId, data) {
            const response = await apiFetch(`${API_URL}/itineraries/${groupId}`, {
                method: 'POST',
                headers: authHeaders(),
                body: JSON.stringify(data),
            });
            if (!response.ok) throw new Error(await parseError(response));
            return await response.json();
        },

        async updateItinerary(groupId, itemId, data) {
            const response = await apiFetch(`${API_URL}/itineraries/${groupId}/${itemId}`, {
                method: 'PUT',
                headers: authHeaders(),
                body: JSON.stringify(data),
            });
            if (!response.ok) throw new Error(await parseError(response));
            return await response.json();
        },

        async deleteItinerary(groupId, itemId) {
            const response = await apiFetch(`${API_URL}/itineraries/${groupId}/${itemId}`, {
                method: 'DELETE',
                headers: authHeaders(),
            });
            if (!response.ok) throw new Error(await parseError(response));
            return await response.json();
        },

        getDocumentUrl(groupId, type) {
            return `${API_URL}/documents/${groupId}/${type}`;
        },

        async getNotifications() {
            const response = await apiFetch(`${API_URL}/notifications`, {
                headers: authHeaders(),
            });
            if (!response.ok) throw new Error(await parseError(response));
            return unwrapData(await response.json());
        },

        async getUnreadNotificationCount() {
            const data = await this.getNotifications();
            return data?.count ?? 0;
        },

        async markNotificationRead(id) {
            const response = await apiFetch(`${API_URL}/notifications/${id}/read`, {
                method: 'PUT',
                headers: authHeaders(),
            });
            if (!response.ok) throw new Error(await parseError(response));
            return unwrapData(await response.json());
        },

        async markAllNotificationsRead() {
            const response = await apiFetch(`${API_URL}/notifications/read-all`, {
                method: 'PUT',
                headers: authHeaders(),
            });
            if (!response.ok) throw new Error(await parseError(response));
            return unwrapData(await response.json());
        },
    };
}
