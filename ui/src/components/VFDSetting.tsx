import { ChangeEvent, useCallback, useEffect, useState } from "react";

import { JsonRpcResponse, useJsonRpc } from "@hooks/useJsonRpc";
import { CheckboxWithLabel } from "@components/Checkbox";
import { InputFieldWithLabel } from "@components/InputField";
import { SettingsSectionHeader } from "@components/SettingsSectionHeader";
import notifications from "@/notifications";
import { m } from "@localizations/messages.js";

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
          notifications.error(
            m.vfd_save_error({ error: String(resp.error.data || m.unknown_error()) }),
          );
          return;
        }
        setVFDConfig(updated);
        notifications.success(m.vfd_saved_restart());
      });
    },
    [send],
  );

  return (
    <div className="space-y-4">
      <div className="h-px w-full bg-slate-800/10 dark:bg-slate-300/20" />

      <SettingsSectionHeader title={m.vfd_display_title()} description={m.vfd_description()} />

      <div className="space-y-4">
        <div>
          <CheckboxWithLabel
            label={m.vfd_enabled()}
            description={m.vfd_enabled_description()}
            checked={vfdConfig.enabled}
            onChange={(e: ChangeEvent<HTMLInputElement>) => {
              saveVFDConfig({ ...vfdConfig, enabled: e.target.checked });
            }}
          />
        </div>

        <div className="grid grid-cols-1 items-end gap-4 md:grid-cols-2">
          <div className="space-y-1">
            <InputFieldWithLabel
              label={m.vfd_device_path()}
              description={m.vfd_device_path_description()}
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
              label={m.vfd_metrics_port()}
              description={m.vfd_metrics_port_description()}
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
