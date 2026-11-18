"use strict";

// Callback IDs for go_invoke_callback (JS to Go canvas pointer bridge).
// Keep the values in sync with the cb* constants in demo/demo.go.
const CALLBACKS = Object.freeze({
  CANVAS_INTERACT: 0,
  DRUM_CANVAS_CLICK: 1,
  CANVAS_FOUR_INTERACT: 2,
});

// Retained draw-command buffer dispatcher (canvas_cmd package). Walks a packed byte buffer
// produced by Go and replays each op onto the 2D context. Wire format documented in
// canvas_cmd/canvas_cmd.go — one JS call per frame. `ctx` is the live CanvasRenderingContext2D;
// `bytes` is a Uint8Array holding the packed commands.
const _cmdGradients = new Map();
const _cmdSprites = new Map();
const _cmdDecoder = new TextDecoder();

// Must be a globalThis property: app.js is loaded as an ES module, so plain top-level
// functions are module-scoped and invisible to Go's js.Global().Get().
globalThis.domCanvasCmdFlush = function (ctx, bytes) {
  if (!ctx || !bytes || bytes.length === 0) return;
  const view = new DataView(bytes.buffer, bytes.byteOffset, bytes.byteLength);
  let p = 0;
  const f = () => { const v = view.getFloat32(p, true); p += 4; return v; };
  const u8 = () => bytes[p++];
  const u16 = () => { const v = view.getUint16(p, true); p += 2; return v; };
  const str = () => {
    const n = view.getUint16(p, true); p += 2;
    const s = _cmdDecoder.decode(new Uint8Array(bytes.buffer, bytes.byteOffset + p, n));
    p += n;
    return s;
  };
  // `c` is the active target: the flush context, or a sprite's offscreen context while baking.
  // Gradients and sprites persist across frames, keyed by app-chosen ids.
  const mainCtx = ctx;
  let c = mainCtx;
  const grads = _cmdGradients;
  grads.clear();
  const sprites = _cmdSprites;
  try {
    while (p < bytes.length) {
      switch (bytes[p++]) {
        case 0x01: c.fillStyle = str(); break;
        case 0x02: c.strokeStyle = str(); break;
        case 0x03: c.lineWidth = f(); break;
        case 0x04: c.font = str(); break;
        case 0x05: c.textAlign = str(); break;
        case 0x06: c.textBaseline = str(); break;
        case 0x07: c.globalAlpha = f(); break;
        case 0x08: c.lineCap = str(); break;
        case 0x10: c.fillRect(f(), f(), f(), f()); break;
        case 0x11: c.strokeRect(f(), f(), f(), f()); break;
        case 0x12: c.clearRect(f(), f(), f(), f()); break;
        case 0x20: c.beginPath(); break;
        case 0x21: c.moveTo(f(), f()); break;
        case 0x22: c.lineTo(f(), f()); break;
        case 0x23: c.closePath(); break;
        case 0x24: { const x = f(), y = f(), r = f(), a0 = f(), a1 = f(); c.arc(x, y, r, a0, a1, u8() !== 0); break; }
        case 0x25: { const x = f(), y = f(), rx = f(), ry = f(), rot = f(), a0 = f(), a1 = f(); c.ellipse(x, y, rx, ry, rot, a0, a1, u8() !== 0); break; }
        case 0x26: c.rect(f(), f(), f(), f()); break;
        case 0x27: c.bezierCurveTo(f(), f(), f(), f(), f(), f()); break;
        case 0x28: c.fill(); break;
        case 0x29: c.stroke(); break;
        case 0x2A: c.clip(); break;
        case 0x30: { const s = str(); c.fillText(s, f(), f()); break; }
        case 0x31: { const s = str(); c.strokeText(s, f(), f()); break; }
        case 0x40: c.save(); break;
        case 0x41: c.restore(); break;
        case 0x42: c.translate(f(), f()); break;
        case 0x43: c.scale(f(), f()); break;
        case 0x44: c.rotate(f()); break;
        case 0x45: c.setTransform(f(), f(), f(), f(), f(), f()); break;
        case 0x46: c.resetTransform(); break;
        case 0x50: { const id = u16(); grads.set(id, c.createLinearGradient(f(), f(), f(), f())); break; }
        case 0x51: { const id = u16(); grads.set(id, c.createRadialGradient(f(), f(), f(), f(), f(), f())); break; }
        case 0x52: { const g = grads.get(u16()); const o = f(); const col = str(); if (g) g.addColorStop(o, col); break; }
        case 0x53: { const g = grads.get(u16()); if (g) c.fillStyle = g; break; }
        case 0x54: { const g = grads.get(u16()); if (g) c.strokeStyle = g; break; }
        case 0x60: { const s = sprites.get(u16()); const dx = f(), dy = f(); if (s) c.drawImage(s.canvas, dx, dy); break; }
        case 0x61: { const s = sprites.get(u16()); const dx = f(), dy = f(), dw = f(), dh = f(); if (s) c.drawImage(s.canvas, dx, dy, dw, dh); break; }
        case 0x62: { const s = sprites.get(u16()); const sx = f(), sy = f(), sw = f(), sh = f(), dx = f(), dy = f(), dw = f(), dh = f(); if (s) c.drawImage(s.canvas, sx, sy, sw, sh, dx, dy, dw, dh); break; }
        case 0x70: { c.shadowColor = str(); c.shadowBlur = f(); break; }
        case 0x71: { c.shadowColor = "rgba(0,0,0,0)"; c.shadowBlur = 0; break; }
        case 0x80: {
          const id = u16(), w = u16(), h = u16();
          let s = sprites.get(id);
          if (!s || s.canvas.width !== w || s.canvas.height !== h) {
            const cv = document.createElement("canvas");
            cv.width = w; cv.height = h;
            s = { canvas: cv, ctx: cv.getContext("2d") };
            sprites.set(id, s);
          } else {
            s.canvas.width = w; // resets all context state (transforms, styles, clip) and clears bitmap
          }
          c = s.ctx;
          break;
        }
        case 0x81: c = mainCtx; break;
        default: return; // unknown opcode — bail rather than desync
      }
    }
  } catch (err) {
    console.warn("domCanvasCmdFlush error:", err);
  }
};

