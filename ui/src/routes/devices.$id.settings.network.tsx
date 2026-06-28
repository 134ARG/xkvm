import { useCallback, useEffect, useState, type ReactNode } from "react";
import { LuCopy, LuInfo } from "react-icons/lu";
import dayjs from "dayjs";
import relativeTime from "dayjs/plugin/relativeTime";

import PublicIPCard from "@components/PublicIPCard";
import { NetworkState, useNetworkStateStore, useRTCStore } from "@hooks/stores";
import { Button } from "@components/Button";
import { GridCard } from "@components/Card";
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

export function FlagLabel({ flag }: { flag: string }) {
  return (
    <span className="ml-2 rounded-sm bg-red-500 px-2 py-1 text-[10px] leading-none font-medium text-white dark:border dark:border-red-700 dark:bg-red-800 dark:text-red-50">
      {flag}
    </span>
  );
}

interface Field {
  label: string;
  value: ReactNode;
}

// Shared card + row primitives — one consistent label-left / value-right style with
// separators, used by every network section so IPv4 and IPv6 look identical.
function InfoCard({
  title,
  action,
  children,
}: {
  title: string;
  action?: ReactNode;
  children: ReactNode;
}) {
  return (
    <GridCard>
      <div className="space-y-3 p-4">
        <div className="flex items-center justify-between">
          <h3 className="text-base font-bold text-slate-900 dark:text-white">{title}</h3>
          {action}
        </div>
        {children}
      </div>
    </GridCard>
  );
}

function FieldRow({ label, value, divider = true }: Field & { divider?: boolean }) {
  return (
    <div
      className={`flex justify-between gap-x-4 ${
        divider ? "border-t border-slate-800/10 pt-2 first:border-t-0 dark:border-slate-300/20" : ""
      }`}
    >
      <span className="text-sm text-slate-600 dark:text-slate-400">{label}</span>
      <span className="text-right text-sm font-medium break-all">{value}</span>
    </div>
  );
}

// Renders fields in two balanced columns, matching the DHCP-lease card layout.
function FieldColumns({ fields }: { fields: Field[] }) {
  const half = Math.ceil(fields.length / 2);
  const columns = [fields.slice(0, half), fields.slice(half)];
  return (
    <div className="grid grid-cols-1 gap-x-8 md:grid-cols-2">
      {columns.map((column, i) => (
        <div key={i} className="space-y-2">
          {column.map(f => (
            <FieldRow key={f.label} label={f.label} value={f.value} />
          ))}
        </div>
      ))}
    </div>
  );
}

