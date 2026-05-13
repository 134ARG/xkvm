import { ChangeEvent, useCallback, useEffect, useState } from "react";

import { JsonRpcResponse, useJsonRpc } from "@hooks/useJsonRpc";
import { CheckboxWithLabel } from "@components/Checkbox";
import { InputFieldWithLabel } from "@components/InputField";
import { SettingsSectionHeader } from "@components/SettingsSectionHeader";
import notifications from "@/notifications";

interface VFDConfig {
  enabled: boolean;
  devicePath: string;
  listenPort: number;
}

const defaultConfig: VFDConfig = {
  enabled: false,
  devicePath: "",
  listenPort: 9101,
};

export function VFDSetting() {
  const { send } = useJsonRpc();
  const [vfdConfig, setVFDConfig] = useState<VFDConfig>(defaultConfig);

  useEffect(() => {
    send("getVFDConfig", {}, (resp: JsonRpcResponse) => {
      if ("error" in resp) return;
      setVFDConfig((resp.result as VFDConfig) || defaultConfig);
    });
  }, [send]);

  const saveVFDConfig = useCallback(
    (updated: VFDConfig) => {
      send("setVFDConfig", { vfdConfig: updated }, (resp: JsonRpcResponse) => {
        if ("error" in resp) {
          notifications.error(`Failed to save VFD config: ${String(resp.error.data || "")}`);
          return;
        }
        setVFDConfig(updated);
      });
    },
    [send],
  );

  return (
    <div className="space-y-4">
      <div className="h-px w-full bg-slate-800/10 dark:bg-slate-300/20" />

      <SettingsSectionHeader
        title="VFD display"
        description="Configure the CH347 driven VFD status display and host metrics receiver."
      />

      <div className="space-y-4">
        <div>
          <CheckboxWithLabel
            label="Enabled"
            description="Start the VFD renderer and metrics receiver."
            checked={vfdConfig.enabled}
            onChange={(e: ChangeEvent<HTMLInputElement>) => {
              saveVFDConfig({ ...vfdConfig, enabled: e.target.checked });
            }}
          />
        </div>

        <div className="grid grid-cols-1 items-end gap-4 md:grid-cols-2">
          <div className="space-y-1">
            <InputFieldWithLabel
              label="Device path"
              description="Leave empty to auto-detect CH347."
              placeholder="/dev/hidrawN"
              value={vfdConfig.devicePath}
              onChange={(e: ChangeEvent<HTMLInputElement>) => {
                setVFDConfig({ ...vfdConfig, devicePath: e.target.value });
              }}
              onBlur={() => saveVFDConfig(vfdConfig)}
            />
          </div>
          <div className="space-y-1">
            <InputFieldWithLabel
              label="Metrics port"
              description="Host agent receiver port."
              type="number"
              min={1}
              max={65535}
              value={vfdConfig.listenPort}
              onChange={(e: ChangeEvent<HTMLInputElement>) => {
                setVFDConfig({ ...vfdConfig, listenPort: Number(e.target.value) });
              }}
              onBlur={() => saveVFDConfig(vfdConfig)}
            />
          </div>
        </div>
      </div>
    </div>
  );
}
