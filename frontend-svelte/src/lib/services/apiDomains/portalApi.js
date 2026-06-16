import { API_URL, authHeaders, apiFetch, parseError } from '../apiCore.js';

function unwrapData(json) {
  if (json && typeof json === 'object' && json.success && json.data !== undefined) return json.data;
  return json;
}

// Jemaah self-service portal (Phase 6) — every endpoint is scoped server-side to
// the signed-in jamaah (role "jamaah"); no ids passed from the client.
export function createPortalApi() {
  const P = `${API_URL}/portal`;
  async function get(path) {
    const res = await apiFetch(`${P}${path}`, { headers: authHeaders() });
    if (!res.ok) throw new Error(await parseError(res));
    return unwrapData(await res.json());
  }
  return {
    portalMe: () => get('/me'),
    portalRegistrations: () => get('/registrations'),
    portalDocuments: () => get('/documents'),
    portalVisa: () => get('/visa'),
    portalPayments: () => get('/payments'),
  };
}
