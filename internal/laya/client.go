package laya

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"strings"
	"time"

	"es-incident-command/internal/domain"
)

var severityLabels = []string{
	"Low severity: minor delay under 60 minutes with no berth loss",
	"Medium severity: single-coach issue absorbable onboard or via simple swap",
	"High severity: major delay over 60 minutes, downgrades, or turnaround risk",
	"Critical severity: full cancellation, blocked shunting, or overnight hotel cascade",
}

// Evaluate calls Laya System 1 (/v1/systemone) with "model": "english" and a 3-second timeout,
// falling back automatically to deterministic heuristic evaluation when LAYA_MODE=mock or unreachable.
// IMPORTANT: Never reads or gates on action.act_probability.
func Evaluate(
	ctx context.Context,
	req domain.IncidentRequest,
	eligibleCount int,
	checks []domain.RuleCheck,
) domain.LayaTriage {
	mode := strings.ToLower(strings.TrimSpace(os.Getenv("LAYA_MODE")))
	if mode == "mock" {
		triage := mockEvaluate(req, eligibleCount, checks)
		return applyRoutingGate(triage, req)
	}

	baseURL := strings.TrimRight(strings.TrimSpace(os.Getenv("LAYA_BASE_URL")), "/")
	if baseURL == "" {
		baseURL = "http://127.0.0.1:8000"
	}

	start := time.Now()
	liveTriage, err := callLiveLaya(ctx, baseURL, req, eligibleCount, checks)
	if err != nil {
		fallback := mockEvaluate(req, eligibleCount, checks)
		fallback.LatencyMs = time.Since(start).Milliseconds()
		if fallback.LatencyMs < 1 {
			fallback.LatencyMs = 1
		}
		return applyRoutingGate(fallback, req)
	}

	liveTriage.LatencyMs = time.Since(start).Milliseconds()
	return applyRoutingGate(liveTriage, req)
}

type systemOneRequest struct {
	Model     string         `json:"model"`
	State     string         `json:"state"`
	Questions map[string]any `json:"questions"`
}

func callLiveLaya(
	parentCtx context.Context,
	baseURL string,
	req domain.IncidentRequest,
	eligibleCount int,
	checks []domain.RuleCheck,
) (domain.LayaTriage, error) {
	ctx, cancel := context.WithTimeout(parentCtx, 3*time.Second)
	defer cancel()

	blockCount := 0
	for _, c := range checks {
		if c.Severity == "BLOCK" {
			blockCount++
		}
	}

	stateSummary, _ := json.Marshal(map[string]any{
		"trainId":            req.TrainID,
		"route":              req.Route,
		"issueType":          req.IssueType,
		"causeCategory":      req.CauseCategory,
		"expectedDelayMin":   req.ExpectedDelayMin,
		"affectedCar":        req.AffectedCar,
		"affectedPax":        req.AffectedPax,
		"eligibleSpareCount": eligibleCount,
		"blockRuleCount":     blockCount,
		"hasPrmPax":          req.HasPRMPax,
		"hasPetPrivateComp":  req.HasPetPrivateComp,
		"hasWomenOnlyComp":   req.HasWomenOnlyComp,
	})

	payload := systemOneRequest{
		Model: "english",
		State: string(stateSummary),
		Questions: map[string]any{
			"recovery_strategy": map[string]any{
				"type":         "choice",
				"instructions": "Select the primary operational recovery strategy for this European Sleeper disruption.",
				"criteria": map[string]string{
					"swap_spare_car":         "Replace the faulty coach using an eligible parked spare car at the depot",
					"downgrade_and_absorb":   "Keep train running without adding a car by moving affected passengers into spare berths or lower classes",
					"move_to_next_train":     "Rebook passengers onto the next scheduled European Sleeper departure on this corridor",
					"hotel_and_daytime_rail": "Provide overnight hotel accommodation and reroute passengers onto daytime international rail",
					"manual_escalation":      "Complex multi-constraint incident requiring senior Night Duty planner decision",
				},
			},
			"operational_severity": map[string]any{
				"type":         "score",
				"instructions": "Rate the operational and passenger severity of this night train disruption.",
				"criteria":     severityLabels,
			},
			"refund_cliff_risk": map[string]any{
				"type":         "noul",
				"instructions": "Is this incident likely to cause arrival delay of 60 minutes or more triggering EU 2021/782 compensation?",
				"criteria": map[string]string{
					"true":  "Expected or cascading delay reaches at least 60 minutes or train is canceled",
					"false": "Arrival delay remains below 60 minutes",
				},
			},
			"turnaround_breach": map[string]any{
				"type":         "noul",
				"instructions": "Is this incident likely to breach the 3-hour cleaning and 2-hour logistics turnaround window for the return train?",
				"criteria": map[string]string{
					"true":  "Turnaround buffer drops below 5 hours or faulty rolling stock requires workshop transfer",
					"false": "Return train turnaround window remains intact",
				},
			},
		},
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return domain.LayaTriage{}, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/v1/systemone", bytes.NewReader(bodyBytes))
	if err != nil {
		return domain.LayaTriage{}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if apiKey := strings.TrimSpace(os.Getenv("LAYA_API_KEY")); apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+apiKey)
	}

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return domain.LayaTriage{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return domain.LayaTriage{}, fmt.Errorf("laya status %d", resp.StatusCode)
	}

	rawBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return domain.LayaTriage{}, err
	}

	return parseLiveResponse(rawBytes)
}

