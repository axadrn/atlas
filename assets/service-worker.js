const CACHE_NAME = "atlas-static-v1";
const OFFLINE_URL = "/offline";
const PRECACHE = [
  OFFLINE_URL,
  "/assets/css/output.css",
  "/assets/icons/app-icon-192-v1.png",
  "/assets/icons/app-icon-512-v1.png",
];

self.addEventListener("install", function (event) {
  event.waitUntil(caches.open(CACHE_NAME).then(function (cache) {
    return cache.addAll(PRECACHE);
  }));
  self.skipWaiting();
});

self.addEventListener("activate", function (event) {
  event.waitUntil(caches.keys().then(function (names) {
    return Promise.all(names.filter(function (name) {
      return name.startsWith("atlas-static-") && name !== CACHE_NAME;
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

  if (!url.pathname.startsWith("/assets/") && !url.pathname.startsWith("/components/")) return;

  event.respondWith(caches.match(request).then(function (cached) {
    if (cached) return cached;
    return fetch(request).then(function (response) {
      if (!response.ok) return response;
      const copy = response.clone();
      caches.open(CACHE_NAME).then(function (cache) {
        cache.put(request, copy);
      });
      return response;
    });
  }));
});
