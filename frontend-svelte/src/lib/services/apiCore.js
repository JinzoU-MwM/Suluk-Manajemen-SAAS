import { mapError } from './toast.svelte.js';

// PROXY CONFIGURATION:
// requests to /api/... will be forwarded to http://localhost:8000/... by Vite
export const API_URL = '/api';

/**
 * Read the stored JWT access token.
 */
function getToken() {
    if (typeof window === 'undefined') return null;
    try {
        return localStorage.getItem('access_token');
    } catch {
        return null;
    }
}

export function getRefreshToken() {
    if (typeof window === 'undefined') return null;
    try {
        return localStorage.getItem('refresh_token');
    } catch {
        return null;
    }
}

/** Persist tokens after login/refresh. */
export function setTokens(accessToken, refreshToken) {
    try {
        if (accessToken) localStorage.setItem('access_token', accessToken);
        if (refreshToken) localStorage.setItem('refresh_token', refreshToken);
    } catch { /* ignore */ }
}

/** Clear tokens (on logout or unrecoverable 401). */
export function clearTokens() {
    try {
        localStorage.removeItem('access_token');
        localStorage.removeItem('refresh_token');
    } catch { /* ignore */ }
}

// A single in-flight refresh shared across concurrent 401s so we don't spawn a
// refresh storm when several requests expire at once.
let refreshInFlight = null;

async function refreshAccessToken() {
    const rt = getRefreshToken();
    if (!rt) return false;
    if (!refreshInFlight) {
        refreshInFlight = fetch(`${API_URL}/auth/refresh`, {
            method: 'POST',
            credentials: 'include',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ refresh_token: rt }),
        })
            .then(async (r) => {
                if (!r.ok) return false;
                const json = await r.json().catch(() => null);
                const data = json && json.success && json.data ? json.data : json;
                if (data && data.access_token) {
                    setTokens(data.access_token, data.refresh_token);
                    return true;
                }
                return false;
            })
            .catch(() => false)
            .finally(() => { refreshInFlight = null; });
    }
    return refreshInFlight;
}

/**
 * Create request headers with JWT Bearer token.
 */
export function authHeaders(extra = {}) {
    const token = getToken();
    const headers = { ...extra };
    if (token) {
        headers['Authorization'] = `Bearer ${token}`;
    }
    return headers;
}

/**
 * Cookie-aware fetch for API requests.
 */
export async function apiFetch(url, options = {}, _allowRetry = true) {
    const token = getToken();
    const headers = { ...(options.headers || {}) };
    if (token && !headers['Authorization']) {
        headers['Authorization'] = `Bearer ${token}`;
    }
    const resp = await fetch(url, {
        credentials: 'include',
        ...options,
        headers,
    });

    // On an expired/invalid access token, transparently refresh once and retry.
    const isAuthFlow = url.includes('/auth/refresh') || url.includes('/auth/login') || url.includes('/auth/register');
    if (resp.status === 401 && _allowRetry && !isAuthFlow && getRefreshToken()) {
        const ok = await refreshAccessToken();
        if (ok) {
            return apiFetch(url, options, false);
        }
        clearTokens();
    }
    return resp;
}

/**
 * Parse API error from response and map to Indonesian messages.
 */
export async function parseError(response) {
    const errText = await response.text();
    let message = errText;
    try {
        const json = JSON.parse(errText);
        const detail = json.detail;
        if (typeof detail === 'string') {
            message = detail;
        } else if (detail && typeof detail === 'object') {
            const detailMessage = typeof detail.message === 'string'
                ? detail.message
                : JSON.stringify(detail);
            if (detail.code === 'bypass_quota_exceeded' && detail.quota) {
                const remaining = detail.quota.remaining_files ?? '-';
                const limit = detail.quota.limit_files ?? '-';
                const suggestedMode = detail.suggested_mode ? ` Try ${detail.suggested_mode} mode.` : '';
                message = `${detailMessage} Remaining ${remaining}/${limit} files in 1h window.${suggestedMode}`;
            } else {
                message = detailMessage;
                if (detail.errors && Array.isArray(detail.errors) && detail.errors.length > 0) {
                    const errorList = detail.errors.slice(0, 10).join('\n');
                    const more = detail.errors.length > 10 ? `\n... dan ${detail.errors.length - 10} error lainnya` : '';
                    message = `${detailMessage}\n\n${errorList}${more}`;
                }
            }
        } else if (typeof json.error === 'string') {
            message = json.error;
        } else if (typeof json.message === 'string') {
            message = json.message;
        } else if (Array.isArray(json.errors) && json.errors.length > 0) {
            const firstError = json.errors[0];
            if (typeof firstError === 'string') {
                message = firstError;
            } else if (firstError && typeof firstError.message === 'string') {
                message = firstError.message;
            }
        }
    } catch (e) { /* ignore */ }
    return mapError(message || 'Request failed');
}