let audioCtx = null;
let synthNode = null;
let gainNode = null;
let audioPlaying = false;
let globalVolume = 0.2;

// Cached DOM elements
let volumeSlider = null;
let volumeLabel = null;

// Cached audio buffers
let clickAudioBuffer = null;
let noiseBuffer = null;

// Pooled GainNodes for drum hits — avoids allocating a new node per hit
const DRUM_GAIN_POOL_SIZE = 8;
let drumGainPool = null;
let drumGainIdx = 0;

function ensureDrumGainPool() {
  if (!drumGainPool) {
    const ctx = ensureAudioCtx();
    drumGainPool = Array.from({ length: DRUM_GAIN_POOL_SIZE }, () => {
      const g = ctx.createGain();
      g.connect(ensureGainNode());
      return g;
    });
  }
  return drumGainPool;
}

function ensureAudioCtx() {
  if (!audioCtx) {
    audioCtx = new (window.AudioContext || window.webkitAudioContext)();
  }
  return audioCtx;
}

function ensureGainNode() {
  const ctx = ensureAudioCtx();
  if (!gainNode) {
    gainNode = ctx.createGain();
    gainNode.gain.setValueAtTime(globalVolume, ctx.currentTime);
    gainNode.connect(ctx.destination);
  }
  return gainNode;
}

function updateVolumeUI() {
  if (!volumeSlider) volumeSlider = document.getElementById("volumeSlider");
  if (!volumeLabel) volumeLabel = document.getElementById("volumeLabel");
  if (volumeSlider) volumeSlider.value = globalVolume;
  if (volumeLabel) {
    const pct = Math.round(globalVolume * 100);
    volumeLabel.innerText = globalVolume === 0 ? "Vol: OFF" : `Vol: ${pct}%`;
  }
}

