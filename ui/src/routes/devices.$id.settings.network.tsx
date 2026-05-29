import { useCallback, useEffect, useState } from "react";
import { LuCopy, LuEthernetPort, LuInfo } from "react-icons/lu";
import dayjs from "dayjs";
import relativeTime from "dayjs/plugin/relativeTime";

import PublicIPCard from "@components/PublicIPCard";
import { NetworkState, useNetworkStateStore, useRTCStore } from "@hooks/stores";
import AutoHeight from "@components/AutoHeight";
import { Button } from "@components/Button";
import DhcpLeaseCard from "@components/DhcpLeaseCard";
import EmptyCard from "@components/EmptyCard";
import { GridCard } from "@components/Card";
import Ipv6NetworkCard from "@components/Ipv6NetworkCard";
import { SettingsItem } from "@components/SettingsItem";
import { SettingsPageHeader } from "@/components/SettingsPageheader";
import { useCopyToClipboard } from "@components/useCopyToClipBoard";
import { getNetworkState } from "@/utils/jsonrpc";
import notifications from "@/notifications";
import { m } from "@localizations/messages";

dayjs.extend(relativeTime);

const resolveOnRtcReady = () => {
  return new Promise(resolve => {
    // Check if RTC is already connected
    const currentState = useRTCStore.getState();
    if (currentState.rpcDataChannel?.readyState === "open") {
      // Already connected, fetch data immediately
      return resolve(void 0);
    }

    // Not connected yet, subscribe to state changes
    const unsubscribe = useRTCStore.subscribe(state => {
      if (state.rpcDataChannel?.readyState === "open") {
        unsubscribe(); // Clean up subscription
        return resolve(void 0);
      }
    });
  });
};

export function LifeTimeLabel({ lifetime }: Readonly<{ lifetime: string }>) {
  const [remaining, setRemaining] = useState<string | null>(null);

  // rrecalculate remaining time every 30 seconds
  useEffect(() => {
    // schedule immediate initial update
    setInterval(() => setRemaining(dayjs(lifetime).fromNow()), 0);

    const interval = setInterval(() => {
      setRemaining(dayjs(lifetime).fromNow());
    }, 1000 * 30);
    return () => clearInterval(interval);
  }, [lifetime]);

  if (lifetime == "") {
    return <strong>{m.not_applicable()}</strong>;
  }

  return (
    <>
      <span className="text-sm font-medium">{remaining && <> {remaining}</>}</span>
      <span className="text-xs text-slate-700 dark:text-slate-300">
        &nbsp;({dayjs(lifetime).format("YYYY-MM-DD HH:mm")})
      </span>
    </>
  );
}

