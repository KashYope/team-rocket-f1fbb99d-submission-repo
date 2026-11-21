import React, { useState } from "react";
import {
  IncidentResponse,
  ScenarioCard,
} from "../../bindings/es-incident-command/internal/domain/models";

interface ResolutionPanelProps {
  response: IncidentResponse | null;
  selectedScenarioId: string;
  onSelectScenario: (id: string) => void;
}

type DrawerTab = "TCS_ORDER" | "PAX_ALERT" | "AUDIT_LOG";

const SCENARIO_BADGE: Record<string, string> = {
  LOWEST_COST: "Lowest Cash Cost",
  PAX_FIRST: "Pax-First Comfort",
  PROTECT_NEXT_TRAIN: "Protect Return Working",
};

export const ResolutionPanel: React.FC<ResolutionPanelProps> = ({
  response,
  selectedScenarioId,
  onSelectScenario,
}) => {
  const [activeTab, setActiveTab] = useState<DrawerTab>("TCS_ORDER");
  const [copiedTab, setCopiedTab] = useState<string | null>(null);

  if (!response) {
    return (
      <section className="rounded-xl border border-slate-200 bg-white p-4 text-xs text-slate-500">
        Awaiting scenario synthesis...
      </section>
    );
  }

  const scenarios = response.scenarios || [];
  const activeScenario: ScenarioCard | undefined =
    scenarios.find((s) => s.id === selectedScenarioId) || scenarios[0];

  const activeTcsOrder =
    activeScenario?.tcsGermanOrder || response.tcsGermanOrder;
  const activePaxAlert =
    activeScenario?.passengerAlert || response.passengerAlert;

  const copyText = async (text: string, tabKey: string) => {
    try {
      if (navigator.clipboard && navigator.clipboard.writeText) {
        await navigator.clipboard.writeText(text);
      } else {
        throw new Error("Clipboard API unavailable");
      }
    } catch {
      const ta = document.createElement("textarea");
      ta.value = text;
      ta.style.position = "fixed";
      ta.style.opacity = "0";
      document.body.appendChild(ta);
      ta.select();
      document.execCommand("copy");
      document.body.removeChild(ta);
    }
    setCopiedTab(tabKey);
    window.setTimeout(() => {
      setCopiedTab((prev) => (prev === tabKey ? null : prev));
    }, 1800);
  };

  return (
    <section className="flex flex-col gap-3">
      {/* 3 Interactive Scenario Cards */}
      <div className="rounded-xl border border-slate-200 bg-white p-3.5 shadow-xs">
        <div className="flex items-center justify-between mb-2">
          <div>
            <h2 className="text-xs font-bold uppercase tracking-wider text-slate-700">
              System 2 Recovery Scenarios ({scenarios.length})
            </h2>
            <p className="text-[11px] text-slate-500">
              Click any scenario card to switch the German TCS Dispatch Order &amp; EU 2021/782 Passenger Alert below
            </p>
          </div>
        </div>

        <div className="grid grid-cols-1 xl:grid-cols-3 gap-2.5">
          {scenarios.map((card) => {
            const isSelected = activeScenario?.id === card.id;
            const riskStyle =
              card.nextTrainRisk === "HIGH"
                ? "bg-rose-100 text-rose-900 border-rose-300"
                : card.nextTrainRisk === "MEDIUM"
                ? "bg-amber-100 text-amber-900 border-amber-300"
                : "bg-emerald-100 text-emerald-900 border-emerald-300";

            return (
              <div
                key={card.id}
                role="button"
                tabIndex={0}
                data-testid={`scenario-card-${card.id}`}
                onClick={() => onSelectScenario(card.id)}
                onKeyDown={(e) => {
                  if (e.key === "Enter" || e.key === " ") {
                    e.preventDefault();
                    onSelectScenario(card.id);
                  }
                }}
                className={`flex flex-col justify-between rounded-xl border p-3 text-left transition-all cursor-pointer ${
                  isSelected
                    ? "border-indigo-600 bg-indigo-50/40 ring-2 ring-indigo-600/30 shadow-xs"
                    : "border-slate-200 bg-slate-50/40 hover:border-slate-300 hover:bg-white"
                }`}
              >
                <div>
                  <div className="flex flex-wrap items-center justify-between gap-1">
                    <span
                      className={`rounded px-1.5 py-0.5 text-[10px] font-bold uppercase tracking-wide ${
                        isSelected
                          ? "bg-indigo-600 text-white"
                          : "bg-slate-200 text-slate-700"
                      }`}
                    >
                      {SCENARIO_BADGE[card.id] || card.id}
                    </span>

                    <span
                      className={`rounded border px-1.5 py-0.5 text-[10px] font-bold uppercase ${riskStyle}`}
                    >
                      Return Risk: {card.nextTrainRisk}
                    </span>
                  </div>

                  {card.wasClampedByGo && (
                    <div
                      data-testid="clamped-by-guardrail-badge"
                      className="mt-2 rounded-md border border-amber-400 bg-amber-100 px-2 py-1 text-[10px] font-extrabold uppercase tracking-wide text-amber-950"
                    >
                      ⚠ Clamped by Rail Guardrail
                    </div>
                  )}

                  <h3 className="mt-2 text-xs font-bold leading-snug text-slate-900">
                    {card.title}
                  </h3>

                  {/* Key Metrics */}
                  <div className="mt-2 grid grid-cols-3 gap-1 rounded-lg border border-slate-200 bg-white p-1.5 text-center">
                    <div>
                      <div className="text-[9px] uppercase text-slate-500">
                        Est. Cost
                      </div>
                      <div className="text-xs font-extrabold text-slate-900 tabular-nums">
                        €{card.estimatedCostEur.toLocaleString("en-US")}
                      </div>
                    </div>
                    <div>
                      <div className="text-[9px] uppercase text-slate-500">
                        Downgrades
                      </div>
                      <div className="text-xs font-extrabold text-slate-900 tabular-nums">
                        {card.downgrades}
                      </div>
                    </div>
                    <div>
                      <div className="text-[9px] uppercase text-slate-500">
                        Canceled
                      </div>
                      <div className="text-xs font-extrabold text-slate-900 tabular-nums">
                        {card.cancellations}
                      </div>
                    </div>
                  </div>

                  {/* Recommended Car */}
                  <div className="mt-2 rounded border border-slate-200 bg-slate-100/80 px-2 py-1">
                    <div className="text-[9px] font-bold uppercase text-slate-500">
                      Recommended Rolling Stock
                    </div>
                    <div className="text-[11px] font-mono font-semibold text-slate-900 break-words">
                      {card.recommendedCar}
                    </div>
                  </div>

                  {/* Action Steps */}
                  <ul className="mt-2 space-y-1 text-[11px] leading-snug text-slate-700 list-disc pl-4">
                    {(card.actionSteps || []).map((step, i) => (
                      <li key={i}>{step}</li>
                    ))}
                  </ul>
                </div>

                {(card.guardrailNotes || []).length > 0 && (
                  <div className="mt-2.5 pt-2 border-t border-slate-200 space-y-1">
                    {card.guardrailNotes.map((note, idx) => (
                      <div
                        key={idx}
                        className={`rounded px-1.5 py-1 text-[10px] leading-tight ${
                          note.includes("CLAMPED")
                            ? "bg-amber-100 font-semibold text-amber-950 border border-amber-300"
                            : "bg-slate-100 text-slate-600"
                        }`}
                      >
                        {note}
                      </div>
                    ))}
                  </div>
                )}
              </div>
            );
          })}
        </div>
      </div>

      {/* Tabbed Dispatch & Comms Drawer */}
      <div className="rounded-xl border border-slate-200 bg-white p-3.5 shadow-xs">
        <div className="flex flex-wrap items-center justify-between gap-2 border-b border-slate-200 pb-2.5">
          <div className="flex items-center gap-1.5">
            <button
              type="button"
              data-testid="tab-tcs-order"
              onClick={() => setActiveTab("TCS_ORDER")}
              className={`rounded-lg px-3 py-1.5 text-xs font-bold transition-colors cursor-pointer ${
                activeTab === "TCS_ORDER"
                  ? "bg-slate-900 text-white"
                  : "bg-slate-100 text-slate-700 hover:bg-slate-200"
              }`}
            >
              TCS Order (Deutsch)
            </button>

            <button
              type="button"
              data-testid="tab-pax-alert"
              onClick={() => setActiveTab("PAX_ALERT")}
              className={`rounded-lg px-3 py-1.5 text-xs font-bold transition-colors cursor-pointer ${
                activeTab === "PAX_ALERT"
                  ? "bg-slate-900 text-white"
                  : "bg-slate-100 text-slate-700 hover:bg-slate-200"
              }`}
            >
              Passenger Alert (EU 2021/782)
            </button>

            <button
              type="button"
              data-testid="tab-audit-trail"
              onClick={() => setActiveTab("AUDIT_LOG")}
              className={`rounded-lg px-3 py-1.5 text-xs font-bold transition-colors cursor-pointer ${
                activeTab === "AUDIT_LOG"
                  ? "bg-slate-900 text-white"
                  : "bg-slate-100 text-slate-700 hover:bg-slate-200"
              }`}
            >
              Audit Trail ({response.auditLog.length})
            </button>
          </div>

          {activeTab !== "AUDIT_LOG" && (
            <div className="flex items-center gap-2">
              {activeScenario && (
                <span className="text-[11px] font-medium text-slate-500">
                  Active:{" "}
                  <strong className="text-slate-800">
                    {SCENARIO_BADGE[activeScenario.id] || activeScenario.id}
                  </strong>
                </span>
              )}
              <button
                type="button"
                data-testid="copy-comms-btn"
                onClick={() =>
                  copyText(
                    activeTab === "TCS_ORDER" ? activeTcsOrder : activePaxAlert,
                    activeTab
                  )
                }
                className="rounded-lg border border-slate-300 bg-slate-50 px-3 py-1.5 text-xs font-bold text-slate-800 hover:bg-slate-100 cursor-pointer transition-colors"
              >
                {copiedTab === activeTab ? "Copied!" : "Copy Text"}
              </button>
            </div>
          )}
        </div>

        <div className="mt-3">
          {activeTab === "TCS_ORDER" && (
            <pre
              data-testid="tcs-order-content"
              className="whitespace-pre-wrap rounded-lg border border-slate-800 bg-slate-950 p-3.5 font-mono text-xs leading-relaxed text-emerald-300 overflow-x-auto"
            >
              {activeTcsOrder}
            </pre>
          )}

          {activeTab === "PAX_ALERT" && (
            <div
              data-testid="pax-alert-content"
              className="rounded-lg border border-indigo-200 bg-indigo-50/50 p-3.5 text-xs leading-relaxed text-slate-900 whitespace-pre-wrap"
            >
              {activePaxAlert}
            </div>
          )}

          {activeTab === "AUDIT_LOG" && (
            <ol
              data-testid="audit-log-content"
              className="space-y-1.5 rounded-lg border border-slate-200 bg-slate-50 p-3 font-mono text-[11px] text-slate-800"
            >
              {response.auditLog.map((entry, idx) => (
                <li
                  key={idx}
                  className={`rounded px-2 py-1 ${
                    entry.includes("POST-CLAMP")
                      ? "bg-amber-100 font-bold text-amber-950 border border-amber-300"
                      : "bg-white border border-slate-200/80"
                  }`}
                >
                  {entry}
                </li>
              ))}
            </ol>
          )}
        </div>
      </div>
    </section>
  );
};
