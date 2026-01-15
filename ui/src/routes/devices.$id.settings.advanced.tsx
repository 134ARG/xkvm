import { useCallback, useEffect, useState } from "react";

import { useSettingsStore } from "@hooks/stores";
import { JsonRpcResponse, useJsonRpc } from "@hooks/useJsonRpc";
import { Button } from "@components/Button";
import Checkbox from "@components/Checkbox";
import { ConfirmDialog } from "@components/ConfirmDialog";
import { SettingsItem } from "@components/SettingsItem";
import { SettingsPageHeader } from "@components/SettingsPageheader";
import { NestedSettingsGroup } from "@components/NestedSettingsGroup";
import notifications from "@/notifications";
import { m } from "@localizations/messages.js";
import { sleep } from "@/utils";

export default function SettingsAdvancedRoute() {
  const { send } = useJsonRpc();

  const [usbEmulationEnabled, setUsbEmulationEnabled] = useState(false);
  const [showLoopbackWarning, setShowLoopbackWarning] = useState(false);
  const [localLoopbackOnly, setLocalLoopbackOnly] = useState(false);
  const [diagnosticsLoading, setDiagnosticsLoading] = useState(false);
  const settings = useSettingsStore();

  useEffect(() => {
    send("getUsbEmulationState", {}, (resp: JsonRpcResponse) => {
      if ("error" in resp) return;
      setUsbEmulationEnabled(resp.result as boolean);
    });

    send("getLocalLoopbackOnly", {}, (resp: JsonRpcResponse) => {
      if ("error" in resp) return;
      setLocalLoopbackOnly(resp.result as boolean);
    });
  }, [send]);

  const getUsbEmulationState = useCallback(() => {
    send("getUsbEmulationState", {}, (resp: JsonRpcResponse) => {
      if ("error" in resp) return;
      setUsbEmulationEnabled(resp.result as boolean);
    });
  }, [send]);

  const handleUsbEmulationToggle = useCallback(
    (enabled: boolean) => {
      send("setUsbEmulationState", { enabled: enabled }, (resp: JsonRpcResponse) => {
        if ("error" in resp) {
          notifications.error(
            enabled
              ? m.advanced_error_usb_emulation_enable({
                  error: resp.error.data || m.unknown_error(),
                })
              : m.advanced_error_usb_emulation_disable({
                  error: resp.error.data || m.unknown_error(),
                }),
          );
          return;
        }
        setUsbEmulationEnabled(enabled);
        getUsbEmulationState();
      });
    },
    [getUsbEmulationState, send],
  );

  const handleResetConfig = useCallback(() => {
    send("resetConfig", {}, (resp: JsonRpcResponse) => {
      if ("error" in resp) {
        notifications.error(
          m.advanced_error_reset_config({ error: resp.error.data || m.unknown_error() }),
        );
        return;
      }
      notifications.success(m.advanced_success_reset_config());
    });
  }, [send]);

  const applyLoopbackOnlyMode = useCallback(
    (enabled: boolean) => {
      send("setLocalLoopbackOnly", { enabled }, (resp: JsonRpcResponse) => {
        if ("error" in resp) {
          notifications.error(
            enabled
              ? m.advanced_error_loopback_enable({ error: resp.error.data || m.unknown_error() })
              : m.advanced_error_loopback_disable({ error: resp.error.data || m.unknown_error() }),
          );
          return;
        }
        setLocalLoopbackOnly(enabled);
        if (enabled) {
          notifications.success(m.advanced_success_loopback_enabled());
        } else {
          notifications.success(m.advanced_success_loopback_disabled());
        }
      });
    },
    [send, setLocalLoopbackOnly],
  );

  const handleLoopbackOnlyModeChange = useCallback(
    (enabled: boolean) => {
      // If trying to enable loopback-only mode, show warning first
      if (enabled) {
        setShowLoopbackWarning(true);
      } else {
        // If disabling, just proceed
        applyLoopbackOnlyMode(false);
      }
    },
    [applyLoopbackOnlyMode, setShowLoopbackWarning],
  );

  const confirmLoopbackModeEnable = useCallback(() => {
    applyLoopbackOnlyMode(true);
    setShowLoopbackWarning(false);
  }, [applyLoopbackOnlyMode, setShowLoopbackWarning]);

  const handleDownloadDiagnostics = useCallback(() => {
    setDiagnosticsLoading(true);

    send("getDiagnostics", {}, async (resp: JsonRpcResponse) => {
      setDiagnosticsLoading(false);

      if ("error" in resp) {
        notifications.error(
          m.advanced_error_download_diagnostics({ error: resp.error.data || m.unknown_error() }),
        );
        return;
      }

      const logContent = resp.result as string;
      const timestamp = new Date().toISOString().replace(/[:.]/g, "-");
      const filename = `xkvm-diagnostics-${timestamp}.txt`;

      const blob = new Blob([logContent], { type: "text/plain" });
      const url = URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = filename;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      URL.revokeObjectURL(url);

      notifications.success(m.advanced_success_download_diagnostics());
    });
  }, [send]);

  // OTA custom version update handlers disabled
  /*
  const handleVersionUpdateError = useCallback((error?: JsonRpcError | string) => {
    notifications.error(
      m.advanced_error_version_update({
        error:
          typeof error === "string" ? error : (error?.data ?? error?.message ?? m.unknown_error()),
      }),
      { duration: 1000 * 15 }, // 15 seconds
    );
    setCustomVersionUpdateLoading(false);
  }, []);

  const handleCustomVersionUpdate = useCallback(async () => {
    const components: UpdateComponents = {};
    if (["app", "both"].includes(updateTarget) && appVersion) components.app = appVersion;
    if (["system", "both"].includes(updateTarget) && systemVersion)
      components.system = systemVersion;
    let versionInfo: SystemVersionInfo | undefined;

    try {
      // we do not need to set it to false if check succeeds,
      // because it will be redirected to the update page later
      setCustomVersionUpdateLoading(true);
      versionInfo = await checkUpdateComponents({ components }, devChannel);
    } catch (error: unknown) {
      const jsonRpcError = error as JsonRpcError;
      handleVersionUpdateError(jsonRpcError);
      return;
    }

    let hasUpdate = false;

    const pageParams = new URLSearchParams();
    if (components.app && versionInfo?.remote?.appVersion && versionInfo?.appUpdateAvailable) {
      hasUpdate = true;
      pageParams.set("custom_app_version", versionInfo.remote?.appVersion);
    }
    if (
      components.system &&
      versionInfo?.remote?.systemVersion &&
      versionInfo?.systemUpdateAvailable
    ) {
      hasUpdate = true;
      pageParams.set("custom_system_version", versionInfo.remote?.systemVersion);
    }
    pageParams.set("reset_config", resetConfig.toString());

    if (!hasUpdate) {
      handleVersionUpdateError("No update available");
      return;
    }

    // Navigate to update page
    navigateTo(`/settings/general/update?${pageParams.toString()}`);
  }, [
    appVersion,
    devChannel,
    handleVersionUpdateError,
    navigateTo,
    resetConfig,
    systemVersion,
    updateTarget,
  ]);
  */

  return (
    <div className="space-y-4">
      <SettingsPageHeader title={m.advanced_title()} description={m.advanced_description()} />

      <div className="space-y-4">
        <SettingsItem
          title={m.advanced_loopback_only_title()}
          description={m.advanced_loopback_only_description()}
        >
          <Checkbox
            checked={localLoopbackOnly}
            onChange={e => handleLoopbackOnlyModeChange(e.target.checked)}
          />
        </SettingsItem>

        <SettingsItem
          title={m.advanced_troubleshooting_mode_title()}
          description={m.advanced_troubleshooting_mode_description()}
        >
          <Checkbox
            defaultChecked={settings.debugMode}
            onChange={e => {
              settings.setDebugMode(e.target.checked);
            }}
          />
        </SettingsItem>

        {settings.debugMode && (
          <NestedSettingsGroup>
            <SettingsItem
              title={m.advanced_usb_emulation_title()}
              description={m.advanced_usb_emulation_description()}
            >
              <Button
                size="SM"
                theme="light"
                text={
                  usbEmulationEnabled
                    ? m.advanced_disable_usb_emulation()
                    : m.advanced_enable_usb_emulation()
                }
                onClick={() => handleUsbEmulationToggle(!usbEmulationEnabled)}
              />
            </SettingsItem>

            <SettingsItem
              title={m.advanced_reset_config_title()}
              description={m.advanced_reset_config_description()}
            >
              <Button
                size="SM"
                theme="light"
                text={m.advanced_reset_config_button()}
                onClick={async () => {
                  handleResetConfig();
                  // Add 2s delay between resetting the configuration and calling reload() to prevent reload from interrupting the RPC call to reset things.
                  await sleep(2000);
                  window.location.reload();
                }}
              />
            </SettingsItem>

            <SettingsItem
              title={m.advanced_download_diagnostics_title()}
              description={m.advanced_download_diagnostics_description()}
            >
              <Button
                size="SM"
                disabled={diagnosticsLoading}
                theme="light"
                text={m.advanced_download_diagnostics_button()}
                loading={diagnosticsLoading}
                onClick={handleDownloadDiagnostics}
              />
            </SettingsItem>
          </NestedSettingsGroup>
        )}
      </div>

      <ConfirmDialog
        open={showLoopbackWarning}
        onClose={() => {
          setShowLoopbackWarning(false);
        }}
        title={m.advanced_loopback_warning_title()}
        description={
          <>
            <p>{m.advanced_loopback_warning_description()}</p>
            <p>{m.advanced_loopback_warning_before()}</p>
            <ul className="list-disc space-y-1 pl-5 text-xs text-slate-700 dark:text-slate-300">
              <li>{m.advanced_loopback_warning_ssh()}</li>
              <li>{m.advanced_loopback_warning_cloud()}</li>
            </ul>
          </>
        }
        variant="warning"
        confirmText={m.advanced_loopback_warning_confirm()}
        onConfirm={confirmLoopbackModeEnable}
      />
    </div>
  );
}
