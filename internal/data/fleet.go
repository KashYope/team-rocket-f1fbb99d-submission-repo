package data

import (
	"strings"

	"es-incident-command/internal/domain"
)

func GetFleetState() domain.FleetState {
	return domain.FleetState{
		TimetableBaseline: "Nov 2026 Timetable Baseline (Reg. EU 2021/782)",
		Corridors:         GetCorridors(),
		SpareCars:         GetSpareCars(),
	}
}

func GetCorridors() []domain.CorridorInfo {
	return []domain.CorridorInfo{
		{
			Route:              domain.RouteMilan,
			TrainIDs:           []string{"401", "400"},
			Title:              "ES 401/400 Brussels – Milano Porta Garibaldi",
			ScheduleSummary:    "3x/week: Outbound Mon/Thu/Sat (ES 401) · Return Wed/Fri/Sun (ES 400) · Historic Gotthard mountain line",
			InfraManagers:      []string{"Infrabel (BE)", "DB InfraGO (DE)", "SBB Infra (CH)", "RFI (IT)"},
			KeyNodes:           "Bruxelles-Midi → Verviers → Aachen Hbf → Köln Hbf → Aarau (Zürich ~30-35m conn.) → Arth-Goldau → Göschenen (~30m dwell) → Bellinzona → Lugano → Como S.G. (Chiasso) → Milano Porta Garibaldi (TCS north / Arenaways IT)",
			MaxCarsCorridor:    13,
			MaxCarsDetail:      "13 cars max (CH/IT Gotthard & RFI cap); 15 BE, 16 DE",
			SecondConductorEUR: 3000,
			WorkshopRoute:      "Karlsruhe – Crailsheim (~€6,000 transfer) or Denderleeuw (ddl) wheel-lathe",
			NextDepartureHours: 48,
			DefaultConsist: []domain.ConsistCar{
				{Position: 1, CarNumber: "57", UIC: "61 80 72-90 031-5", Owner: "RDC", CarType: "WL30", ProductClass: "Comfort Plus", BookedPax: 26, Capacity: 30, IsCommercial: true, SpecialNotes: "Door orientation fixed"},
				{Position: 2, CarNumber: "56", UIC: "61 80 72-90 034-9", Owner: "RDC", CarType: "WL33", ProductClass: "Comfort Plus", BookedPax: 30, Capacity: 33, IsCommercial: true, SpecialNotes: "1 Pet Private Comp"},
				{Position: 3, CarNumber: "55", UIC: "61 80 50-91 021-1", Owner: "RDC", CarType: "Bvcmz5", ProductClass: "Comfort Standard", BookedPax: 28, Capacity: 30, IsCommercial: true},
				{Position: 4, CarNumber: "54", UIC: "61 80 50-91 022-9", Owner: "RDC", CarType: "Bvcmz5", ProductClass: "Comfort Standard", BookedPax: 30, Capacity: 30, IsCommercial: true},
				{Position: 5, CarNumber: "53", UIC: "61 80 50-91 024-5", Owner: "RDC", CarType: "Bocmh6", ProductClass: "Classic", BookedPax: 44, Capacity: 48, IsCommercial: true},
				{Position: 6, CarNumber: "52", UIC: "61 80 50-91 026-0", Owner: "RDC", CarType: "Bocmh6", ProductClass: "Classic", BookedPax: 42, Capacity: 48, IsCommercial: true, SpecialNotes: "Women-Only Comp"},
				{Position: 7, CarNumber: "51", UIC: "61 80 29-90 011-7", Owner: "RDC", CarType: "Am66", ProductClass: "Budget", BookedPax: 44, Capacity: 54, IsCommercial: true},
				{Position: 8, CarNumber: "50", UIC: "61 80 22-90 072-3", Owner: "RDC", CarType: "Bm72", ProductClass: "Budget", BookedPax: 46, Capacity: 60, IsCommercial: true, SpecialNotes: "8-car single-conductor limit"},
			},
		},
		{
			Route:              domain.RoutePrague,
			TrainIDs:           []string{"453", "452"},
			Title:              "ES 453/452 Brussels – Praha hl.n.",
			ScheduleSummary:    "3x/week: Outbound Mon/Wed/Fri (ES 453) · Return Tue/Thu/Sun (ES 452)",
			InfraManagers:      []string{"Infrabel (BE)", "ProRail (NL)", "DB InfraGO (DE)", "Správa železnic (CZ)"},
			KeyNodes:           "Bruxelles-Midi → Antwerp → Roosendaal → Rotterdam → Den Haag HS → Amsterdam → Amersfoort → Deventer → Bad Bentheim (loco handover) → Berlin Gesundbrunnen → Dresden Hbf → Bad Schandau → Děčín → Ústí nad Labem → Praha hl.n.",
			MaxCarsCorridor:    15,
			MaxCarsDetail:      "15 cars max BE/NL (ProRail 2027 Network Statement flags ES 453 15-coach platform track-length conflict); 16 DE",
			SecondConductorEUR: 1600,
			WorkshopRoute:      "Bad Bentheim – Münster (emstg) – Lengerich (~€5,000 transfer) or Berlin Wustermark",
			NextDepartureHours: 48,
			DefaultConsist: []domain.ConsistCar{
				{Position: 1, CarNumber: "18", UIC: "61 80 72-90 032-3", Owner: "RDC", CarType: "WL33", ProductClass: "Comfort Plus", BookedPax: 31, Capacity: 33, IsCommercial: true},
				{Position: 2, CarNumber: "17", UIC: "61 80 50-91 023-7", Owner: "RDC", CarType: "Bvcmz5", ProductClass: "Comfort Standard", BookedPax: 29, Capacity: 30, IsCommercial: true, SpecialNotes: "Foldable PRM pax"},
				{Position: 3, CarNumber: "16", UIC: "61 84 50-70 602-2", Owner: "EE", CarType: "Bvcmz6", ProductClass: "Classic", BookedPax: 48, Capacity: 60, IsCommercial: true},
				{Position: 4, CarNumber: "15", UIC: "61 84 50-70 603-0", Owner: "EE", CarType: "Bvcmz6", ProductClass: "Classic", BookedPax: 46, Capacity: 60, IsCommercial: true, SpecialNotes: "Women-Only Comp"},
				{Position: 5, CarNumber: "14", UIC: "61 80 82-90 035-7", Owner: "RDC", CarType: "Bdcm36", ProductClass: "Classic", BookedPax: 34, Capacity: 36, IsCommercial: true},
				{Position: 6, CarNumber: "13", UIC: "51 80 22-94 301-4", Owner: "GFF", CarType: "Bimz", ProductClass: "Budget", BookedPax: 52, Capacity: 60, IsCommercial: true},
				{Position: 7, CarNumber: "12", UIC: "51 80 22-94 304-8", Owner: "GFF", CarType: "Bimz", ProductClass: "Budget", BookedPax: 50, Capacity: 60, IsCommercial: true},
				{Position: 8, CarNumber: "11", UIC: "51 80 22-94 307-1", Owner: "GFF", CarType: "Bimz", ProductClass: "Budget", BookedPax: 50, Capacity: 60, IsCommercial: true, SpecialNotes: "8-car single-conductor limit"},
			},
		},
		{
			Route:              domain.RouteParis,
			TrainIDs:           []string{"475", "474"},
			Title:              "ES 475/474 Paris Nord – Brussels – Berlin",
			ScheduleSummary:    "3x/week: Outbound Tue/Thu/Sun from Paris Nord (ES 475) · Return Mon/Wed/Fri from Berlin (ES 474)",
			InfraManagers:      []string{"SNCF Réseau (FR)", "Infrabel (BE)", "DB InfraGO (DE)"},
			KeyNodes:           "Paris Nord (Le Bourget shunt restriction) → Aulnoye-Aymeries → Mons → Bruxelles-Midi (onboard crew handover hub) → Liège-Guillemins → Hamburg-Harburg → Berlin Hbf (multi-system loco through-hauled)",
			MaxCarsCorridor:    13,
			MaxCarsDetail:      "13 cars max FR (SNCF Réseau cap); 15 BE, 16 DE · Active consist already at 9 commercial cars (2 conductors active)",
			SecondConductorEUR: 1600,
			WorkshopRoute:      "Brussels Schaarbeek / Denderleeuw (ddl) or Berlin Wustermark (Paris Le Bourget shunting restricted)",
			NextDepartureHours: 48,
			DefaultConsist: []domain.ConsistCar{
				{Position: 1, CarNumber: "38", UIC: "61 80 72-90 030-7", Owner: "RDC", CarType: "WL30", ProductClass: "Comfort Plus", BookedPax: 28, Capacity: 30, IsCommercial: true},
				{Position: 2, CarNumber: "37", UIC: "61 80 72-90 035-6", Owner: "RDC", CarType: "WL33", ProductClass: "Comfort Plus", BookedPax: 31, Capacity: 33, IsCommercial: true},
				{Position: 3, CarNumber: "36", UIC: "61 80 50-91 027-8", Owner: "RDC", CarType: "Bvcmz5", ProductClass: "Comfort Standard", BookedPax: 28, Capacity: 30, IsCommercial: true},
				{Position: 4, CarNumber: "35", UIC: "61 80 50-91 028-6", Owner: "RDC", CarType: "Bvcmz5", ProductClass: "Comfort Standard", BookedPax: 29, Capacity: 30, IsCommercial: true},
				{Position: 5, CarNumber: "34", UIC: "61 80 50-91 029-4", Owner: "RDC", CarType: "Bvcmz5", ProductClass: "Classic", BookedPax: 50, Capacity: 54, IsCommercial: true, SpecialNotes: "Women-Only Comp"},
				{Position: 6, CarNumber: "33", UIC: "61 84 50-70 604-8", Owner: "EE", CarType: "Bvcmz6", ProductClass: "Classic", BookedPax: 54, Capacity: 60, IsCommercial: true},
				{Position: 7, CarNumber: "32", UIC: "61 84 50-70 605-5", Owner: "EE", CarType: "Bvcmz6", ProductClass: "Classic", BookedPax: 58, Capacity: 60, IsCommercial: true},
				{Position: 8, CarNumber: "31", UIC: "51 80 22-94 310-5", Owner: "GFF", CarType: "Bimz", ProductClass: "Budget", BookedPax: 66, Capacity: 66, IsCommercial: true, SpecialNotes: "8-car single-conductor boundary"},
				{Position: 9, CarNumber: "30", UIC: "51 80 22-94 311-3", Owner: "GFF", CarType: "Bimz", ProductClass: "Budget", BookedPax: 66, Capacity: 66, IsCommercial: true, ExceedsSingleConductor: true, SpecialNotes: "9th car (+€1,600 2nd conductor active)"},
			},
		},
	}
}

