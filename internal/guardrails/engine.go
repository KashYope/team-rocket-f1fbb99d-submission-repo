package guardrails

import (
	"fmt"
	"math"
	"slices"
	"strings"

	"es-incident-command/internal/data"
	"es-incident-command/internal/domain"
)

// EvaluatePreChecks executes deterministic European Sleeper rolling-stock homologation,
// train-length, shunting, crew-cost, product-protection, and Regulation (EU) 2021/782 rules
// before any System 1 (Laya) or System 2 (Gemini) model is invoked.
func EvaluatePreChecks(
	req domain.IncidentRequest,
	spares []domain.SpareCar,
) ([]domain.RuleCheck, []domain.SpareCar, []domain.SpareCar, domain.EU2021782Assessment) {
	corridor := data.FindCorridor(req.Route)
	consist := data.BuildActiveConsist(req)
	isCanceled := req.IssueType == "CANCELED" || req.ExpectedDelayMin >= 999

	var checks []domain.RuleCheck
	var eligible []domain.SpareCar
	var rejected []domain.SpareCar

	// 1. Rolling-Stock Homologation Filter
	var blockedLabels []string
	for _, car := range spares {
		carCopy := car
		blockedReason := checkHomologation(req.Route, carCopy)
		if blockedReason != "" {
			carCopy.BlockedReason = blockedReason
			rejected = append(rejected, carCopy)
			blockedLabels = append(blockedLabels, fmt.Sprintf("%s %s (%s)", carCopy.Owner, carCopy.CarType, carCopy.UIC))
		} else {
			eligible = append(eligible, carCopy)
		}
	}

	if len(rejected) > 0 {
		checks = append(checks, domain.RuleCheck{
			ID:       "HOMOLOGATION_FILTER",
			Category: "HOMOLOGATION",
			Severity: "BLOCK",
			Summary:  fmt.Sprintf("%d spare car(s) blocked for %s corridor homologation", len(rejected), req.Route),
			Detail:   fmt.Sprintf("Blocked from %s: %s. Only RIC/homologated stock (%d eligible spare(s), RDC universal fleet) may be dispatched.", req.Route, strings.Join(blockedLabels, "; "), len(eligible)),
		})
	} else {
		checks = append(checks, domain.RuleCheck{
			ID:       "HOMOLOGATION_FILTER",
			Category: "HOMOLOGATION",
			Severity: "PASS",
			Summary:  fmt.Sprintf("All %d reserve coaches homologated for %s corridor", len(eligible), req.Route),
			Detail:   "RDC, EE, and GFF reserve stock in Denderleeuw (ddl) and Schaarbeek hold valid BE/NL/DE/CZ approvals.",
		})
	}

	// 2. Train Length, Shunting & HV Safety Constraints
	activeCount := len(consist)
	if !isCanceled && activeCount+1 > corridor.MaxCarsCorridor {
		checks = append(checks, domain.RuleCheck{
			ID:       "CORRIDOR_MAX_LENGTH",
			Category: "SAFETY",
			Severity: "BLOCK",
			Summary:  fmt.Sprintf("Adding coach without detaching exceeds %d-car %s cap", corridor.MaxCarsCorridor, req.Route),
			Detail:   fmt.Sprintf("Active consist has %d cars; %s.", activeCount, corridor.MaxCarsDetail),
		})
	} else {
		sev := "PASS"
		summary := fmt.Sprintf("Consist length (%d/%d cars) within %s corridor limit", activeCount, corridor.MaxCarsCorridor, req.Route)
		if req.Route == domain.RoutePrague {
			sev = "WARN"
			summary = fmt.Sprintf("Consist %d/15 cars — ProRail ES 453 15-coach track-length note", activeCount)
		}
		checks = append(checks, domain.RuleCheck{
			ID:       "CORRIDOR_MAX_LENGTH",
			Category: "SAFETY",
			Severity: sev,
			Summary:  summary,
			Detail:   corridor.MaxCarsDetail,
		})
	}

	// Passenger onboard shunting ban
	if !isCanceled {
		shuntSev := "WARN"
		if req.IssueType == "SHUNT_BLOCKED" {
			shuntSev = "BLOCK"
		}
		checks = append(checks, domain.RuleCheck{
			ID:       "NO_PAX_SHUNTING",
			Category: "SAFETY",
			Severity: shuntSev,
			Summary:  "No shunting permitted with passengers onboard",
			Detail:   "Affected coach must be evacuated prior to yard uncoupling/beistellen, or locked out empty in-consist if safe to run.",
		})
	}

	// Paris Le Bourget shunting restriction & Bruxelles-Midi handover
	if req.Route == domain.RouteParis && !isCanceled {
		checks = append(checks, domain.RuleCheck{
			ID:       "PARIS_LE_BOURGET_SHUNT",
			Category: "SAFETY",
			Severity: "WARN",
			Summary:  "Paris Le Bourget shunting restricted · Bruxelles-Midi crew hub risk",
			Detail:   "SNCF Paris Nord / Le Bourget shunting slots restrict ad-hoc coach insertion; evaluate onboard absorption to Bruxelles-Midi crew handover hub or swap at Brussels Schaarbeek/ddl.",
		})
	}

	// WL30/WL33 orientation, Gotthard/Aarau connection & NO_HV rear-only rule
	upperCar := strings.ToUpper(req.AffectedCar)
	upperExtra := strings.ToUpper(req.ExtraInfo)
	if req.IssueType == "NO_HV" || strings.Contains(upperExtra, "HV") {
		checks = append(checks, domain.RuleCheck{
			ID:       "NO_HV_REAR_ONLY",
			Category: "SAFETY",
			Severity: "BLOCK",
			Summary:  "NO_HV coach restricted to extreme rear · zero loco reversal",
			Detail:   "Coach lacking 1000V/1500V/3000V UIC high-voltage train-line continuity cannot sit mid-consist or traverse reversal terminals.",
		})
	} else if strings.Contains(upperCar, "WL30") || strings.Contains(upperCar, "WL33") || req.Route == domain.RouteMilan {
		detail := "WL30/WL33 sleeper door/vestibule orientation must match platform boarding side during yard replacement."
		if req.Route == domain.RouteMilan {
			detail += " Note ES 401 serves Zürich via Aarau (~30–35m domestic connection) to avoid Zürich HB loco reversal before Gotthard climb."
		}
		checks = append(checks, domain.RuleCheck{
			ID:       "WL_DOOR_AND_ROUTING",
			Category: "SAFETY",
			Severity: "WARN",
			Summary:  "WL30/WL33 door orientation & Aarau/Zürich connection check",
			Detail:   detail,
		})
	}

	// 3. Crew Economics & Turnaround Thresholds
	commercialCars := 0
	for _, c := range consist {
		if c.IsCommercial {
			commercialCars++
		}
	}
	if !isCanceled {
		if commercialCars == 8 {
			checks = append(checks, domain.RuleCheck{
				ID:       "CONDUCTOR_CLIFF_8_CAR",
				Category: "ECONOMICS",
				Severity: "WARN",
				Summary:  fmt.Sprintf("At 8-car single-conductor boundary (+€%d if 9th occupied car added)", corridor.SecondConductorEUR),
				Detail:   fmt.Sprintf("1-for-1 coach swap keeps 8 occupied commercial cars (1 conductor). Attaching a 9th commercial car without detaching triggers mandatory 2nd conductor (+€%d on %s).", corridor.SecondConductorEUR, req.Route),
			})
		} else if commercialCars >= 9 {
			checks = append(checks, domain.RuleCheck{
				ID:       "CONDUCTOR_CLIFF_ACTIVE",
				Category: "ECONOMICS",
				Severity: "WARN",
				Summary:  fmt.Sprintf("Consist at %d commercial cars (+€%d 2nd-conductor tier active)", commercialCars, corridor.SecondConductorEUR),
				Detail:   fmt.Sprintf("Locking out faulty Car %s and absorbing onboard drops occupied commercial cars to 8, avoiding extra shunting delay while preserving crew margin.", req.AffectedCar),
			})
		}
	}

	if req.ExpectedDelayMin >= 60 || isCanceled {
		checks = append(checks, domain.RuleCheck{
			ID:       "TURNAROUND_AND_WORKSHOP",
			Category: "ECONOMICS",
			Severity: "WARN",
			Summary:  "5h turnaround buffer (3h cleaning + 2h logistics) & workshop transfer risk",
			Detail:   fmt.Sprintf("Delay of %d min risks return working turnaround. Workshop routing: %s.", req.ExpectedDelayMin, corridor.WorkshopRoute),
		})
	}

	// 4. Product & Accessibility Protections
	if req.HasPetPrivateComp {
		checks = append(checks, domain.RuleCheck{
			ID:       "PET_PRIVATE_COMP",
			Category: "PRODUCT",
			Severity: "BLOCK",
			Summary:  "Pet Private Compartment: shared downgrade prohibited",
			Detail:   "European Sleeper 2026 travel rules permit pets strictly in private compartments. Affected pet party must receive a private compartment in an eligible coach or dedicated rerouting.",
		})
	}

	if req.HasWomenOnlyComp {
		checks = append(checks, domain.RuleCheck{
			ID:       "WOMEN_ONLY_COMP",
			Category: "PRODUCT",
			Severity: "WARN",
			Summary:  "Women-Only compartment protection active",
			Detail:   "Passengers booked in Women-Only couchette/sleeper compartments cannot be reallocated into mixed-gender berths without explicit individual consent.",
		})
	}

	if req.HasPRMPax {
		checks = append(checks, domain.RuleCheck{
			ID:       "PRM_PRIORITY_ART_20_5",
			Category: "PRODUCT",
			Severity: "BLOCK",
			Summary:  "PRM passenger onboard: mandatory HUMAN_REVIEW & Art. 20(5) priority",
			Detail:   "Vintage fleet accommodates foldable wheelchairs only (no full PRM cabin). EU 2021/782 Art. 20(5) requires priority accessible rerouting/hotel assistance uncapped by force-majeure night limits.",
		})
	}

	checks = append(checks, domain.RuleCheck{
		ID:       "PRODUCT_2026_BASELINE",
		Category: "PRODUCT",
		Severity: "PASS",
		Summary:  "2026 product baseline: zero bicycle bookings · international tickets only",
		Detail:   "Bicycles are non-bookable across 2026 timetables; domestic-only BE/NL/FR/IT/CH/CZ travel is restricted.",
	})

	// 5. Regulation (EU) 2021/782 Statutory Liability Calculator
	eu := computeEU2021782(req, corridor, isCanceled)
	euSev := "PASS"
	if eu.CompensationBandPct >= 50 || eu.HotelRequired {
		euSev = "BLOCK"
	} else if eu.CompensationBandPct >= 25 || eu.RerouteObligation {
		euSev = "WARN"
	}
	checks = append(checks, domain.RuleCheck{
		ID:       "EU_2021_782_LIABILITY",
		Category: "EU_2021_782",
		Severity: euSev,
		Summary:  fmt.Sprintf("EU 2021/782: %d%% comp (€%.0f est.) · 100m self-reroute=%t · Hotel=%t", eu.CompensationBandPct, eu.EstimatedRefundEUR, eu.SelfReroute100Min, eu.HotelRequired),
		Detail:   eu.LegalBasisNotes,
	})

	return checks, eligible, rejected, eu
}

