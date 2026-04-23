// GoNote Service Worker
// Minimal service worker for PWA install support
//
// PWA 功能已禁用，如需重新启用请取消下方注释

/*
// Cache version - automatically uses app version from VERSION file
// Cache is invalidated when app version changes (e.g., 0.10.4 -> 0.10.5)
// This forces users to download fresh files when you release a new version.
const CACHE_NAME = 'gonote-__APP_VERSION__';

// Assets to cache for faster repeat visits
const PRECACHE_ASSETS = [
  '/static/logo.svg',
  '/static/favicon.svg',
  '/static/app.js'
];

// Install event - cache essential assets
self.addEventListener('install', (event) => {
  event.waitUntil(
    caches.open(CACHE_NAME)
      .then((cache) => cache.addAll(PRECACHE_ASSETS))
      .then(() => self.skipWaiting())
  );
});

// Activate event - clean up old caches
self.addEventListener('activate', (event) => {
  event.waitUntil(
    caches.keys().then((cacheNames) => {
      return Promise.all(
        cacheNames
          .filter((name) => name !== CACHE_NAME)
          .map((name) => caches.delete(name))
      );
    }).then(() => self.clients.claim())
  );
});

// Fetch event - network first, fallback to cache for assets
self.addEventListener('fetch', (event) => {
  const url = new URL(event.request.url);
  
  // Only handle same-origin requests
  if (url.origin !== location.origin) {
    return;
  }
  
  // For API calls, always go to network
  if (url.pathname.startsWith('/api/')) {
    return;
  }
  
  // Skip manifest.json - let browser handle it directly
  // This avoids CORS issues when behind proxy auth
  if (url.pathname === '/static/manifest.json') {
    return;
  }
  
  // For static assets, try cache first then network
  if (url.pathname.startsWith('/static/')) {
    event.respondWith(
      caches.match(event.request)
        .then((cached) => cached || fetch(event.request))
    );
    return;
  }
  
  // For everything else, network first
  event.respondWith(
    fetch(event.request)
      .catch(() => caches.match(event.request))
  );
});
*/

