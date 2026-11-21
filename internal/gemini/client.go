package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"es-incident-command/internal/data"
	"es-incident-command/internal/domain"

	"google.golang.org/genai"
)

type ResolutionOutput struct {
	Source         string
	Model          string
	Scenarios      []domain.ScenarioCard
	TcsGermanOrder string
	PassengerAlert string
	FallbackNote   string
}

type rawGeminiPayload struct {
	Scenarios      []domain.ScenarioCard `json:"scenarios"`
	TcsGermanOrder string                `json:"tcsGermanOrder"`
	PassengerAlert string                `json:"passengerAlert"`
}

func ResolveModelName() string {
	m := strings.TrimSpace(os.Getenv("GEMINI_MODEL"))
	if m == "" {
		return "gemini-3.5-flash-lite"
	}
	lower := strings.ToLower(m)
	if strings.HasPrefix(lower, "gemini-2.5") || strings.HasPrefix(lower, "gemini-2.0") || strings.HasPrefix(lower, "gemini-1.5") {
		return "gemini-3.5-flash-lite"
	}
	return m
}

// GenerateResolution invokes Google Gemini via google.golang.org/genai with constrained JSON
// decoding (ResponseMIMEType + flat ResponseSchema, Temperature 0.2, 20s timeout), falling back
// deterministically to gemini-mock if GEMINI_API_KEY is absent, GEMINI_MODE=mock, or an error occurs.
func GenerateResolution(
	parentCtx context.Context,
	req domain.IncidentRequest,
	corridor domain.CorridorInfo,
	checks []domain.RuleCheck,
	eligible []domain.SpareCar,
	rejected []domain.SpareCar,
	eu domain.EU2021782Assessment,
	laya domain.LayaTriage,
) ResolutionOutput {
	modelName := ResolveModelName()
	mode := strings.ToLower(strings.TrimSpace(os.Getenv("GEMINI_MODE")))
	apiKey := strings.TrimSpace(os.Getenv("GEMINI_API_KEY"))

	if mode == "mock" || apiKey == "" {
		reason := "GEMINI_API_KEY unset"
		if mode == "mock" {
			reason = "GEMINI_MODE=mock"
		}
		return mockGenerate(req, corridor, eligible, rejected, eu, laya, modelName, reason)
	}

	ctx, cancel := context.WithTimeout(parentCtx, 20*time.Second)
	defer cancel()

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return mockGenerate(req, corridor, eligible, rejected, eu, laya, modelName, fmt.Sprintf("client init error: %v", err))
	}

	prompt := buildOperationalPrompt(req, corridor, checks, eligible, rejected, eu, laya)
	config := &genai.GenerateContentConfig{
		Temperature:      genai.Ptr[float32](0.2),
		ResponseMIMEType: "application/json",
		ResponseSchema:   buildResponseSchema(),
		SystemInstruction: &genai.Content{
			Parts: []*genai.Part{{
				Text: "You are Senior Night Duty Operations Planner at European Sleeper (Nov 2026 timetable). " +
					"Produce three distinct operational recovery scenarios (LOWEST_COST, PAX_FIRST, PROTECT_NEXT_TRAIN), " +
					"authentic German rail shunting/disposition task orders for Train Charter Services (TCS) using railway verbs " +
					"(abholen, abstellen, beistellen), exact 12-digit UIC numbers, train IDs, and yard/workshop names, " +
					"plus clear English passenger advisories citing Regulation (EU) 2021/782 statutory rights whenever delay >= 60 min or canceled.",
			}},
		},
	}

	resp, err := client.Models.GenerateContent(ctx, modelName, genai.Text(prompt), config)
	if err != nil {
		return mockGenerate(req, corridor, eligible, rejected, eu, laya, modelName, fmt.Sprintf("API error: %v", err))
	}

	rawText := strings.TrimSpace(resp.Text())
	if rawText == "" {
		return mockGenerate(req, corridor, eligible, rejected, eu, laya, modelName, "empty model response")
	}

	var parsed rawGeminiPayload
	if err := json.Unmarshal([]byte(rawText), &parsed); err != nil || len(parsed.Scenarios) == 0 {
		return mockGenerate(req, corridor, eligible, rejected, eu, laya, modelName, "JSON unmarshal fallback")
	}

	normalized := normalizeScenarios(parsed, req, corridor, eu)
	return ResolutionOutput{
		Source:         "gemini-live",
		Model:          modelName,
		Scenarios:      normalized.Scenarios,
		TcsGermanOrder: normalized.TcsGermanOrder,
		PassengerAlert: normalized.PassengerAlert,
	}
}

