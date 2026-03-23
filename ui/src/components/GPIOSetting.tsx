import { useCallback, useEffect, useState } from "react";

import { JsonRpcResponse, useJsonRpc } from "@hooks/useJsonRpc";
import { m } from "@localizations/messages.js";
import { SelectMenuBasic } from "@components/SelectMenuBasic";
import { SettingsSectionHeader } from "@components/SettingsSectionHeader";
import notifications from "@/notifications";

interface HardwarePort {
  path: string;
  type: string;
  info: string;
}

interface GPIOConfig {
  pwrChip: string;
  pwrLine: number;
  pwrActiveHigh: boolean;
  rstChip: string;
  rstLine: number;
  rstActiveHigh: boolean;
  pwrLedChip: string;
  pwrLedLine: number;
  pwrLedActiveHigh: boolean;
  hddLedChip: string;
  hddLedLine: number;
  hddLedActiveHigh: boolean;
}

export function GPIOSetting() {
  const { send } = useJsonRpc();
  const [gpioChips, setGpioChips] = useState<HardwarePort[]>([]);
  const [gpioConfig, setGpioConfig] = useState<GPIOConfig>({
    pwrChip: "",
    pwrLine: -1,
    pwrActiveHigh: true,
    rstChip: "",
    rstLine: -1,
    rstActiveHigh: true,
    pwrLedChip: "",
    pwrLedLine: -1,
    pwrLedActiveHigh: false,
    hddLedChip: "",
    hddLedLine: -1,
    hddLedActiveHigh: false,
  });
  const [pwrLineCount, setPwrLineCount] = useState(0);
  const [rstLineCount, setRstLineCount] = useState(0);
  const [pwrLedLineCount, setPwrLedLineCount] = useState(0);
  const [hddLedLineCount, setHddLedLineCount] = useState(0);

  const fetchLineCount = useCallback(
    (chip: string, setter: (n: number) => void) => {
      if (!chip) {
        setter(0);
        return;
      }
      send("getGPIOChipLines", { chip }, (resp: JsonRpcResponse) => {
        if ("error" in resp) {
          setter(0);
          return;
        }
        setter(resp.result as number);
      });
    },
    [send],
  );

  useEffect(() => {
    send("getGPIOChips", {}, (resp: JsonRpcResponse) => {
      if ("error" in resp) return;
      setGpioChips((resp.result as HardwarePort[]) || []);
    });
    send("getGPIOConfig", {}, (resp: JsonRpcResponse) => {
      if ("error" in resp) {
        notifications.error(
          m.hardware_gpio_load_error({ error: String(resp.error.data || m.unknown_error()) }),
        );
        return;
      }
      const cfg = resp.result as GPIOConfig;
      setGpioConfig(cfg);
      fetchLineCount(cfg.pwrChip, setPwrLineCount);
      fetchLineCount(cfg.rstChip, setRstLineCount);
      fetchLineCount(cfg.pwrLedChip, setPwrLedLineCount);
      fetchLineCount(cfg.hddLedChip, setHddLedLineCount);
    });
  }, [send, fetchLineCount]);

  const saveGPIOConfig = useCallback(
    (updated: GPIOConfig) => {
      send("setGPIOConfig", { gpioConfig: updated }, (resp: JsonRpcResponse) => {
        if ("error" in resp) {
          notifications.error(
            m.hardware_gpio_save_error({ error: String(resp.error.data || m.unknown_error()) }),
          );
          return;
        }
        setGpioConfig(updated);
      });
    },
    [send],
  );

  const chipOptions = [
    { label: String(m.hardware_gpio_none()), value: "" },
    ...gpioChips.map(c => ({ label: `${c.path} — ${c.info}`, value: c.path })),
  ];

  const lineOptions = (count: number) => {
    const opts = [{ label: String(m.hardware_gpio_none()), value: "-1" }];
    for (let i = 0; i < count; i++) {
      opts.push({ label: String(i), value: String(i) });
    }
    return opts;
  };

  const activeLevelOptions = [
    { label: String(m.hardware_gpio_active_high()), value: "true" },
    { label: String(m.hardware_gpio_active_low()), value: "false" },
  ];

  return (
    <div className="space-y-4">
      <div className="h-px w-full bg-slate-800/10 dark:bg-slate-300/20" />

      <SettingsSectionHeader
        title={m.hardware_gpio_title()}
        description={m.hardware_gpio_description()}
      />

      <div className="grid grid-cols-3 gap-4">
        <SelectMenuBasic
          label={m.hardware_gpio_pwr_chip()}
          options={chipOptions}
          value={gpioConfig.pwrChip}
          onChange={(e: React.ChangeEvent<HTMLSelectElement>) => {
            const chip = e.target.value;
            const updated = { ...gpioConfig, pwrChip: chip, pwrLine: -1 };
            saveGPIOConfig(updated);
            fetchLineCount(chip, setPwrLineCount);
          }}
        />
        <SelectMenuBasic
          label={m.hardware_gpio_pwr_line()}
          options={lineOptions(pwrLineCount)}
          value={String(gpioConfig.pwrLine)}
          disabled={!gpioConfig.pwrChip}
          onChange={(e: React.ChangeEvent<HTMLSelectElement>) => {
            const updated = { ...gpioConfig, pwrLine: parseInt(e.target.value, 10) };
            saveGPIOConfig(updated);
          }}
        />
        <SelectMenuBasic
          label={m.hardware_gpio_pwr_active_level()}
          options={activeLevelOptions}
          value={String(gpioConfig.pwrActiveHigh)}
          disabled={!gpioConfig.pwrChip}
          onChange={(e: React.ChangeEvent<HTMLSelectElement>) => {
            const updated = { ...gpioConfig, pwrActiveHigh: e.target.value === "true" };
            saveGPIOConfig(updated);
          }}
        />
        <SelectMenuBasic
          label={m.hardware_gpio_rst_chip()}
          options={chipOptions}
          value={gpioConfig.rstChip}
          onChange={(e: React.ChangeEvent<HTMLSelectElement>) => {
            const chip = e.target.value;
            const updated = { ...gpioConfig, rstChip: chip, rstLine: -1 };
            saveGPIOConfig(updated);
            fetchLineCount(chip, setRstLineCount);
          }}
        />
        <SelectMenuBasic
          label={m.hardware_gpio_rst_line()}
          options={lineOptions(rstLineCount)}
          value={String(gpioConfig.rstLine)}
          disabled={!gpioConfig.rstChip}
          onChange={(e: React.ChangeEvent<HTMLSelectElement>) => {
            const updated = { ...gpioConfig, rstLine: parseInt(e.target.value, 10) };
            saveGPIOConfig(updated);
          }}
        />
        <SelectMenuBasic
          label={m.hardware_gpio_rst_active_level()}
          options={activeLevelOptions}
          value={String(gpioConfig.rstActiveHigh)}
          disabled={!gpioConfig.rstChip}
          onChange={(e: React.ChangeEvent<HTMLSelectElement>) => {
            const updated = { ...gpioConfig, rstActiveHigh: e.target.value === "true" };
            saveGPIOConfig(updated);
          }}
        />
        <SelectMenuBasic
          label={m.hardware_gpio_pwr_led_chip()}
          options={chipOptions}
          value={gpioConfig.pwrLedChip}
          onChange={(e: React.ChangeEvent<HTMLSelectElement>) => {
            const chip = e.target.value;
            const updated = { ...gpioConfig, pwrLedChip: chip, pwrLedLine: -1 };
            saveGPIOConfig(updated);
            fetchLineCount(chip, setPwrLedLineCount);
          }}
        />
        <SelectMenuBasic
          label={m.hardware_gpio_pwr_led_line()}
          options={lineOptions(pwrLedLineCount)}
          value={String(gpioConfig.pwrLedLine)}
          disabled={!gpioConfig.pwrLedChip}
          onChange={(e: React.ChangeEvent<HTMLSelectElement>) => {
            const updated = { ...gpioConfig, pwrLedLine: parseInt(e.target.value, 10) };
            saveGPIOConfig(updated);
          }}
        />
        <SelectMenuBasic
          label={m.hardware_gpio_pwr_led_active_level()}
          options={activeLevelOptions}
          value={String(gpioConfig.pwrLedActiveHigh)}
          disabled={!gpioConfig.pwrLedChip}
          onChange={(e: React.ChangeEvent<HTMLSelectElement>) => {
            const updated = { ...gpioConfig, pwrLedActiveHigh: e.target.value === "true" };
            saveGPIOConfig(updated);
          }}
        />
        <SelectMenuBasic
          label={m.hardware_gpio_hdd_led_chip()}
          options={chipOptions}
          value={gpioConfig.hddLedChip}
          onChange={(e: React.ChangeEvent<HTMLSelectElement>) => {
            const chip = e.target.value;
            const updated = { ...gpioConfig, hddLedChip: chip, hddLedLine: -1 };
            saveGPIOConfig(updated);
            fetchLineCount(chip, setHddLedLineCount);
          }}
        />
        <SelectMenuBasic
          label={m.hardware_gpio_hdd_led_line()}
          options={lineOptions(hddLedLineCount)}
          value={String(gpioConfig.hddLedLine)}
          disabled={!gpioConfig.hddLedChip}
          onChange={(e: React.ChangeEvent<HTMLSelectElement>) => {
            const updated = { ...gpioConfig, hddLedLine: parseInt(e.target.value, 10) };
            saveGPIOConfig(updated);
          }}
        />
        <SelectMenuBasic
          label={m.hardware_gpio_hdd_led_active_level()}
          options={activeLevelOptions}
          value={String(gpioConfig.hddLedActiveHigh)}
          disabled={!gpioConfig.hddLedChip}
          onChange={(e: React.ChangeEvent<HTMLSelectElement>) => {
            const updated = { ...gpioConfig, hddLedActiveHigh: e.target.value === "true" };
            saveGPIOConfig(updated);
          }}
        />
      </div>
    </div>
  );
}
