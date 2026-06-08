import { API_URL, authHeaders, parseError, apiFetch } from '../apiCore.js';

function unwrapData(json) {
  if (json && json.success === true && json.data !== undefined) {
    return json.data;
  }
  return json;
}

function mapPagination(json) {
  const data = unwrapData(json);
  if (data && Array.isArray(data)) return data;
  if (json && json.meta) return json;
  return data;
}

export function createVendorApi({ cacheInvalidate }) {
  return {
    // ── Vendors ────────────────────────────────────────────
    async listVendors({ type = '', search = '', page = 1, pageSize = 50 } = {}) {
      const params = new URLSearchParams({
        page: String(page),
        page_size: String(pageSize),
      });
      if (type) params.set('type', type);
      if (search) params.set('search', search);
      const response = await apiFetch(`${API_URL}/vendors/?${params.toString()}`, {
        headers: authHeaders(),
      });
      if (!response.ok) throw new Error(await parseError(response));
      return mapPagination(await response.json());
    },

    async getVendor(vendorId) {
      const response = await apiFetch(`${API_URL}/vendors/${vendorId}`, {
        headers: authHeaders(),
      });
      if (!response.ok) throw new Error(await parseError(response));
      return unwrapData(await response.json());
    },

    async createVendor(data) {
      const response = await apiFetch(`${API_URL}/vendors/`, {
        method: 'POST',
        headers: authHeaders({ 'Content-Type': 'application/json' }),
        body: JSON.stringify(data),
      });
      if (!response.ok) throw new Error(await parseError(response));
      cacheInvalidate?.('vendors:');
      return unwrapData(await response.json());
    },

    async updateVendor(vendorId, data) {
      const response = await apiFetch(`${API_URL}/vendors/${vendorId}`, {
        method: 'PUT',
        headers: authHeaders({ 'Content-Type': 'application/json' }),
        body: JSON.stringify(data),
      });
      if (!response.ok) throw new Error(await parseError(response));
      cacheInvalidate?.('vendors:');
      return unwrapData(await response.json());
    },

    async deleteVendor(vendorId) {
      const response = await apiFetch(`${API_URL}/vendors/${vendorId}`, {
        method: 'DELETE',
        headers: authHeaders(),
      });
      if (!response.ok) throw new Error(await parseError(response));
      cacheInvalidate?.('vendors:');
      return unwrapData(await response.json());
    },

    // ── Bills ──────────────────────────────────────────────
    async listBills({ vendorId = '', packageId = '', status = '', page = 1, pageSize = 50 } = {}) {
      const params = new URLSearchParams({
        page: String(page),
        page_size: String(pageSize),
      });
      if (vendorId) params.set('vendor_id', vendorId);
      if (packageId) params.set('package_id', packageId);
      if (status) params.set('status', status);
      const response = await apiFetch(`${API_URL}/vendors/bills?${params.toString()}`, {
        headers: authHeaders(),
      });
      if (!response.ok) throw new Error(await parseError(response));
      return mapPagination(await response.json());
    },

    async getBill(billId) {
      const response = await apiFetch(`${API_URL}/vendors/bills/${billId}`, {
        headers: authHeaders(),
      });
      if (!response.ok) throw new Error(await parseError(response));
      return unwrapData(await response.json());
    },

    async createBill(data) {
      const response = await apiFetch(`${API_URL}/vendors/bills`, {
        method: 'POST',
        headers: authHeaders({ 'Content-Type': 'application/json' }),
        body: JSON.stringify(data),
      });
      if (!response.ok) throw new Error(await parseError(response));
      cacheInvalidate?.('vendors:bills:');
      return unwrapData(await response.json());
    },

    async updateBill(billId, data) {
      const response = await apiFetch(`${API_URL}/vendors/bills/${billId}`, {
        method: 'PUT',
        headers: authHeaders({ 'Content-Type': 'application/json' }),
        body: JSON.stringify(data),
      });
      if (!response.ok) throw new Error(await parseError(response));
      cacheInvalidate?.('vendors:bills:');
      return unwrapData(await response.json());
    },

    async deleteBill(billId) {
      const response = await apiFetch(`${API_URL}/vendors/bills/${billId}`, {
        method: 'DELETE',
        headers: authHeaders(),
      });
      if (!response.ok) throw new Error(await parseError(response));
      cacheInvalidate?.('vendors:bills:');
      return unwrapData(await response.json());
    },

    async getOverdueBills() {
      const response = await apiFetch(`${API_URL}/vendors/bills/overdue`, {
        headers: authHeaders(),
      });
      if (!response.ok) throw new Error(await parseError(response));
      return unwrapData(await response.json());
    },

    async getBillsDueSoon(days = 7) {
      const response = await apiFetch(`${API_URL}/vendors/bills/due-soon?days=${days}`, {
        headers: authHeaders(),
      });
      if (!response.ok) throw new Error(await parseError(response));
      return unwrapData(await response.json());
    },

    async getDebtSummary(vendorId = '') {
      const params = vendorId ? `?vendor_id=${vendorId}` : '';
      const response = await apiFetch(`${API_URL}/vendors/bills/summary${params}`, {
        headers: authHeaders(),
      });
      if (!response.ok) throw new Error(await parseError(response));
      return unwrapData(await response.json());
    },

    async getPackageBillSummary(packageId) {
      const response = await apiFetch(`${API_URL}/vendors/bills/package/${packageId}`, {
        headers: authHeaders(),
      });
      if (!response.ok) throw new Error(await parseError(response));
      return unwrapData(await response.json());
    },

    // ── Payments ──────────────────────────────────────────
    async createPayment(data) {
      const response = await apiFetch(`${API_URL}/vendors/bills/${data.vendor_bill_id}/payments`, {
        method: 'POST',
        headers: authHeaders({ 'Content-Type': 'application/json' }),
        body: JSON.stringify(data),
      });
      if (!response.ok) throw new Error(await parseError(response));
      cacheInvalidate?.('vendors:bills:');
      return unwrapData(await response.json());
    },

    async listPaymentsByBill(billId) {
      const response = await apiFetch(`${API_URL}/vendors/bills/${billId}/payments`, {
        headers: authHeaders(),
      });
      if (!response.ok) throw new Error(await parseError(response));
      return unwrapData(await response.json());
    },

    async deletePayment(paymentId) {
      const response = await apiFetch(`${API_URL}/vendors/payments/${paymentId}`, {
        method: 'DELETE',
        headers: authHeaders(),
      });
      if (!response.ok) throw new Error(await parseError(response));
      cacheInvalidate?.('vendors:bills:');
      return unwrapData(await response.json());
    },
  };
}