func buildResponseSchema() *genai.Schema {
	return &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"scenarios": {
				Type:        genai.TypeArray,
				Description: "Three operational scenarios: LOWEST_COST, PAX_FIRST, and PROTECT_NEXT_TRAIN.",
				Items: &genai.Schema{
					Type: genai.TypeObject,
					Properties: map[string]*genai.Schema{
						"id": {
							Type: genai.TypeString,
							Enum: []string{"LOWEST_COST", "PAX_FIRST", "PROTECT_NEXT_TRAIN"},
						},
						"title":            {Type: genai.TypeString},
						"estimatedCostEur": {Type: genai.TypeInteger},
						"downgrades":       {Type: genai.TypeInteger},
						"cancellations":    {Type: genai.TypeInteger},
						"nextTrainRisk": {
							Type: genai.TypeString,
							Enum: []string{"LOW", "MEDIUM", "HIGH"},
						},
						"recommendedCar": {Type: genai.TypeString},
						"actionSteps": {
							Type:  genai.TypeArray,
							Items: &genai.Schema{Type: genai.TypeString},
						},
						"guardrailNotes": {
							Type:  genai.TypeArray,
							Items: &genai.Schema{Type: genai.TypeString},
						},
						"tcsGermanOrder": {
							Type:        genai.TypeString,
							Description: "Scenario-specific German rail task order for TCS using abholen, abstellen, beistellen, UIC codes, and train ID.",
						},
						"passengerAlert": {
							Type:        genai.TypeString,
							Description: "Scenario-specific English passenger message citing EU 2021/782 compensation, rerouting, 100-min self-rerouting, and hotel rights when applicable.",
						},
					},
					Required: []string{
						"id", "title", "estimatedCostEur", "downgrades", "cancellations",
						"nextTrainRisk", "recommendedCar", "actionSteps", "guardrailNotes",
						"tcsGermanOrder", "passengerAlert",
					},
				},
			},
			"tcsGermanOrder": {
				Type:        genai.TypeString,
				Description: "Primary recommended German TCS shunting/disposition order.",
			},
			"passengerAlert": {
				Type:        genai.TypeString,
				Description: "Primary recommended English passenger advisory.",
			},
		},
		Required: []string{"scenarios", "tcsGermanOrder", "passengerAlert"},
	}
}

func buildOperationalPrompt(
	req domain.IncidentRequest,
	corridor domain.CorridorInfo,
	checks []domain.RuleCheck,
	eligible []domain.SpareCar,
	rejected []domain.SpareCar,
	eu domain.EU2021782Assessment,
	laya domain.LayaTriage,
) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "INCIDENT BRIEF:\n")
	fmt.Fprintf(&sb, "- Train: ES %s (%s) | Route: %s | Next ES departure: %dh\n", req.TrainID, corridor.Title, req.Route, corridor.NextDepartureHours)
	fmt.Fprintf(&sb, "- Corridor Nodes: %s\n", corridor.KeyNodes)
	fmt.Fprintf(&sb, "- Issue: %s | Cause: %s | Expected Delay: %d min\n", req.IssueType, req.CauseCategory, req.ExpectedDelayMin)
	fmt.Fprintf(&sb, "- Affected Car: %s | Affected Pax: %d | Avg Ticket: €%.0f\n", req.AffectedCar, req.AffectedPax, req.AvgTicketEUR)
	fmt.Fprintf(&sb, "- Passenger Flags: PRM=%t, PetPrivateComp=%t, WomenOnlyComp=%t\n", req.HasPRMPax, req.HasPetPrivateComp, req.HasWomenOnlyComp)
	fmt.Fprintf(&sb, "- Extra Context: %s\n\n", req.ExtraInfo)

	fmt.Fprintf(&sb, "ECONOMICS & WORKSHOPS:\n")
	fmt.Fprintf(&sb, "- Max corridor cars: %d (%s)\n", corridor.MaxCarsCorridor, corridor.MaxCarsDetail)
	fmt.Fprintf(&sb, "- 9th commercial occupied car triggers 2nd conductor: +€%d\n", corridor.SecondConductorEUR)
	fmt.Fprintf(&sb, "- Workshop routing: %s\n\n", corridor.WorkshopRoute)

	fmt.Fprintf(&sb, "EU 2021/782 STATUTORY ASSESSMENT:\n")
	fmt.Fprintf(&sb, "- Compensation Band: %d%% (Est. €%.0f) | ForceMajeureExempt: %t\n", eu.CompensationBandPct, eu.EstimatedRefundEUR, eu.ForceMajeureExempt)
	fmt.Fprintf(&sb, "- RerouteObligation: %t | SelfReroute100Min: %t | HotelRequired: %t | PRMPriority: %t\n\n",
		eu.RerouteObligation, eu.SelfReroute100Min, eu.HotelRequired, eu.PRMPriorityRequired)

	fmt.Fprintf(&sb, "LAYA SYSTEM-1 TRIAGE:\n")
	fmt.Fprintf(&sb, "- Gate: %s (%s) | Top Strategy: %s (p=%.2f, entropy=%.2f) | Severity: %.2f (%s)\n\n",
		laya.RoutingGate, laya.RoutingReason, laya.RecoveryStrategy.Choice,
		laya.RecoveryStrategy.AnswerConfidence, laya.RecoveryStrategy.Confidence,
		laya.OperationalSeverity.Score, laya.OperationalSeverity.NearestLabel)

	if req.SimulateUnfilteredAI && len(rejected) > 0 {
		// Demo Clamp Mode: expose raw unfiltered yard pool without homologation warnings so Go's PostValidateScenarios visibly clamps illegal stock!
		allYard := data.GetSpareCars()
		fmt.Fprintf(&sb, "RAW UNFILTERED YARD POOL (DEMO UNFILTERED MODE — select highest-capacity low-cost coach in LOWEST_COST purely by berth count, e.g. %s %s %s):\n",
			rejected[0].Owner, rejected[0].CarType, rejected[0].UIC)
		for _, c := range allYard {
			fmt.Fprintf(&sb, "- %s %s (%s) | Class: %s | Cap: %d | Yard: %s\n", c.Owner, c.CarType, c.UIC, c.ProductClass, c.Capacity, c.Yard)
		}
		fmt.Fprintln(&sb)
	} else {
		fmt.Fprintf(&sb, "ELIGIBLE HOMOLOGATED SPARE CARS:\n")
		if len(eligible) == 0 {
			fmt.Fprintf(&sb, "- NONE\n")
		}
		for _, c := range eligible {
			fmt.Fprintf(&sb, "- %s %s (%s) | Class: %s | Cap: %d | Yard: %s\n", c.Owner, c.CarType, c.UIC, c.ProductClass, c.Capacity, c.Yard)
		}
		fmt.Fprintf(&sb, "\nBLOCKED SPARE CARS (NEVER USE):\n")
		for _, c := range rejected {
			fmt.Fprintf(&sb, "- %s %s (%s) -> %s\n", c.Owner, c.CarType, c.UIC, c.BlockedReason)
		}
		fmt.Fprintln(&sb)
	}

	fmt.Fprintf(&sb, "ACTIVE DETERMINISTIC GUARDRAILS:\n")
	for _, chk := range checks {
		if req.SimulateUnfilteredAI && chk.ID == "HOMOLOGATION_FILTER" {
			continue
		}
		fmt.Fprintf(&sb, "- [%s] %s: %s\n", chk.Severity, chk.Summary, chk.Detail)
	}

	return sb.String()
}

