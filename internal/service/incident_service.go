package service

import (
	"es-incident-command/internal/data"
	"es-incident-command/internal/domain"
)

type IncidentService struct{}

func NewIncidentService() *IncidentService {
	return &IncidentService{}
}

func (s *IncidentService) GetPresets() []domain.DemoPreset {
	return data.GetPresets()
}

func (s *IncidentService) GetFleetState() domain.FleetState {
	return data.GetFleetState()
}

func (s *IncidentService) AnalyzeIncident(req domain.IncidentRequest) (domain.IncidentResponse, error) {
	corridor := data.FindCorridor(req.Route)
	consist := data.BuildActiveConsist(req)
	spares := data.GetSpareCars()

	return domain.IncidentResponse{
		Request:            req,
		TrainSummary:       corridor.Title,
		CorridorNotes:      corridor.KeyNodes,
		NextDepartureHours: corridor.NextDepartureHours,
		MaxCorridorCars:    corridor.MaxCarsCorridor,
		ActiveConsist:      consist,
		RuleChecks: []domain.RuleCheck{
			{
				ID:       "STUB_INIT",
				Category: "HOMOLOGATION",
				Severity: "PASS",
				Summary:  "Part 1 scaffold initialized",
				Detail:   "Full rail guardrails, Laya client, and Gemini client wire in Part 2.",
			},
		},
		EligibleSpares: spares,
		RejectedSpares: []domain.SpareCar{},
		EU2021782: domain.EU2021782Assessment{
			CompensationBandPct: 25,
			EstimatedRefundEUR:  float64(req.AffectedPax) * req.AvgTicketEUR * 0.25,
			RerouteObligation:   req.ExpectedDelayMin >= 60,
			LegalBasisNotes:     "Regulation (EU) 2021/782 Art. 18–20 scaffold placeholder",
		},
		Laya: domain.LayaTriage{
			Source:        "laya-mock",
			Model:         "english",
			LatencyMs:     1,
			RoutingGate:   "HUMAN_REVIEW",
			RoutingReason: "Part 1 scaffold stub",
			RecoveryStrategy: domain.LayaChoiceAnswer{
				Choice:           "swap_spare_car",
				Probabilities:    map[string]float64{"swap_spare_car": 0.78, "downgrade_and_absorb": 0.14, "manual_escalation": 0.08},
				Confidence:       0.64,
				AnswerConfidence: 0.78,
			},
			OperationalSeverity: domain.LayaScoreAnswer{
				Score:            1.8,
				NearestLabel:     "High severity",
				Probabilities:    map[string]float64{"0": 0.05, "1": 0.25, "2": 0.55, "3": 0.15},
				Confidence:       0.61,
				AnswerConfidence: 0.55,
			},
		},
		GeminiSource: "gemini-mock",
		GeminiModel:  "gemini-3.5-flash-lite",
		Scenarios:    []domain.ScenarioCard{},
		AuditLog:     []string{"Part 1 scaffold active"},
	}, nil
}
