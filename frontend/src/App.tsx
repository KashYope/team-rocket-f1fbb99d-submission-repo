import { useCallback, useEffect, useState } from "react";
import {
  AnalyzeIncident,
  GetFleetState,
  GetPresets,
} from "../bindings/es-incident-command/internal/service/incidentservice";
import {
  DemoPreset,
  FleetState,
  IncidentRequest,
  IncidentResponse,
  RouteCode,
} from "../bindings/es-incident-command/internal/domain/models";
import { TrainConsistStrip } from "./components/TrainConsistStrip";
import { IncidentForm } from "./components/IncidentForm";
import { TriageGuardrailsPanel } from "./components/TriageGuardrailsPanel";
import { ResolutionPanel } from "./components/ResolutionPanel";

function pickDefaultScenarioId(resp: IncidentResponse): string {
  const scenarios = resp.scenarios || [];
  if (scenarios.length === 0) return "PAX_FIRST";
  const clamped = scenarios.find((s) => s.wasClampedByGo);
  if (clamped) return clamped.id;
  if (resp.laya?.recoveryStrategy?.choice === "downgrade_and_absorb") {
    const lc = scenarios.find((s) => s.id === "LOWEST_COST");
    if (lc) return lc.id;
  }
  const pf = scenarios.find((s) => s.id === "PAX_FIRST");
  return pf ? pf.id : scenarios[0].id;
}