func GetSpareCars() []domain.SpareCar {
	return []domain.SpareCar{
		{
			UIC:            "61 80 72-90 033-1",
			Owner:          "RDC",
			CarType:        "WL33",
			ProductClass:   "Comfort Plus",
			Capacity:       33,
			Yard:           "ddl",
			StatusCode:     "X",
			ApprovedRoutes: []string{"PRAGUE", "MILAN", "PARIS"},
		},
		{
			UIC:            "61 80 50-91 025-2",
			Owner:          "RDC",
			CarType:        "Bvcmz5",
			ProductClass:   "Comfort Standard",
			Capacity:       30,
			Yard:           "ddl",
			StatusCode:     "X",
			ApprovedRoutes: []string{"PRAGUE", "MILAN", "PARIS"},
		},
		{
			UIC:            "61 84 50-70 601-4",
			Owner:          "EE",
			CarType:        "Bvcmz6",
			ProductClass:   "Classic",
			Capacity:       60,
			Yard:           "schaarbeek",
			StatusCode:     "X",
			ApprovedRoutes: []string{"PRAGUE", "PARIS"},
		},
		{
			UIC:            "51 80 22-94 309-7",
			Owner:          "GFF",
			CarType:        "Bimz 309",
			ProductClass:   "Budget",
			Capacity:       60,
			Yard:           "ddl",
			StatusCode:     "X",
			ApprovedRoutes: []string{"PRAGUE", "PARIS"},
		},
		{
			UIC:            "61 80 82-90 036-5",
			Owner:          "RDC",
			CarType:        "Bdcm36",
			ProductClass:   "Classic",
			Capacity:       36,
			Yard:           "schaarbeek",
			StatusCode:     "X",
			ApprovedRoutes: []string{"PRAGUE", "MILAN"},
		},
	}
}

