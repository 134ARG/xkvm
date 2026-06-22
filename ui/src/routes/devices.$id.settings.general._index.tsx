import { useState, useMemo, useEffect, useCallback, useRef, type ReactNode } from "react";

import { SelectMenuBasic } from "@components/SelectMenuBasic";
import { SettingsItem } from "@components/SettingsItem";
import { SettingsPageHeader } from "@components/SettingsPageheader";
import { SettingsSectionHeader } from "@components/SettingsSectionHeader";
import { Button } from "@components/Button";
import ExtLink from "@components/ExtLink";
import Pill, { PillTheme } from "@components/Pill";
import LoadingSpinner from "@components/LoadingSpinner";
import { useRTCStore } from "@/hooks/stores";
import { useJsonRpc } from "@/hooks/useJsonRpc";
import { isNative } from "@/main";
import {
  BackendUpdateInfo,
  cancelBackendUpdate,
  checkBackendUpdate,
  tryUpdateBackend,
} from "@/utils/jsonrpc";
import notifications from "@/notifications";
import {
  getLocale,
  setLocale,
  locales,
  baseLocale,
  cookieName,
  localStorageKey,
} from "@localizations/runtime.js";
import { m } from "@localizations/messages.js";
import { deleteCookie, formatters, map_locale_code_to_name } from "@/utils";

// JSON-RPC internal errors put the useful text in `data`; `message` is a
// generic "Internal error".
function rpcErrorMessage(error: unknown): string {
  const e = error as { data?: unknown; message?: string };
  if (typeof e?.data === "string" && e.data) return e.data;
  return e?.message || m.unknown_error();
}

// Inactivity watchdog: reset on every progress/installing event, so a slow but
// moving download never trips it. Fires only when events stop arriving and the
// connection hasn't cycled — i.e. something is genuinely stuck.
const UPDATE_WATCHDOG_MS = 90000;

export default function SettingsGeneralRoute() {
  const [currentLocale, setCurrentLocale] = useState(getLocale());

  const localeOptions = useMemo(() => {
    return ["", ...locales].map(code => {
      const [localizedName, nativeName] = map_locale_code_to_name(currentLocale, code);
      // don't repeat the name if it's the same in both locales (or blank)
      const label =
        nativeName && nativeName !== localizedName
          ? `${localizedName} - ${nativeName}`
          : localizedName;
      return { value: code, label: label };
    });
  }, [currentLocale]);

  const handleLocaleChange = (newLocale: string) => {
    if (newLocale === currentLocale) return;

    let validLocale = newLocale as (typeof locales)[number];

    if (newLocale !== "") {
      if (!locales.includes(validLocale)) {
        validLocale = baseLocale;
      }

      setLocale(validLocale); // tell the i18n system to change locale
    } else {
      localStorage.removeItem(localStorageKey);
      deleteCookie(cookieName, "", "/");
      window.location.reload();
    }

    setCurrentLocale(validLocale);
    notifications.success(m.locale_change_success({ locale: validLocale || m.locale_auto() }));
  };

  return (
    <div className="space-y-4">
      <SettingsPageHeader title={m.general_title()} description={m.general_page_description()} />

      <div className="space-y-4">
        <div className="space-y-4 pb-2">
          <div className="space-y-4">
            <SettingsItem
              badge={m.beta()}
              badgeTheme="info"
              title={m.user_interface_language_title()}
              description={m.user_interface_language_description()}
            >
              <SelectMenuBasic
                size="SM"
                label=""
                value={currentLocale}
                options={localeOptions}
                onChange={e => {
                  handleLocaleChange(e.target.value);
                }}
              />
            </SettingsItem>
          </div>

          <UpdateSection />
        </div>
      </div>
    </div>
  );
}

