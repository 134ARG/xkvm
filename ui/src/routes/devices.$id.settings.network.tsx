import { useCallback, useEffect, useState } from "react";
import { LuCopy, LuEthernetPort, LuInfo } from "react-icons/lu";
import dayjs from "dayjs";
import relativeTime from "dayjs/plugin/relativeTime";

import PublicIPCard from "@components/PublicIPCard";
import { NetworkSettings, NetworkState, useNetworkStateStore, useRTCStore } from "@hooks/stores";
import AutoHeight from "@components/AutoHeight";
import { Button } from "@components/Button";
import DhcpLeaseCard from "@components/DhcpLeaseCard";
import EmptyCard from "@components/EmptyCard";
import { GridCard } from "@components/Card";
import Ipv6NetworkCard from "@components/Ipv6NetworkCard";
import { SettingsItem } from "@components/SettingsItem";
import { SettingsPageHeader } from "@/components/SettingsPageheader";
import StaticIpv4Card from "@components/StaticIpv4Card";
import StaticIpv6Card from "@components/StaticIpv6Card";
import { useCopyToClipboard } from "@components/useCopyToClipBoard";
import { getNetworkSettings, getNetworkState } from "@/utils/jsonrpc";
import notifications from "@/notifications";
import { m } from "@localizations/messages";

dayjs.extend(relativeTime);

const isLLDPAvailable = false; // LLDP is not supported yet

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
  const [networkSettings, setNetworkSettings] = useState<NetworkSettings | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  const fetchNetworkData = useCallback(async () => {
    try {
      setIsLoading(true);
      console.log("Fetching network data...");

      const [settings, state] = (await Promise.all([getNetworkSettings(), getNetworkState()])) as [
        NetworkSettings,
        NetworkState,
      ];

      setNetworkState(state);
      setNetworkSettings(settings);
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
            <span>Read-only - Use OS network tools to configure</span>
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
              {/* Network Configuration Information */}
              <GridCard>
                <div className="p-4 space-y-4">
                  <h3 className="text-base font-bold text-slate-900 dark:text-white">
                    Network Configuration
                  </h3>
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm">
                    <div>
                      <span className="text-slate-600 dark:text-slate-400">Hostname:</span>
                      <span className="ml-2 font-mono">{networkSettings?.hostname || networkState?.hostname || "N/A"}</span>
                    </div>
                    <div>
                      <span className="text-slate-600 dark:text-slate-400">Domain:</span>
                      <span className="ml-2 font-mono">{networkSettings?.domain || "N/A"}</span>
                    </div>
                    <div>
                      <span className="text-slate-600 dark:text-slate-400">DHCP Client:</span>
                      <span className="ml-2 font-mono">{networkSettings?.dhcp_client || "N/A"}</span>
                    </div>
                    <div>
                      <span className="text-slate-600 dark:text-slate-400">IPv4 Mode:</span>
                      <span className="ml-2 font-mono">{networkSettings?.ipv4_mode || "N/A"}</span>
                    </div>
                    <div>
                      <span className="text-slate-600 dark:text-slate-400">IPv6 Mode:</span>
                      <span className="ml-2 font-mono">{networkSettings?.ipv6_mode || "N/A"}</span>
                    </div>
                    <div>
                      <span className="text-slate-600 dark:text-slate-400">mDNS Mode:</span>
                      <span className="ml-2 font-mono">{networkSettings?.mdns_mode || "N/A"}</span>
                    </div>
                    {networkSettings?.http_proxy && (
                      <div className="md:col-span-2">
                        <span className="text-slate-600 dark:text-slate-400">HTTP Proxy:</span>
                        <span className="ml-2 font-mono">{networkSettings.http_proxy}</span>
                      </div>
                    )}
                  </div>
                </div>
              </GridCard>

              <PublicIPCard />

              {/* IPv4 Information */}
              <div>
                <AutoHeight>
                  {networkSettings?.ipv4_mode === "static" ? (
                    <StaticIpv4Card ipv4Static={networkSettings.ipv4_static} />
                  ) : networkSettings?.ipv4_mode === "dhcp" ? (
                    <DhcpLeaseCard
                      networkState={networkState}
                      setShowRenewLeaseConfirm={() => {
                        // DHCP lease renewal is disabled - show notification
                        notifications.info("DHCP lease renewal is disabled. Use OS network management tools.");
                      }}
                    />
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
                  {networkSettings?.ipv6_mode === "static" ? (
                    <StaticIpv6Card ipv6Static={networkSettings.ipv6_static} />
                  ) : (
                    <Ipv6NetworkCard networkState={networkState || undefined} />
                  )}
                </AutoHeight>
              </div>

              {/* Information Notice */}
              <GridCard>
                <div className="p-4">
                  <div className="flex items-start gap-3">
                    <LuInfo className="h-5 w-5 text-blue-500 mt-0.5 flex-shrink-0" />
                    <div className="space-y-2">
                      <h4 className="font-medium text-slate-900 dark:text-white">
                        Network Configuration is Read-Only
                      </h4>
                      <p className="text-sm text-slate-600 dark:text-slate-400">
                        Network settings cannot be modified through this interface. 
                        Use your operating system's network management tools such as:
                      </p>
                      <ul className="text-sm text-slate-600 dark:text-slate-400 list-disc list-inside space-y-1 ml-2">
                        <li>NetworkManager (nmcli, nmtui)</li>
                        <li>systemd-networkd</li>
                        <li>Manual configuration files (/etc/network/interfaces, /etc/netplan/)</li>
                        <li>Your distribution's network configuration tools</li>
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
