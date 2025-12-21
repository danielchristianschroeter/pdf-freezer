<script>
  import { onMount } from "svelte";
  import { Events } from "@wailsio/runtime";

  import {
    CheckDeps,
    ProcessFile,
    SelectFile,
    GetCurrentNumber,
    SetNumberOverride,
    GetConfig,
    SetPrefix,
    SetOverlayPosition,
    SetCompressionLevel,
  } from "../bindings/pdf-freezer/pkg/app/app.js";

  let status = "Ready";
  let isProcessing = false;
  let counter = 0;
  let prefix = "AR";
  let missingDeps = false;
  let showSettings = false;

  // Config options
  let overlayEnabled = true;
  let overlayPosition = "bottom-right";
  let fileSuffix = "_frozen";
  let overwriteMode = false;
  let compressionLevel = "none";

  onMount(async () => {
    try {
      await CheckDeps();
      counter = await GetCurrentNumber();

      // Load config
      try {
        const cfg = await GetConfig();
        if (cfg) {
          if (cfg.prefix) prefix = cfg.prefix;
          if (cfg.overlay_position) overlayPosition = cfg.overlay_position;
          if (cfg.file_suffix) fileSuffix = cfg.file_suffix;
          if (typeof cfg.overwrite_mode === "boolean")
            overwriteMode = cfg.overwrite_mode;
          if (cfg.compression_level) compressionLevel = cfg.compression_level;
        }
      } catch (e) {
        console.error("Config load error", e);
      }

      // Override the Wails runtime file drop handler to intercept files
      // This is called by the native layer with the file paths
      const originalHandler = window._wails?.handlePlatformFileDrop;
      if (window._wails) {
        window._wails.handlePlatformFileDrop = async (filenames, x, y) => {
          console.log("handlePlatformFileDrop intercepted:", filenames, x, y);

          // Filter for PDF files
          const pdfFiles = filenames.filter(
            (f) => typeof f === "string" && f.toLowerCase().endsWith(".pdf"),
          );

          if (pdfFiles.length > 0) {
            console.log("Processing dropped PDFs:", pdfFiles);
            await processMultiple(pdfFiles);
          }

          // Still call the original handler to maintain Wails functionality
          if (originalHandler) {
            return originalHandler.call(window._wails, filenames, x, y);
          }
        };
      }

      // Also listen for Wails file drop events as fallback
      const dropEvents = [
        "files-dropped",
        "wails:window:filedropped",
        "windows:WindowFilesDropped",
        "mac:WindowFilesDropped",
        "common:WindowFilesDropped",
      ];
      dropEvents.forEach((name) => {
        Events.On(name, async (event) => {
          console.log("Drop event received:", name, event);
          const data = event?.data || event;
          const files = Array.isArray(data)
            ? data
            : data?.filenames || data?.paths || [];
          if (Array.isArray(files) && files.length > 0) {
            const pdfPaths = files.filter(
              (p) => typeof p === "string" && p.toLowerCase().endsWith(".pdf"),
            );
            if (pdfPaths.length > 0) {
              await processMultiple(pdfPaths);
            }
          }
        });
      });
    } catch {
      status = "Ghostscript not found";
      missingDeps = true;
    }
  });

  async function handleSelect() {
    if (isProcessing || missingDeps) return;
    try {
      const file = await SelectFile();
      if (file) process(file);
    } catch (err) {
      status = "Error: " + err;
    }
  }

  async function process(path) {
    status = "Processing...";
    isProcessing = true;
    try {
      const result = await ProcessFile(
        path,
        overlayEnabled,
        prefix,
        overlayPosition,
        fileSuffix,
        overwriteMode,
        compressionLevel,
      );
      const filename = result.split("/").pop();
      status = `✓ Saved: ${filename}`;
      counter = await GetCurrentNumber();
    } catch (err) {
      status = "✗ " + err;
    } finally {
      isProcessing = false;
    }
  }

  // Process multiple files sequentially (same as clicking "Select File" for each)
  async function processMultiple(paths) {
    if (!paths || paths.length === 0) return;

    for (let i = 0; i < paths.length; i++) {
      const path = paths[i];
      status = `Processing ${i + 1}/${paths.length}...`;
      isProcessing = true;
      try {
        const result = await ProcessFile(
          path,
          overlayEnabled,
          prefix,
          overlayPosition,
          fileSuffix,
          overwriteMode,
          compressionLevel,
        );
        const filename = result.split("/").pop();
        if (i === paths.length - 1) {
          status = `✓ Saved ${paths.length} file(s)`;
        }
        counter = await GetCurrentNumber();
      } catch (err) {
        status = `✗ Error on file ${i + 1}: ` + err;
        break;
      }
    }
    isProcessing = false;
  }

  // Handle native browser drop event
  async function handleDrop(e) {
    isDragging = false;
    if (isProcessing || missingDeps) return;

    const files = Array.from(e.dataTransfer.files);
    // In Wails v3, file.path contains the absolute path
    const paths = files
      .filter((f) => f.name.toLowerCase().endsWith(".pdf"))
      .map((f) => f.path)
      .filter((p) => p); // Filter out empty paths

    if (paths.length > 0) {
      await processMultiple(paths);
    }
  }

  async function updateCounter() {
    try {
      await SetNumberOverride(counter);
      status = "Counter updated";
      setTimeout(() => (status = "Ready"), 1500);
    } catch (err) {
      status = "Error: " + err;
    }
  }

  async function savePrefix() {
    try {
      await SetPrefix(prefix);
      status = "Prefix saved";
      setTimeout(() => (status = "Ready"), 1500);
    } catch (err) {
      status = "Error: " + err;
    }
  }

  async function savePosition() {
    try {
      await SetOverlayPosition(overlayPosition);
      status = "Position saved";
      setTimeout(() => (status = "Ready"), 1500);
    } catch (err) {
      status = "Error: " + err;
    }
  }

  async function saveCompression() {
    try {
      await SetCompressionLevel(compressionLevel);
      status = "Compression saved";
      setTimeout(() => (status = "Ready"), 1500);
    } catch (err) {
      status = "Error: " + err;
    }
  }

  let isDragging = false;
