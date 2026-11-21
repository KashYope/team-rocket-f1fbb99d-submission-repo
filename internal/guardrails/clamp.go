package guardrails

import (
	"fmt"
	"strings"

	"es-incident-command/internal/domain"
)

// PostValidateScenarios scans every AI-generated ScenarioCard (including RecommendedCar,
// ActionSteps, and TcsGermanOrder) for non-homologated / rejected rolling stock.
// Any illegal coach reference is deterministically clamped to an eligible RDC spare
// (or "NONE — Absorb / Reroute"), setting WasClampedByGo = true.
func PostValidateScenarios(
	scenarios []domain.ScenarioCard,
	rejected []domain.SpareCar,
	eligible []domain.SpareCar,
	req domain.IncidentRequest,
) ([]domain.ScenarioCard, []string) {
	if len(scenarios) == 0 {
		return scenarios, nil
	}

	fallbackLabel := selectFallbackCarLabel(eligible, req)
	var auditEntries []string
	out := make([]domain.ScenarioCard, len(scenarios))

	for i, sc := range scenarios {
		card := sc
		var matchedRejected []domain.SpareCar

		combinedText := strings.ToUpper(strings.Join([]string{
			card.RecommendedCar,
			card.Title,
			strings.Join(card.ActionSteps, " "),
			card.TcsGermanOrder,
		}, " | "))

		for _, rej := range rejected {
			if referencesRejectedCar(combinedText, rej, req.Route) {
				matchedRejected = append(matchedRejected, rej)
			}
		}

		if len(matchedRejected) > 0 {
			card.WasClampedByGo = true
			origCar := card.RecommendedCar
			card.RecommendedCar = fallbackLabel

			for _, rej := range matchedRejected {
				card.Title = replaceIllegalTokens(card.Title, rej, fallbackLabel)
				for sIdx := range card.ActionSteps {
					card.ActionSteps[sIdx] = replaceIllegalTokens(card.ActionSteps[sIdx], rej, fallbackLabel)
				}
				card.TcsGermanOrder = replaceIllegalTokens(card.TcsGermanOrder, rej, fallbackLabel)

				note := fmt.Sprintf(
					"CLAMPED BY GO GUARDRAIL: AI proposed %s %s (%s) which is blocked on %s (%s). Replaced with %s.",
					rej.Owner, rej.CarType, rej.UIC, req.Route, rej.BlockedReason, fallbackLabel,
				)
				card.GuardrailNotes = append(card.GuardrailNotes, note)
				auditEntries = append(auditEntries, fmt.Sprintf(
					"[POST-CLAMP] Scenario %s referenced blocked coach %s %s (%s, orig=%q) -> clamped to %q",
					card.ID, rej.Owner, rej.CarType, rej.UIC, origCar, fallbackLabel,
				))
			}
		}

		out[i] = card
	}

	return out, auditEntries
}

func selectFallbackCarLabel(eligible []domain.SpareCar, req domain.IncidentRequest) string {
	if req.IssueType == "CANCELED" || len(eligible) == 0 {
		return "NONE — Absorb / Reroute"
	}
	upperAffected := strings.ToUpper(req.AffectedCar)
	for _, el := range eligible {
		if strings.Contains(upperAffected, strings.ToUpper(el.CarType)) {
			return fmt.Sprintf("%s %s (%s, %s)", el.Owner, el.CarType, el.UIC, el.Yard)
		}
	}
	first := eligible[0]
	return fmt.Sprintf("%s %s (%s, %s)", first.Owner, first.CarType, first.UIC, first.Yard)
}

func referencesRejectedCar(upperCombined string, rej domain.SpareCar, route domain.RouteCode) bool {
	uicUpper := strings.ToUpper(rej.UIC)
	uicCompact := strings.ReplaceAll(uicUpper, " ", "")
	combinedCompact := strings.ReplaceAll(upperCombined, " ", "")

	if uicUpper != "" && (strings.Contains(upperCombined, uicUpper) || strings.Contains(combinedCompact, uicCompact)) {
		return true
	}
	carTypeUpper := strings.ToUpper(rej.CarType)
	if carTypeUpper != "" && strings.Contains(upperCombined, carTypeUpper) {
		return true
	}
	if strings.Contains(carTypeUpper, "309") && (strings.Contains(upperCombined, "GFF 309") || strings.Contains(upperCombined, "BIMZ 309") || strings.Contains(upperCombined, "WAGEN 309")) {
		return true
	}
	if route == domain.RouteMilan {
		if rej.Owner == "EE" && (strings.Contains(upperCombined, "EE BVCMZ") || strings.Contains(upperCombined, "EURO-EXPRESS")) {
			return true
		}
		if rej.Owner == "GFF" && (strings.Contains(upperCombined, "GFF BIMZ") || strings.Contains(upperCombined, "GFF ")) {
			return true
		}
	}
	return false
}

func replaceIllegalTokens(input string, rej domain.SpareCar, replacement string) string {
	if input == "" {
		return input
	}
	replacerPairs := []string{
		fmt.Sprintf("%s %s (%s)", rej.Owner, rej.CarType, rej.UIC), replacement,
		fmt.Sprintf("%s %s", rej.Owner, rej.CarType), replacement,
		rej.UIC, replacement,
		rej.CarType, replacement,
	}
	if strings.Contains(rej.CarType, "309") {
		replacerPairs = append(replacerPairs, "GFF 309", replacement, "Gff 309", replacement)
	}
	r := strings.NewReplacer(replacerPairs...)
	return r.Replace(input)
}
