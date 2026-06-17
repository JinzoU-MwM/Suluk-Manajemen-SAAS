import { API_URL, authHeaders, parseError, apiFetch } from '../apiCore.js';

function unwrapData(json) {
    if (json && typeof json === 'object' && json.success === true && json.data !== undefined) {
        return json.data;
    }
    return json;
}

export const documentExcelApi = {
    /**
     * Get OCR runtime/provider/cache status (auth required)
     */
    async getOcrStatus() {
        try {
            const response = await apiFetch(`${API_URL}/ocr/status`, {
                headers: authHeaders(),
            });

            if (!response.ok) {
                throw new Error(await parseError(response));
            }

            return unwrapData(await response.json());
        } catch (error) {
            if (error.message.includes('fetch')) {
                throw new Error(`Failed to get OCR status: ${error.message}`);
            }
            throw error;
        }
    },

    /**
     * Upload documents for OCR processing (auth required)
     */
    async uploadDocuments(files, sessionId = null, options = {}) {
        const formData = new FormData();
        files.forEach((file) => {
            formData.append('files', file);
        });

        const normalizedCacheMode = String(options.cacheMode || 'default').trim().toLowerCase();
        if (!['default', 'refresh', 'bypass'].includes(normalizedCacheMode)) {
            throw new Error(`Invalid cache mode: ${normalizedCacheMode}`);
        }
        const params = new URLSearchParams();
        if (sessionId) params.set('session_id', sessionId);
        params.set('cache_mode', normalizedCacheMode);
        const query = params.toString();
        const url = query
            ? `${API_URL}/process-documents/?${query}`
            : `${API_URL}/process-documents/`;

        try {
            const response = await apiFetch(url, {
                method: 'POST',
                headers: authHeaders(),
                body: formData,
            });

            if (!response.ok) {
                throw new Error(await parseError(response));
            }

            return unwrapData(await response.json());
        } catch (error) {
            if (error.message.includes('fetch')) {
                throw new Error(`Connection failed: ${error.message}. Is the backend running?`);
            }
            throw error;
        }
    },

    /**
     * Stream progress updates via SSE
     */
    streamProgress(sessionId, onProgress) {
        // No live progress stream: /process-documents is synchronous and there
        // is no backend SSE endpoint. Returning a no-op handle (with .close())
        // keeps callers working without firing a /progress request that 404s
        // and trips the service worker. Real status comes from the POST result.
        void sessionId;
        void onProgress;
        return { close() {} };
    },

    async generateExcel(data) {
        try {
            const response = await apiFetch(`${API_URL}/generate-excel/`, {
                method: 'POST',
                headers: authHeaders({ 'Content-Type': 'application/json' }),
                body: JSON.stringify({ data }),
            });

            if (!response.ok) {
                throw new Error(await parseError(response));
            }

            const blob = await response.blob();
            if (blob.size < 100) {
                throw new Error('Downloaded file appears to be corrupted (too small)');
            }
            return blob;
        } catch (error) {
            if (error.message.includes('fetch')) {
                throw new Error(`Excel generation failed: ${error.message}`);
            }
            throw error;
        }
    },
};
