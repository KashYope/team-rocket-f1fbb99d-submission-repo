package service

import (
	"context"
	"fmt"
	"time"

	"es-incident-command/internal/data"
	"es-incident-command/internal/domain"
	"es-incident-command/internal/gemini"
	"es-incident-command/internal/guardrails"
	"es-incident-command/internal/laya"

	"github.com/joho/godotenv"
)

type IncidentService struct{}

func NewIncidentService() *IncidentService {
	_ = godotenv.Load()
	return &IncidentService{}
}

func (s *IncidentService) GetPresets() []domain.DemoPreset {
	return data.GetPresets()
}

func (s *IncidentService) GetFleetState() domain.FleetState {
	return data.GetFleetState()
}

func (s *IncidentService) AnalyzeIncident(req domain.IncidentRequest) (domain.IncidentResponse, error) {
	_ = godotenv.Load()
	ctx := context.Background()
	started := time.Now()

	corridor := data.FindCorridor(req.Route)
	consist := data.BuildActiveConsist(req)
	spares := data.GetSpareCars()

	var audit []string
	audit = append(audit, fmt.Sprintf(
		"[STAGE 1 · PRE-CHECKS] Train ES %s (%s) | Issue=%s | Cause=%s | Delay=%dm | Pax=%d | UnfilteredDemo=%t",
		req.TrainID, req.Route, req.IssueType, req.CauseCategory, req.ExpectedDelayMin, req.AffectedPax, req.SimulateUnfilteredAI,
	))

	// Stage 1: Deterministic Rail Guardrails & EU 2021/782 Statutory Pre-Checks
	checks, eligible, rejected, eu := guardrails.EvaluatePreChecks(req, spares)
	audit = append(audit, fmt.Sprintf(
		"[STAGE 1 · HOMOLOGATION] %d eligible spare(s), %d rejected spare(s) for %s corridor",
		len(eligible), len(rejected), req.Route,
	))
	audit = append(audit, fmt.Sprintf(
		"[STAGE 1 · EU 2021/782] Band=%d%% (Est. €%.0f) | ForceMajeureExempt=%t | 100mSelfReroute=%t | Hotel=%t | PRMPriority=%t",
		eu.CompensationBandPct, eu.EstimatedRefundEUR, eu.ForceMajeureExempt, eu.SelfReroute100Min, eu.HotelRequired, eu.PRMPriorityRequired,
	))

	// Stage 2: Laya System 1 Non-Autoregressive Triage (/v1/systemone or laya-mock)
	layaTriage := laya.Evaluate(ctx, req, len(eligible), checks)
	audit = append(audit, fmt.Sprintf(
		"[STAGE 2 · LAYA %s] model=%s (%dms) | strategy=%s (max_p=%.2f, entropy=%.2f) | severity=%.2f (%s) | gate=%s",
		layaTriage.Source, layaTriage.Model, layaTriage.LatencyMs,
		layaTriage.RecoveryStrategy.Choice, layaTriage.RecoveryStrategy.AnswerConfidence, layaTriage.RecoveryStrategy.Confidence,
		layaTriage.OperationalSeverity.Score, layaTriage.OperationalSeverity.NearestLabel, layaTriage.RoutingGate,
	))

	// Stage 3: Gemini System 2 Structured JSON Scenario & Dispatch Generation
	gemStart := time.Now()
	gemOut := gemini.GenerateResolution(ctx, req, corridor, checks, eligible, rejected, eu, layaTriage)
	gemMs := time.Since(gemStart).Milliseconds()
	if gemOut.FallbackNote != "" {
		audit = append(audit, fmt.Sprintf(
			"[STAGE 3 · GEMINI %s] model=%s (%dms) | fallback reason: %s",
			gemOut.Source, gemOut.Model, gemMs, gemOut.FallbackNote,
		))
	} else {
		audit = append(audit, fmt.Sprintf(
			"[STAGE 3 · GEMINI %s] model=%s (%dms) | constrained-decoding JSON validated (%d scenarios)",
			gemOut.Source, gemOut.Model, gemMs, len(gemOut.Scenarios),
		))
	}

	// Stage 4: Deterministic Post-Validation Clamp (blocks illegal UICs/owners even if AI hallucinated)
	clampedScenarios, clampAudit := guardrails.PostValidateScenarios(gemOut.Scenarios, rejected, eligible, req)
	if len(clampAudit) > 0 {
		audit = append(audit, clampAudit...)
	} else {
		audit = append(audit, "[STAGE 4 · POST-CLAMP] All 3 scenarios passed homologation & safety verification without override.")
	}

	// Keep top-level TcsGermanOrder & PassengerAlert aligned with clamped scenarios
	topOrder := gemOut.TcsGermanOrder
	topAlert := gemOut.PassengerAlert
	primaryIdx := selectRecommendedScenarioIndex(layaTriage, clampedScenarios)
	if primaryIdx >= 0 && primaryIdx < len(clampedScenarios) {
		topOrder = clampedScenarios[primaryIdx].TcsGermanOrder
		topAlert = clampedScenarios[primaryIdx].PassengerAlert
	}

	audit = append(audit, fmt.Sprintf(
		"[COMPLETE] Pipeline finished in %dms.",
		time.Since(started).Milliseconds(),
	))

	return domain.IncidentResponse{
		Request:            req,
		TrainSummary:       fmt.Sprintf("%s · %s", corridor.Title, corridor.ScheduleSummary),
		CorridorNotes:      corridor.KeyNodes,
		NextDepartureHours: corridor.NextDepartureHours,
		MaxCorridorCars:    corridor.MaxCarsCorridor,
		ActiveConsist:      consist,
		RuleChecks:         checks,
		EligibleSpares:     eligible,
		RejectedSpares:     rejected,
		EU2021782:          eu,
		Laya:               layaTriage,
		GeminiSource:       gemOut.Source,
		GeminiModel:        gemOut.Model,
		Scenarios:          clampedScenarios,
		TcsGermanOrder:     topOrder,
		PassengerAlert:     topAlert,
		AuditLog:           audit,
	}, nil
}

func selectRecommendedScenarioIndex(layaTriage domain.LayaTriage, scenarios []domain.ScenarioCard) int {
	if len(scenarios) == 0 {
		return -1
	}
	targetID := "PAX_FIRST"
	if layaTriage.RecoveryStrategy.Choice == "downgrade_and_absorb" {
		targetID = "LOWEST_COST"
	}
	for i, sc := range scenarios {
		if sc.ID == targetID {
			return i
		}
	}
	return 0
}