func checkHomologation(route domain.RouteCode, car domain.SpareCar) string {
	routeStr := string(route)
	if route == domain.RouteMilan {
		if car.Owner == "EE" || car.Owner == "GFF" || strings.Contains(car.CarType, "309") {
			return fmt.Sprintf("Blocked for Italy/Gotthard (MILAN): %s %s lacks RFI/SBB homologation; only RDC cars permitted.", car.Owner, car.CarType)
		}
	}
	if route == domain.RouteParis {
		if strings.Contains(car.CarType, "Bdcm36") {
			return "Blocked for France (PARIS): Bdcm36 lacks SNCF Réseau corridor homologation."
		}
	}
	if !slices.Contains(car.ApprovedRoutes, routeStr) {
		return fmt.Sprintf("Blocked for %s: approved routes are [%s].", routeStr, strings.Join(car.ApprovedRoutes, ", "))
	}
	return ""
}

func computeEU2021782(req domain.IncidentRequest, corridor domain.CorridorInfo, isCanceled bool) domain.EU2021782Assessment {
	bandPct := 0
	if isCanceled || req.ExpectedDelayMin >= 120 {
		bandPct = 50
	} else if req.ExpectedDelayMin >= 60 {
		bandPct = 25
	}

	// Art. 19(10): Only extreme third-party events exempt Art. 19 compensation.
	// Infrastructure manager faults (Infrabel, ProRail, DB InfraGO, SNCF Réseau, SBB, RFI)
	// and operator technical faults are NEVER force-majeure exempt.
	forceMajeure := req.CauseCategory == "THIRD_PARTY_EXTREME"

	estimatedRefund := 0.0
	if !forceMajeure && bandPct > 0 {
		perTicket := req.AvgTicketEUR * (float64(bandPct) / 100.0)
		if perTicket >= 4.0 { // Art. 19(8) €4 minimum threshold
			estimatedRefund = math.Round(float64(req.AffectedPax) * perTicket)
		}
	}

	reroute := isCanceled || req.ExpectedDelayMin >= 60
	selfReroute100 := isCanceled || req.ExpectedDelayMin >= 60
	hotelReq := isCanceled || req.ExpectedDelayMin >= 120
	prmPriority := req.HasPRMPax

	var notes []string
	if forceMajeure {
		notes = append(notes, "Art. 19(10) extraordinary third-party circumstance exempts delay compensation, AFTER verifying duty of care.")
	} else {
		notes = append(notes, fmt.Sprintf("Art. 19 non-exempt (%s — infra manager & operator technical causes never qualify as force majeure): %d%% ticket compensation mandatory in cash/bank transfer upon request.", req.CauseCategory, bandPct))
	}
	if reroute {
		notes = append(notes, "Art. 18(1) & (3): Offer refund or earliest comparable rerouting; failure to communicate options within 100 min of scheduled departure activates passenger self-rerouting reimbursement.")
	}
	if hotelReq {
		notes = append(notes, fmt.Sprintf("Art. 20(2)(b): Next European Sleeper departure in %dh (3x/week service) — overnight hotel + station transfer required plus meals/refreshments.", corridor.NextDepartureHours))
	}
	if prmPriority {
		notes = append(notes, "Art. 20(5): PRM travellers receive priority accessible transport & accommodation without night-cap limitations.")
	}

	return domain.EU2021782Assessment{
		CompensationBandPct: bandPct,
		EstimatedRefundEUR:  estimatedRefund,
		ForceMajeureExempt:  forceMajeure,
		RerouteObligation:   reroute,
		SelfReroute100Min:   selfReroute100,
		HotelRequired:       hotelReq,
		PRMPriorityRequired: prmPriority,
		LegalBasisNotes:     strings.Join(notes, " "),
	}
}