func parseLiveResponse(rawBytes []byte) (domain.LayaTriage, error) {
	var raw map[string]any
	if err := json.Unmarshal(rawBytes, &raw); err != nil {
		return domain.LayaTriage{}, err
	}

	answersMap, _ := raw["answers"].(map[string]any)
	if answersMap == nil {
		answersMap, _ = raw["questions"].(map[string]any)
	}
	if answersMap == nil {
		return domain.LayaTriage{}, fmt.Errorf("missing answers in laya response")
	}

	recProbs := extractProbMap(answersMap["recovery_strategy"])
	if len(recProbs) == 0 {
		return domain.LayaTriage{}, fmt.Errorf("empty recovery_strategy probabilities")
	}
	choice, ansConf, entConf := summarizeChoice(recProbs)

	sevProbs := extractProbMap(answersMap["operational_severity"])
	scoreVal, nearest, sevAnsConf, sevEntConf := summarizeScore(sevProbs, answersMap["operational_severity"])

	refundTrue := extractNoulTrue(answersMap["refund_cliff_risk"])
	turnTrue := extractNoulTrue(answersMap["turnaround_breach"])

	return domain.LayaTriage{
		Source: "laya-live",
		Model:  "english",
		RecoveryStrategy: domain.LayaChoiceAnswer{
			Choice:           choice,
			Probabilities:    recProbs,
			Confidence:       entConf,
			AnswerConfidence: ansConf,
		},
		OperationalSeverity: domain.LayaScoreAnswer{
			Score:            scoreVal,
			NearestLabel:     nearest,
			Probabilities:    sevProbs,
			Confidence:       sevEntConf,
			AnswerConfidence: sevAnsConf,
		},
		RefundCliffRisk:  refundTrue,
		TurnaroundBreach: turnTrue,
	}, nil
}

func extractProbMap(entry any) map[string]float64 {
	out := map[string]float64{}
	m, ok := entry.(map[string]any)
	if !ok {
		return out
	}
	probsRaw, _ := m["probabilities"].(map[string]any)
	if probsRaw == nil {
		probsRaw, _ = m["distribution"].(map[string]any)
	}
	for k, v := range probsRaw {
		if f, ok := v.(float64); ok {
			out[k] = round4(f)
		}
	}
	return out
}

func extractNoulTrue(entry any) float64 {
	probs := extractProbMap(entry)
	if v, ok := probs["true"]; ok {
		return round4(v)
	}
	if m, ok := entry.(map[string]any); ok {
		if p, ok := m["p_true"].(float64); ok {
			return round4(p)
		}
	}
	return 0.5
}

