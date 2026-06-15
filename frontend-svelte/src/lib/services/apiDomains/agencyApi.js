import { API_URL, authHeaders, apiFetch, parseError } from '../apiCore.js';

function unwrapData(json) {
  if (json && typeof json === 'object' && json.success && json.data !== undefined) return json.data;
  return json;
}

// B2B agent portal — every endpoint is scoped server-side to the signed-in
// agent (role "agent"); no ids are passed from the client.
export function createAgencyApi() {
  const B2B = `${API_URL}/b2b`;
  return {
    async myProfile() {
      const res = await apiFetch(`${B2B}/me`, { headers: authHeaders() });
      if (!res.ok) throw new Error(await parseError(res));
      return unwrapData(await res.json());
    },
    async myDashboard() {
      const res = await apiFetch(`${B2B}/dashboard`, { headers: authHeaders() });
      if (!res.ok) throw new Error(await parseError(res));
      return unwrapData(await res.json());
    },
    async myDownline() {
      const res = await apiFetch(`${B2B}/downline`, { headers: authHeaders() });
      if (!res.ok) throw new Error(await parseError(res));
      return unwrapData(await res.json());
    },
    async myCommissions() {
      const res = await apiFetch(`${B2B}/commissions`, { headers: authHeaders() });
      if (!res.ok) throw new Error(await parseError(res));
      return unwrapData(await res.json());
    },
  };
}
