import { ChangeEvent, useCallback, useEffect, useMemo, useState } from "react";

import { JsonRpcResponse, useJsonRpc } from "@hooks/useJsonRpc";
import { SelectMenuBasic } from "@components/SelectMenuBasic";
import { SettingsSectionHeader } from "@components/SettingsSectionHeader";
import notifications from "@/notifications";
import { m } from "@localizations/messages.js";

interface SensorFeature {
  path: string;
  label: string;
  value: number;
}

interface SensorChip {
  chip: string;
  adapter: string;
  features: SensorFeature[];
}

interface SensorConfig {
  envChip: string;
  tempFeature: string;
  humFeature: string;
  socChip: string;
  socFeature: string;
}

const emptyConfig: SensorConfig = {
  envChip: "",
  tempFeature: "",
  humFeature: "",
  socChip: "",
  socFeature: "",
};

export function SensorSetting() {
  const { send } = useJsonRpc();
  const [chips, setChips] = useState<SensorChip[]>([]);
  const [sensorConfig, setSensorConfig] = useState<SensorConfig>(emptyConfig);

  useEffect(() => {
    send("getSensorChips", {}, (resp: JsonRpcResponse) => {
      if ("error" in resp) return;
      setChips((resp.result as SensorChip[]) || []);
    });
    send("getSensorConfig", {}, (resp: JsonRpcResponse) => {
      if ("error" in resp) return;
      setSensorConfig((resp.result as SensorConfig) || emptyConfig);
    });
  }, [send]);

  const saveSensorConfig = useCallback(
    (updated: SensorConfig) => {
      send("setSensorConfig", { sensorConfig: updated }, (resp: JsonRpcResponse) => {
        if ("error" in resp) {
          notifications.error(
            m.sensor_save_error({ error: String(resp.error.data || m.unknown_error()) }),
          );
          return;
        }
        setSensorConfig(updated);
      });
    },
    [send],
  );

  const chipOptions = useMemo(
    () => [
      { label: m.sensor_none(), value: "" },
      ...chips.map(c => ({ label: `${c.chip} - ${c.adapter}`, value: c.chip })),
    ],
    [chips],
  );

  const featureOptions = useCallback(
    (chipName: string) => {
      const chip = chips.find(c => c.chip === chipName);
      return [
        { label: m.sensor_none(), value: "" },
        ...(chip?.features.map(f => ({
          label: `${f.label} (${f.value.toFixed(1)})`,
          value: f.path,
        })) || []),
      ];
    },
    [chips],
  );

  return (
    <div className="space-y-4">
      <div className="h-px w-full bg-slate-800/10 dark:bg-slate-300/20" />

      <SettingsSectionHeader
        title={m.sensor_case_sensors_title()}
        description={m.sensor_case_sensors_description()}
      />

      <div className="grid grid-cols-3 gap-4">
        <SelectMenuBasic
          label={m.sensor_environment_chip()}
          options={chipOptions}
          value={sensorConfig.envChip}
          onChange={(e: ChangeEvent<HTMLSelectElement>) => {
            const updated = {
              ...sensorConfig,
              envChip: e.target.value,
              tempFeature: "",
              humFeature: "",
            };
            saveSensorConfig(updated);
          }}
        />
        <SelectMenuBasic
          label={m.sensor_temperature()}
          options={featureOptions(sensorConfig.envChip)}
          value={sensorConfig.tempFeature}
          disabled={!sensorConfig.envChip}
          onChange={(e: ChangeEvent<HTMLSelectElement>) => {
            saveSensorConfig({ ...sensorConfig, tempFeature: e.target.value });
          }}
        />
        <SelectMenuBasic
          label={m.sensor_humidity()}
          options={featureOptions(sensorConfig.envChip)}
          value={sensorConfig.humFeature}
          disabled={!sensorConfig.envChip}
          onChange={(e: ChangeEvent<HTMLSelectElement>) => {
            saveSensorConfig({ ...sensorConfig, humFeature: e.target.value });
          }}
        />
        <SelectMenuBasic
          label={m.sensor_soc_chip()}
          options={chipOptions}
          value={sensorConfig.socChip}
          onChange={(e: ChangeEvent<HTMLSelectElement>) => {
            const updated = { ...sensorConfig, socChip: e.target.value, socFeature: "" };
            saveSensorConfig(updated);
          }}
        />
        <SelectMenuBasic
          label={m.sensor_soc_temperature()}
          options={featureOptions(sensorConfig.socChip)}
          value={sensorConfig.socFeature}
          disabled={!sensorConfig.socChip}
          onChange={(e: ChangeEvent<HTMLSelectElement>) => {
            saveSensorConfig({ ...sensorConfig, socFeature: e.target.value });
          }}
        />
      </div>
    </div>
  );
}