function ValueList({ items }: { items: string[] }) {
  return (
    <div className="space-y-1">
      {items.map(item => (
        <div key={item}>{item}</div>
      ))}
    </div>
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

  const copyMac = async () => {
    const mac = networkState?.mac_address || "";
    if (await copy(mac)) {
      notifications.success(m.network_mac_address_copy_success({ mac }));
    } else {
      notifications.error(m.network_mac_address_copy_error());
    }
  };

  // ----- Build field sets -------------------------------------------------
  const summaryFields: Field[] = [
    {
      label: m.network_interface_label(),
      value: networkState?.interface_name || m.not_available(),
    },
    { label: m.network_hostname_title(), value: networkState?.hostname || m.not_available() },
    {
      label: m.network_status_label(),
      value: networkState?.online ? (
        <span className="text-green-600 dark:text-green-400">{m.online()}</span>
      ) : networkState?.up ? (
        <span className="text-yellow-600 dark:text-yellow-400">
          {m.network_status_up_no_internet()}
        </span>
      ) : (
        <span className="text-red-600 dark:text-red-400">{m.network_status_down()}</span>
      ),
    },
    {
      label: m.network_mac_address_title(),
      value: (
        <span className="inline-flex items-center gap-2">
          {networkState?.mac_address || m.not_available()}
          {networkState?.mac_address && (
            <Button size="XS" type="button" theme="light" LeadingIcon={LuCopy} onClick={copyMac} />
          )}
        </span>
      ),
    },
  ];

  const lease = networkState?.dhcp_lease;
  const primaryIpv4 = lease?.ip || networkState?.ipv4_address;
  // Exclude the primary IPv4 even when it appears in CIDR form (e.g. "192.168.1.111/24").
  const additionalIpv4 = (networkState?.ipv4_addresses || []).filter(
    a => a.split("/")[0] !== primaryIpv4,
  );

  const ipv4Fields: Field[] = [];
  if (primaryIpv4) ipv4Fields.push({ label: m.ip_address(), value: primaryIpv4 });
  if (lease?.netmask) ipv4Fields.push({ label: m.subnet_mask(), value: lease.netmask });
  if (lease?.routers?.length)
    ipv4Fields.push({ label: m.dhcp_lease_gateway(), value: <ValueList items={lease.routers} /> });
  if (lease?.server_id) ipv4Fields.push({ label: m.dhcp_server(), value: lease.server_id });
  if (lease?.dns_servers?.length)
    ipv4Fields.push({ label: m.dns_servers(), value: <ValueList items={lease.dns_servers} /> });
  if (lease?.lease_expiry)
    ipv4Fields.push({
      label: m.dhcp_lease_lease_expires(),
      value: <LifeTimeLabel lifetime={`${lease.lease_expiry}`} />,
    });
  if (lease?.dhcp_client)
    ipv4Fields.push({ label: m.network_dhcp_client_title(), value: lease.dhcp_client });
  if (additionalIpv4.length)
    ipv4Fields.push({
      label: m.network_additional_addresses(),
      value: <ValueList items={additionalIpv4} />,
    });

  const ipv6Fields: Field[] = [
    {
      label: m.ipv6_link_local(),
      value: networkState?.ipv6_link_local || m.network_not_configured(),
    },
    { label: m.ipv6_gateway(), value: networkState?.ipv6_gateway || m.network_not_configured() },
  ];
  const ipv6Addresses = networkState?.ipv6_addresses ?? [];

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
        <div className="space-y-4">
          {/* Summary — identity and status only, no addresses */}
          <InfoCard title={m.network_status_title()}>
            <FieldColumns fields={summaryFields} />
          </InfoCard>

          <PublicIPCard />

          {/* IPv4 — DHCP-lease fields unpacked, no nested card */}
          <InfoCard
            title={m.network_ipv4_information()}
            action={
              <span className="text-xs text-slate-500 dark:text-slate-400">{m.read_only()}</span>
            }
          >
            {ipv4Fields.length > 0 ? (
              <FieldColumns fields={ipv4Fields} />
            ) : (
              <p className="text-sm text-slate-600 dark:text-slate-400">
                {m.network_not_configured()}
              </p>
            )}
          </InfoCard>

          {/* IPv6 — same style; one nested sub-card per address */}
          <InfoCard title={m.ipv6_information()}>
            <FieldColumns fields={ipv6Fields} />

            {ipv6Addresses.length > 0 && (
              <div className="space-y-3 pt-2">
                <h4 className="text-sm font-semibold">{m.network_ipv6_addresses_header()}</h4>
                {ipv6Addresses.map(addr => (
                  <div
                    key={addr.address}
                    className="space-y-2 rounded-md border border-slate-500/10 border-l-blue-700/50 bg-white p-4 dark:bg-transparent"
                  >
                    <FieldRow
                      label={m.ipv6_address_label()}
                      value={
                        <span className="inline-flex items-center font-mono text-xs">
                          <span>{addr.address}</span>
                          {addr.flag_deprecated && (
                            <FlagLabel flag={m.network_ipv6_flag_deprecated()} />
                          )}
                          {addr.flag_dad_failed && (
                            <FlagLabel flag={m.network_ipv6_flag_dad_failed()} />
                          )}
                        </span>
                      }
                    />
                    {addr.prefix && (
                      <FieldRow
                        label={m.network_ipv6_prefix()}
                        value={<span className="font-mono text-xs">{addr.prefix}</span>}
                      />
                    )}
                    {(addr.valid_lifetime != null || addr.preferred_lifetime != null) && (
                      <div className="grid grid-cols-1 gap-x-8 border-t border-slate-800/10 pt-2 md:grid-cols-2 dark:border-slate-300/20">
                        {addr.valid_lifetime != null && (
                          <FieldRow
                            divider={false}
                            label={m.ipv6_valid_lifetime()}
                            value={
                              <LifeTimeLabel
                                lifetime={dayjs().add(addr.valid_lifetime, "second").toISOString()}
                              />
                            }
                          />
                        )}
                        {addr.preferred_lifetime != null && (
                          <FieldRow
                            divider={false}
                            label={m.ipv6_preferred_lifetime()}
                            value={
                              <LifeTimeLabel
                                lifetime={dayjs()
                                  .add(addr.preferred_lifetime, "second")
                                  .toISOString()}
                              />
                            }
                          />
                        )}
                      </div>
                    )}
                  </div>
                ))}
              </div>
            )}
          </InfoCard>

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
        </div>
      )}
    </div>
  );
}
