import { API_URL, authHeaders, parseError, apiFetch } from '../apiCore.js';

export function unwrapData(json) {
    if (json && typeof json === 'object' && json.success === true && json.data !== undefined) {
        return json.data;
    }
    return json;
}

export { API_URL, authHeaders, parseError, apiFetch };
