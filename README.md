# pdf-freezer

**pdf-freezer** is a secure, cross-platform desktop tool designed to "freeze" PDF documents. It renders pages as high-resolution images and re-assembles them into a non-editable format, ensuring the document's visual fidelity while preventing text modification. It also applies a sequential serial number overlay for accounting and audit purposes.

## Features

- **Immutable Output**: Converts vector-based PDFs into high-quality rasterized equivalents (300 DPI).
- **Audit Overlay**: Automatically stamps each document with a unique serial number (e.g., `AR0001`, `AR0002`).
- **Persistent Configuration**: Counters and settings are preserved across re-starts.
- **Enterprise Logging**: Maintains a detailed audit log in `~/Library/Application Support/pdf-freezer/app.log`.
- **Security Check**: Auto-detects dependencies and validates input paths to prevent traversals.

## Project Structure

```bash
pdf-freezer/
├── cmd/                # Main entry point (Wails bootstrap)
├── internal/
│   ├── app/            # Application logic & bridge
│   ├── config/         # Persistent settings & logging
│   ├── counter/        # Thread-safe number management
│   └── engine/         # Ghostscript wrapper & pipeline
├── frontend/           # Svelte + Vite UI
└── build/              # Output binaries
```

## Prerequisites

- **Go** 1.25+
- **Node.js** 24+
- **Ghostscript**: Required for rendering engine.
  - **macOS**: `brew install ghostscript`
  - **Windows**: Install [Ghostscript](https://ghostscript.com/releases/gsdnld.html).

## Development

1. **Clone the repository**:
   ```bash
   git clone https://github.com/your/pdf-freezer.git
   cd pdf-freezer
   ```

2. **Run in Dev Mode**:
   ```bash
   wails dev
   ```
   This starts the backend and the Svelte dev server with hot-reload.

3. **## Building

### Prerequisites
- Go 1.25+
- Node.js 20+
- Wails v3 CLI (`go install github.com/wailsapp/wails/v3/cmd/wails3@latest`)

### Build Steps

To build the application for production:

```bash
~/go/bin/wails3 build
```
This will automatically compile both the frontend and backend into the `pdf-freezer` executable.

### Troubleshooting (Manual Build)

If the automatic build fails due to environment issues, you can build manually:

1. Build Frontend:
```bash
cd frontend
npm install
npm run build
cd ..
```

2. Build Backend:
```bash
go build -o pdf-freezer ./cmd/pdf-freezer
```

3. Run:
```bash
./pdf-freezer
```

## Configuration & Logs

- **Config**: JSON settings are stored in `~/Library/Application Support/pdf-freezer/config.json` (Mac) or `%APPDATA%\pdf-freezer\config.json` (Windows).
- **Logs**: Operation logs are written to `app.log` in the same directory.

## License

MIT