function applyVolume() {
  if (gainNode && audioCtx) {
    gainNode.gain.setValueAtTime(globalVolume, audioCtx.currentTime);
  }
}

function adjustVolume(delta) {
  globalVolume = Math.max(0, Math.min(1, Math.round((globalVolume + delta) * 100) / 100));
  applyVolume();
  updateVolumeUI();
}

globalThis.changeVolume = function (val) {
  globalVolume = parseFloat(val);
  applyVolume();
  updateVolumeUI();
};

globalThis.toggleAudio = async function () {
  const ctx = ensureAudioCtx();

  if (!synthNode) {
    // Load the AudioWorkletProcessor module (dedicated audio thread)
    await ctx.audioWorklet.addModule('synth-worklet.js');

    // Create the synth AudioWorkletNode and connect through shared gain
    synthNode = new AudioWorkletNode(ctx, 'synth-worklet');
    synthNode.connect(ensureGainNode());
  }

  if (ctx.state === "suspended") {
    await ctx.resume();
  }

  audioPlaying = !audioPlaying;

  // Send play state to the worklet's dedicated audio thread
  synthNode.port.postMessage({ type: 'play', value: audioPlaying });

  const btn = document.getElementById("audioToggleButton");
  if (btn) {
    btn.innerText = audioPlaying ? "🔊 Mute Soundtrack" : "🔇 Play Soundtrack";
    btn.classList.toggle("playing", audioPlaying);
  }
};

// Click sound — pre-rendered 50ms UI click from Go WASM

globalThis.playClickSound = function () {
  const ctx = ensureAudioCtx();
  if (ctx.state === "suspended") {
    ctx.resume();
  }

  // Lazy AudioBuffer creation from pre-rendered WASM samples
  if (!clickAudioBuffer) {
    const samples = go_get_click_buffer();
    clickAudioBuffer = ctx.createBuffer(1, samples.length, 44100);
    clickAudioBuffer.copyToChannel(samples, 0);
  }

  const source = ctx.createBufferSource();
  source.buffer = clickAudioBuffer;
  source.connect(ensureGainNode()); // route through shared volume gain
  source.start();
};

// Drum machine — tone map per track
const DRUM_FREQS = [60, 200, 8000, 6000, 1200, 400]; // kick, snare, hh-c, hh-o, clap, rim
const DRUM_TYPES = ["sine", "triangle", "square", "square", "noise", "noise"];

globalThis.drumHit = function (track) {
  const ctx = ensureAudioCtx();
  if (ctx.state === "suspended") {
    ctx.resume();
  }

  const pool = ensureDrumGainPool();
  const gain = pool[drumGainIdx];
  drumGainIdx = (drumGainIdx + 1) % DRUM_GAIN_POOL_SIZE;
  gain.gain.cancelScheduledValues(ctx.currentTime);
  gain.gain.setValueAtTime(0.4, ctx.currentTime);
  gain.gain.exponentialRampToValueAtTime(0.001, ctx.currentTime + 0.08);

  if (DRUM_TYPES[track] === "noise") {
    if (!noiseBuffer) {
      noiseBuffer = ctx.createBuffer(1, Math.floor(ctx.sampleRate * 0.05), ctx.sampleRate);
      const data = noiseBuffer.getChannelData(0);
      for (let i = 0; i < data.length; i += 1) data[i] = Math.random() * 2 - 1;
    }

    const src = ctx.createBufferSource();
    src.buffer = noiseBuffer;
    src.connect(gain);
    src.start();
    src.stop(ctx.currentTime + 0.06);
    return;
  }

  const osc = ctx.createOscillator();
  osc.type = DRUM_TYPES[track] || "sine";
  osc.frequency.setValueAtTime(DRUM_FREQS[track] || 220, ctx.currentTime);
  osc.frequency.exponentialRampToValueAtTime(20, ctx.currentTime + 0.06);
  osc.connect(gain);
  osc.start();
  osc.stop(ctx.currentTime + 0.1);
};

