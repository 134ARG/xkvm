import { useCallback, useEffect, useState } from "react";
import { useLoaderData, type LoaderFunction } from "react-router";

import { useDeviceUiNavigation } from "@hooks/useAppNavigation";
import { JsonRpcResponse, useJsonRpc } from "@hooks/useJsonRpc";
import { Button } from "@components/Button";
import { SelectMenuBasic } from "@components/SelectMenuBasic";
import { SettingsItem } from "@components/SettingsItem";
import { SettingsPageHeader } from "@components/SettingsPageheader";
import { SettingsSectionHeader } from "@components/SettingsSectionHeader";
import { NestedSettingsGroup } from "@components/NestedSettingsGroup";
import { TextAreaWithLabel } from "@components/TextArea";
import api from "@/api";
import notifications from "@/notifications";
import { getDeviceAPI } from "@/ui.config";
import { isOnDevice, isNative } from "@/main";
import { m } from "@localizations/messages.js";

import { LocalDevice } from "./devices.$id";

export interface TLSState {
  mode: "self-signed" | "custom" | "disabled";
  certificate?: string;
  privateKey?: string;
}

const loader: LoaderFunction = async () => {
  // On-device serves /device at the same origin; native (Tauri) reaches it via the
  // configured connection URL. getDeviceAPI() returns "" on-device and that URL in native.
  if (isOnDevice || isNative) {
    const status = await api
      .GET(`${getDeviceAPI()}/device`)
      .then(res => res.json() as Promise<LocalDevice>);
    return status;
  }
  return null;
};

export default function SettingsAccessIndexRoute() {
  const loaderData = useLoaderData() as LocalDevice | null;

  const { navigateTo } = useDeviceUiNavigation();

  const { send } = useJsonRpc();

  const [tlsMode, setTlsMode] = useState<string>("unknown");
  const [tlsCert, setTlsCert] = useState<string>("");
  const [tlsKey, setTlsKey] = useState<string>("");

  const getTLSState = useCallback(() => {
    send("getTLSState", {}, (resp: JsonRpcResponse) => {
      if ("error" in resp) return console.error(resp.error);
      const tlsState = resp.result as TLSState;

      setTlsMode(tlsState.mode);
      if (tlsState.certificate) setTlsCert(tlsState.certificate);
      if (tlsState.privateKey) setTlsKey(tlsState.privateKey);
    });
  }, [send]);

  // Function to update TLS state - accepts a mode parameter
  const updateTlsState = useCallback(
    (mode: string, cert?: string, key?: string) => {
      const state = { mode } as TLSState;
      if (cert && key) {
        state.certificate = cert;
        state.privateKey = key;
      }

      send("setTLSState", { state }, (resp: JsonRpcResponse) => {
        if ("error" in resp) {
          notifications.error(
            m.access_failed_update_tls({ error: resp.error.data || m.unknown_error() }),
          );
          return;
        }

        notifications.success(m.access_tls_updated());
      });
    },
    [send],
  );

  // Handle TLS mode change
  const handleTlsModeChange = (value: string) => {
    setTlsMode(value);

    // For "disabled" and "self-signed" modes, immediately apply the settings
    if (value !== "custom") {
      updateTlsState(value);
    }
  };

  const handleTlsCertChange = (value: string) => {
    setTlsCert(value);
  };

  const handleTlsKeyChange = (value: string) => {
    setTlsKey(value);
  };

  // Update the custom TLS settings button click handler
  const handleCustomTlsUpdate = () => {
    updateTlsState(tlsMode, tlsCert, tlsKey);
  };

  // Fetch TLS state on component mount
  useEffect(() => {
    getTLSState();
  }, [send, getTLSState]);

  return (
    <div className="space-y-4">
      <SettingsPageHeader title={m.access_title()} description={m.access_description()} />

      {loaderData?.authMode && (
        <>
          <div className="space-y-4">
            <SettingsSectionHeader
              title={m.access_local_title()}
              description={m.access_local_description()}
            />
            <>
              <SettingsItem
                title={m.access_https_mode_title()}
                badge={m.experimental()}
                description={m.access_https_description()}
              >
                <SelectMenuBasic
                  size="SM"
                  value={tlsMode}
                  onChange={e => handleTlsModeChange(e.target.value)}
                  disabled={tlsMode === "unknown"}
                  options={[
                    { value: "disabled", label: m.access_tls_disabled() },
                    { value: "self-signed", label: m.access_tls_self_signed() },
                    { value: "custom", label: m.access_tls_custom() },
                  ]}
                />
              </SettingsItem>

              {tlsMode === "custom" && (
                <NestedSettingsGroup className="mt-4">
                  <SettingsItem
                    title={m.access_tls_certificate_title()}
                    description={m.access_tls_certificate_description()}
                  />
                  <TextAreaWithLabel
                    label={m.access_certificate_label()}
                    rows={3}
                    placeholder={"-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----"}
                    value={tlsCert}
                    onChange={e => handleTlsCertChange(e.target.value)}
                  />
                  <TextAreaWithLabel
                    label={m.access_private_key_label()}
                    description={m.access_private_key_description()}
                    rows={3}
                    placeholder={"-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----"}
                    value={tlsKey}
                    onChange={e => handleTlsKeyChange(e.target.value)}
                  />
                  <div className="flex items-center gap-x-2">
                    <Button
                      size="SM"
                      theme="primary"
                      text={m.access_update_tls_settings()}
                      onClick={handleCustomTlsUpdate}
                    />
                  </div>
                </NestedSettingsGroup>
              )}

              <SettingsItem
                title={m.access_authentication_mode_title()}
                description={
                  loaderData.authMode === "password"
                    ? m.access_auth_mode_password()
                    : m.access_auth_mode_no_password()
                }
              >
                {loaderData.authMode === "password" ? (
                  <Button
                    size="SM"
                    theme="light"
                    text={m.access_disable_protection()}
                    onClick={() => {
                      navigateTo("./local-auth", { state: { init: "deletePassword" } });
                    }}
                  />
                ) : (
                  <Button
                    size="SM"
                    theme="light"
                    text={m.access_enable_password()}
                    onClick={() => {
                      navigateTo("./local-auth", { state: { init: "createPassword" } });
                    }}
                  />
                )}
              </SettingsItem>
            </>

            {loaderData.authMode === "password" && (
              <SettingsItem
                title={m.access_change_password_title()}
                description={m.access_change_password_description()}
              >
                <Button
                  size="SM"
                  theme="light"
                  text={m.access_change_password_button()}
                  onClick={() => {
                    navigateTo("./local-auth", { state: { init: "updatePassword" } });
                  }}
                />
              </SettingsItem>
            )}
          </div>
        </>
      )}
    </div>
  );
}

SettingsAccessIndexRoute.loader = loader;
