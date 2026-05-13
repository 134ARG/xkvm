import { LuCpu, LuDroplets, LuThermometer } from "react-icons/lu";

import { cx } from "@/cva.config";
import { useEnvironmentMetrics } from "@hooks/useEnvironmentMetrics";

function formatMetric(value: number | null | undefined, suffix: string) {
  return value == null ? null : `${value.toFixed(1)}${suffix}`;
}

export function EnvironmentMetricsBadge() {
  const { metrics, available, stale } = useEnvironmentMetrics();

  if (!available || !metrics) return null;

  const caseTemp = formatMetric(metrics.caseTemperatureC, "°C");
  const humidity = formatMetric(metrics.caseHumidityPercent, "%");
  const socTemp = formatMetric(metrics.socTemperatureC, "°C");

  return (
    <div
      title="Case environment"
      className={cx(
        "hidden h-[24.5px] items-center gap-x-2 rounded-sm border border-slate-800/20 px-2 text-xs font-medium text-slate-700 md:flex dark:border-slate-300/20 dark:text-slate-200",
        stale && "opacity-60",
      )}
    >
      {caseTemp && (
        <span className="flex items-center gap-x-1">
          <LuThermometer className="h-3.5 w-3.5 text-red-500" />
          {caseTemp}
        </span>
      )}
      {humidity && (
        <span className="flex items-center gap-x-1">
          <LuDroplets className="h-3.5 w-3.5 text-blue-500" />
          {humidity}
        </span>
      )}
      {socTemp && (
        <span className="hidden items-center gap-x-1 xl:flex">
          <LuCpu className="h-3.5 w-3.5 text-slate-500" />
          {socTemp}
        </span>
      )}
    </div>
  );
}
