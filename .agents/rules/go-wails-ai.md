---
trigger: glob
globs: "**/*.go, go.mod, Taskfile.yml"
description: "Strict Go 1.26+, Wails v3 service bindings, Laya /v1/systemone, and google.golang.org/genai v1 structured output conventions."
---

# Go, Wails v3, Laya, and Gemini Rules

## Wails v3 Service Pattern
- Register services via `application.NewService(svc)` in `main.go`. Do NOT thread `context.Context` as a struct field for runtime calls.
- Export clean PascalCase methods on `IncidentService`: `GetPresets()`, `GetFleetState()`, `AnalyzeIncident(req domain.IncidentRequest) (domain.IncidentResponse, error)`.
- After any signature or struct change in `internal/domain` or `internal/service`, immediately run `wails3 generate bindings -ts`.

## Gemini Official Go SDK
- ALWAYS import `google.golang.org/genai`. NEVER import deprecated `github.com/google/generative-ai-go`.
- Initialize via `genai.NewClient(ctx, &genai.ClientConfig{APIKey: key, Backend: genai.BackendGeminiAPI})`.
- Default model: `os.Getenv("GEMINI_MODEL")` falling back to `"gemini-3.5-flash-lite"` (also supports `"gemini-3.8-flash"` and `"gemini-4-argon"`). NEVER use deprecated `gemini-2.5-*`, `gemini-2.0-*`, or `gemini-1.5-*`.
- Structured JSON MUST set BOTH `ResponseMIMEType: "application/json"` AND `ResponseSchema: &genai.Schema{...}` with `Temperature: genai.Ptr[float32](0.2)`.
- Keep `ResponseSchema` flat; do NOT use `minLength`, `maxLength`, or `uniqueItems`; do NOT duplicate the JSON schema inside the prompt prose.
- Enforce a 20-second `context.WithTimeout` and fall back to deterministic `gemini-mock` on missing key, timeout, or JSON parse failure.

## Laya System 1 HTTP Client
- Target `LAYA_BASE_URL` (default `http://127.0.0.1:8000`) at `POST /v1/systemone` with a 3-second timeout and optional `Authorization: Bearer $LAYA_API_KEY`.
- Pin `"model": "english"` in requests to avoid multilingual level-0 score bias.
- NEVER read or gate on `action.act_probability`. Gate `AUTO_CANDIDATE` vs `HUMAN_REVIEW` strictly using `answer_confidence >= 0.75`, `confidence >= 0.55`, `operational_severity.score < 2.2`, `choice != "manual_escalation"`, no PRM pax (`!HasPRMPax`), and `ExpectedDelayMin < 120`.
- If `LAYA_MODE=mock` or the HTTP call fails, run the deterministic Go heuristic mock returning realistic `choice`, `score`, and `noul` probabilities.