func summarizeChoice(probs map[string]float64) (string, float64, float64) {
	bestKey := ""
	bestVal := -1.0
	for k, v := range probs {
		if v > bestVal || (v == bestVal && k < bestKey) {
			bestVal = v
			bestKey = k
		}
	}
	return bestKey, round4(bestVal), round4(normalizedEntropyConfidence(probs))
}

func summarizeScore(probs map[string]float64, raw any) (float64, string, float64, float64) {
	if len(probs) == 0 {
		probs = map[string]float64{"0": 0.10, "1": 0.40, "2": 0.35, "3": 0.15}
	}
	expected := 0.0
	maxP := 0.0
	for idx := 0; idx < 4; idx++ {
		key := fmt.Sprintf("%d", idx)
		p := probs[key]
		if p == 0 && idx < len(severityLabels) {
			p = probs[severityLabels[idx]]
		}
		expected += float64(idx) * p
		if p > maxP {
			maxP = p
		}
	}
	if m, ok := raw.(map[string]any); ok {
		if s, ok := m["score"].(float64); ok && s > 0 {
			expected = s
		}
	}
	nearestIdx := int(math.Round(expected))
	if nearestIdx < 0 {
		nearestIdx = 0
	}
	if nearestIdx >= len(severityLabels) {
		nearestIdx = len(severityLabels) - 1
	}
	shortLabels := []string{"Low", "Medium", "High", "Critical"}
	return round4(expected), shortLabels[nearestIdx], round4(maxP), round4(normalizedEntropyConfidence(probs))
}

func normalizedEntropyConfidence(probs map[string]float64) float64 {
	k := len(probs)
	if k <= 1 {
		return 1.0
	}
	h := 0.0
	for _, p := range probs {
		if p > 1e-12 {
			h -= p * math.Log(p)
		}
	}
	maxH := math.Log(float64(k))
	if maxH <= 0 {
		return 0
	}
	conf := 1.0 - (h / maxH)
	if conf < 0 {
		return 0
	}
	if conf > 1 {
		return 1
	}
	return conf
}

func mockEvaluate(req domain.IncidentRequest, eligibleCount int, _ []domain.RuleCheck) domain.LayaTriage {
	isCanceled := req.IssueType == "CANCELED" || req.ExpectedDelayMin >= 999

	var recProbs map[string]float64
	var sevProbs map[string]float64
	var refundRisk, turnaroundRisk float64

	switch {
	case isCanceled:
		// Full train cancellation (e.g., Preset 2 Prague ES 453)
		recProbs = map[string]float64{
			"swap_spare_car":         0.02,
			"downgrade_and_absorb":   0.03,
			"move_to_next_train":     0.14,
			"hotel_and_daytime_rail": 0.64,
			"manual_escalation":      0.17,
		}
		sevProbs = map[string]float64{"0": 0.01, "1": 0.03, "2": 0.14, "3": 0.82}
		refundRisk = 0.99
		turnaroundRisk = 0.94

	case req.Route == domain.RouteParis && (req.IssueType == "NO_HEATING" || req.IssueType == "SHUNT_BLOCKED"):
		// Paris Nord Le Bourget shunt constraint vs onboard absorption (Preset 3 ES 475)
		recProbs = map[string]float64{
			"swap_spare_car":         0.24,
			"downgrade_and_absorb":   0.56,
			"move_to_next_train":     0.05,
			"hotel_and_daytime_rail": 0.04,
			"manual_escalation":      0.11,
		}
		sevProbs = map[string]float64{"0": 0.04, "1": 0.28, "2": 0.56, "3": 0.12}
		refundRisk = 0.84
		turnaroundRisk = 0.62

	case eligibleCount > 0 && req.ExpectedDelayMin < 120 && !req.HasPRMPax:
		// Clean single-coach swap candidate (e.g., Preset 1 Milan ES 401 Flat Wheel with RDC WL33 at ddl)
		recProbs = map[string]float64{
			"swap_spare_car":         0.81,
			"downgrade_and_absorb":   0.10,
			"move_to_next_train":     0.02,
			"hotel_and_daytime_rail": 0.03,
			"manual_escalation":      0.04,
		}
		sevProbs = map[string]float64{"0": 0.06, "1": 0.34, "2": 0.48, "3": 0.12}
		if req.ExpectedDelayMin >= 60 {
			refundRisk = 0.88
			turnaroundRisk = 0.71
		} else {
			refundRisk = 0.31
			turnaroundRisk = 0.28
		}

	default:
		recProbs = map[string]float64{
			"swap_spare_car":         0.38,
			"downgrade_and_absorb":   0.31,
			"move_to_next_train":     0.07,
			"hotel_and_daytime_rail": 0.10,
			"manual_escalation":      0.14,
		}
		sevProbs = map[string]float64{"0": 0.05, "1": 0.25, "2": 0.50, "3": 0.20}
		if req.ExpectedDelayMin >= 60 {
			refundRisk = 0.86
			turnaroundRisk = 0.68
		} else {
			refundRisk = 0.29
			turnaroundRisk = 0.24
		}
	}

	choice, ansConf, entConf := summarizeChoice(recProbs)
	scoreVal, nearest, sevAnsConf, sevEntConf := summarizeScore(sevProbs, nil)

	return domain.LayaTriage{
		Source:    "laya-mock",
		Model:     "english",
		LatencyMs: 2,
		RecoveryStrategy: domain.LayaChoiceAnswer{
			Choice:           choice,
			Probabilities:    recProbs,
			Confidence:       entConf,
			AnswerConfidence: ansConf,
		},
		OperationalSeverity: domain.LayaScoreAnswer{
			Score:            scoreVal,
			NearestLabel:     nearest,
			Probabilities:    sevProbs,
			Confidence:       sevEntConf,
			AnswerConfidence: sevAnsConf,
		},
		RefundCliffRisk:  refundRisk,
		TurnaroundBreach: turnaroundRisk,
	}
}

