"use strict";

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

  const gain = ctx.createGain();
  gain.connect(ensureGainNode());
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

function initializeInteractions() {
  // Wire click sound to action buttons (skip audioToggleButton — that's the synth)
  ['addSomethingButton', 'clearAsideButton', 'refreshButton', 'drumPlayButton'].forEach(function (id) {
    const btn = document.getElementById(id);
    if (btn) btn.addEventListener('click', playClickSound);
  });

  // Touch / Click interaction for the physics canvas (canvas two)
  const canvasTwoDiv = document.getElementById("canvasTwoDiv");
  if (canvasTwoDiv) {
    let canvas = null;
    const handleInteraction = (clientX, clientY) => {
      if (!canvas) canvas = canvasTwoDiv.querySelector("canvas");
      if (!canvas) return;
      const rect = canvas.getBoundingClientRect();
      if (rect.width === 0 || rect.height === 0) return;
      // Map from CSS pixels to canvas pixel space
      const scaleX = canvas.width / rect.width;
      const scaleY = canvas.height / rect.height;
      const x = Math.round((clientX - rect.left) * scaleX);
      const y = Math.round((clientY - rect.top) * scaleY);
      go_set_interaction(x, y);
      go_invoke_callback(4); // 4 = onCanvasInteraction
    };

    canvasTwoDiv.addEventListener("mousedown", (e) => {
      if (e.target.tagName === "CANVAS") {
        e.preventDefault();
        handleInteraction(e.clientX, e.clientY);
      }
    });

    canvasTwoDiv.addEventListener("touchstart", (e) => {
      if (e.target.tagName === "CANVAS") {
        e.preventDefault();
        const touch = e.touches[0];
        handleInteraction(touch.clientX, touch.clientY);
      }
    }, { passive: false });
  }

  // Touch / Click interaction for the sequencer canvas (canvas three)
  const canvasThreeDiv = document.getElementById("canvasThreeDiv");
  if (canvasThreeDiv) {
    let canvas = null;
    const handleInteraction = (clientX, clientY) => {
      if (!canvas) canvas = canvasThreeDiv.querySelector("canvas");
      if (!canvas) return;
      const rect = canvas.getBoundingClientRect();
      if (rect.width === 0 || rect.height === 0) return;
      // Map from CSS pixels to canvas pixel space
      const scaleX = canvas.width / rect.width;
      const scaleY = canvas.height / rect.height;
      const x = Math.round((clientX - rect.left) * scaleX);
      const y = Math.round((clientY - rect.top) * scaleY);
      go_set_interaction(x, y);
      go_invoke_callback(6); // 6 = onDrumCanvasClick
    };

    canvasThreeDiv.addEventListener("mousedown", (e) => {
      if (e.target.tagName === "CANVAS") {
        e.preventDefault();
        handleInteraction(e.clientX, e.clientY);
      }
    });

    canvasThreeDiv.addEventListener("touchstart", (e) => {
      if (e.target.tagName === "CANVAS") {
        e.preventDefault();
        const touch = e.touches[0];
        handleInteraction(touch.clientX, touch.clientY);
      }
    }, { passive: false });
  }

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
