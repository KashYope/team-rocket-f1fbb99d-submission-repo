package guardrails

import (
	"strings"
	"testing"

	"es-incident-command/internal/data"
	"es-incident-command/internal/domain"
)

func TestEvaluatePreChecks_PresetsAndEdgeCases(t *testing.T) {
	spares := data.GetSpareCars()
	presets := data.GetPresets()

	if len(presets) != 3 {
		t.Fatalf("expected 3 presets, got %d", len(presets))
	}

	t.Run("Preset1_Milan_Homologation_And_25Pct_EU2021782", func(t *testing.T) {
		req := presets[0].Request
		checks, eligible, rejected, eu := EvaluatePreChecks(req, spares)

		if len(eligible) != 3 {
			t.Errorf("expected 3 eligible RDC spares for MILAN (WL33, Bvcmz5, Bdcm36), got %d", len(eligible))
		}
		if len(rejected) != 2 {
			t.Errorf("expected 2 rejected spares for MILAN (EE Bvcmz6, GFF Bimz 309), got %d", len(rejected))
		}
		for _, r := range rejected {
			if r.Owner != "EE" && r.Owner != "GFF" {
				t.Errorf("unexpected rejected owner on MILAN: %s", r.Owner)
			}
		}
		if eu.CompensationBandPct != 25 {
			t.Errorf("expected 25%% compensation for 95m delay, got %d%%", eu.CompensationBandPct)
		}
		if eu.ForceMajeureExempt {
			t.Errorf("OPERATOR_TECH must never be force-majeure exempt")
		}
		if !eu.RerouteObligation || !eu.SelfReroute100Min {
			t.Errorf("expected RerouteObligation and SelfReroute100Min true at 95m")
		}
		if !hasRuleID(checks, "PET_PRIVATE_COMP", "BLOCK") {
			t.Errorf("expected PET_PRIVATE_COMP BLOCK check")
		}
		if !hasRuleID(checks, "CONDUCTOR_CLIFF_8_CAR", "WARN") {
			t.Errorf("expected CONDUCTOR_CLIFF_8_CAR WARN check")
		}
	})

	t.Run("Preset2_Prague_InfraManager_NonExempt_50Pct_PRM", func(t *testing.T) {
		req := presets[1].Request
		checks, _, _, eu := EvaluatePreChecks(req, spares)

		if eu.ForceMajeureExempt {
			t.Errorf("INFRA_MANAGER cause must NEVER be force-majeure exempt under EU 2021/782 Art. 19(10)")
		}
		if eu.CompensationBandPct != 50 {
			t.Errorf("expected 50%% compensation for canceled train, got %d%%", eu.CompensationBandPct)
		}
		if eu.EstimatedRefundEUR <= 0 {
			t.Errorf("expected positive EstimatedRefundEUR for non-exempt cancellation, got %v", eu.EstimatedRefundEUR)
		}
		if !eu.HotelRequired || !eu.SelfReroute100Min || !eu.PRMPriorityRequired {
			t.Errorf("expected HotelRequired, SelfReroute100Min, and PRMPriorityRequired all true: %+v", eu)
		}
		if !hasRuleID(checks, "PRM_PRIORITY_ART_20_5", "BLOCK") {
			t.Errorf("expected PRM_PRIORITY_ART_20_5 BLOCK check")
		}
	})

	t.Run("Preset3_Paris_Bdcm36Blocked_LeBourgetShunt_WomenOnly", func(t *testing.T) {
		req := presets[2].Request
		checks, eligible, rejected, eu := EvaluatePreChecks(req, spares)

		if len(rejected) != 1 || !strings.Contains(rejected[0].CarType, "Bdcm36") {
			t.Fatalf("expected exactly RDC Bdcm36 rejected on PARIS, got %+v", rejected)
		}
		if len(eligible) != 4 {
			t.Errorf("expected 4 eligible spares on PARIS, got %d", len(eligible))
		}
		if eu.CompensationBandPct != 25 {
			t.Errorf("expected 25%% band for 75m delay, got %d%%", eu.CompensationBandPct)
		}
		if !hasRuleID(checks, "PARIS_LE_BOURGET_SHUNT", "WARN") {
			t.Errorf("expected PARIS_LE_BOURGET_SHUNT WARN check")
		}
		if !hasRuleID(checks, "WOMEN_ONLY_COMP", "WARN") {
			t.Errorf("expected WOMEN_ONLY_COMP WARN check")
		}
	})

	t.Run("ForceMajeure_ThirdPartyExtreme_KeepsCareAndRerouting", func(t *testing.T) {
		req := presets[1].Request
		req.CauseCategory = "THIRD_PARTY_EXTREME"
		_, _, _, eu := EvaluatePreChecks(req, spares)

		if !eu.ForceMajeureExempt {
			t.Errorf("expected ForceMajeureExempt true for THIRD_PARTY_EXTREME")
		}
		if eu.EstimatedRefundEUR != 0 {
			t.Errorf("expected €0 Art. 19 compensation under THIRD_PARTY_EXTREME, got %v", eu.EstimatedRefundEUR)
		}
		if !eu.RerouteObligation || !eu.SelfReroute100Min || !eu.HotelRequired {
			t.Errorf("Art. 18 & 20 rerouting and hotel duties must remain TRUE even under force majeure: %+v", eu)
		}
	})
}

func TestPostValidateScenarios_ClampsIllegalMilanCoach(t *testing.T) {
	spares := data.GetSpareCars()
	req := data.GetPresets()[0].Request // Milan Flat Wheel
	_, eligible, rejected, _ := EvaluatePreChecks(req, spares)

	illegalScenarios := []domain.ScenarioCard{
		{
			ID:             "PAX_FIRST",
			Title:          "Swap EE Bvcmz6 from Schaarbeek",
			EstimatedCost:  1800,
			RecommendedCar: "EE Bvcmz6 (61 84 50-70 601-4)",
			ActionSteps: []string{
				"Uncouple RDC WL33 Car 56 at Brussels",
				"Attach EE Bvcmz6 (61 84 50-70 601-4) for Milan service",
			},
			TcsGermanOrder: "Zug ES 401: Wagen 56 abstellen; Ersatzwagen EE Bvcmz6 61 84 50-70 601-4 beistellen.",
			PassengerAlert: "Replacement coach attached.",
		},
	}

	clamped, audit := PostValidateScenarios(illegalScenarios, rejected, eligible, req)
	if len(clamped) != 1 {
		t.Fatalf("expected 1 scenario, got %d", len(clamped))
	}
	if !clamped[0].WasClampedByGo {
		t.Fatalf("expected WasClampedByGo=true when EE Bvcmz6 was proposed on MILAN")
	}
	if strings.Contains(clamped[0].RecommendedCar, "EE") || strings.Contains(clamped[0].RecommendedCar, "Bvcmz6") {
		t.Errorf("RecommendedCar still contains illegal coach: %s", clamped[0].RecommendedCar)
	}
	if !strings.Contains(clamped[0].RecommendedCar, "RDC WL33") {
		t.Errorf("expected clamp to select matching RDC WL33 spare, got %s", clamped[0].RecommendedCar)
	}
	if strings.Contains(clamped[0].TcsGermanOrder, "61 84 50-70 601-4") {
		t.Errorf("TcsGermanOrder still contains illegal UIC: %s", clamped[0].TcsGermanOrder)
	}
	if len(audit) == 0 {
		t.Errorf("expected post-clamp audit entry")
	}
}

func hasRuleID(checks []domain.RuleCheck, id string, severity string) bool {
	for _, c := range checks {
		if c.ID == id && c.Severity == severity {
			return true
		}
	}
	return false
}
