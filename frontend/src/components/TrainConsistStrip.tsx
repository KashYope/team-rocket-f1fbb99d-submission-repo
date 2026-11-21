import React from "react";
import { ConsistCar, RouteCode } from "../../bindings/es-incident-command/internal/domain/models";

interface TrainConsistStripProps {
  trainId: string;
  route: RouteCode;
  consist: ConsistCar[];
  maxCorridorCars: number;
  nextDepartureHours: number;
  issueType: string;
}

export const TrainConsistStrip: React.FC<TrainConsistStripProps> = ({
  trainId,
  route,
  consist,
  maxCorridorCars,
  nextDepartureHours,
  issueType,
}) => {
  const tractionLabel =
    route === RouteCode.RouteMilan
      ? "TCS / Arenaways"
      : route === RouteCode.RoutePrague
      ? "Swap Bad Bentheim"
      : "Multi-Sys FR/BE/DE";

  const commercialCount = consist.filter((c) => c.isCommercial).length;
  const isCanceled = issueType === "CANCELED";

  return (
    <section className="bg-white border-b border-slate-200 px-4 py-2.5 shadow-2xs">
      <div className="flex flex-wrap items-center justify-between gap-2 mb-1.5">
        <div className="flex items-center gap-2 text-xs">
          <span className="font-semibold uppercase tracking-wider text-slate-700">
            Active Consist · ES {trainId} ({route})
          </span>
          <span className="rounded bg-slate-100 px-2 py-0.5 font-medium text-slate-700 tabular-nums">
            {consist.length} / {maxCorridorCars} Corridor Max Cars
          </span>
          <span
            className={`rounded px-2 py-0.5 font-medium tabular-nums ${
              commercialCount > 8
                ? "bg-amber-100 text-amber-900 border border-amber-300"
                : "bg-emerald-50 text-emerald-800 border border-emerald-200"
            }`}
          >
            {commercialCount} Commercial Cars ({commercialCount > 8 ? "2 Conductors Active" : "1 Conductor Limit = 8"})
          </span>
        </div>

        <div className="flex items-center gap-2 text-[11px] text-slate-600">
          <span className="rounded border border-slate-200 bg-slate-50 px-2 py-0.5 font-medium">
            NO_HV Rule: Rear-Only · No Loco Reversal
          </span>
          <span className="rounded border border-indigo-200 bg-indigo-50 px-2 py-0.5 font-semibold text-indigo-900 tabular-nums">
            Next ES Departure: +{nextDepartureHours}h (3x/wk)
          </span>
        </div>
      </div>

      <div className="flex items-stretch gap-1.5 overflow-x-auto pb-0.5">
        {/* Locomotive Head */}
        <div className="flex shrink-0 flex-col justify-between rounded-l-lg rounded-r border border-slate-700 bg-slate-800 px-2.5 py-1.5 text-white min-w-[108px]">
          <div className="flex items-center justify-between gap-1">
            <span className="text-[10px] font-bold uppercase tracking-wider text-amber-300">
              LOCO
            </span>
            <span className="text-[10px] text-slate-300 font-mono">ES {trainId}</span>
          </div>
          <div className="text-[11px] font-medium leading-tight text-slate-200 truncate">
            {tractionLabel}
          </div>
        </div>

        {consist.map((car, index) => {
          const showConductorDivider = index === 8;
          return (
            <React.Fragment key={`${car.position}-${car.carNumber}`}>
              {showConductorDivider && (
                <div
                  className="flex shrink-0 flex-col items-center justify-center px-1"
                  title="9th occupied commercial car triggers mandatory 2nd conductor (+€3,000 Milan / +€1,600 Prague & Paris)"
                >
                  <div className="h-full border-l-2 border-dashed border-amber-500" />
                  <span className="mt-0.5 rounded bg-amber-100 px-1 text-[9px] font-bold uppercase text-amber-900 whitespace-nowrap">
                    &gt;8 Cars: +2nd Cond.
                  </span>
                </div>
              )}

              <div
                title={`Car ${car.carNumber} · ${car.owner} ${car.carType} (${car.uic}) · ${car.productClass} · ${car.bookedPax}/${car.capacity} pax${
                  car.specialNotes ? ` · ${car.specialNotes}` : ""
                }`}
                className={`flex shrink-0 flex-col justify-between rounded border px-2 py-1 min-w-[112px] transition-colors ${
                  car.isFaulty
                    ? "border-rose-500 bg-rose-50 text-rose-950 ring-2 ring-rose-500/25"
                    : car.exceedsSingleConductor
                    ? "border-amber-300 bg-amber-50/60 text-slate-900"
                    : "border-slate-200 bg-slate-50 text-slate-800"
                }`}
              >
                <div className="flex items-center justify-between gap-1">
                  <span className="text-[11px] font-bold tabular-nums">
                    #{car.carNumber}{" "}
                    <span className="font-semibold text-slate-600">
                      {car.owner} {car.carType}
                    </span>
                  </span>
                  {car.isFaulty && (
                    <span className="rounded bg-rose-600 px-1 py-0.2 text-[9px] font-bold uppercase text-white">
                      {isCanceled ? "CANC" : "FAULT"}
                    </span>
                  )}
                </div>

                <div className="flex items-center justify-between gap-1 mt-0.5 text-[10px]">
                  <span className="truncate text-slate-600 max-w-[72px]">
                    {car.productClass}
                  </span>
                  <span className="font-mono font-medium tabular-nums text-slate-700">
                    {car.bookedPax}/{car.capacity}p
                  </span>
                </div>
              </div>
            </React.Fragment>
          );
        })}
      </div>
    </section>
  );
};
