import React from "react";
import {
  DemoPreset,
  IncidentRequest,
  RouteCode,
} from "../../bindings/es-incident-command/internal/domain/models";

interface IncidentFormProps {
  presets: DemoPreset[];
  activePresetId: string;
  request: IncidentRequest;
  isAnalyzing: boolean;
  onSelectPreset: (preset: DemoPreset) => void;
  onChangeRequest: (updated: IncidentRequest) => void;
  onToggleUnfilteredDemo: (enabled: boolean) => void;
  onSubmit: () => void;
}

const ROUTE_DEFAULT_TRAIN: Record<string, string> = {
  MILAN: "401",
  PRAGUE: "453",
  PARIS: "475",
};

export const IncidentForm: React.FC<IncidentFormProps> = ({
  presets,
  activePresetId,
  request,
  isAnalyzing,
  onSelectPreset,
  onChangeRequest,
  onToggleUnfilteredDemo,
  onSubmit,
}) => {
  const updateField = <K extends keyof IncidentRequest>(
    key: K,
    value: IncidentRequest[K]
  ) => {
    onChangeRequest({
      ...request,
      presetId: "CUSTOM",
      [key]: value,
    });
  };

  const handleRouteChange = (newRouteStr: string) => {
    const routeEnum = newRouteStr as RouteCode;
    onChangeRequest({
      ...request,
      presetId: "CUSTOM",
      route: routeEnum,
      trainId: ROUTE_DEFAULT_TRAIN[newRouteStr] || request.trainId,
    });
  };

  return (
    <aside className="flex flex-col gap-3 rounded-xl border border-slate-200 bg-white p-3.5 shadow-xs">
      {/* Preset Buttons */}
      <div>
        <div className="flex items-center justify-between mb-2">
          <h2 className="text-xs font-bold uppercase tracking-wider text-slate-700">
            One-Click Demo Presets
          </h2>
          <span className="text-[11px] text-slate-500">Auto-runs pipeline</span>
        </div>

        <div className="flex flex-col gap-1.5">
          {presets.map((preset) => {
            const isSelected = activePresetId === preset.id;
            return (
              <button
                key={preset.id}
                type="button"
                disabled={isAnalyzing}
                onClick={() => onSelectPreset(preset)}
                data-testid={`preset-btn-${preset.id}`}
                className={`group w-full rounded-lg border p-2.5 text-left transition-all cursor-pointer disabled:opacity-60 ${
                  isSelected
                    ? "border-indigo-600 bg-indigo-50/70 ring-1 ring-indigo-600"
                    : "border-slate-200 bg-slate-50/70 hover:border-slate-300 hover:bg-slate-100/70"
                }`}
              >
                <div className="flex items-center justify-between gap-1.5">
                  <span className="text-xs font-bold text-slate-900">
                    {preset.label}
                  </span>
                  <span
                    className={`shrink-0 rounded px-1.5 py-0.5 text-[9px] font-bold uppercase tracking-wide ${
                      isSelected
                        ? "bg-indigo-600 text-white"
                        : "bg-slate-200 text-slate-700"
                    }`}
                  >
                    {preset.badge}
                  </span>
                </div>
                <p className="mt-1 text-[11px] leading-snug text-slate-600">
                  {preset.subtitle}
                </p>
              </button>
            );
          })}
        </div>
      </div>

      <hr className="border-slate-200" />

      {/* Intake Controls */}
      <div className="flex flex-col gap-2.5">
        <h2 className="text-xs font-bold uppercase tracking-wider text-slate-700">
          Disruption Intake Parameters
        </h2>

        <div className="grid grid-cols-2 gap-2">
          <label className="flex flex-col gap-1">
            <span className="text-[11px] font-medium text-slate-600">Corridor</span>
            <select
              value={request.route}
              onChange={(e) => handleRouteChange(e.target.value)}
              className="rounded-md border border-slate-300 bg-white px-2 py-1.5 text-xs font-medium text-slate-900 focus:border-indigo-500 focus:outline-none"
            >
              <option value={RouteCode.RouteMilan}>MILAN (ES 401/400)</option>
              <option value={RouteCode.RoutePrague}>PRAGUE (ES 453/452)</option>
              <option value={RouteCode.RouteParis}>PARIS–BERLIN (475/474)</option>
            </select>
          </label>

          <label className="flex flex-col gap-1">
            <span className="text-[11px] font-medium text-slate-600">Train ID</span>
            <select
              value={request.trainId}
              onChange={(e) => updateField("trainId", e.target.value)}
              className="rounded-md border border-slate-300 bg-white px-2 py-1.5 text-xs font-mono font-semibold text-slate-900 focus:border-indigo-500 focus:outline-none"
            >
              <option value="401">ES 401 (Southbound)</option>
              <option value="400">ES 400 (Northbound)</option>
              <option value="453">ES 453 (Outbound)</option>
              <option value="452">ES 452 (Return)</option>
              <option value="475">ES 475 (Paris→Berlin)</option>
              <option value="474">ES 474 (Berlin→Paris)</option>
            </select>
          </label>
        </div>

        <div className="grid grid-cols-2 gap-2">
          <label className="flex flex-col gap-1">
            <span className="text-[11px] font-medium text-slate-600">Issue Type</span>
            <select
              value={request.issueType}
              onChange={(e) => updateField("issueType", e.target.value)}
              className="rounded-md border border-slate-300 bg-white px-2 py-1.5 text-xs font-medium text-slate-900 focus:border-indigo-500 focus:outline-none"
            >
              <option value="FLAT_WHEEL">FLAT_WHEEL (&gt;6mm)</option>
              <option value="CANCELED">CANCELED</option>
              <option value="NO_HEATING">NO_HEATING</option>
              <option value="NO_HV">NO_HV (Train-Line)</option>
              <option value="SHUNT_BLOCKED">SHUNT_BLOCKED</option>
            </select>
          </label>

          <label className="flex flex-col gap-1">
            <span className="text-[11px] font-medium text-slate-600">Cause Category</span>
            <select
              value={request.causeCategory}
              onChange={(e) => updateField("causeCategory", e.target.value)}
              className="rounded-md border border-slate-300 bg-white px-2 py-1.5 text-xs font-medium text-slate-900 focus:border-indigo-500 focus:outline-none"
            >
              <option value="OPERATOR_TECH">OPERATOR_TECH</option>
              <option value="INFRA_MANAGER">INFRA_MANAGER</option>
              <option value="THIRD_PARTY_EXTREME">THIRD_PARTY_EXTREME</option>
            </select>
          </label>
        </div>

        <div className="grid grid-cols-2 gap-2">
          <label className="flex flex-col gap-1">
            <span className="text-[11px] font-medium text-slate-600">
              Est. Delay (min)
            </span>
            <input
              type="number"
              min={0}
              max={1440}
              value={request.expectedDelayMin}
              onChange={(e) =>
                updateField("expectedDelayMin", Number(e.target.value) || 0)
              }
              className="rounded-md border border-slate-300 bg-white px-2 py-1.5 text-xs font-mono tabular-nums text-slate-900 focus:border-indigo-500 focus:outline-none"
            />
          </label>

          <label className="flex flex-col gap-1">
            <span className="text-[11px] font-medium text-slate-600">Affected Car</span>
            <input
              type="text"
              value={request.affectedCar}
              onChange={(e) => updateField("affectedCar", e.target.value)}
              className="rounded-md border border-slate-300 bg-white px-2 py-1.5 text-xs font-mono text-slate-900 focus:border-indigo-500 focus:outline-none"
            />
          </label>
        </div>

        <div className="grid grid-cols-2 gap-2">
          <label className="flex flex-col gap-1">
            <span className="text-[11px] font-medium text-slate-600">Affected Pax</span>
            <input
              type="number"
              min={0}
              max={900}
              value={request.affectedPax}
              onChange={(e) =>
                updateField("affectedPax", Number(e.target.value) || 0)
              }
              className="rounded-md border border-slate-300 bg-white px-2 py-1.5 text-xs font-mono tabular-nums text-slate-900 focus:border-indigo-500 focus:outline-none"
            />
          </label>

          <label className="flex flex-col gap-1">
            <span className="text-[11px] font-medium text-slate-600">
              Avg Ticket (€)
            </span>
            <input
              type="number"
              min={0}
              max={1000}
              value={request.avgTicketEur}
              onChange={(e) =>
                updateField("avgTicketEur", Number(e.target.value) || 0)
              }
              className="rounded-md border border-slate-300 bg-white px-2 py-1.5 text-xs font-mono tabular-nums text-slate-900 focus:border-indigo-500 focus:outline-none"
            />
          </label>
        </div>

        {/* Protected Passenger Toggles */}
        <div className="rounded-lg border border-slate-200 bg-slate-50 p-2 flex flex-col gap-1.5">
          <span className="text-[10px] font-bold uppercase tracking-wider text-slate-600">
            Protected Bookings & Accessibility
          </span>

          <label className="flex items-center justify-between text-xs text-slate-800 cursor-pointer">
            <span>PRM Pax (Foldable Wheelchair · Art. 20(5))</span>
            <input
              type="checkbox"
              checked={request.hasPrmPax}
              onChange={(e) => updateField("hasPrmPax", e.target.checked)}
              className="h-3.5 w-3.5 rounded border-slate-300 text-indigo-600"
            />
          </label>

          <label className="flex items-center justify-between text-xs text-slate-800 cursor-pointer">
            <span>Pet Private Compartment (No Shared)</span>
            <input
              type="checkbox"
              checked={request.hasPetPrivateComp}
              onChange={(e) => updateField("hasPetPrivateComp", e.target.checked)}
              className="h-3.5 w-3.5 rounded border-slate-300 text-indigo-600"
            />
          </label>

          <label className="flex items-center justify-between text-xs text-slate-800 cursor-pointer">
            <span>Women-Only Compartment Protected</span>
            <input
              type="checkbox"
              checked={request.hasWomenOnlyComp}
              onChange={(e) => updateField("hasWomenOnlyComp", e.target.checked)}
              className="h-3.5 w-3.5 rounded border-slate-300 text-indigo-600"
            />
          </label>
        </div>

        {/* Live Clamp Demo Toggle */}
        <label
          data-testid="toggle-unfiltered-ai"
          className={`flex items-start justify-between gap-2 rounded-lg border p-2 cursor-pointer transition-colors ${
            request.simulateUnfilteredAi
              ? "border-amber-400 bg-amber-50 text-amber-950"
              : "border-slate-200 bg-slate-50/70 text-slate-800 hover:bg-slate-100/70"
          }`}
        >
          <div className="flex flex-col">
            <span className="text-xs font-bold">
              Simulate Unfiltered AI (Demo Clamp)
            </span>
            <span className="text-[10px] leading-tight text-slate-600 mt-0.5">
              Exposes raw yard list to AI so Go&apos;s PostValidateScenarios visibly clamps illegal stock.
            </span>
          </div>
          <input
            type="checkbox"
            checked={request.simulateUnfilteredAi}
            disabled={isAnalyzing}
            onChange={(e) => onToggleUnfilteredDemo(e.target.checked)}
            className="mt-0.5 h-4 w-4 shrink-0 rounded border-amber-400 text-amber-600"
          />
        </label>

        <label className="flex flex-col gap-1">
          <span className="text-[11px] font-medium text-slate-600">
            Operational Notes
          </span>
          <textarea
            rows={2}
            value={request.extraInfo}
            onChange={(e) => updateField("extraInfo", e.target.value)}
            className="rounded-md border border-slate-300 bg-white px-2 py-1.5 text-xs text-slate-800 focus:border-indigo-500 focus:outline-none resize-none"
          />
        </label>

        <button
          type="button"
          onClick={onSubmit}
          disabled={isAnalyzing}
          data-testid="run-pipeline-btn"
          className="mt-0.5 flex w-full items-center justify-center gap-2 rounded-lg bg-indigo-600 px-4 py-2.5 text-xs font-bold uppercase tracking-wider text-white shadow-xs hover:bg-indigo-700 disabled:bg-indigo-400 cursor-pointer transition-colors"
        >
          {isAnalyzing ? (
            <>
              <span className="h-3.5 w-3.5 animate-spin rounded-full border-2 border-white border-t-transparent" />
              Running Guardrails + Laya + Gemini...
            </>
          ) : (
            "Run Laya + Gemini"
          )}
        </button>
      </div>
    </aside>
  );
};
