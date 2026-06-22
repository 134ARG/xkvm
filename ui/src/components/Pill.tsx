import React from "react";

import { cx } from "@/cva.config";

export type PillTheme = "neutral" | "success" | "info" | "warning";

const themes: Record<PillTheme, string> = {
  neutral: "bg-slate-100 text-slate-700 dark:bg-slate-700 dark:text-slate-200",
  success: "bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200",
  info: "bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200",
  warning: "bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200",
};

export default function Pill({
  theme = "neutral",
  className,
  children,
}: {
  theme?: PillTheme;
  className?: string;
  children: React.ReactNode;
}) {
  return (
    <span
      className={cx(
        "rounded-full px-2 py-1 text-xs font-medium whitespace-nowrap",
        themes[theme],
        className,
      )}
    >
      {children}
    </span>
  );
}
