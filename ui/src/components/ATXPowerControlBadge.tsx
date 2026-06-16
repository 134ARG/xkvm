import { Fragment, type ReactNode, useCallback, useEffect, useRef, useState } from "react";
import { Popover, PopoverButton, PopoverPanel } from "@headlessui/react";
import { LuHardDrive, LuPower, LuRotateCcw } from "react-icons/lu";

import { cx } from "@/cva.config";
import notifications from "@/notifications";
import { Button } from "@components/Button";
import {
  statusBadgeButtonClassName,
  statusPopoverPanelClassName,
} from "@components/EnvironmentMetricsBadge";
import { JsonRpcResponse, useJsonRpc } from "@hooks/useJsonRpc";
import { m } from "@localizations/messages.js";

const LONG_PRESS_DURATION = 3000;

interface ATXState {
  power: boolean;
  hdd: boolean;
}

function ATXStatusItem({
  icon,
  label,
  active,
  activeClassName,
  pulse,
}: {
  icon: ReactNode;
  label: string;
  active: boolean;
  activeClassName: string;
  pulse?: boolean;
}) {
  return (
    <span
      className={cx(
        "flex items-center gap-x-1 rounded-sm border border-slate-800/10 px-2 py-1 text-xs font-medium text-slate-900 dark:border-slate-300/20 dark:text-white",
      )}
    >
      <span
        className={cx(
          active ? activeClassName : "text-slate-300 dark:text-slate-600",
          pulse && active && "animate-pulse",
        )}
      >
        {icon}
      </span>
      {label}
    </span>
  );
}

export function ATXPowerControlBadge({
  onOpen,
  onOpenChange,
}: {
  onOpen?: () => void;
  onOpenChange?: (open: boolean) => void;
}) {
  const [atxState, setAtxState] = useState<ATXState | null>(null);
  const [isPowerPressed, setIsPowerPressed] = useState(false);
  const powerPressTimer = useRef<ReturnType<typeof setTimeout> | null>(null);
  const didLongPress = useRef(false);

  const { send } = useJsonRpc(resp => {
    if (resp.method === "atxState") setAtxState(resp.params as ATXState);
  });

  useEffect(() => {
    send("getATXState", {}, (resp: JsonRpcResponse) => {
      if ("error" in resp) {
        notifications.error(
          m.atx_power_control_get_state_error({ error: resp.error.data || m.unknown_error() }),
        );
        return;
      }
      setAtxState(resp.result as ATXState);
    });
  }, [send]);

  const sendAction = useCallback(
    (action: "power-short" | "power-long" | "reset", actionLabel: string) => {
      send("setATXPowerAction", { action }, (resp: JsonRpcResponse) => {
        if ("error" in resp) {
          notifications.error(
            m.atx_power_control_send_action_error({
              action: actionLabel,
              error: resp.error.data || m.unknown_error(),
            }),
          );
        }
      });
    },
    [send],
  );

  const clearPowerTimer = useCallback(() => {
    if (!powerPressTimer.current) return;
    clearTimeout(powerPressTimer.current);
    powerPressTimer.current = null;
  }, []);

  const handlePowerPress = useCallback(
    (pressed: boolean) => {
      if (pressed) {
        if (powerPressTimer.current) return;
        didLongPress.current = false;
        setIsPowerPressed(true);
        powerPressTimer.current = setTimeout(() => {
          powerPressTimer.current = null;
          didLongPress.current = true;
          setIsPowerPressed(false);
          sendAction("power-long", m.atx_power_control_long_power_button());
        }, LONG_PRESS_DURATION);
        return;
      }

      if (!isPowerPressed && !powerPressTimer.current) return;
      clearPowerTimer();
      setIsPowerPressed(false);
      if (!didLongPress.current) {
        sendAction("power-short", m.atx_power_control_short_power_button());
      }
    },
    [clearPowerTimer, isPowerPressed, sendAction],
  );

  useEffect(() => clearPowerTimer, [clearPowerTimer]);

  const powerActive = atxState?.power ?? false;
  const hddActive = atxState?.hdd ?? false;

  return (
    <Popover>
      <PopoverButton as={Fragment}>
        <button
          type="button"
          title={m.extensions_atx_power_control()}
          className={cx(statusBadgeButtonClassName, "gap-x-1.5")}
          onClick={onOpen}
        >
          <LuPower
            className={cx("h-3.5 w-3.5", powerActive ? "text-green-600" : "text-slate-300")}
            strokeWidth={3}
          />
          <LuHardDrive
            className={cx(
              "h-3.5 w-3.5",
              hddActive ? "animate-pulse text-blue-500" : "text-slate-300",
            )}
            strokeWidth={3}
          />
        </button>
      </PopoverButton>
      <PopoverPanel
        anchor="bottom end"
        transition
        className={cx(
          statusPopoverPanelClassName,
          "w-[280px] p-3",
          "origin-top transition duration-200 ease-out data-closed:translate-y-2 data-closed:opacity-0",
        )}
      >
        {({ open }) => {
          onOpenChange?.(open);
          return (
            <div className="space-y-3">
              <div className="flex items-center justify-between">
                <h3 className="text-sm font-semibold text-slate-950 dark:text-white">
                  {m.extensions_atx_power_control()}
                </h3>
              </div>
              <div className="grid grid-cols-2 gap-2">
                <ATXStatusItem
                  icon={<LuPower className="h-3.5 w-3.5" strokeWidth={3} />}
                  label="PWR"
                  active={powerActive}
                  activeClassName="text-green-600"
                />
                <ATXStatusItem
                  icon={<LuHardDrive className="h-3.5 w-3.5" strokeWidth={3} />}
                  label="HDD"
                  active={hddActive}
                  activeClassName="text-blue-500"
                  pulse
                />
              </div>
              <div className="h-px bg-slate-800/10 dark:bg-slate-300/20" />
              <div className="flex items-center gap-x-2">
                <Button
                  size="SM"
                  theme="light"
                  LeadingIcon={LuPower}
                  text={
                    powerActive
                      ? m.dc_power_control_power_off_button()
                      : m.dc_power_control_power_on_button()
                  }
                  onMouseDown={() => handlePowerPress(true)}
                  onMouseUp={() => handlePowerPress(false)}
                  onMouseLeave={() => handlePowerPress(false)}
                  className={isPowerPressed ? "opacity-75" : ""}
                />
                <Button
                  size="SM"
                  theme="light"
                  LeadingIcon={LuRotateCcw}
                  text={m.atx_power_control_reset_button()}
                  onClick={() => sendAction("reset", m.atx_power_control_reset_button())}
                />
              </div>
            </div>
          );
        }}
      </PopoverPanel>
    </Popover>
  );
}
