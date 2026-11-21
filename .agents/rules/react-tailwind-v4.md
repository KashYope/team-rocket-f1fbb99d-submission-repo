---
trigger: glob
globs: "frontend/**/*.ts, frontend/**/*.tsx, frontend/**/*.css, frontend/vite.config.ts"
description: "React + TypeScript + Tailwind CSS v4 + Wails v3 bindings rules for the ES Incident Command dashboard."
---

# React + TypeScript + Tailwind CSS v4 Rules

- **Tailwind v4 ONLY**: Use `@tailwindcss/vite` plugin in `frontend/vite.config.ts` and `@import "tailwindcss";` in `frontend/src/index.css`. Do NOT create `tailwind.config.js` or `postcss.config.js`.
- Keep `vite.config.ts` dev server bound to `host: "127.0.0.1"`, `port: Number(process.env.WAILS_VITE_PORT) || 9245`, `strictPort: true`.
- Import backend methods ONLY from generated `../bindings/...` TypeScript modules. Never use `fetch("/api/...")` or Wails v2 `wailsjs` imports.
- Single-screen 3-column command center layout (`1440x900`) with a visual Train Consist Strip:
  - **Left (25%)**: Incident Intake + 3 Demo Presets + `"Simulate Unfiltered AI (Demo Clamp)"` toggle.
  - **Center (30%)**: Laya Probability Bars + EU 2021/782 Statutory Box + Deterministic Rail Guardrail Pills + Eligible/Rejected Spare Fleet.
  - **Right (45%)**: 3 Interactive Scenario Cards (`LOWEST_COST`, `PAX_FIRST`, `PROTECT_NEXT_TRAIN`) updating the Tabbed German TCS / Passenger Alert / Audit Trail drawer with 1-click Copy.
- Surface explicit live-vs-mock source badges (`LAYA LIVE` / `LAYA MOCK`, `GEMINI LIVE` / `GEMINI MOCK`) and a visible `"Clamped by Rail Guardrail"` amber badge whenever `wasClampedByGo` is `true`.
