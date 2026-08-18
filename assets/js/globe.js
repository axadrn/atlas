// Landing page globe. Canvas 2D, zero dependencies.
// Dotted continents from a rasterized landmask (landmask.js), nomad hubs as
// markers, a few great circle arcs. Horizontal drag spins the globe 1:1,
// vertical drag tilts a little and springs back, hover pauses the spin.
(function () {
  const canvas = document.getElementById("globe");
  if (!canvas) return;
  const ctx = canvas.getContext("2d");

  let PLACES = [];

  const D2R = Math.PI / 180;
  function toVec(lat, lon) {
    const la = lat * D2R, lo = lon * D2R;
    return [Math.cos(la) * Math.cos(lo), Math.sin(la), Math.cos(la) * Math.sin(lo)];
  }
  let placeVecs = [];

  // Land dots: a fibonacci sphere sampled against the landmask, so the dots
  // are evenly spaced with no visible latitude rows.
  const landVecs = [];
  const mask = window.ATLAS_LAND;
  if (mask) {
    const bit = (i) => (parseInt(mask.hex[i >> 2], 16) >> (i & 3)) & 1;
    const isLand = (lat, lon) => {
      const row = Math.min(mask.h - 1, Math.max(0, Math.floor((90 - lat) / mask.step)));
      const col = Math.min(mask.w - 1, Math.max(0, Math.floor((lon + 180) / mask.step)));
      return bit(row * mask.w + col);
    };
    const N = 22000;
    const GA = Math.PI * (3 - Math.sqrt(5));
    for (let i = 0; i < N; i++) {
      const y = 1 - (2 * i + 1) / N;
      const lat = Math.asin(y) / D2R;
      if (Math.abs(lat) > 84) continue;
      const lon = (((i * GA) / D2R) % 360) - 180;
      if (isLand(lat, lon)) landVecs.push(toVec(lat, lon));
    }
  }

  function slerp(a, b, t) {
    let dot = a[0] * b[0] + a[1] * b[1] + a[2] * b[2];
    dot = Math.min(1, Math.max(-1, dot));
    const th = Math.acos(dot);
    if (th < 1e-6) return a.slice();
    const s = Math.sin(th);
    const f1 = Math.sin((1 - t) * th) / s, f2 = Math.sin(t * th) / s;
    return [a[0] * f1 + b[0] * f2, a[1] * f1 + b[1] * f2, a[2] * f1 + b[2] * f2];
  }

  function themeColor(name, fallback) {
    const v = getComputedStyle(document.documentElement).getPropertyValue(name).trim();
    return v || fallback;
  }
  let colFg, colPrimary, colMuted;
  function loadColors() {
    colFg = themeColor("--foreground", "#111");
    colPrimary = themeColor("--primary", "#111");
    colMuted = themeColor("--muted-foreground", "#888");
  }
  loadColors();

  // U scales every absolute pixel value (markers, labels, arcs) with the
  // globe, so the drawing looks identical at any size.
  let W = 0, H = 0, R = 0, CX = 0, CY = 0, U = 1;
  function resize() {
    const parent = canvas.parentElement;
    // Fill the available box so the one page hero never scrolls.
    const size = Math.max(0, Math.min(parent.clientWidth, parent.clientHeight));
    const dpr = window.devicePixelRatio || 1;
    canvas.style.width = size + "px";
    canvas.style.height = size + "px";
    canvas.width = size * dpr;
    canvas.height = size * dpr;
    ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
    W = size; H = size;
    CX = W / 2; CY = H / 2;
    R = size * 0.48;
    U = size / 620;
  }
  resize();
  new ResizeObserver(resize).observe(canvas.parentElement);
  window.addEventListener("resize", resize);

  // ry spins around the axis. The tilt is fixed at a pleasant angle plus a
  // small springy offset from vertical dragging.
  const TILT = 0.3;
  const IDLE_SPIN = 0.0016;
  let ry = -0.8, vy = IDLE_SPIN, tiltOffset = 0;
  let dragging = false, lastX = 0, lastY = 0;
  let mouseX = -1, mouseY = -1;

  function rotate(v) {
    const cy = Math.cos(ry), sy = Math.sin(ry);
    const x = v[0] * cy - v[2] * sy;
    let z = v[0] * sy + v[2] * cy;
    const rx = TILT + tiltOffset;
    const cx = Math.cos(rx), sx = Math.sin(rx);
    const y = v[1] * cx - z * sx;
    z = v[1] * sx + z * cx;
    return [x, y, z];
  }
  function project(v) {
    return [CX + v[0] * R, CY - v[1] * R, v[2]];
  }

  canvas.addEventListener("pointerdown", (e) => {
    dragging = true; lastX = e.clientX; lastY = e.clientY;
    canvas.setPointerCapture(e.pointerId);
  });
  canvas.addEventListener("pointermove", (e) => {
    mouseX = e.offsetX; mouseY = e.offsetY;
    if (!dragging) return;
    const dx = e.clientX - lastX, dy = e.clientY - lastY;
    lastX = e.clientX; lastY = e.clientY;
    // 1:1 feel: dragging by the globe radius moves the surface by one radian.
    const d = dx / R;
    ry -= d;
    vy = vy * 0.7 - d * 0.3;
    tiltOffset = Math.min(0.35, Math.max(-0.35, tiltOffset + dy / R * 0.6));
  });
  canvas.addEventListener("pointerup", () => { dragging = false; });
  canvas.addEventListener("pointerleave", () => { mouseX = mouseY = -1; });
  canvas.addEventListener("click", () => {
    if (hovered >= 0) {
      window.location.href = "/places/" + encodeURIComponent(PLACES[hovered][4]);
    }
  });

  // A few arcs between random hubs.
  const arcs = [];
  function spawnArc() {
    if (PLACES.length < 2) return;
    const a = Math.floor(Math.random() * PLACES.length);
    let b = Math.floor(Math.random() * PLACES.length);
    if (b === a) b = (b + 7) % PLACES.length;
    arcs.push({ a: placeVecs[a], b: placeVecs[b], t: 0, speed: 0.004 + Math.random() * 0.003 });
  }

  function loadPlaces() {
    try {
      const element = document.getElementById("atlas-globe-places");
      if (!element) return;
      const data = JSON.parse(element.textContent);
      PLACES = data
        .map((place) => [
          place.name,
          place.latitude,
          place.longitude,
          [place.country_code, place.population ? new Intl.NumberFormat().format(place.population) + " people" : ""]
            .filter(Boolean)
            .join(" · "),
          place.slug,
          place.rank || 0,
        ]);
      placeVecs = PLACES.map((place) => toVec(place[1], place[2]));
      for (let i = 0; i < 3; i++) spawnArc();
    } catch (_) {
      // The globe remains interactive when embedded place data is unavailable.
    }
  }
  loadPlaces();

  // A quiet sphere body so the globe reads as a ball, not a point cloud.
  function drawSphere() {
    ctx.fillStyle = colMuted;
    ctx.globalAlpha = 0.06;
    ctx.beginPath();
    ctx.arc(CX, CY, R, 0, Math.PI * 2);
    ctx.fill();
    ctx.globalAlpha = 0.2;
    ctx.strokeStyle = colMuted;
    ctx.lineWidth = 1 * U;
    ctx.stroke();
    ctx.globalAlpha = 1;
  }

  function drawLand() {
    // Depth quantized into alpha buckets so the dots batch into a handful
    // of paths instead of thousands of draw calls. Small dots at low alpha
    // read as a soft land texture, not a dot grid; the hotspots carry the
    // visual weight.
    const N = 6;
    const buckets = Array.from({ length: N }, () => []);
    for (const v of landVecs) {
      const p = project(rotate(v));
      if (p[2] <= 0.02) continue;
      buckets[Math.min(N - 1, Math.floor(p[2] * N))].push(p);
    }
    const dot = W / 520;
    ctx.fillStyle = colMuted;
    for (let b = 0; b < N; b++) {
      ctx.globalAlpha = 0.12 + (b / (N - 1)) * 0.24;
      ctx.beginPath();
      for (const p of buckets[b]) {
        ctx.moveTo(p[0] + dot, p[1]);
        ctx.arc(p[0], p[1], dot, 0, Math.PI * 2);
      }
      ctx.fill();
    }
    ctx.globalAlpha = 1;
  }

  let hovered = -1;
  function drawPlaces() {
    hovered = -1;
    let best = (13 * U) ** 2;
    const pts = [];
    for (let i = 0; i < placeVecs.length; i++) {
      const p = project(rotate(placeVecs[i]));
      pts.push(p);
      if (p[2] > 0 && mouseX >= 0) {
        const d = (p[0] - mouseX) ** 2 + (p[1] - mouseY) ** 2;
        if (d < best) { best = d; hovered = i; }
      }
    }
    ctx.save();
    ctx.shadowColor = colPrimary;
    ctx.shadowBlur = 8 * U;
    for (let i = 0; i < pts.length; i++) {
      const p = pts[i];
      if (p[2] <= 0) continue;
      // Everything on the globe is a curated hotspot; the top ten lead.
      const top = PLACES[i][5] > 0 && PLACES[i][5] <= 10;
      const base = top ? 3.4 + p[2] * 1.6 : 2.4 + p[2] * 1.1;
      const r = (i === hovered ? 5.4 : base) * U;
      ctx.fillStyle = colPrimary;
      ctx.globalAlpha = 0.7 + 0.3 * p[2];
      ctx.beginPath();
      ctx.arc(p[0], p[1], r, 0, Math.PI * 2);
      ctx.fill();
    }
    ctx.restore();
    ctx.globalAlpha = 1;
    if (hovered >= 0) {
      const p = pts[hovered];
      ctx.font = `600 ${Math.round(13 * U)}px Inter, ui-sans-serif, system-ui, sans-serif`;
      ctx.fillStyle = colFg;
      const label = PLACES[hovered][0];
      const w = ctx.measureText(label).width;
      ctx.fillText(label, Math.min(Math.max(p[0] - w / 2, 4), W - w - 4), p[1] - 12 * U);
      canvas.style.cursor = "pointer";
    } else {
      canvas.style.cursor = dragging ? "grabbing" : "grab";
    }
  }

  function drawArcs() {
    ctx.lineWidth = 1.5 * U;
    ctx.strokeStyle = colPrimary;
    for (const arc of arcs) {
      arc.t += arc.speed;
      const tail = Math.max(0, arc.t - 0.3);
      let prev = null, head = null;
      for (let t = tail; t <= Math.min(arc.t, 1); t += 0.02) {
        const v = slerp(arc.a, arc.b, t);
        const lift = 1 + 0.08 * Math.sin(Math.PI * t);
        const p = project(rotate([v[0] * lift, v[1] * lift, v[2] * lift]));
        if (prev && p[2] > 0 && prev[2] > 0) {
          ctx.globalAlpha = 0.45 * p[2] * ((t - tail) / 0.3);
          ctx.beginPath();
          ctx.moveTo(prev[0], prev[1]);
          ctx.lineTo(p[0], p[1]);
          ctx.stroke();
        }
        prev = p; head = p;
      }
      if (head && head[2] > 0 && arc.t <= 1) {
        ctx.globalAlpha = 0.85 * head[2];
        ctx.fillStyle = colPrimary;
        ctx.beginPath();
        ctx.arc(head[0], head[1], 1.8 * U, 0, Math.PI * 2);
        ctx.fill();
      }
    }
    ctx.globalAlpha = 1;
    for (let i = arcs.length - 1; i >= 0; i--) {
      if (arcs[i].t > 1.4) { arcs.splice(i, 1); spawnArc(); }
    }
  }

  function frame() {
    ctx.clearRect(0, 0, W, H);
    drawSphere();
    drawLand();
    drawArcs();
    drawPlaces();

    const hovering = mouseX >= 0;
    if (!dragging) {
      // Tilt springs back to the base angle.
      tiltOffset *= 0.9;
      if (hovering) {
        // Rest while the pointer is over the globe.
        vy *= 0.85;
      } else {
        // Inertia fades into the idle spin.
        vy += (IDLE_SPIN - vy) * 0.02;
        ry += vy;
      }
    }
    requestAnimationFrame(frame);
  }

  new MutationObserver(loadColors).observe(document.documentElement, {
    attributes: true, attributeFilter: ["class", "data-theme"],
  });

  requestAnimationFrame(frame);
})();