function UpdateSection() {
  const [info, setInfo] = useState<BackendUpdateInfo | null>(null);
  const [checking, setChecking] = useState(true);
  const [updating, setUpdating] = useState(false);
  const [installing, setInstalling] = useState(false);
  const [progress, setProgress] = useState<{ downloaded: number; total: number } | null>(null);
  const [connectorVersion, setConnectorVersion] = useState<string | null>(null);

  const peerConnectionState = useRTCStore(s => s.peerConnectionState);
  const sawDisconnectRef = useRef(false);
  const watchdogRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const resetUpdate = useCallback(() => {
    if (watchdogRef.current) clearTimeout(watchdogRef.current);
    watchdogRef.current = null;
    sawDisconnectRef.current = false;
    setUpdating(false);
    setInstalling(false);
    setProgress(null);
  }, []);

  // Feed the inactivity watchdog; an update with no events and no reconnect for
  // this long is treated as stuck.
  const feedWatchdog = useCallback(() => {
    if (watchdogRef.current) clearTimeout(watchdogRef.current);
    watchdogRef.current = setTimeout(() => {
      notifications.error(m.general_backend_update_timeout());
      resetUpdate();
    }, UPDATE_WATCHDOG_MS);
  }, [resetUpdate]);

  useEffect(() => () => clearTimeout(watchdogRef.current ?? undefined), []);

  // Backend update lifecycle events: progress while downloading, then either a
  // restart (installing) or a failure.
  useJsonRpc(
    useCallback(
      payload => {
        switch (payload.method) {
          case "backendUpdateProgress":
            setProgress(payload.params as { downloaded: number; total: number });
            feedWatchdog();
            break;
          case "backendUpdateInstalling":
            setInstalling(true);
            setProgress(null);
            feedWatchdog();
            break;
          case "backendUpdateFailed":
            notifications.error(
              m.updates_failed_check({
                error: (payload.params as { error?: string })?.error || m.unknown_error(),
              }),
            );
            resetUpdate();
            break;
          case "backendUpdateCanceled":
            resetUpdate();
            break;
        }
      },
      [feedWatchdog, resetUpdate],
    ),
  );

  const check = useCallback(async () => {
    setChecking(true);
    try {
      const result = await checkBackendUpdate();
      setInfo(result);
    } catch (error) {
      notifications.error(m.updates_failed_check({ error: rpcErrorMessage(error) }));
    } finally {
      setChecking(false);
    }
  }, []);

  useEffect(() => {
    check();
  }, [check]);

  // Read the Tauri Connector version when running as the native app.
  useEffect(() => {
    if (!isNative) return;
    import("@tauri-apps/api/app")
      .then(({ getVersion }) => getVersion())
      .then(setConnectorVersion)
      .catch(() => setConnectorVersion(null));
  }, []);

  // Only once the install starts does the service restart and the connection
  // drop; reload after it cycles. Gating on `installing` (not `updating`) avoids
  // a transient WebRTC blip during the download triggering a premature reload.
  useEffect(() => {
    if (!installing) return;
    if (peerConnectionState && peerConnectionState !== "connected") {
      sawDisconnectRef.current = true;
    } else if (sawDisconnectRef.current && peerConnectionState === "connected") {
      window.location.reload();
    }
  }, [installing, peerConnectionState]);

  const onUpdateNow = useCallback(async () => {
    setUpdating(true);
    feedWatchdog();
    try {
      await tryUpdateBackend();
    } catch (error) {
      notifications.error(m.updates_failed_check({ error: rpcErrorMessage(error) }));
      resetUpdate();
    }
  }, [feedWatchdog, resetUpdate]);

  const onCancel = useCallback(async () => {
    try {
      await cancelBackendUpdate();
    } catch {
      // Best-effort; the backend may already be past the cancelable phase.
    }
    resetUpdate();
  }, [resetUpdate]);

  const connectorUpdateAvailable =
    isNative &&
    !!connectorVersion &&
    !!info?.latestVersion &&
    connectorVersion !== info.latestVersion;

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between gap-x-8">
        <SettingsSectionHeader
          title={m.general_versions_title()}
          description={m.general_versions_description()}
        />
        <Button
          size="SM"
          theme="light"
          text={m.general_check_for_updates()}
          onClick={check}
          loading={checking}
          disabled={checking || updating}
        />
      </div>

      <VersionRow
        name={m.general_backend_name()}
        version={info?.currentVersion}
        status={
          updating
            ? { theme: "info", label: m.general_status_updating() }
            : info?.updateAvailable
              ? { theme: "info", label: m.general_status_update_available() }
              : info
                ? { theme: "success", label: m.general_status_up_to_date() }
                : undefined
        }
        description={updating ? m.general_backend_updating() : undefined}
        loading={updating}
      >
        {updating ? (
          // During the download the update button doubles as cancel; once
          // installing there's no going back, so no button.
          !installing ? (
            <Button size="SM" theme="light" text={m.cancel()} onClick={onCancel} />
          ) : null
        ) : info?.updateAvailable ? (
          info.canAutoUpdate ? (
            <Button size="SM" theme="primary" text={m.general_update_now()} onClick={onUpdateNow} />
          ) : (
            <ExtLink href={info.releaseUrl}>
              <Button size="SM" theme="light" text={m.general_download_update()} />
            </ExtLink>
          )
        ) : null}
      </VersionRow>

      {updating && <DownloadProgress progress={progress} />}

      {isNative && (
        <VersionRow
          name={m.general_connector_name()}
          version={connectorVersion ?? undefined}
          status={
            connectorUpdateAvailable
              ? { theme: "info", label: m.general_status_update_available() }
              : connectorVersion
                ? { theme: "success", label: m.general_status_up_to_date() }
                : undefined
          }
        >
          {connectorUpdateAvailable && info ? (
            <ExtLink href={info.releaseUrl}>
              <Button size="SM" theme="light" text={m.general_download_update()} />
            </ExtLink>
          ) : null}
        </VersionRow>
      )}
    </div>
  );
}

