// Minimal service worker: PWA installability plus an offline fallback page.
// No manual cache versioning. Network always wins; the cache is only used
// when the network is unreachable, so assets can never go stale.
const OFFLINE_CACHE = "atlas-offline";
const OFFLINE_URL = "/offline";
const OFFLINE_ASSETS = [
  OFFLINE_URL,
  "/assets/css/output.css",
  "/assets/icons/app-icon-192.png",
  "/assets/icons/app-icon-512.png",
];

self.addEventListener("install", function (event) {
  event.waitUntil(caches.open(OFFLINE_CACHE).then(function (cache) {
    return cache.addAll(OFFLINE_ASSETS.map(function (url) {
      return new Request(url, { cache: "reload" });
    }));
  }));
  self.skipWaiting();
});

self.addEventListener("activate", function (event) {
  // One-time cleanup of the old versioned caches (atlas-static-*).
  event.waitUntil(caches.keys().then(function (names) {
    return Promise.all(names.filter(function (name) {
      return name !== OFFLINE_CACHE;
    }).map(function (name) {
      return caches.delete(name);
    }));
  }));
  self.clients.claim();
});

self.addEventListener("fetch", function (event) {
  const request = event.request;
  if (request.method !== "GET") return;

  const url = new URL(request.url);
  if (url.origin !== self.location.origin) return;

  if (request.mode === "navigate") {
    event.respondWith(fetch(request).catch(function () {
      return caches.match(OFFLINE_URL);
    }));
    return;
  }

  // Only the offline page's own assets get a fallback, and only when
  // the network fails. Everything else goes straight to the network.
  if (OFFLINE_ASSETS.indexOf(url.pathname) === -1) return;

  event.respondWith(fetch(request).catch(function () {
    return caches.match(request);
  }));
});
