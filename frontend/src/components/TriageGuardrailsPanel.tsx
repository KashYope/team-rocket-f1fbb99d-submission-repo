import React from "react";
import { IncidentResponse } from "../../bindings/es-incident-command/internal/domain/models";

interface TriageGuardrailsPanelProps {
  response: IncidentResponse | null;
}

const STRATEGY_LABELS: Record<string, string> = {
  swap_spare_car: "Swap Eligible Spare Car",
  downgrade_and_absorb: "Downgrade & Absorb Onboard",
  hotel_and_daytime_rail: "Hotel + Daytime Rail Reroute",
  move_to_next_train: "Move to Next ES Departure",
  manual_escalation: "Manual Senior Escalation",
};

const STRATEGY_ORDER = [
  "swap_spare_car",
  "downgrade_and_absorb",
  "hotel_and_daytime_rail",
  "move_to_next_train",
  "manual_escalation",
];

const SEVERITY_STEPS = [
  { idx: 0, label: "0 · Low" },
  { idx: 1, label: "1 · Medium" },
  { idx: 2, label: "2 · High" },
  { idx: 3, label: "3 · Critical" },
];

export const TriageGuardrailsPanel: React.FC<TriageGuardrailsPanelProps> = ({
  response,
}) => {
  if (!response) {
    return (
      <section className="rounded-xl border border-slate-200 bg-white p-4 text-xs text-slate-500">
        Select a preset or run the pipeline to view Laya System 1 triage and deterministic European Sleeper guardrails.
      </section>
    );
  }

  const { laya, eu2021782 } = response;
  const checks = response.ruleChecks ?? [];
  const eligible = response.eligibleSpares ?? [];
  const rejected = response.rejectedSpares ?? [];
  const isAuto = laya.routingGate === "AUTO_CANDIDATE";
  const probs = laya.recoveryStrategy.probabilities || {};
  const nearestSevIdx = Math.min(3, Math.max(0, Math.round(laya.operationalSeverity.score)));

  return (
    <section className="flex flex-col gap-3">
      {/* 1. Laya System 1 Triage Card */}
      <div className="rounded-xl border border-slate-200 bg-white p-3.5 shadow-xs">
        <div className="flex items-start justify-between gap-2">
          <div>
            <h2 className="text-xs font-bold uppercase tracking-wider text-slate-700">
              Laya System 1 Triage (/v1/systemone)
            </h2>
            <p className="text-[11px] text-slate-500">
              Model: <span className="font-mono font-medium">{laya.model}</span> · Gated on{" "}
              <span className="font-mono">max(p)</span> &amp; Shannon entropy
            </p>
          </div>

          <span
            data-testid="laya-routing-gate"
            className={`shrink-0 rounded-md px-2.5 py-1 text-xs font-extrabold uppercase tracking-wide border ${
              isAuto
                ? "border-emerald-300 bg-emerald-100 text-emerald-900"
                : "border-amber-400 bg-amber-100 text-amber-950"
            }`}
          >
            {laya.routingGate}
          </span>
        </div>

        <div
          className={`mt-2 rounded-md px-2.5 py-1.5 text-[11px] leading-snug ${
            isAuto
              ? "bg-emerald-50 text-emerald-900 border border-emerald-200"
              : "bg-amber-50 text-amber-950 border border-amber-200"
          }`}
        >
          {laya.routingReason}
        </div>

        {/* Recovery Strategy Probability Bars */}
        <div className="mt-3">
          <div className="flex items-center justify-between text-[11px] mb-1.5">
            <span className="font-semibold text-slate-700">
              Recovery Strategy Distribution
            </span>
            <span className="font-mono text-slate-600 tabular-nums">
              ansConf={(laya.recoveryStrategy.answerConfidence * 100).toFixed(0)}% · entropyConf=
              {(laya.recoveryStrategy.confidence * 100).toFixed(0)}%
            </span>
          </div>

          <div className="flex flex-col gap-1.5">
            {STRATEGY_ORDER.map((key) => {
              const p = probs[key] ?? 0;
              const pct = Math.round(p * 100);
              const isWinner = laya.recoveryStrategy.choice === key;
              return (
                <div key={key} className="flex flex-col gap-0.5">
                  <div className="flex items-center justify-between text-[11px]">
                    <span
                      className={
                        isWinner ? "font-bold text-slate-900" : "text-slate-600"
                      }
                    >
                      {STRATEGY_LABELS[key] || key}
                    </span>
                    <span className="font-mono font-semibold tabular-nums text-slate-800">
                      {pct}%
                    </span>
                  </div>
                  <div className="h-2 w-full overflow-hidden rounded-full bg-slate-100">
                    <div
                      style={{ width: `${Math.max(pct, 2)}%` }}
                      className={`h-full rounded-full transition-all ${
                        isWinner ? "bg-indigo-600" : "bg-slate-400"
                      }`}
                    />
                  </div>
                </div>
              );
            })}
          </div>
        </div>

        {/* Ordinal Severity Gauge + Noul Risks */}
        <div className="mt-3 pt-2.5 border-t border-slate-100">
          <div className="flex items-center justify-between text-[11px] mb-1">
            <span className="font-semibold text-slate-700">
              Operational Severity (0.0–3.0)
            </span>
            <span className="font-mono font-bold text-slate-900 tabular-nums">
              {laya.operationalSeverity.score.toFixed(2)} ({laya.operationalSeverity.nearestLabel})
            </span>
          </div>

          <div className="grid grid-cols-4 gap-1">
            {SEVERITY_STEPS.map((step) => {
              const active = step.idx === nearestSevIdx;
              const colorCls =
                step.idx === 3
                  ? active
                    ? "bg-rose-600 text-white border-rose-700"
                    : "bg-rose-50 text-rose-800 border-rose-200"
                  : step.idx === 2
                  ? active
                    ? "bg-amber-500 text-white border-amber-600"
                    : "bg-amber-50 text-amber-800 border-amber-200"
                  : step.idx === 1
                  ? active
                    ? "bg-sky-600 text-white border-sky-700"
                    : "bg-sky-50 text-sky-800 border-sky-200"
                  : active
                  ? "bg-emerald-600 text-white border-emerald-700"
                  : "bg-emerald-50 text-emerald-800 border-emerald-200";
              return (
                <div
                  key={step.idx}
                  className={`rounded border py-1 text-center text-[10px] font-bold ${colorCls}`}
                >
                  {step.label}
                </div>
              );
            })}
          </div>

          <div className="mt-2 grid grid-cols-2 gap-1.5">
            <div className="flex items-center justify-between rounded-md border border-slate-200 bg-slate-50 px-2 py-1 text-[11px]">
              <span className="text-slate-600">Refund Cliff (&ge;60m)</span>
              <span className="font-mono font-bold tabular-nums text-slate-900">
                {Math.round(laya.refundCliffRisk * 100)}%
              </span>
            </div>
            <div className="flex items-center justify-between rounded-md border border-slate-200 bg-slate-50 px-2 py-1 text-[11px]">
              <span className="text-slate-600">Turnaround Breach</span>
              <span className="font-mono font-bold tabular-nums text-slate-900">
                {Math.round(laya.turnaroundBreach * 100)}%
              </span>
            </div>
          </div>
        </div>
      </div>

      {/* 2. EU 2021/782 Statutory Liability Card */}
      <div
        data-testid="eu-2021-782-card"
        className="rounded-xl border border-slate-200 bg-white p-3.5 shadow-xs"
      >
        <div className="flex items-center justify-between">
          <h2 className="text-xs font-bold uppercase tracking-wider text-slate-700">
            EU 2021/782 Statutory Liability
          </h2>
          <span
            className={`rounded px-2 py-0.5 text-[10px] font-bold uppercase ${
              eu2021782.forceMajeureExempt
                ? "bg-slate-200 text-slate-800"
                : "bg-rose-100 text-rose-900 border border-rose-300"
            }`}
          >
            {eu2021782.forceMajeureExempt
              ? "Force Majeure Exempt (Art. 19(10))"
              : "Non-Exempt Liability"}
          </span>
        </div>

        <div className="mt-2 grid grid-cols-2 gap-2">
          <div className="rounded-lg border border-slate-200 bg-slate-50 p-2">
            <div className="text-[10px] uppercase font-semibold text-slate-500">
              Art. 19 Comp. Band
            </div>
            <div
              data-testid="eu-comp-band"
              className="text-lg font-extrabold text-slate-900 tabular-nums"
            >
              {eu2021782.compensationBandPct}%
            </div>
          </div>
          <div className="rounded-lg border border-slate-200 bg-slate-50 p-2">
            <div className="text-[10px] uppercase font-semibold text-slate-500">
              Est. Cash Exposure
            </div>
            <div className="text-lg font-extrabold text-slate-900 tabular-nums">
              €{eu2021782.estimatedRefundEur.toLocaleString("en-US")}
            </div>
          </div>
        </div>

        <div className="mt-2 flex flex-wrap gap-1.5">
          <span
            className={`rounded-full px-2 py-0.5 text-[10px] font-bold border ${
              eu2021782.selfReroute100Min
                ? "border-amber-300 bg-amber-100 text-amber-950"
                : "border-slate-200 bg-slate-100 text-slate-500"
            }`}
          >
            100m Self-Reroute: {eu2021782.selfReroute100Min ? "ACTIVE" : "OFF"}
          </span>
          <span
            className={`rounded-full px-2 py-0.5 text-[10px] font-bold border ${
              eu2021782.hotelRequired
                ? "border-rose-300 bg-rose-100 text-rose-900"
                : "border-slate-200 bg-slate-100 text-slate-500"
            }`}
          >
            Art. 20 Hotel Req: {eu2021782.hotelRequired ? "YES" : "NO"}
          </span>
          <span
            className={`rounded-full px-2 py-0.5 text-[10px] font-bold border ${
              eu2021782.prmPriorityRequired
                ? "border-purple-300 bg-purple-100 text-purple-900"
                : "border-slate-200 bg-slate-100 text-slate-500"
            }`}
          >
            PRM Art. 20(5): {eu2021782.prmPriorityRequired ? "PRIORITY" : "N/A"}
          </span>
        </div>

        <p className="mt-2 text-[11px] leading-snug text-slate-600">
          {eu2021782.legalBasisNotes}
        </p>
      </div>

      {/* 3. Deterministic Rail Guardrails & Spare Fleet */}
      <div className="rounded-xl border border-slate-200 bg-white p-3.5 shadow-xs">
        <h2 className="text-xs font-bold uppercase tracking-wider text-slate-700 mb-2">
          Deterministic Rail Guardrails (Go Pre-Filter)
        </h2>

        <div className="flex flex-col gap-1.5">
          {checks.map((chk) => {
            const badgeStyle =
              chk.severity === "BLOCK"
                ? "bg-rose-600 text-white"
                : chk.severity === "WARN"
                ? "bg-amber-500 text-white"
                : "bg-emerald-600 text-white";
            const rowStyle =
              chk.severity === "BLOCK"
                ? "border-rose-200 bg-rose-50/70"
                : chk.severity === "WARN"
                ? "border-amber-200 bg-amber-50/60"
                : "border-emerald-200 bg-emerald-50/50";
            return (
              <div
                key={chk.id}
                className={`rounded-lg border p-2 ${rowStyle}`}
              >
                <div className="flex items-center gap-1.5">
                  <span
                    className={`shrink-0 rounded px-1.5 py-0.5 text-[9px] font-extrabold uppercase ${badgeStyle}`}
                  >
                    {chk.severity}
                  </span>
                  <span className="text-xs font-bold text-slate-900">
                    {chk.summary}
                  </span>
                </div>
                <p className="mt-0.5 text-[11px] leading-snug text-slate-700">
                  {chk.detail}
                </p>
              </div>
            );
          })}
        </div>

        {/* Eligible & Rejected Spare Pool */}
        <div className="mt-3 pt-2.5 border-t border-slate-200">
          <div className="text-[11px] font-bold uppercase tracking-wide text-slate-700 mb-1.5">
            Parked Spare Fleet ({eligible.length} Eligible · {rejected.length} Blocked)
          </div>

          <div className="flex flex-col gap-1">
            {eligible.map((car) => (
              <div
                key={car.uic}
                className="flex items-center justify-between rounded border border-emerald-200 bg-emerald-50/60 px-2 py-1 text-[11px]"
              >
                <div>
                  <span className="font-bold text-emerald-950">
                    {car.owner} {car.carType}
                  </span>{" "}
                  <span className="font-mono text-slate-600">({car.uic})</span>
                </div>
                <span className="text-[10px] font-semibold text-emerald-800">
                  {car.productClass} · {car.capacity}p · {car.yard}
                </span>
              </div>
            ))}

            {rejected.map((car) => (
              <div
                key={car.uic}
                data-testid={`rejected-spare-${car.owner}-${car.carType.replace(/\s+/g, "-")}`}
                className="rounded border border-rose-300 bg-rose-50 px-2 py-1 text-[11px]"
              >
                <div className="flex items-center justify-between">
                  <div>
                    <span className="font-bold text-rose-950">
                      {car.owner} {car.carType}
                    </span>{" "}
                    <span className="font-mono text-rose-800">({car.uic})</span>
                  </div>
                  <span className="rounded bg-rose-600 px-1.5 py-0.5 text-[9px] font-bold uppercase text-white">
                    BLOCKED
                  </span>
                </div>
                <div className="mt-0.5 text-[10px] text-rose-900">
                  {car.blockedReason}
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    </section>
  );
};