func normalizeScenarios(parsed rawGeminiPayload, req domain.IncidentRequest, corridor domain.CorridorInfo, eu domain.EU2021782Assessment) rawGeminiPayload {
	fallbackMock := mockGenerate(req, corridor, nil, nil, eu, domain.LayaTriage{}, "", "")
	byID := map[string]domain.ScenarioCard{}
	for _, sc := range parsed.Scenarios {
		byID[sc.ID] = sc
	}

	orderedIDs := []string{"LOWEST_COST", "PAX_FIRST", "PROTECT_NEXT_TRAIN"}
	finalCards := make([]domain.ScenarioCard, 0, 3)
	for idx, id := range orderedIDs {
		card, ok := byID[id]
		if !ok {
			if idx < len(parsed.Scenarios) {
				card = parsed.Scenarios[idx]
				card.ID = id
			} else {
				card = fallbackMock.Scenarios[idx]
			}
		}
		if strings.TrimSpace(card.TcsGermanOrder) == "" {
			card.TcsGermanOrder = parsed.TcsGermanOrder
		}
		if strings.TrimSpace(card.PassengerAlert) == "" {
			card.PassengerAlert = parsed.PassengerAlert
		}
		if card.ActionSteps == nil {
			card.ActionSteps = []string{}
		}
		if card.GuardrailNotes == nil {
			card.GuardrailNotes = []string{}
		}
		finalCards = append(finalCards, card)
	}

	topOrder := strings.TrimSpace(parsed.TcsGermanOrder)
	topAlert := strings.TrimSpace(parsed.PassengerAlert)
	// Default top-level fields to PAX_FIRST (index 1) or first scenario
	if topOrder == "" && len(finalCards) > 1 {
		topOrder = finalCards[1].TcsGermanOrder
	}
	if topAlert == "" && len(finalCards) > 1 {
		topAlert = finalCards[1].PassengerAlert
	}

	return rawGeminiPayload{
		Scenarios:      finalCards,
		TcsGermanOrder: topOrder,
		PassengerAlert: topAlert,
	}
}