export default function SettingsNetworkRoute() {
  const networkState = useNetworkStateStore(state => state);
  const setNetworkState = useNetworkStateStore(state => state.setNetworkState);
  const [isLoading, setIsLoading] = useState(true);

  const fetchNetworkData = useCallback(async () => {
    try {
      setIsLoading(true);
      console.log("Fetching network state...");

      const state = (await getNetworkState()) as NetworkState;

      setNetworkState(state);
      setIsLoading(false);
    } catch (err) {
      setIsLoading(false);
      notifications.error(
        m.network_settings_load_error({
          error: err instanceof Error ? err.message : m.unknown_error(),
        }),
      );
      throw err;
    }
  }, [setNetworkState]);

  // Fetch data on component mount
  useEffect(() => {
    const loadData = async () => {
      // Ensure data channel is ready, before fetching network data from the device
      await resolveOnRtcReady();
      await fetchNetworkData();
    };
    loadData();
  }, [fetchNetworkData]);

  const { copy } = useCopyToClipboard();

  return (
    <div className="space-y-4">
      <SettingsPageHeader
        title={m.network_title()}
        description={m.network_description()}
        action={
          <div className="flex items-center gap-2 text-sm text-slate-600 dark:text-slate-400">
            <LuInfo className="h-4 w-4" />
            <span>{m.network_os_tools_note()}</span>
          </div>
        }
      />

      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <SettingsItem
            title={m.network_mac_address_title()}
            description={m.network_mac_address_description()}
          />
          <div className="flex items-center">
            <GridCard cardClassName="rounded-r-none">
              <div className="flex h-[34px] items-center px-3 font-mono text-xs text-black select-all dark:text-white">
                {networkState?.mac_address}{" "}
              </div>
            </GridCard>
            <Button
              className="rounded-l-none border-l-slate-800/30 dark:border-slate-300/20"
              size="SM"
              type="button"
              theme="light"
              LeadingIcon={LuCopy}
              onClick={async () => {
                const mac = networkState?.mac_address || "";
                if (await copy(mac)) {
                  notifications.success(m.network_mac_address_copy_success({ mac: mac }));
                } else {
                  notifications.error(m.network_mac_address_copy_error());
                }
              }}
            />
          </div>
        </div>

        <div className="space-y-4">
          {/* Read-only network information display */}
          {isLoading ? (
            <GridCard>
              <div className="p-4">
                <div className="space-y-4">
                  <div className="h-6 w-1/3 animate-pulse rounded bg-slate-200 dark:bg-slate-700" />
                  <div className="animate-pulse space-y-2">
                    <div className="h-4 w-1/4 rounded bg-slate-200 dark:bg-slate-700" />
                    <div className="h-4 w-1/2 rounded bg-slate-200 dark:bg-slate-700" />
                    <div className="h-4 w-1/3 rounded bg-slate-200 dark:bg-slate-700" />
                  </div>
                </div>
              </div>
            </GridCard>
          ) : (
            <>
              {/* Network State Information */}
              <GridCard>
                <div className="space-y-4 p-4">
                  <h3 className="text-base font-bold text-slate-900 dark:text-white">
                    {m.network_status_title()}
                  </h3>
                  <div className="grid grid-cols-1 gap-4 text-sm md:grid-cols-2">
                    <div>
                      <span className="text-slate-600 dark:text-slate-400">
                        {m.network_interface_label()}:
                      </span>
                      <span className="ml-2 font-mono">
                        {networkState?.interface_name || m.not_available()}
                      </span>
                    </div>
                    <div>
                      <span className="text-slate-600 dark:text-slate-400">
                        {m.network_hostname_title()}:
                      </span>
                      <span className="ml-2 font-mono">
                        {networkState?.hostname || m.not_available()}
                      </span>
                    </div>
                    <div>
                      <span className="text-slate-600 dark:text-slate-400">
                        {m.network_status_label()}:
                      </span>
                      <span className="ml-2 font-mono">
                        {networkState?.online ? (
                          <span className="text-green-600 dark:text-green-400">{m.online()}</span>
                        ) : networkState?.up ? (
                          <span className="text-yellow-600 dark:text-yellow-400">
                            {m.network_status_up_no_internet()}
                          </span>
                        ) : (
                          <span className="text-red-600 dark:text-red-400">
                            {m.network_status_down()}
                          </span>
                        )}
                      </span>
                    </div>
                    <div>
                      <span className="text-slate-600 dark:text-slate-400">IPv4:</span>
                      <span className="ml-2 font-mono">
                        {networkState?.ipv4_address || m.network_not_configured()}
                      </span>
                    </div>
                  </div>
                  {/* IPv6 addresses on separate lines to prevent overflow */}
                  <div className="space-y-2 text-sm">
                    {networkState?.ipv6_address && (
                      <div>
                        <span className="text-slate-600 dark:text-slate-400">IPv6:</span>
                        <div className="ml-2 font-mono text-xs break-all">
                          {networkState.ipv6_address}
                        </div>
                      </div>
                    )}
                    {!networkState?.ipv6_address && (
                      <div>
                        <span className="text-slate-600 dark:text-slate-400">IPv6:</span>
                        <span className="ml-2 font-mono">{m.network_not_configured()}</span>
                      </div>
                    )}
                    {networkState?.ipv6_link_local && (
                      <div>
                        <span className="text-slate-600 dark:text-slate-400">
                          {m.ipv6_link_local()}:
                        </span>
                        <div className="ml-2 font-mono text-xs break-all">
                          {networkState.ipv6_link_local}
                        </div>
                      </div>
                    )}
                  </div>
                </div>
              </GridCard>

              <PublicIPCard />

              {/* IPv4 Information */}
              <div>
                <AutoHeight>
                  {networkState?.dhcp_lease ? (
                    <DhcpLeaseCard
                      networkState={networkState}
                      setShowRenewLeaseConfirm={() => {
                        // DHCP lease renewal is disabled - show notification
                        notifications.error(m.network_dhcp_lease_renew_disabled());
                      }}
                    />
                  ) : networkState?.ipv4_address ? (
                    <GridCard>
                      <div className="space-y-4 p-4">
                        <h3 className="text-base font-bold text-slate-900 dark:text-white">
                          {m.network_ipv4_configuration()}
                        </h3>
                        <div className="space-y-2 text-sm">
                          <div>
                            <span className="text-slate-600 dark:text-slate-400">
                              {m.ip_address()}:
                            </span>
                            <span className="ml-2 font-mono">{networkState.ipv4_address}</span>
                          </div>
                          {networkState.ipv4_addresses &&
                            networkState.ipv4_addresses.length > 1 && (
                              <div>
                                <span className="text-slate-600 dark:text-slate-400">
                                  {m.network_all_addresses()}:
                                </span>
                                <div className="ml-2 space-y-1 font-mono text-xs">
                                  {networkState.ipv4_addresses.map((addr, idx) => (
                                    <div key={idx}>{addr}</div>
                                  ))}
                                </div>
                              </div>
                            )}
                        </div>
                      </div>
                    </GridCard>
                  ) : (
                    <EmptyCard
                      IconElm={LuEthernetPort}
                      headline={m.network_no_information_headline()}
                      description={m.network_no_information_description()}
                    />
                  )}
                </AutoHeight>
              </div>

              {/* IPv6 Information */}
              <div className="space-y-4">
                <AutoHeight>
                  <Ipv6NetworkCard networkState={networkState || undefined} />
                </AutoHeight>
              </div>

              {/* Information Notice */}
              <GridCard>
                <div className="p-4">
                  <div className="flex items-start gap-3">
                    <LuInfo className="mt-0.5 h-5 w-5 shrink-0 text-blue-500" />
                    <div className="space-y-2">
                      <h4 className="font-medium text-slate-900 dark:text-white">
                        {m.network_configuration_read_only_title()}
                      </h4>
                      <p className="text-sm text-slate-600 dark:text-slate-400">
                        {m.network_configuration_read_only_description()}
                      </p>
                      <ul className="ml-2 list-inside list-disc space-y-1 text-sm text-slate-600 dark:text-slate-400">
                        <li>NetworkManager (nmcli, nmtui)</li>
                        <li>{m.network_systemd_networkd()}</li>
                        <li>{m.network_manual_config_files()}</li>
                        <li>{m.network_distribution_tools()}</li>
                      </ul>
                    </div>
                  </div>
                </div>
              </GridCard>
            </>
          )}
        </div>
      </div>
    </div>
  );
}
