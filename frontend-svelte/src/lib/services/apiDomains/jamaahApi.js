import { API_URL, authHeaders, parseError, apiFetch } from '../apiCore.js';

function unwrapData(json) {
  if (json && json.success === true && json.data !== undefined) {
    return json.data;
  }
  return json;
}

export function createJamaahApi({ cacheInvalidate }) {
  return {
    async listJamaah({ search = '', page = 1, pageSize = 50 } = {}) {
      const params = new URLSearchParams({ page: String(page), page_size: String(pageSize) });
      if (search) params.set('search', search);
      const response = await apiFetch(`${API_URL}/jamaah/?${params.toString()}`, { headers: authHeaders() });
      if (!response.ok) throw new Error(await parseError(response));
      const json = await response.json();
      return unwrapData(json);
    },

    async getJamaah(id) {
      const response = await apiFetch(`${API_URL}/jamaah/${id}`, { headers: authHeaders() });
      if (!response.ok) throw new Error(await parseError(response));
      return unwrapData(await response.json());
    },

    async listDocuments(jamaahId) {
      const response = await apiFetch(`${API_URL}/jamaah/${jamaahId}/documents`, { headers: authHeaders() });
      if (!response.ok) throw new Error(await parseError(response));
      return unwrapData(await response.json());
    },

    async uploadDocument(jamaahId, data, file) {
      const formData = new FormData();
      formData.append('doc_type', data.doc_type);
      if (data.status) formData.append('status', data.status);
      if (data.notes) formData.append('notes', data.notes);
      if (file) formData.append('file', file);
      const response = await apiFetch(`${API_URL}/jamaah/${jamaahId}/documents`, {
        method: 'POST',
        headers: authHeaders(),
        body: formData,
      });
      if (!response.ok) throw new Error(await parseError(response));
      cacheInvalidate?.('jamaah:documents:');
      return unwrapData(await response.json());
    },

    async updateDocumentStatus(jamaahId, docId, data) {
      const response = await apiFetch(`${API_URL}/jamaah/${jamaahId}/documents/${docId}/status`, {
        method: 'PATCH',
        headers: authHeaders({ 'Content-Type': 'application/json' }),
        body: JSON.stringify(data),
      });
      if (!response.ok) throw new Error(await parseError(response));
      cacheInvalidate?.('jamaah:documents:');
      return unwrapData(await response.json());
    },

    async getDashboardAlerts() {
      const response = await apiFetch(`${API_URL}/jamaah/dashboard/alerts`, { headers: authHeaders() });
      if (!response.ok) throw new Error(await parseError(response));
      return unwrapData(await response.json());
    },
  };
}
