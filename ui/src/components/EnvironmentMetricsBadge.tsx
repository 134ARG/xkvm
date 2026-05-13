import { Popover, PopoverButton, PopoverPanel } from "@headlessui/react";
import {
  LuArrowDown,
  LuArrowUp,
  LuClock3,
  LuCpu,
  LuDroplets,
  LuThermometer,
  LuTriangleAlert,
} from "react-icons/lu";

import { cx } from "@/cva.config";
import { useEnvironmentMetrics } from "@hooks/useEnvironmentMetrics";
import { useVFDHostMetrics } from "@hooks/useVFDHostMetrics";

function formatMetric(value: number | null | undefined, suffix: string) {
  return value == null ? "-" : `${value.toFixed(1)}${suffix}`;
}

function formatPlain(value: number | null | undefined, suffix = "") {
  return value == null ? "-" : `${value}${suffix}`;
}

function formatBytes(value: number | null | undefined) {
  if (value == null) return "-";
  if (value < 1024) return `${value.toFixed(0)} B`;
  if (value < 1024 * 1024) return `${(value / 1024).toFixed(1)} KB`;
  if (value < 1024 * 1024 * 1024) return `${(value / 1024 / 1024).toFixed(1)} MB`;
  return `${(value / 1024 / 1024 / 1024).toFixed(1)} GB`;
}

function formatUptime(value: number | null | undefined) {
  if (value == null) return "-";
  const hours = Math.floor(value / 3600);
  const minutes = Math.floor((value % 3600) / 60);
  return `${hours}h ${minutes}m`;
}

function HostDetail({
  icon,
  label,
  value,
  alert,
}: {
  icon: React.ReactNode;
  label: string;
  value: string;
  alert?: boolean;
}) {
  return (
    <div className="rounded-sm border border-slate-800/10 p-2 dark:border-slate-300/20">
      <div className="flex items-center gap-x-1.5 text-xs text-slate-600 dark:text-slate-400">
        {icon}
        <span className="truncate">{label}</span>
      </div>
      <div
        className={cx(
          "mt-1 truncate font-mono text-sm font-semibold text-slate-950 dark:text-white",
          alert && "text-red-600 dark:text-red-400",
        )}
      >
        {value}
      </div>
    </div>
  );
}

function UsageBar({
  label,
  value,
  tone = "blue",
}: {
  label: string;
  value: number | null | undefined;
  tone?: "blue" | "green" | "violet";
}) {
  const percent = value == null ? 0 : Math.max(0, Math.min(100, value));
  const fillClass =
    tone === "green" ? "bg-green-500" : tone === "violet" ? "bg-violet-500" : "bg-blue-500";

  return (
    <div className="space-y-1.5">
      <div className="flex items-center justify-between gap-x-3 text-sm">
        <span className="flex items-center gap-x-1.5 text-slate-700 dark:text-slate-300">
          <LuCpu className="h-3.5 w-3.5 text-slate-500" />
          {label}
        </span>
        <span className="font-mono text-slate-950 dark:text-white">
          {value == null ? "-" : `${value.toFixed(1)}%`}
        </span>
      </div>
      <div className="h-1.5 overflow-hidden rounded-full bg-slate-200 dark:bg-slate-700">
        <div className={cx("h-full rounded-full", fillClass)} style={{ width: `${percent}%` }} />
      </div>
    </div>
  );
}

function MetricTile({
  icon,
  label,
  value,
}: {
  icon: React.ReactNode;
  label: string;
  value: string;
}) {
  return (
    <div className="rounded-sm border border-slate-800/10 p-2 dark:border-slate-300/20">
      <div className="flex items-center gap-x-1.5 text-xs text-slate-600 dark:text-slate-400">
        {icon}
        {label}
      </div>
      <div className="mt-1 font-mono text-sm font-semibold text-slate-950 dark:text-white">
        {value}
      </div>
    </div>
  );
}

