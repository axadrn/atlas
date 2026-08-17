(() => {
  "use strict";

  const workerURL = "/assets/vendor/maplibre-gl-csp-worker-5.24.0.js";
  const storageKey = "atlas-map-style";
  const styles = {
    liberty: "https://tiles.openfreemap.org/styles/liberty",
    positron: "https://tiles.openfreemap.org/styles/positron",
    dark: "https://tiles.openfreemap.org/styles/dark",
  };

  function storedStyle() {
    try {
      const value = localStorage.getItem(storageKey);
      if (styles[value]) return value;
    } catch (_) {
      // Storage can be disabled without disabling the map.
    }
    return document.documentElement.classList.contains("dark") ? "dark" : "liberty";
  }

  function number(root, name) {
    const value = Number.parseFloat(root.dataset[name]);
    return Number.isFinite(value) ? value : null;
  }

  function initialize(root) {
    const canvas = root.querySelector("[data-map-canvas]");
    const styleSelect = root.querySelector("[data-map-style]");
    if (!canvas || !styleSelect) return;

    const styleName = storedStyle();
    styleSelect.value = styleName;

    const options = {
      container: canvas,
      style: styles[styleName],
      attributionControl: false,
      cooperativeGestures: true,
      maplibreLogo: false,
    };

    const west = number(root, "west");
    const south = number(root, "south");
    const east = number(root, "east");
    const north = number(root, "north");
    const longitude = number(root, "longitude");
    const latitude = number(root, "latitude");

    if ([west, south, east, north].every((value) => value !== null)) {
      options.bounds = [[west, south], [east, north]];
      options.fitBoundsOptions = { padding: 48, maxZoom: 7 };
    } else if (longitude !== null && latitude !== null) {
      options.center = [longitude, latitude];
      options.zoom = 10;
    } else {
      return;
    }

    const map = new maplibregl.Map(options);
    map.addControl(new maplibregl.NavigationControl({ showCompass: false }), "top-left");
    map.addControl(new maplibregl.AttributionControl({ compact: true }), "bottom-right");

    if (longitude !== null && latitude !== null) {
      new maplibregl.Marker({ color: "#0f766e" })
        .setLngLat([longitude, latitude])
        .addTo(map);
    }

    styleSelect.addEventListener("change", () => {
      const nextStyle = styleSelect.value;
      if (!styles[nextStyle]) return;
      map.setStyle(styles[nextStyle]);
      try {
        localStorage.setItem(storageKey, nextStyle);
      } catch (_) {
        // The selected style still works for the current page.
      }
    });
  }

  document.addEventListener("DOMContentLoaded", () => {
    if (typeof maplibregl === "undefined") return;
    maplibregl.setWorkerUrl(workerURL);
    document.querySelectorAll("[data-place-map]").forEach(initialize);
  });
})();