// Arrow key volume control (↑ = +1%, ↓ = −1%) & Space crash
document.addEventListener('keydown', function (e) {
  if (e.key === 'ArrowUp') {
    e.preventDefault();
    adjustVolume(+0.01);
  } else if (e.key === 'ArrowDown') {
    e.preventDefault();
    adjustVolume(-0.01);
  } else if (e.code === 'Space') {
    // Easter egg: crash cymbal accent
    if (synthNode && synthNode.port) {
      e.preventDefault();
      synthNode.port.postMessage({ type: 'crash' });
    }
  }
});

function bindCanvasInteraction(containerId, callbackId, allowDrag) {
  const container = document.getElementById(containerId);
  if (!container) return;

  let canvas = null;
  let isDragging = false;
  let cachedRect = null;

  window.addEventListener("resize", () => { cachedRect = null; });
  window.addEventListener("scroll", () => { cachedRect = null; }, { passive: true });

  const handlePoint = (clientX, clientY) => {
    if (!canvas) canvas = container.querySelector("canvas");
    if (!canvas) return;
    if (!cachedRect) cachedRect = canvas.getBoundingClientRect();
    if (cachedRect.width === 0 || cachedRect.height === 0) return;

    const scaleX = canvas.width / cachedRect.width;
    const scaleY = canvas.height / cachedRect.height;
    const x = Math.round((clientX - cachedRect.left) * scaleX);
    const y = Math.round((clientY - cachedRect.top) * scaleY);
    go_invoke_callback(callbackId, x, y);
  };

  container.addEventListener("mousedown", (e) => {
    if (e.target.tagName === "CANVAS") {
      e.preventDefault();
      if (allowDrag) isDragging = true;
      handlePoint(e.clientX, e.clientY);
    }
  });

  if (allowDrag) {
    container.addEventListener("mousemove", (e) => {
      if (isDragging) {
        e.preventDefault();
        handlePoint(e.clientX, e.clientY);
      }
    });
    document.addEventListener("mouseup", () => { isDragging = false; });
  }

  container.addEventListener("touchstart", (e) => {
    if (e.target.tagName === "CANVAS") {
      e.preventDefault();
      const touch = e.touches[0];
      handlePoint(touch.clientX, touch.clientY);
    }
  }, { passive: false });

  if (allowDrag) {
    container.addEventListener("touchmove", (e) => {
      if (e.target.tagName === "CANVAS") {
        e.preventDefault();
        const touch = e.touches[0];
        handlePoint(touch.clientX, touch.clientY);
      }
    }, { passive: false });
  }
}

function initializeInteractions() {
  // Wire click sound to action buttons (skip audioToggleButton, which controls the synth)
  ["addSomethingButton", "clearAsideButton", "refreshButton", "drumPlayButton", "gravityButton"].forEach((id) => {
    const btn = document.getElementById(id);
    if (btn) btn.addEventListener("click", playClickSound);
  });

  bindCanvasInteraction("canvasTwoDiv", CALLBACKS.CANVAS_INTERACT, true);
  bindCanvasInteraction("canvasThreeDiv", CALLBACKS.DRUM_CANVAS_CLICK, false);
  bindCanvasInteraction("canvasFourDiv", CALLBACKS.CANVAS_FOUR_INTERACT, false);

  // Show controls as flex
  const controls = document.getElementById("controls");
  if (controls) controls.style.display = "flex";
}

// Instantiate the Go WASM module
const go = new Go();
WebAssembly.instantiateStreaming(fetch("godom.wasm"), go.importObject).then((result) => {
  document.getElementById("loading").style.display = "none";
  go.run(result.instance);
  initializeInteractions();
}).catch((err) => {
  console.error("Godom failed to load:", err);
  document.getElementById("loading").innerText = "Failed to load WASM: " + err.message;
});