</script>

<main class="app">
  <header>
    <h1>PDF Freezer</h1>
    <span class="counter">Next: {prefix}{String(counter).padStart(4, "0")}</span
    >
  </header>

  <div
    id="pdf-dropzone"
    data-wails-dropzone
    class="dropzone"
    class:dragging={isDragging}
    class:disabled={isProcessing || missingDeps}
    on:dragover|preventDefault={() => (isDragging = true)}
    on:dragleave={() => (isDragging = false)}
    on:drop|preventDefault={handleDrop}
    role="button"
    tabindex="0"
  >
    {#if isProcessing}
      <div class="spinner"></div>
      <p>Processing...</p>
    {:else if missingDeps}
      <p class="error">⚠ Ghostscript not found</p>
    {:else}
      <div class="icon">📄</div>
      <p>Drop PDF here</p>
      <button class="btn-primary" on:click={handleSelect}>Select File</button>
    {/if}
  </div>

  <div
    class="status"
    class:success={status.includes("✓")}
    class:error={status.includes("✗") || status.includes("not found")}
  >
    {status}
  </div>

  <button
    class="toggle-settings"
    on:click={() => (showSettings = !showSettings)}
  >
    {showSettings ? "▲ Hide Settings" : "▼ Settings"}
  </button>

  {#if showSettings}
    <div class="settings">
      <div class="setting-row">
        <label for="prefix">Prefix</label>
        <input
          id="prefix"
          type="text"
          bind:value={prefix}
          maxlength="10"
          on:blur={savePrefix}
        />
      </div>
      <div class="setting-row">
        <label for="counter">Counter</label>
        <input id="counter" type="number" bind:value={counter} min="1" />
        <button class="btn-sm" on:click={updateCounter}>Set</button>
      </div>
      <div class="setting-row">
        <label for="position">Position</label>
        <select
          id="position"
          bind:value={overlayPosition}
          on:change={savePosition}
        >
          <option value="top-left">Top Left</option>
          <option value="top-right">Top Right</option>
          <option value="bottom-left">Bottom Left</option>
          <option value="bottom-right">Bottom Right</option>
        </select>
      </div>
      <div class="setting-row">
        <label for="compression">Compression</label>
        <select
          id="compression"
          bind:value={compressionLevel}
          on:change={saveCompression}
        >
          <option value="none">None (High Quality)</option>
          <option value="low">Low (Good Quality)</option>
          <option value="medium">Medium (Standard)</option>
          <option value="high">High (Smallest File)</option>
        </select>
      </div>
      <div class="setting-row">
        <label for="file-suffix">File Suffix</label>
        <input
          id="file-suffix"
          type="text"
          bind:value={fileSuffix}
          maxlength="15"
          disabled={overwriteMode}
        />
      </div>
      <div class="setting-row checkbox">
        <input type="checkbox" id="overwrite" bind:checked={overwriteMode} />
        <label for="overwrite">Overwrite Original</label>
      </div>
      <div class="setting-row checkbox">
        <input type="checkbox" id="overlay" bind:checked={overlayEnabled} />
        <label for="overlay">Add Serial Overlay</label>
      </div>
    </div>
  {/if}
</main>

<style>
  :root {
    --bg: #0f0f14;
    --surface: #1a1a24;
    --border: #2d2d3a;
    --primary: #6366f1;
    --primary-glow: rgba(99, 102, 241, 0.15);
    --accent: #8b5cf6;
    --text: #f1f5f9;
    --muted: #94a3b8;
    --success: #10b981;
    --error: #ef4444;
  }

  * {
    box-sizing: border-box;
    margin: 0;
    padding: 0;
  }

  .app {
    background: var(--bg);
    min-height: 100vh;
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
    color: var(--text);
    display: flex;
    flex-direction: column;
    padding: 1rem;
    gap: 0.75rem;
  }

  header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 0.5rem 0;
  }

  h1 {
    font-size: 1.1rem;
    font-weight: 600;
  }

  .counter {
    font-size: 0.85rem;
    color: var(--primary);
    font-weight: 500;
  }

  .dropzone {
    flex: 1;
    min-height: 160px;
    border: 2px dashed var(--border);
    border-radius: 12px;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 0.75rem;
    transition: all 0.2s;
    cursor: pointer;
  }

  .dropzone:hover,
  .dropzone.dragging {
    border-color: var(--primary);
    background: var(--primary-glow);
  }

  .dropzone.disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }

  .dropzone .icon {
    font-size: 2.5rem;
  }

  .dropzone p {
    color: var(--muted);
    font-size: 0.9rem;
  }

  .spinner {
    width: 32px;
    height: 32px;
    border: 3px solid var(--border);
    border-top-color: var(--primary);
    border-radius: 50%;
    animation: spin 1s linear infinite;
  }

  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }

  .btn-primary {
    background: var(--primary);
    color: white;
    border: none;
    padding: 0.6rem 1.2rem;
    border-radius: 6px;
    font-size: 0.85rem;
    font-weight: 500;
    cursor: pointer;
    transition: opacity 0.2s;
  }

  .btn-primary:hover {
    opacity: 0.9;
  }

  .btn-sm {
    background: var(--surface);
    border: 1px solid var(--border);
    color: var(--text);
    padding: 0.35rem 0.6rem;
    border-radius: 4px;
    font-size: 0.75rem;
    cursor: pointer;
  }

  .status {
    text-align: center;
    padding: 0.5rem;
    font-size: 0.8rem;
    color: var(--muted);
    border-radius: 6px;
    background: var(--surface);
  }

  .status.success {
    color: var(--success);
    background: rgba(52, 211, 153, 0.1);
  }

  .status.error {
    color: var(--error);
    background: rgba(248, 113, 113, 0.1);
  }

  .toggle-settings {
    background: none;
    border: none;
    color: var(--muted);
    font-size: 0.75rem;
    cursor: pointer;
    padding: 0.25rem;
  }

  .toggle-settings:hover {
    color: var(--text);
  }

  .settings {
    background: var(--surface);
    border-radius: 8px;
    padding: 0.75rem;
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .setting-row {
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }

  .setting-row label {
    font-size: 0.75rem;
    color: var(--muted);
    min-width: 70px;
  }

  .setting-row input[type="text"],
  .setting-row input[type="number"],
  .setting-row select {
    flex: 1;
    background: var(--bg);
    border: 1px solid var(--border);
    color: var(--text);
    padding: 0.35rem 0.5rem;
    border-radius: 4px;
    font-size: 0.8rem;
  }

  .setting-row.checkbox {
    flex-direction: row-reverse;
    justify-content: flex-end;
  }

  .setting-row.checkbox label {
    min-width: auto;
  }

  .setting-row input[type="checkbox"] {
    width: 16px;
    height: 16px;
    accent-color: var(--primary);
  }

  .error {
    color: var(--error);
  }
</style>
