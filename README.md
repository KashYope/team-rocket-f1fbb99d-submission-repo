# ES Incident Command — Night Train Ops PoC

Disruption decision-support desktop & web application for **European Sleeper** night train operations across three November 2026 corridors:
- **ES 401 / 400**: Brussels – Milano Porta Garibaldi (via historic Gotthard mountain line & Aarau)
- **ES 453 / 452**: Brussels – Praha hl.n. (via Amsterdam, Bad Bentheim, Berlin & Dresden)
- **ES 475 / 474**: Paris Nord – Brussels – Berlin (via Aulnoye-Aymeries, Bruxelles-Midi & Hamburg-Harburg)

## Architecture
- **Shell**: Wails v3 (`v3.1.0`, Go `1.26+`) dual-mode architecture — compiles into both a native Windows WebView2 `.exe` (`1440x900`) and a headless HTTP/WebSocket server (`-tags server` at `http://localhost:8080` with `/health`).
- **Frontend**: React + TypeScript + Vite + Tailwind CSS v4 (`@tailwindcss/vite`), wired strictly through generated Wails v3 TypeScript bindings (`frontend/bindings/`).
- **4-Stage Backend Pipeline (`internal/`)**:
  1. **Deterministic Pre-Checks (`internal/guardrails/engine.go`)**: Italy/France rolling-stock homologation (`EE` & `GFF 309` blocked on Milan; `Bdcm36` blocked on Paris), corridor car caps (`13` CH/IT/FR, `15` BE/NL, `16` DE), 8-car single-conductor cliff (`+€3,000` Milan / `+€1,600` Prague & Paris), Pet Private / Women-Only / PRM protections, and statutory **Regulation (EU) 2021/782** Art. 18–20 calculator.
  2. **System 1 Triage (`internal/laya/client.go`)**: Laya `POST /v1/systemone` (`"model": "english"`, gated on `answerConfidence` & Shannon entropy confidence, never `act_probability`) with automatic `laya-mock` fallback.
  3. **System 2 Synthesis (`internal/gemini/client.go`)**: `google.golang.org/genai` (`v1.79.1`) constrained JSON decoding (`gemini-3.5-flash-lite` default) generating 3 interactive scenarios (`LOWEST_COST`, `PAX_FIRST`, `PROTECT_NEXT_TRAIN`), German TCS shunting orders (`abholen`/`abstellen`/`beistellen`), and EU 2021/782 passenger advisories, with `gemini-mock` offline fallback.
  4. **Deterministic Post-Clamp (`internal/guardrails/clamp.go`)**: Scans AI outputs and clamps any non-homologated UIC/coach to approved `RDC` stock (`WasClampedByGo = true`).

## Quick Start (Windows PowerShell)

### 1. Configure Environment (Optional — runs 100% offline without keys)
Copy `.env.example` to `.env` and set `GEMINI_API_KEY` if testing live Gemini:
```powershell
Copy-Item .env.example .env
```

### 2. Run Go Unit & Guardrail Tests
```powershell
go test -v ./...
```

### 3. Build Native Desktop `.exe` & Server Binary
```powershell
wails3 build
go build -tags server -o bin/es-incident-server.exe .
```

### 4. Launch Modes
- **Native Desktop Window (uses `.env` if present, auto-mocks otherwise)**:
  ```powershell
  .\bin\es-incident-command.exe
  ```
- **Browser / Server Mode (`http://localhost:8080`)**:
  ```powershell
  .\bin\es-incident-server.exe
  ```
- **Pure Zero-WiFi Offline Mock Mode**:
  ```powershell
  $env:LAYA_MODE="mock"; $env:GEMINI_MODE="mock"; .\bin\es-incident-command.exe
  ```
