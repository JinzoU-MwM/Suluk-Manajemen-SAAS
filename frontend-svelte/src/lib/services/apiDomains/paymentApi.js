import { API_URL, authHeaders, parseError, apiFetch } from '../apiCore.js';

function unwrapData(json) {
    if (json && typeof json === 'object' && json.success === true && json.data !== undefined) {
        return json.data;
    }
    return json;
}

export const paymentApi = {
    async createPaymentOrder(planType = 'monthly') {
        const response = await apiFetch(`${API_URL}/payment/create-order?plan_type=${planType}`, {
            method: 'POST',
            headers: authHeaders({ 'Content-Type': 'application/json' }),
        });
        if (!response.ok) throw new Error(await parseError(response));
        return unwrapData(await response.json());
    },

    async checkPaymentStatus(orderId) {
        const response = await apiFetch(`${API_URL}/payment/status/${orderId}`, {
            headers: authHeaders(),
        });
        if (!response.ok) throw new Error(await parseError(response));
        return unwrapData(await response.json());
    },
};
