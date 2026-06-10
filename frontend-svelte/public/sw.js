// Service Worker — Suluk PWA
// Cache static assets only; keep API/dynamic requests network-first to avoid stale data.

const CACHE_NAME = 'jamaah-v17';
const STATIC_ASSETS = [
    '/',
    '/index.html',
];

const STATIC_DESTINATIONS = new Set(['style', 'script', 'image', 'font', 'worker']);

// Install: cache shell
self.addEventListener('install', (event) => {
    event.waitUntil(
        caches.open(CACHE_NAME).then((cache) => cache.addAll(STATIC_ASSETS))
    );
    self.skipWaiting();
});

// Activate: clean old caches
self.addEventListener('activate', (event) => {
    event.waitUntil(
        caches.keys().then((keys) =>
            Promise.all(keys.filter((k) => k !== CACHE_NAME).map((k) => caches.delete(k)))
        )
    );
    self.clients.claim();
});

// Fetch strategy
self.addEventListener('fetch', (event) => {
    const { request } = event;
    const url = new URL(request.url);

    // Skip non-GET
    if (request.method !== 'GET') return;

    // Navigations: network-first with app shell fallback.
    if (request.mode === 'navigate') {
        event.respondWith(fetch(request).catch(() => caches.match('/index.html')));
        return;
    }

    // Never cache API/dynamic endpoints; this avoids stale JSON for auth/groups/subscription, etc.
    const isSameOrigin = url.origin === self.location.origin;
    const isLikelyStatic = STATIC_DESTINATIONS.has(request.destination) || STATIC_ASSETS.includes(url.pathname);
    if (!isSameOrigin || !isLikelyStatic) {
        event.respondWith(fetch(request));
        return;
    }

    // Scripts & styles: stale-while-revalidate. Serve the cached copy instantly
    // for speed, but always refetch in the background and update the cache so a
    // new deploy is picked up on the next load even if CACHE_NAME wasn't bumped
    // (guards against the stale-bundle problem). Content-hashed filenames still
    // miss-then-cache as before.
    if (request.destination === 'script' || request.destination === 'style') {
        event.respondWith(
            caches.match(request).then((cached) => {
                const network = fetch(request).then((response) => {
                    if (response.ok) {
                        const clone = response.clone();
                        caches.open(CACHE_NAME).then((cache) => cache.put(request, clone));
                    }
                    return response;
                }).catch(() => cached);
                return cached || network;
            })
        );
        return;
    }

    // Other static assets (images, fonts): cache-first.
    event.respondWith(
        caches.match(request).then((cached) => {
            if (cached) return cached;
            return fetch(request).then((response) => {
                if (response.ok) {
                    const clone = response.clone();
                    caches.open(CACHE_NAME).then((cache) => cache.put(request, clone));
                }
                return response;
            });
        })
    );
});