func GetPresets() []domain.DemoPreset {
	return []domain.DemoPreset{
		{
			ID:       "MILAN_FLAT_WHEEL",
			Label:    "1. ES 401 Milan — Flat Wheel (>6mm)",
			Subtitle: "Car 56 (RDC WL33, 30 pax, Pet Private Comp) · 95m delay · Italy homologation & +€3,000 conductor cliff",
			Badge:    "ITALY HOMOLOGATION",
			Request: domain.IncidentRequest{
				PresetID:             "MILAN_FLAT_WHEEL",
				TrainID:              "401",
				Route:                domain.RouteMilan,
				IssueType:            "FLAT_WHEEL",
				CauseCategory:        "OPERATOR_TECH",
				ExpectedDelayMin:     95,
				AffectedCar:          "56 (RDC WL33)",
				AffectedPax:          30,
				AvgTicketEUR:         165.0,
				HasPRMPax:            false,
				HasPetPrivateComp:    true,
				HasWomenOnlyComp:     false,
				SimulateUnfilteredAI: false,
				ExtraInfo:            "Wheel flat >6.2mm detected during pre-departure brake test at Brussels. ES 401 operates at 8 commercial cars (single conductor limit). Aarau ~35m domestic connection to Zürich HB at risk.",
			},
		},
		{
			ID:       "PRAGUE_FULL_CANCEL",
			Label:    "2. ES 453 Prague — Full Cancellation",
			Subtitle: "ALL 8 cars (340 pax, PRM wheelchair pax) · Canceled 3h pre-dep · INFRA_MANAGER fault (non-exempt)",
			Badge:    "EU 2021/782 + PRM",
			Request: domain.IncidentRequest{
				PresetID:             "PRAGUE_FULL_CANCEL",
				TrainID:              "453",
				Route:                domain.RoutePrague,
				IssueType:            "CANCELED",
				CauseCategory:        "INFRA_MANAGER",
				ExpectedDelayMin:     999,
				AffectedCar:          "ALL",
				AffectedPax:          340,
				AvgTicketEUR:         128.0,
				HasPRMPax:            true,
				HasPetPrivateComp:    false,
				HasWomenOnlyComp:     true,
				SimulateUnfilteredAI: false,
				ExtraInfo:            "Overhead catenary & multi-system HV supply failure at Bad Bentheim handover reported 3h pre-departure by DB InfraGO / ProRail. Next ES 453 departure in 48h. 2 foldable-wheelchair PRM travellers booked in Car 17.",
			},
		},
		{
			ID:       "PARIS_HEATING_SHUNT",
			Label:    "3. ES 475 Paris–Berlin — Heating & Shunt Cliff",
			Subtitle: "Car 34 (Bvcmz5, 50 pax, Women-Only comp) · 75m delay at Paris Nord · Le Bourget shunt & FR Bdcm36 ban",
			Badge:    "FR SHUNT + WOMEN-ONLY",
			Request: domain.IncidentRequest{
				PresetID:             "PARIS_HEATING_SHUNT",
				TrainID:              "475",
				Route:                domain.RouteParis,
				IssueType:            "NO_HEATING",
				CauseCategory:        "OPERATOR_TECH",
				ExpectedDelayMin:     75,
				AffectedCar:          "34 (RDC Bvcmz5)",
				AffectedPax:          50,
				AvgTicketEUR:         115.0,
				HasPRMPax:            false,
				HasPetPrivateComp:    false,
				HasWomenOnlyComp:     true,
				SimulateUnfilteredAI: false,
				ExtraInfo:            "HVAC / converter failure on Car 34 at Paris Nord in sub-5°C ambient conditions. Le Bourget shunting slot restricted; active consist has 9 commercial cars (+€1,600 2nd conductor). Bruxelles-Midi crew handover buffer tight.",
			},
		},
	}
}

func FindCorridor(route domain.RouteCode) domain.CorridorInfo {
	for _, c := range GetCorridors() {
		if c.Route == route {
			return c
		}
	}
	return GetCorridors()[0]
}

func BuildActiveConsist(req domain.IncidentRequest) []domain.ConsistCar {
	corridor := FindCorridor(req.Route)
	out := make([]domain.ConsistCar, len(corridor.DefaultConsist))
	copy(out, corridor.DefaultConsist)

	affectedUpper := strings.ToUpper(strings.TrimSpace(req.AffectedCar))
	cancelAll := affectedUpper == "ALL" || req.IssueType == "CANCELED"

	commercialIdx := 0
	for i := range out {
		if out[i].IsCommercial {
			commercialIdx++
			out[i].ExceedsSingleConductor = commercialIdx > 8
		}
		if cancelAll {
			out[i].IsFaulty = true
			continue
		}
		if affectedUpper != "" && (strings.HasPrefix(affectedUpper, out[i].CarNumber+" ") || affectedUpper == out[i].CarNumber || strings.Contains(affectedUpper, " "+out[i].CarNumber+" ")) {
			out[i].IsFaulty = true
		}
	}
	return out
}
