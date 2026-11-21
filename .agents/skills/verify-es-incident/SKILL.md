---
name: verify-es-incident
description: Builds and verifies ES Incident Command end-to-end using Go unit tests, Wails v3 TypeScript binding generation, Wails server-mode HTTP host on localhost:8080, and Antigravity browser QA across all three European Sleeper demo presets. Use after backend or UI milestones.
---

# Verify ES Incident Command End-to-End

Execute these steps sequentially. Stop and fix any failure before advancing.

1. **Go Backend & Guardrail Tests**
   Run `go test -v ./...` and confirm all rail guardrail, EU 2021/782, Laya mock, and Gemini post-clamp tests pass.

2. **Regenerate Bindings & Build Frontend**
   Run `wails3 generate bindings -ts` from the workspace root, then run `npm --prefix frontend run build` and confirm zero TypeScript or Vite errors.

3. **Build & Start Wails Server Mode**
   Run `go build -tags server -o bin/es-incident-server.exe .` and start `./bin/es-incident-server.exe` in a background task. Poll `http://localhost:8080/health` until it returns `{"status":"ok"}`.

4. **Browser Subagent QA (`/browser`)**
   Open `http://localhost:8080` at a `1440x900` viewport with the browser subagent and verify:
   - Zero uncaught console errors or failed network calls.
   - **Preset 1 (Milan Flat Wheel)**: Verify `EE Bvcmz6` and `GFF Bimz 309` appear as BLOCKED for Italy, 25% EU 2021/782 compensation displays, Laya bars render, 3 scenario cards render, clicking cards updates the TCS German Order tab, and 1-click Copy works.
   - **Preset 2 (Prague Full Cancellation)**: Verify `HUMAN_REVIEW` badge, 50% EU 2021/782 band, 100-min self-rerouting warning, non-exempt `INFRA_MANAGER` status, and PRM Art. 20(5) priority flag.
   - **Preset 3 (Paris–Brussels–Berlin Heating & Shunt)**: Verify `RDC Bdcm36` blocked for France, Women-Only protection note, Paris Le Bourget shunt warning, and +€1,600 conductor cliff note.
   - **Live Clamp Demo**: Toggle `"Simulate Unfiltered AI (Demo Clamp)"` on Preset 1 and confirm the `"Clamped by Rail Guardrail"` badge appears.
   - Capture screenshots of verified preset states for the Walkthrough artifact.

5. **Teardown & Native Desktop Check**
   Stop the background `es-incident-server` process, then build the native Windows desktop binary into `bin/`.
