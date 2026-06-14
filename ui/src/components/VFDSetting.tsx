import { ChangeEvent, useCallback, useEffect, useState } from "react";

import { JsonRpcResponse, useJsonRpc } from "@hooks/useJsonRpc";
import { CheckboxWithLabel } from "@components/Checkbox";
import { InputFieldWithLabel } from "@components/InputField";
import { SettingsSectionHeader } from "@components/SettingsSectionHeader";
import notifications from "@/notifications";
import { m } from "@localizations/messages.js";

interface VFDConfig {
  enabled: boolean;
  hostMetricsEnabled: boolean;
  devicePath: string;
  hostMetricsListenPort: number;
}

const defaultConfig: VFDConfig = {
  enabled: false,
  hostMetricsEnabled: false,
  devicePath: "",
  hostMetricsListenPort: 9101,
};

export function VFDSetting() {
  const { send } = useJsonRpc();
  const [vfdConfig, setVFDConfig] = useState<VFDConfig>(defaultConfig);

  useEffect(() => {
    send("getVFDConfig", {}, (resp: JsonRpcResponse) => {
      if ("error" in resp) return;
      setVFDConfig({ ...defaultConfig, ...((resp.result as Partial<VFDConfig>) || {}) });
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

  const hostMetricsEnabled = vfdConfig.hostMetricsEnabled || vfdConfig.enabled;

  return (
    <div className="space-y-4">
      <div className="h-px w-full bg-slate-800/10 dark:bg-slate-300/20" />

      <SettingsSectionHeader title={m.vfd_display_title()} description={m.vfd_description()} />

      <div className="space-y-4">
        <div>
          <CheckboxWithLabel
            label={m.vfd_host_metrics_enabled()}
            description={
              vfdConfig.enabled
                ? m.vfd_host_metrics_required_description()
                : m.vfd_host_metrics_enabled_description()
            }
            checked={hostMetricsEnabled}
            disabled={vfdConfig.enabled}
            onChange={(e: ChangeEvent<HTMLInputElement>) => {
              saveVFDConfig({ ...vfdConfig, hostMetricsEnabled: e.target.checked });
            }}
          />
        </div>

        <div>
          <CheckboxWithLabel
            label={m.vfd_enabled()}
            description={m.vfd_enabled_description()}
            checked={vfdConfig.enabled}
            onChange={(e: ChangeEvent<HTMLInputElement>) => {
              saveVFDConfig({
                ...vfdConfig,
                enabled: e.target.checked,
                hostMetricsEnabled: e.target.checked ? true : vfdConfig.hostMetricsEnabled,
              });
            }}
          />
        </div>

        <div className="grid grid-cols-1 items-end gap-4 md:grid-cols-2">
          <div className="space-y-1">
            <InputFieldWithLabel
              label={m.host_metrics_listen_port()}
              description={m.host_metrics_listen_port_description()}
              type="number"
              min={1}
              max={65535}
              value={vfdConfig.hostMetricsListenPort}
              onChange={(e: ChangeEvent<HTMLInputElement>) => {
                setVFDConfig({ ...vfdConfig, hostMetricsListenPort: Number(e.target.value) });
              }}
              onBlur={() => saveVFDConfig(vfdConfig)}
            />
          </div>
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
        </div>
      </div>
    </div>
  );
}