export function EnvironmentMetricsBadge() {
  const { metrics, supported, stale } = useEnvironmentMetrics();
  const { metrics: hostMetrics } = useVFDHostMetrics();

  const caseTemp = formatMetric(metrics.caseTemperatureC, "°C");
  const humidity = formatMetric(metrics.caseHumidityPercent, "%");
  const socTemp = formatMetric(metrics.socTemperatureC, "°C");

  return (
    <Popover>
      <PopoverButton
        title="Metrics"
        className={cx(
          "hidden h-[24.5px] items-center gap-x-2 rounded-sm border border-slate-800/20 px-2 text-xs font-medium text-slate-700 transition hover:bg-blue-50/80 md:flex dark:border-slate-300/20 dark:text-slate-200 dark:hover:bg-slate-800",
          stale && "opacity-60",
        )}
      >
        <span className="flex items-center gap-x-1">
          <LuThermometer className="h-3.5 w-3.5 text-red-500" />
          {caseTemp}
        </span>
        <span className="flex items-center gap-x-1">
          <LuDroplets className="h-3.5 w-3.5 text-blue-500" />
          {humidity}
        </span>
        <span className="flex items-center gap-x-1">
          <LuCpu className="h-3.5 w-3.5 text-slate-500" />
          {socTemp}
        </span>
      </PopoverButton>
      <PopoverPanel
        anchor="bottom end"
        transition
        className={cx(
          "z-10 w-[390px] rounded-sm border border-slate-800/20 bg-white p-4 shadow-lg dark:border-slate-300/20 dark:bg-slate-900",
          "origin-top transition duration-200 ease-out data-closed:translate-y-2 data-closed:opacity-0",
        )}
      >
        <div className="space-y-4">
          <section className="space-y-2">
            <div className="flex items-center justify-between">
              <h3 className="text-sm font-semibold text-slate-950 dark:text-white">Device</h3>
              <span className="text-xs text-slate-500 dark:text-slate-400">
                {!supported ? "unsupported" : stale ? "stale" : "live"}
              </span>
            </div>
            <div className="grid grid-cols-3 gap-2">
              <MetricTile
                icon={<LuThermometer className="h-3.5 w-3.5 text-red-500" />}
                label="Case"
                value={caseTemp}
              />
              <MetricTile
                icon={<LuDroplets className="h-3.5 w-3.5 text-blue-500" />}
                label="Humidity"
                value={humidity}
              />
              <MetricTile
                icon={<LuCpu className="h-3.5 w-3.5 text-slate-500" />}
                label="SoC"
                value={socTemp}
              />
            </div>
          </section>

          <div className="h-px bg-slate-800/10 dark:bg-slate-300/20" />

          <section className="space-y-2">
            <div className="flex items-center justify-between">
              <h3 className="text-sm font-semibold text-slate-950 dark:text-white">Host</h3>
              <span
                className={cx(
                  "inline-flex items-center gap-x-1.5 text-xs",
                  hostMetrics.connected
                    ? "text-green-600 dark:text-green-400"
                    : "text-slate-500 dark:text-slate-400",
                )}
              >
                <span
                  className={cx(
                    "h-1.5 w-1.5 rounded-full",
                    hostMetrics.connected ? "bg-green-500" : "bg-slate-400",
                  )}
                />
                {hostMetrics.connected ? "connected" : "disconnected"}
              </span>
            </div>

            <div className="space-y-3">
              <UsageBar label="CPU" value={hostMetrics.cpuUtil} tone="blue" />
              <UsageBar label="RAM" value={hostMetrics.ramUtil} tone="green" />
              <UsageBar label="GPU" value={hostMetrics.gpuUtil} tone="violet" />
            </div>

            <div className="grid grid-cols-3 gap-2 pt-1">
              <HostDetail
                icon={<LuThermometer className="h-3.5 w-3.5 text-red-500" />}
                label="CPU"
                value={formatMetric(hostMetrics.cpuTemp, "°C")}
              />
              <HostDetail
                icon={<LuThermometer className="h-3.5 w-3.5 text-red-500" />}
                label="GPU"
                value={formatMetric(hostMetrics.gpuTemp, "°C")}
              />
              <HostDetail
                icon={<LuClock3 className="h-3.5 w-3.5 text-slate-500" />}
                label="Uptime"
                value={formatUptime(hostMetrics.uptimeSec)}
              />
              <HostDetail
                icon={<LuArrowDown className="h-3.5 w-3.5 text-blue-500" />}
                label="Received"
                value={formatBytes(hostMetrics.netRxBytes)}
              />
              <HostDetail
                icon={<LuArrowUp className="h-3.5 w-3.5 text-green-500" />}
                label="Sent"
                value={formatBytes(hostMetrics.netTxBytes)}
              />
              <HostDetail
                icon={
                  <LuTriangleAlert
                    className={cx(
                      "h-3.5 w-3.5",
                      (hostMetrics.failedUnits ?? 0) > 0 ? "text-red-500" : "text-slate-500",
                    )}
                  />
                }
                label="Failed units"
                value={formatPlain(hostMetrics.failedUnits)}
                alert={(hostMetrics.failedUnits ?? 0) > 0}
              />
            </div>
          </section>
        </div>
      </PopoverPanel>
    </Popover>
  );
}