function VersionRow({
  name,
  version,
  status,
  description,
  loading,
  children,
}: {
  name: string;
  version?: string;
  status?: { theme: PillTheme; label: string };
  description?: string;
  loading?: boolean;
  children?: ReactNode;
}) {
  return (
    <div className="flex items-center justify-between gap-x-8">
      <div className="space-y-1">
        <div className="flex items-center gap-x-2">
          <span className="text-base font-semibold text-black dark:text-white">{name}</span>
          {version && <Pill theme="neutral">v{version}</Pill>}
          {status && <Pill theme={status.theme}>{status.label}</Pill>}
          {loading && <LoadingSpinner className="h-4 w-4 text-blue-500" />}
        </div>
        {description && (
          <div className="text-sm text-slate-700 dark:text-slate-300">{description}</div>
        )}
      </div>
      {children ? <div>{children}</div> : null}
    </div>
  );
}

function DownloadProgress({
  progress,
}: {
  progress: { downloaded: number; total: number } | null;
}) {
  const hasTotal = !!progress && progress.total > 0;
  const percent = hasTotal ? Math.min((progress.downloaded / progress.total) * 100, 100) : 0;

  return (
    <div className="space-y-1">
      <div className="h-2 w-full overflow-hidden rounded-full bg-slate-200 dark:bg-slate-700">
        <div
          className={
            hasTotal
              ? "h-2 bg-blue-700 transition-all duration-300 ease-out"
              : "h-2 w-full animate-pulse bg-blue-700"
          }
          style={hasTotal ? { width: `${percent}%` } : undefined}
        />
      </div>
      <div className="text-xs text-slate-600 dark:text-slate-300">
        {progress
          ? hasTotal
            ? `${formatters.bytes(progress.downloaded)} / ${formatters.bytes(progress.total)} (${Math.round(percent)}%)`
            : formatters.bytes(progress.downloaded)
          : m.general_status_updating()}
      </div>
    </div>
  );
}