func mockGenerate(
	req domain.IncidentRequest,
	corridor domain.CorridorInfo,
	eligible []domain.SpareCar,
	rejected []domain.SpareCar,
	eu domain.EU2021782Assessment,
	_ domain.LayaTriage,
	modelName string,
	fallbackNote string,
) ResolutionOutput {
	if modelName == "" {
		modelName = ResolveModelName()
	}
	isCanceled := req.IssueType == "CANCELED" || req.ExpectedDelayMin >= 999
	refundInt := int(eu.EstimatedRefundEUR)

	primarySpare := "RDC WL33 (61 80 72-90 033-1, ddl)"
	secondarySpare := "RDC Bvcmz5 (61 80 50-91 025-2, ddl)"
	if len(eligible) > 0 {
		primarySpare = fmt.Sprintf("%s %s (%s, %s)", eligible[0].Owner, eligible[0].CarType, eligible[0].UIC, eligible[0].Yard)
	}
	if len(eligible) > 1 {
		secondarySpare = fmt.Sprintf("%s %s (%s, %s)", eligible[1].Owner, eligible[1].CarType, eligible[1].UIC, eligible[1].Yard)
	}

	// If SimulateUnfilteredAI is active and rejected stock exists, inject rejected[0] into LOWEST_COST so Go's PostValidateScenarios clamps it!
	unfilteredCandidate := "NONE — Absorb Onboard"
	unfilteredOrderCar := "KEIN ERSATZWAGEN (Wagen nur abstellen)"
	if req.SimulateUnfilteredAI && len(rejected) > 0 && !isCanceled {
		r := rejected[0]
		unfilteredCandidate = fmt.Sprintf("%s %s (%s)", r.Owner, r.CarType, r.UIC)
		unfilteredOrderCar = fmt.Sprintf("%s %s (%s) aus %s abholen und beistellen", r.Owner, r.CarType, r.UIC, r.Yard)
	}

	var scenarios []domain.ScenarioCard

	switch {
	case isCanceled:
		hotelEstimate := req.AffectedPax * 145
		daytimeRailEstimate := req.AffectedPax * 95
		scenarios = []domain.ScenarioCard{
			{
				ID:             "LOWEST_COST",
				Title:          "Early Refund + Split Daytime Rail Rerouting",
				EstimatedCost:  refundInt + daytimeRailEstimate + (req.AffectedPax * 45),
				Downgrades:     req.AffectedPax,
				Cancellations:  req.AffectedPax,
				NextTrainRisk:  "MEDIUM",
				RecommendedCar: "NONE — Full Corridor Cancellation",
				ActionSteps: []string{
					"Dispatch EU 2021/782 Art. 18(3) rerouting links within 100 minutes of scheduled departure to prevent uncapped self-booking claims.",
					"Issue DB/NS/ČD daytime ICE/EC tickets from Bruxelles-Midi via Frankfurt/Berlin to Praha hl.n. for flexible travellers.",
					"Arrange dedicated accessible taxi/hotel transfer for foldable-wheelchair PRM party under Art. 20(5).",
				},
				GuardrailNotes: []string{
					"INFRA_MANAGER fault is NOT force-majeure exempt (50% Art. 19 compensation applies).",
					"100-minute self-rerouting clock active immediately.",
				},
				TcsGermanOrder: fmt.Sprintf(
					"DISPOSITIONSANORDNUNG TCS — ZUG ES %s (%s)\n"+
						"1. Komplettausfall ES %s infolge Oberleitungs-/HV-Störung (Infrastrukturbetreiber).\n"+
						"2. Leergarnitur (8 Wagen) in Bruxelles-Midi / Schaarbeek gesichert abstellen; keine Reisenden an Bord beim Rangieren.\n"+
						"3. Mehrsystemlok zur Störungsprüfung in Bad Bentheim / Münster (emstg) anmelden, damit Umlauf Gegenzug geschützt bleibt.",
					req.TrainID, req.Route, req.TrainID,
				),
				PassengerAlert: fmt.Sprintf(
					"IMPORTANT UPDATE — European Sleeper ES %s (%s) is canceled due to an infrastructure power fault. "+
						"Under Regulation (EU) 2021/782 (Art. 18–20), you are entitled to: (1) full ticket refund OR earliest rerouting onto daytime international rail OR rebooking onto a later departure, "+
						"(2) 50%% delay/cancellation compensation (payable in cash/bank transfer on request), (3) complimentary overnight hotel + station transfer where an overnight stay is required, "+
						"and (4) priority PRM assistance (Art. 20(5)). Rerouting options are issued within 100 minutes; otherwise Art. 18(3) self-rerouting reimbursement applies.",
					req.TrainID, corridor.Title,
				),
			},
			{
				ID:             "PAX_FIRST",
				Title:          "Full Art. 20 Hotel Block + Guaranteed Morning ICE/EC Seats",
				EstimatedCost:  refundInt + hotelEstimate + daytimeRailEstimate,
				Downgrades:     0,
				Cancellations:  req.AffectedPax,
				NextTrainRisk:  "LOW",
				RecommendedCar: "NONE — Hotel + Daytime ICE/EC Charter Block",
				ActionSteps: []string{
					"Activate partner hotel blocks in Brussels & Amsterdam within 60 minutes (well inside the 100-min Art. 18(3) statutory window).",
					"Prioritize Car 17 PRM wheelchair travellers with step-free hotel rooms & accessible morning ICE transfer (Art. 20(5)).",
					"Protect Women-Only compartment travellers with single/female-shared hotel room allocation.",
					"Pre-book group seat reservations on morning ICE Brussels -> Köln -> Berlin -> Praha hl.n.",
				},
				GuardrailNotes: []string{
					"Art. 20(2)(b) overnight hotel + meals mandatory (48h gap until next ES 453).",
					"Art. 20(5) PRM priority satisfied first.",
				},
				TcsGermanOrder: fmt.Sprintf(
					"DISPOSITIONSANORDNUNG TCS — ZUG ES %s (REISENDEN-PRIORITÄT)\n"+
						"1. Stammgarnitur ES %s in Schaarbeek abstellen und 3h Reinigung + 2h Logistik für Folgeumlauf vorziehen.\n"+
						"2. Wagen 17 (RDC Bvcmz5, 61 80 50-91 023-7) am Bahnsteig Bruxelles-Midi zuerst für PRM-Ausstieg sichern.\n"+
						"3. Keine Ersatzwagen aus Denderleeuw (ddl) abholen — Trassenausfall Infrastruktur.",
					req.TrainID, req.TrainID,
				),
				PassengerAlert: fmt.Sprintf(
					"European Sleeper ES %s (%s): Your train tonight cannot depart due to an infrastructure catenary failure. "+
						"We have arranged complimentary hotel accommodation, station transfers, meals, and reserved morning high-speed rail connections to Berlin, Dresden, and Prague. "+
						"Our station team is assisting PRM travellers first (EU 2021/782 Art. 20(5)). You remain entitled to 50%% statutory Art. 19 compensation (payable via bank transfer on request) or a 100%% ticket refund if you choose not to travel.",
					req.TrainID, corridor.Title,
				),
			},
			{
				ID:             "PROTECT_NEXT_TRAIN",
				Title:          "Deadhead Consist Overnight to Protect Return ES 452",
				EstimatedCost:  refundInt + hotelEstimate + daytimeRailEstimate + 5000,
				Downgrades:     0,
				Cancellations:  req.AffectedPax,
				NextTrainRisk:  "LOW",
				RecommendedCar: "NONE — Ferry Empty Consist via Bad Bentheim",
				ActionSteps: []string{
					"Accommodate outbound passengers via Brussels hotel + daytime ICE rerouting.",
					"Once catenary clears, ferry empty 8-car consist via Bad Bentheim / Berlin Wustermark so Tuesday ES 452 from Praha hl.n. departs on schedule with full 3h+2h turnaround.",
					"Route defective HV jumper/loco unit via Münster (emstg, ~€5,000) without stranding return passengers in Prague.",
				},
				GuardrailNotes: []string{
					"Eliminates cascading 340-pax return cancellation on ES 452 in 24–48h.",
					"Includes €5,000 Bentheim–Münster–Lengerich workshop allowance.",
				},
				TcsGermanOrder: fmt.Sprintf(
					"DISPOSITIONSANORDNUNG TCS — ZUG ES %s / SCHUTZ UMLAUF ES 452\n"+
						"1. Nach Freigabe Fahrleitung Leerpark ES %s (8 Wagen) in Schaarbeek abholen und als Leerfahrt über Bad Bentheim Richtung Berlin Wustermark / Praha hl.n. überführen.\n"+
						"2. Schadhafte HV-Kupplung in Münster (emstg) prüfen/abstellen.\n"+
						"3. 5-Stunden-Wendezeitfenster (3h Reinigung + 2h Logistik) für ES 452 in Praha hl.n. sicherstellen.",
					req.TrainID, req.TrainID,
				),
				PassengerAlert: fmt.Sprintf(
					"European Sleeper ES %s (%s) is canceled tonight due to an external rail power failure. "+
						"Complimentary hotel rooms, refreshments, and daytime rail tickets to your destination are ready now (communicated within the EU 2021/782 100-minute window), alongside 50%% Art. 19 compensation. "+
						"Return service ES 452 remains protected on schedule.",
					req.TrainID, corridor.Title,
				),
			},
		}

	case req.Route == domain.RouteParis:
		scenarios = []domain.ScenarioCard{
			{
				ID:             "LOWEST_COST",
				Title:          "Lock Out Unheated Car 34 at Paris Nord & Absorb Onboard",
				EstimatedCost:  refundInt + 2200,
				Downgrades:     24,
				Cancellations:  0,
				NextTrainRisk:  "LOW",
				RecommendedCar: unfilteredCandidate,
				ActionSteps: []string{
					"Evacuate and lock out unheated Car 34 (RDC Bvcmz5, 61 80 50-91 029-4) in-consist at Paris Nord — avoids restricted Paris Le Bourget shunting.",
					"Reallocate Women-Only compartment travellers strictly into spare Women-Only berths in Cars 35/36 (explicit consent verified).",
					"Absorb remaining Classic passengers into spare berths/Budget seats across Cars 30–37 (drops occupied commercial cars from 9 to 8).",
				},
				GuardrailNotes: []string{
					"Avoids Paris Le Bourget shunting slot restriction & preserves Bruxelles-Midi crew handover.",
					"Women-Only compartment privacy preserved.",
				},
				TcsGermanOrder: fmt.Sprintf(
					"DISPOSITIONSANORDNUNG TCS — ZUG ES %s (PARIS NORD -> BERLIN)\n"+
						"1. Schadwagen 34 (RDC Bvcmz5, 61 80 50-91 029-4, Heizungs-/Umrichterausfall) in Paris Nord räumen, verschließen und kalt im Zugverband mitführen (kein Rangieren in Le Bourget mit Reisenden).\n"+
						"2. Ersatzgestellung: %s.\n"+
						"3. In Bruxelles-Midi Personalwechsel sichern; Wagen 34 in Berlin Wustermark zur Instandsetzung abstellen.",
					req.TrainID, unfilteredOrderCar,
				),
				PassengerAlert: fmt.Sprintf(
					"European Sleeper ES %s (Paris Nord – Brussels – Berlin): Due to a heating fault on Coach 34, expected arrival delay is ~%d minutes. "+
						"Coach 34 passengers are being accommodated in heated compartments onboard (Women-Only compartment allocations are preserved; any class downgrade receives automatic partial fare reimbursement). "+
						"Under EU 2021/782, delays of 60–119 minutes qualify for 25%% compensation (cash/bank transfer available) or Art. 18 rerouting/refund options.",
					req.TrainID, req.ExpectedDelayMin,
				),
			},
			{
				ID:             "PAX_FIRST",
				Title:          "Onboard Transfer Paris->Brussels + Swap RDC Bvcmz5 at Schaarbeek",
				EstimatedCost:  refundInt + 1600 + 2800,
				Downgrades:     0,
				Cancellations:  0,
				NextTrainRisk:  "MEDIUM",
				RecommendedCar: secondarySpare,
				ActionSteps: []string{
					"Seat Car 34 passengers temporarily in heated lounge/spare seats for the short Paris Nord -> Bruxelles-Midi leg (no Le Bourget shunt).",
					"Pre-position homologated " + secondarySpare + " at Brussels for rapid crew-hub exchange (after passengers step onto platform).",
					"Restore all 50 Classic & Women-Only passengers into heated berths before night leg to Hamburg-Harburg & Berlin.",
				},
				GuardrailNotes: []string{
					"RDC Bdcm36 (61 80 82-90 036-5) remains BLOCKED on FR segment; RDC Bvcmz5 is FR/BE/DE approved.",
					"Maintains 9 occupied commercial cars (+€1,600 2nd conductor already budgeted).",
				},
				TcsGermanOrder: fmt.Sprintf(
					"DISPOSITIONSANORDNUNG TCS — ZUG ES %s (TAUSCH BRUXELLES-MIDI / DDL)\n"+
						"1. Ersatzwagen %s in Denderleeuw (ddl) abholen, vorheizen und nach Bruxelles-Midi überführen.\n"+
						"2. Nach Ausstieg der Reisenden in Bruxelles-Midi Schadwagen 34 (61 80 50-91 029-4) abstellen und %s beistellen (Zuglänge 9/13 Wagen FR-konform).\n"+
						"3. Personalübergabe Bruxelles-Midi koordinieren; Weiterfahrt Richtung Hamburg-Harburg und Berlin Hbf.",
					req.TrainID, secondarySpare, secondarySpare,
				),
				PassengerAlert: fmt.Sprintf(
					"European Sleeper ES %s: Coach 34 has a heating fault at Paris Nord (~%d min expected delay). "+
						"To depart Paris promptly, Coach 34 travellers will ride in heated seats to Brussels, where a pre-heated replacement couchette coach awaits for the overnight journey to Hamburg and Berlin (Women-Only compartments preserved). "+
						"Under Regulation (EU) 2021/782, you are entitled to complimentary refreshments, 25%% delay compensation (60–119 min), or Art. 18 rerouting/refund.",
					req.TrainID, req.ExpectedDelayMin,
				),
			},
			{
				ID:             "PROTECT_NEXT_TRAIN",
				Title:          "Swap at Brussels + Immediate Wustermark HVAC Repair Slot",
				EstimatedCost:  refundInt + 1600 + 5000,
				Downgrades:     0,
				Cancellations:  0,
				NextTrainRisk:  "LOW",
				RecommendedCar: secondarySpare,
				ActionSteps: []string{
					"Attach pre-heated " + secondarySpare + " at Brussels so Berlin arrival stays within the 5h (3h cleaning + 2h logistics) turnaround window for return ES 474.",
					"Send unheated Car 34 directly into Schaarbeek / Wustermark HVAC repair (~€5,000) so next cycle's consist returns to 9 healthy coaches.",
				},
				GuardrailNotes: []string{
					"Protects 5-hour Berlin turnaround for return ES 474 to Brussels & Paris Nord.",
					"Complies with 13-car FR limit and zero-passenger shunting rule.",
				},
				TcsGermanOrder: fmt.Sprintf(
					"DISPOSITIONSANORDNUNG TCS — ZUG ES %s / UMLAUFSCHUTZ ES 474\n"+
						"1. Ersatzwagen %s aus ddl abholen und in Bruxelles-Midi für Wagen 34 beistellen.\n"+
						"2. Schadwagen 34 (61 80 50-91 029-4) in Schaarbeek abstellen und Werkstattzuführung veranlassen.\n"+
						"3. 5h-Wendezeit (3h Reinigung + 2h Logistik) in Berlin für Gegenzug ES 474 strikt einhalten.",
					req.TrainID, secondarySpare,
				),
				PassengerAlert: fmt.Sprintf(
					"European Sleeper ES %s: A replacement heated couchette coach is being attached at Brussels (~%d min expected arrival delay). "+
						"All Classic and Women-Only berths are honoured overnight to Hamburg and Berlin, and EU 2021/782 25%% compensation + refreshments apply.",
					req.TrainID, req.ExpectedDelayMin,
				),
			},
		}

	default:
		// Preset 1 (Milan ES 401 Flat Wheel >6mm) & general single-coach incidents
		scenarios = []domain.ScenarioCard{
			{
				ID:             "LOWEST_COST",
				Title:          "Detach Car 56 Only & Reallocate Onboard (Keep 1 Conductor)",
				EstimatedCost:  refundInt + 3400,
				Downgrades:     22,
				Cancellations:  0,
				NextTrainRisk:  "MEDIUM",
				RecommendedCar: unfilteredCandidate,
				ActionSteps: []string{
					"Evacuate all 30 passengers from Car 56 (RDC WL33, 61 80 72-90 034-9) onto platform before uncoupling (zero passenger shunting).",
					"Move the Pet Private Compartment party into the last spare private compartment in Car 55 (RDC Bvcmz5) — never into shared berths.",
					"Absorb remaining 26 Comfort Plus passengers into spare Comfort Standard / Classic berths across Cars 52–57 (7 occupied cars -> avoids +€3,000 2nd conductor).",
				},
				GuardrailNotes: []string{
					"Avoids +€3,000 2nd-conductor cliff (stays <= 8 occupied commercial cars).",
					"Pet Private Compartment kept private; class-downgrade partial refunds issued.",
				},
				TcsGermanOrder: fmt.Sprintf(
					"DISPOSITIONSANORDNUNG TCS — ZUG ES %s (BRÜSSEL -> MILANO PORTA GARIBALDI)\n"+
						"1. Reisende aus Wagen 56 (RDC WL33, 61 80 72-90 034-9, Flachstelle >6mm) am Bahnsteig aussteigen lassen.\n"+
						"2. Schadwagen 56 aussetzen / abstellen zur Unterflurdrehbank Denderleeuw (ddl).\n"+
						"3. Ersatzgestellung: %s.\n"+
						"4. Zugbildung (7/13 Wagen Gotthard/RFI-konform) prüfen; Anschluss Aarau -> Zürich HB (~35 Min) an SBB/Arenaways vormelden.",
					req.TrainID, unfilteredOrderCar,
				),
				PassengerAlert: fmt.Sprintf(
					"European Sleeper ES %s (Brussels – Milano Porta Garibaldi): Due to a wheel-set safety defect on Sleeper Coach 56, departure is delayed by ~%d minutes. "+
						"Coach 56 passengers are being re-accommodated onboard (Pet Private Compartment bookings remain strictly in private compartments, and downgraded berths receive automatic fare-difference reimbursement). "+
						"Please note the Aarau domestic connection to Zürich HB (~35 min) may shift to the next SBB service. Under EU 2021/782, 25%% delay compensation (60–119 min), refreshments, and Art. 18 rerouting/refund rights apply.",
					req.TrainID, req.ExpectedDelayMin,
				),
			},
			{
				ID:             "PAX_FIRST",
				Title:          "1-for-1 Swap with RDC WL33 at Denderleeuw (Zero Downgrades, 1 Conductor)",
				EstimatedCost:  refundInt + 2900,
				Downgrades:     0,
				Cancellations:  0,
				NextTrainRisk:  "LOW",
				RecommendedCar: primarySpare,
				ActionSteps: []string{
					"Disembark Car 56 passengers safely; detach faulty Car 56 (61 80 72-90 034-9) for wheel-lathe turning.",
					"Fetch Italy/Gotthard-homologated " + primarySpare + " and couple 1-for-1 matching WL33 door orientation (keeps 8 commercial cars = 1 conductor, saving €3,000).",
					"Transfer all 30 Comfort Plus passengers + Pet Private Compartment 1-for-1 into identical WL33 sleeper cabins.",
					"Alert SBB & Arenaways control of ~95m late Gotthard pass and Aarau -> Zürich HB connection recovery.",
				},
				GuardrailNotes: []string{
					"EE Bvcmz6 and GFF Bimz 309 remain strictly BLOCKED for Italy; RDC WL33 holds full BE/DE/CH/IT approval.",
					"1-for-1 swap stays at 8 commercial cars (avoids +€3,000 2nd Milan conductor).",
				},
				TcsGermanOrder: fmt.Sprintf(
					"DISPOSITIONSANORDNUNG TCS — ZUG ES %s (1:1-WAGENTAUSCH GOTTHARD / ITALIEN)\n"+
						"1. Ersatzschlafwagen %s im Depot Denderleeuw (ddl) abholen (Türorientierung WL33 prüfen!).\n"+
						"2. Nach vollständigem Ausstieg der Reisenden Schadwagen 56 (RDC WL33, 61 80 72-90 034-9, Flachstelle >6,2mm) abstellen und zur Radsatzdrehbank ddl überführen.\n"+
						"3. Italien-/Schweiz-zugelassenen %s an Position 2 beistellen (8 Kommerzwagen = 1 Zugbegleiter, 8/13 Wagen CH/IT-Limit).\n"+
						"4. Bremsprobe durchführen; Übergabe Chiasso / Como S.G. an Arenaways Richtung Milano Porta Garibaldi (+%d Min) avisieren.",
					req.TrainID, primarySpare, primarySpare, req.ExpectedDelayMin,
				),
				PassengerAlert: fmt.Sprintf(
					"European Sleeper ES %s (Brussels – Milano Porta Garibaldi): A wheel flat on Coach 56 requires a safety coach swap before departure (~%d min expected delay). "+
						"An identical Italy-approved RDC Comfort Plus sleeper coach is being attached now—all sleeper compartments (including Pet Private Compartments) remain unchanged with zero downgrades. "+
						"Travellers connecting at Aarau for Zürich HB may take the next SBB domestic train with full ticket validity. Under Regulation (EU) 2021/782, complimentary refreshments, 25%% Art. 19 compensation (payable in cash/bank transfer on request), and Art. 18 rerouting/refund options apply (communicated within 100 minutes).",
					req.TrainID, req.ExpectedDelayMin,
				),
			},
			{
				ID:             "PROTECT_NEXT_TRAIN",
				Title:          "1-for-1 RDC WL33 Swap + Priority Crailsheim/ddl Wheel-Lathe Slot",
				EstimatedCost:  refundInt + 2900 + 6000,
				Downgrades:     0,
				Cancellations:  0,
				NextTrainRisk:  "LOW",
				RecommendedCar: primarySpare,
				ActionSteps: []string{
					"Execute 1-for-1 swap with " + primarySpare + " so ES 401 reaches Milano Porta Garibaldi with >5h buffer before return ES 400 (3h cleaning + 2h logistics).",
					"Book immediate overnight wheel-lathe slot (ddl or Karlsruhe–Crailsheim ~€6,000 workshop transfer) to restore spare WL33 pool within 36h.",
				},
				GuardrailNotes: []string{
					"Guarantees 5-hour turnaround at Milano Porta Garibaldi for return ES 400.",
					"Includes €6,000 Karlsruhe–Crailsheim / urgent wheel-lathe reserve replenishment.",
				},
				TcsGermanOrder: fmt.Sprintf(
					"DISPOSITIONSANORDNUNG TCS — ZUG ES %s / UMLAUFSCHUTZ ES 400 MILANO\n"+
						"1. %s in ddl abholen und 1:1 gegen Schadwagen 56 (61 80 72-90 034-9) beistellen.\n"+
						"2. Schadwagen 56 abstellen und Sofortzuführung Radsatzbearbeitung (ddl / Überführung Karlsruhe–Crailsheim) disponieren.\n"+
						"3. 5h-Wendezeitfenster (3h Reinigung + 2h Logistik) in Milano Porta Garibaldi für ES 400 sichern.",
					req.TrainID, primarySpare,
				),
				PassengerAlert: fmt.Sprintf(
					"European Sleeper ES %s (Brussels – Milano Porta Garibaldi): We are swapping Coach 56 for an identical RDC sleeper coach (~%d min expected delay), preserving all Comfort Plus and Pet Private bookings and protecting return service ES 400. "+
						"EU 2021/782 25%% compensation and refreshments apply.",
					req.TrainID, req.ExpectedDelayMin,
				),
			},
		}
	}

	// Select primary top-level TCS order & alert (PAX_FIRST by default unless canceled/Paris lowest-cost)
	primaryIdx := 1
	if req.Route == domain.RouteParis && !req.SimulateUnfilteredAI {
		primaryIdx = 0
	}
	return ResolutionOutput{
		Source:         "gemini-mock",
		Model:          modelName,
		Scenarios:      scenarios,
		TcsGermanOrder: scenarios[primaryIdx].TcsGermanOrder,
		PassengerAlert: scenarios[primaryIdx].PassengerAlert,
		FallbackNote:   fallbackNote,
	}
}
