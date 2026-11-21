package service

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/joho/godotenv"
)

func TestIncidentService_OfflineMockPresetsAndLiveClamp(t *testing.T) {
	t.Setenv("LAYA_MODE", "mock")
	t.Setenv("GEMINI_MODE", "mock")

	svc := NewIncidentService()
	presets := svc.GetPresets()
	if len(presets) != 3 {
		t.Fatalf("expected 3 demo presets, got %d", len(presets))
	}

	tests := []struct {
		name              string
		presetIdx         int
		wantGate          string
		wantBandPct       int
		wantRejectedCount int
	}{
		{
			name:              "Preset1_Milan_AutoCandidate_25Pct_2Blocked",
			presetIdx:         0,
			wantGate:          "AUTO_CANDIDATE",
			wantBandPct:       25,
			wantRejectedCount: 2,
		},
		{
			name:              "Preset2_Prague_HumanReview_50Pct_PRM",
			presetIdx:         1,
			wantGate:          "HUMAN_REVIEW",
			wantBandPct:       50,
			wantRejectedCount: 0,
		},
		{
			name:              "Preset3_Paris_HumanReview_25Pct_Bdcm36Blocked",
			presetIdx:         2,
			wantGate:          "HUMAN_REVIEW",
			wantBandPct:       25,
			wantRejectedCount: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := svc.AnalyzeIncident(presets[tc.presetIdx].Request)
			if err != nil {
				t.Fatalf("AnalyzeIncident error: %v", err)
			}
			if resp.Laya.Source != "laya-mock" {
				t.Errorf("expected laya-mock, got %s", resp.Laya.Source)
			}
			if resp.GeminiSource != "gemini-mock" {
				t.Errorf("expected gemini-mock, got %s", resp.GeminiSource)
			}
			if resp.Laya.RoutingGate != tc.wantGate {
				t.Errorf("expected RoutingGate=%s, got %s (%s)", tc.wantGate, resp.Laya.RoutingGate, resp.Laya.RoutingReason)
			}
			if resp.EU2021782.CompensationBandPct != tc.wantBandPct {
				t.Errorf("expected EU2021782 band=%d%%, got %d%%", tc.wantBandPct, resp.EU2021782.CompensationBandPct)
			}
			if len(resp.RejectedSpares) != tc.wantRejectedCount {
				t.Errorf("expected %d rejected spares, got %d", tc.wantRejectedCount, len(resp.RejectedSpares))
			}
			if len(resp.Scenarios) != 3 {
				t.Fatalf("expected 3 scenarios, got %d", len(resp.Scenarios))
			}
			for _, sc := range resp.Scenarios {
				if strings.TrimSpace(sc.TcsGermanOrder) == "" || strings.TrimSpace(sc.PassengerAlert) == "" {
					t.Errorf("scenario %s missing per-card TcsGermanOrder or PassengerAlert", sc.ID)
				}
			}
		})
	}

	t.Run("SimulateUnfilteredAI_TriggersPostClamp", func(t *testing.T) {
		req := presets[0].Request // Milan
		req.SimulateUnfilteredAI = true
		resp, err := svc.AnalyzeIncident(req)
		if err != nil {
			t.Fatalf("AnalyzeIncident error: %v", err)
		}
		foundClamped := false
		for _, sc := range resp.Scenarios {
			if sc.WasClampedByGo {
				foundClamped = true
			}
		}
		if !foundClamped {
			t.Fatalf("expected at least one scenario clamped when SimulateUnfilteredAI=true on Milan")
		}
	})
}

func TestIncidentService_LiveGeminiSmokeIfConfigured(t *testing.T) {
	// Load workspace root .env if running from internal/service
	rootEnv := filepath.Join("..", "..", ".env")
	_ = godotenv.Load(rootEnv)

	if strings.TrimSpace(os.Getenv("GEMINI_API_KEY")) == "" {
		t.Skip("GEMINI_API_KEY not set in .env; skipping live Gemini smoke test")
	}

	t.Setenv("LAYA_MODE", "mock")
	t.Setenv("GEMINI_MODE", "live")

	svc := NewIncidentService()
	req := svc.GetPresets()[0].Request
	resp, err := svc.AnalyzeIncident(req)
	if err != nil {
		t.Fatalf("AnalyzeIncident live error: %v", err)
	}
	if resp.GeminiSource != "gemini-live" {
		t.Fatalf("expected gemini-live when GEMINI_API_KEY is configured, got %s (audit: %v)", resp.GeminiSource, resp.AuditLog)
	}
	if len(resp.Scenarios) != 3 {
		t.Fatalf("expected 3 live Gemini scenarios, got %d", len(resp.Scenarios))
	}
}
