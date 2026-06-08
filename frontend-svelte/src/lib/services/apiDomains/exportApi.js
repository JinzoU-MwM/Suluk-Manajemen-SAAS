import { API_URL } from '../apiCore.js';

export const exportLinks = {
    getInvoicePDFUrl(id) {
        return `${API_URL}/invoices/${id}/pdf`;
    },
    getSlipPDFUrl(id) {
        return `${API_URL}/payroll/slips/${id}/pdf`;
    },
    getPnLExportUrl(packageId) {
        const p = new URLSearchParams();
        if (packageId) p.set('package_id', packageId);
        return `${API_URL}/finance/export/pnl?${p}`;
    },
    getExpensesExportUrl() {
        return `${API_URL}/finance/export/expenses`;
    },
};
