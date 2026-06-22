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
import { isNative } from "@/main";
import { BackendUpdateInfo, checkBackendUpdate, tryUpdateBackend } from "@/utils/jsonrpc";
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
import { deleteCookie, map_locale_code_to_name } from "@/utils";

// JSON-RPC internal errors put the useful text in `data`; `message` is a
// generic "Internal error".
function rpcErrorMessage(error: unknown): string {
  const e = error as { data?: unknown; message?: string };
  if (typeof e?.data === "string" && e.data) return e.data;
  return e?.message || m.unknown_error();
}

// Backend restart + reconnect should complete well within this window; if not,
// surface an error rather than spinning forever.
const UPDATE_TIMEOUT_MS = 120000;

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
  const [connectorVersion, setConnectorVersion] = useState<string | null>(null);

  const peerConnectionState = useRTCStore(s => s.peerConnectionState);
  const sawDisconnectRef = useRef(false);

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

  // While a backend update is in progress, the service restarts and the
  // connection drops; once it reconnects, force a full reload.
  useEffect(() => {
    if (!updating) return;
    if (peerConnectionState && peerConnectionState !== "connected") {
      sawDisconnectRef.current = true;
    } else if (sawDisconnectRef.current && peerConnectionState === "connected") {
      window.location.reload();
    }
  }, [updating, peerConnectionState]);

  // If the install fails server-side or the new backend never comes back, the
  // connection won't cycle — bail out with an error instead of hanging.
  useEffect(() => {
    if (!updating) return;
    const timer = setTimeout(() => {
      setUpdating(false);
      sawDisconnectRef.current = false;
      notifications.error(m.general_backend_update_timeout());
    }, UPDATE_TIMEOUT_MS);
    return () => clearTimeout(timer);
  }, [updating]);

  const onUpdateNow = useCallback(async () => {
    sawDisconnectRef.current = false;
    setUpdating(true);
    try {
      await tryUpdateBackend();
    } catch (error) {
      notifications.error(m.updates_failed_check({ error: rpcErrorMessage(error) }));
      setUpdating(false);
    }
  }, []);

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
        {!updating && info?.updateAvailable ? (
          info.canAutoUpdate ? (
            <Button size="SM" theme="primary" text={m.general_update_now()} onClick={onUpdateNow} />
          ) : (
            <ExtLink href={info.releaseUrl}>
              <Button size="SM" theme="light" text={m.general_download_update()} />
            </ExtLink>
          )
        ) : null}
      </VersionRow>

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