// applyRoutingGate enforces the exact System 1 -> System 2 routing invariant without touching act_probability.
func applyRoutingGate(triage domain.LayaTriage, req domain.IncidentRequest) domain.LayaTriage {
	var reasons []string

	if triage.RecoveryStrategy.Choice == "manual_escalation" {
		reasons = append(reasons, "top strategy is manual_escalation")
	}
	if triage.RecoveryStrategy.AnswerConfidence < 0.75 {
		reasons = append(reasons, fmt.Sprintf("answerConfidence %.2f < 0.75", triage.RecoveryStrategy.AnswerConfidence))
	}
	if triage.RecoveryStrategy.Confidence < 0.55 {
		reasons = append(reasons, fmt.Sprintf("entropy confidence %.2f < 0.55", triage.RecoveryStrategy.Confidence))
	}
	if triage.OperationalSeverity.Score >= 2.2 {
		reasons = append(reasons, fmt.Sprintf("severity score %.2f >= 2.20", triage.OperationalSeverity.Score))
	}
	if req.HasPRMPax {
		reasons = append(reasons, "PRM passenger onboard (EU 2021/782 Art. 20(5))")
	}
	if req.ExpectedDelayMin >= 120 || req.IssueType == "CANCELED" {
		reasons = append(reasons, fmt.Sprintf("expected delay %d min >= 120m cliff", req.ExpectedDelayMin))
	}

	if len(reasons) == 0 {
		triage.RoutingGate = "AUTO_CANDIDATE"
		triage.RoutingReason = fmt.Sprintf(
			"All 6 System-1 gates passed: %s (max p=%.2f >= 0.75, entropy=%.2f >= 0.55, severity=%.2f < 2.20, delay=%dm < 120m, no PRM)",
			triage.RecoveryStrategy.Choice,
			triage.RecoveryStrategy.AnswerConfidence,
			triage.RecoveryStrategy.Confidence,
			triage.OperationalSeverity.Score,
			req.ExpectedDelayMin,
		)
	} else {
		triage.RoutingGate = "HUMAN_REVIEW"
		triage.RoutingReason = "Escalated to Night Duty Planner: " + strings.Join(reasons, "; ")
	}
	return triage
}

func round4(v float64) float64 {
	return math.Round(v*10000) / 10000
}