export default function App() {
  const [presets, setPresets] = useState<DemoPreset[]>([]);
  const [fleet, setFleet] = useState<FleetState | null>(null);
  const [activePresetId, setActivePresetId] = useState<string>("");
  const [request, setRequest] = useState<IncidentRequest>(
    new IncidentRequest({
      trainId: "401",
      route: RouteCode.RouteMilan,
      issueType: "FLAT_WHEEL",
      causeCategory: "OPERATOR_TECH",
      expectedDelayMin: 95,
      affectedCar: "56 (RDC WL33)",
      affectedPax: 30,
      avgTicketEur: 165,
      hasPetPrivateComp: true,
    })
  );
  const [response, setResponse] = useState<IncidentResponse | null>(null);
  const [selectedScenarioId, setSelectedScenarioId] =
    useState<string>("PAX_FIRST");
  const [isAnalyzing, setIsAnalyzing] = useState<boolean>(false);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);

  const runAnalysis = useCallback(async (reqToRun: IncidentRequest) => {
    setIsAnalyzing(true);
    setErrorMsg(null);
    try {
      const res = await AnalyzeIncident(reqToRun);
      setResponse(res);
      setSelectedScenarioId(pickDefaultScenarioId(res));
    } catch (err: unknown) {
      const message =
        err instanceof Error ? err.message : "Failed to analyze disruption";
      setErrorMsg(message);
    } finally {
      setIsAnalyzing(false);
    }
  }, []);

  useEffect(() => {
    let mounted = true;
    (async () => {
      try {
        const [loadedPresets, loadedFleet] = await Promise.all([
          GetPresets(),
          GetFleetState(),
        ]);
        if (!mounted) return;
        setPresets(loadedPresets);
        setFleet(loadedFleet);
        if (loadedPresets.length > 0) {
          const initial = loadedPresets[0];
          setActivePresetId(initial.id);
          setRequest(initial.request);
          await runAnalysis(initial.request);
        }
      } catch (err: unknown) {
        if (!mounted) return;
        setErrorMsg(
          err instanceof Error ? err.message : "Failed to load initial fleet state"
        );
      }
    })();
    return () => {
      mounted = false;
    };
  }, [runAnalysis]);

  const handleSelectPreset = (preset: DemoPreset) => {
    const nextReq = new IncidentRequest({
      ...preset.request,
      simulateUnfilteredAi: false,
    });
    setActivePresetId(preset.id);
    setRequest(nextReq);
    void runAnalysis(nextReq);
  };

  const handleToggleUnfilteredDemo = (enabled: boolean) => {
    const nextReq = new IncidentRequest({
      ...request,
      simulateUnfilteredAi: enabled,
    });
    setRequest(nextReq);
    void runAnalysis(nextReq);
  };

  const layaIsLive = response?.laya?.source === "laya-live";
  const geminiIsLive = response?.geminiSource === "gemini-live";

  return (
    <div className="min-h-screen bg-slate-100 text-slate-900 flex flex-col">
      {/* Top Dark-Slate Operational Status Bar */}
      <header className="bg-slate-900 text-white border-b border-slate-800 px-4 py-2.5">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div className="flex items-center gap-3">
            <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-indigo-500 font-extrabold text-sm tracking-tight shadow-xs">
              ES
            </div>
            <div>
              <div className="flex items-center gap-2">
                <h1 className="text-sm font-bold tracking-tight">
                  ES Incident Command — Night Train Ops PoC
                </h1>
                <span className="rounded bg-slate-800 border border-slate-700 px-2 py-0.5 text-[10px] font-semibold text-slate-300">
                  {fleet?.timetableBaseline || "Nov 2026 Timetable Baseline"}
                </span>
              </div>
              <p className="text-[11px] text-slate-400 truncate max-w-2xl">
                {response?.trainSummary ||
                  "European Sleeper Disruption Decision Support · Deterministic Guardrails + Laya System 1 + Gemini System 2"}
              </p>
            </div>
          </div>

          {/* Runtime Source Badges */}
          <div className="flex flex-wrap items-center gap-2 text-[11px]">
            <span
              data-testid="badge-laya-source"
              className={`inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 font-bold uppercase tracking-wide border ${
                layaIsLive
                  ? "border-emerald-500/40 bg-emerald-500/20 text-emerald-300"
                  : "border-sky-500/40 bg-sky-500/15 text-sky-300"
              }`}
            >
              <span
                className={`h-2 w-2 rounded-full ${
                  layaIsLive ? "bg-emerald-400" : "bg-sky-400"
                }`}
              />
              {layaIsLive ? "LAYA LIVE" : "LAYA MOCK"}
              {response?.laya ? (
                <span className="font-mono font-normal text-slate-300 tabular-nums">
                  ({response.laya.latencyMs}ms)
                </span>
              ) : null}
            </span>

            <span
              data-testid="badge-gemini-source"
              className={`inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 font-bold uppercase tracking-wide border ${
                geminiIsLive
                  ? "border-emerald-500/40 bg-emerald-500/20 text-emerald-300"
                  : "border-amber-500/40 bg-amber-500/15 text-amber-300"
              }`}
            >
              <span
                className={`h-2 w-2 rounded-full ${
                  geminiIsLive ? "bg-emerald-400" : "bg-amber-400"
                }`}
              />
              {geminiIsLive ? "GEMINI LIVE" : "GEMINI MOCK"}
              {response?.geminiModel ? (
                <span className="font-mono font-normal text-slate-300 lowercase">
                  ({response.geminiModel})
                </span>
              ) : null}
            </span>
          </div>
        </div>
      </header>

      {/* Visual Train Consist Strip */}
      {response && response.activeConsist && (
        <TrainConsistStrip
          trainId={response.request.trainId}
          route={response.request.route}
          consist={response.activeConsist}
          maxCorridorCars={response.maxCorridorCars}
          nextDepartureHours={response.nextDepartureHours}
          issueType={response.request.issueType}
        />
      )}

      {errorMsg && (
        <div className="mx-4 mt-3 rounded-lg border border-rose-300 bg-rose-50 px-3 py-2 text-xs font-semibold text-rose-900">
          Backend Error: {errorMsg}
        </div>
      )}

      {/* 3-Column Command Grid (25% / 30% / 45%) */}
      <main className="flex-1 p-3.5 grid grid-cols-1 lg:grid-cols-[25fr_30fr_45fr] gap-3.5 items-start">
        <IncidentForm
          presets={presets}
          activePresetId={activePresetId}
          request={request}
          isAnalyzing={isAnalyzing}
          onSelectPreset={handleSelectPreset}
          onChangeRequest={(updated) => {
            setActivePresetId("CUSTOM");
            setRequest(updated);
          }}
          onToggleUnfilteredDemo={handleToggleUnfilteredDemo}
          onSubmit={() => void runAnalysis(request)}
        />

        <TriageGuardrailsPanel response={response} />

        <ResolutionPanel
          response={response}
          selectedScenarioId={selectedScenarioId}
          onSelectScenario={setSelectedScenarioId}
        />
      </main>
    </div>
  );
}
