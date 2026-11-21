# ES Incident Command — Project Invariants

## Mission
Build a high-reliability Proof of Concept desktop application ("ES Incident Command") for European Sleeper night train disruption management across three corridors:
- **Brussels–Prague** (`ES 453` / `ES 452`)
- **Brussels–Milan Porta Garibaldi** (`ES 401` / `ES 400` via Gotthard & Aarau)
- **Paris Nord–Brussels–Berlin** (`ES 475` / `ES 474` via Aulnoye-Aymeries, Bruxelles-Midi & Hamburg-Harburg)

## Non-Negotiable Architecture Rules
1. **Wails v3 ONLY**: Import `github.com/wailsapp/wails/v3/pkg/application` and use the `wails3` CLI (`v3.1.0+`). NEVER use Wails v2 (`wails/v2`, `wails.json`, `frontend/wailsjs`).
2. **Dual-Mode Entrypoint**: `main.go` MUST configure `Server: application.ServerOptions{Host: "localhost", Port: 8080}` alongside the native Webview window (`1440x900`) and expose `GET /health` (`{"status":"ok"}`) so `go build -tags server` serves the embedded UI + live Go bindings at `http://localhost:8080` for Chrome `/browser` verification.
3. **Backend-Only AI Calls**: React NEVER calls Laya or Gemini directly. All AI calls, API keys (`GEMINI_API_KEY`, `LAYA_API_KEY`), mock fallbacks, and rail guardrails live strictly in Go under `internal/`.
4. **Offline Fallback**: The app MUST operate end-to-end with zero external API keys set, automatically falling back to deterministic realistic mocks for Laya (`laya-mock`) and Gemini (`gemini-mock`) and displaying the active source badge in the UI.
5. **Secrets Safety**: Never commit `.env` or hardcode API keys. Provide `.env.example` with blank secret placeholders only.
6. **Verification Before Done**: Never mark a stage complete without running the stage's exact terminal or `/browser` verification commands and confirming zero errors.
