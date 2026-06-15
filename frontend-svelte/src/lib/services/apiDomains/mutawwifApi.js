import { API_URL, authHeaders, apiFetch, parseError } from '../apiCore.js';

function unwrapData(json) {
  if (json && typeof json === 'object' && json.success === true && json.data !== undefined) return json.data;
  return json;
}

// Mutawwif / tour-guide management (Phase 5B).
export function createMutawwifApi() {
  const M = `${API_URL}/mutawwif`;
  return {
    async listGuides(search = '') {
      const q = search ? `?search=${encodeURIComponent(search)}` : '';
      const res = await apiFetch(`${M}/guides${q}`, { headers: authHeaders() });
      if (!res.ok) throw new Error(await parseError(res));
      return unwrapData(await res.json());
    },
    async createGuide(data) {
      const res = await apiFetch(`${M}/guides`, { method: 'POST', headers: authHeaders({ 'Content-Type': 'application/json' }), body: JSON.stringify(data) });
      if (!res.ok) throw new Error(await parseError(res));
      return unwrapData(await res.json());
    },
    async updateGuide(id, data) {
      const res = await apiFetch(`${M}/guides/${id}`, { method: 'PUT', headers: authHeaders({ 'Content-Type': 'application/json' }), body: JSON.stringify(data) });
      if (!res.ok) throw new Error(await parseError(res));
      return unwrapData(await res.json());
    },
    async deleteGuide(id) {
      const res = await apiFetch(`${M}/guides/${id}`, { method: 'DELETE', headers: authHeaders() });
      if (!res.ok) throw new Error(await parseError(res));
      return unwrapData(await res.json());
    },
    async assignGuide({ guide_id, group_id, role }) {
      const res = await apiFetch(`${M}/assignments`, { method: 'POST', headers: authHeaders({ 'Content-Type': 'application/json' }), body: JSON.stringify({ guide_id, group_id, role }) });
      if (!res.ok) throw new Error(await parseError(res));
      return unwrapData(await res.json());
    },
    async listGroupGuides(groupId) {
      const res = await apiFetch(`${M}/assignments/group/${groupId}`, { headers: authHeaders() });
      if (!res.ok) throw new Error(await parseError(res));
      return unwrapData(await res.json());
    },
    async unassignGuide(groupId, guideId) {
      const res = await apiFetch(`${M}/assignments/${groupId}/${guideId}`, { method: 'DELETE', headers: authHeaders() });
      if (!res.ok) throw new Error(await parseError(res));
      return unwrapData(await res.json());
    },
  };
}
