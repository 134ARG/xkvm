// TODO: Remove this component - it's orphaned and no longer used.
// Replaced by inline implementation in devices.$id.settings.network.tsx

import { GridCard } from "@components/Card";
import { NetworkSettings } from "@hooks/stores";
import { m } from "@localizations/messages.js";

interface StaticIpv4CardProps {
  ipv4Static?: NetworkSettings["ipv4_static"];
}

export default function StaticIpv4Card({ ipv4Static }: StaticIpv4CardProps) {
  if (!ipv4Static) {
    return (
      <GridCard>
        <div className="p-4">
          <h3 className="text-base font-bold text-slate-900 dark:text-white">
            {m.network_static_ipv4_header()}
          </h3>
          <p className="mt-2 text-sm text-slate-600 dark:text-slate-400">
            No static IPv4 configuration available
          </p>
        </div>
      </GridCard>
    );
  }

  return (
    <GridCard>
      <div className="animate-fadeIn p-4 text-black opacity-0 animation-duration-500 dark:text-white">
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <h3 className="text-base font-bold text-slate-900 dark:text-white">
              {m.network_static_ipv4_header()}
            </h3>
            <div className="text-xs text-slate-500 dark:text-slate-400">Read-only</div>
          </div>

          <div className="grid grid-cols-1 gap-4 text-sm md:grid-cols-2">
            {ipv4Static.address && (
              <div>
                <span className="text-slate-600 dark:text-slate-400">
                  {m.network_ipv4_address()}:
                </span>
                <span className="ml-2 font-mono">{ipv4Static.address}</span>
              </div>
            )}

            {ipv4Static.netmask && (
              <div>
                <span className="text-slate-600 dark:text-slate-400">
                  {m.network_ipv4_netmask()}:
                </span>
                <span className="ml-2 font-mono">{ipv4Static.netmask}</span>
              </div>
            )}

            {ipv4Static.gateway && (
              <div className="md:col-span-2">
                <span className="text-slate-600 dark:text-slate-400">
                  {m.network_ipv4_gateway()}:
                </span>
                <span className="ml-2 font-mono">{ipv4Static.gateway}</span>
              </div>
            )}

            {ipv4Static.dns && ipv4Static.dns.length > 0 && (
              <div className="md:col-span-2">
                <span className="text-slate-600 dark:text-slate-400">{m.network_ipv4_dns()}:</span>
                <div className="ml-2 font-mono">
                  {ipv4Static.dns
                    .filter(dns => dns && dns.trim() !== "")
                    .map((dns, index) => (
                      <div key={index}>{dns}</div>
                    ))}
                </div>
              </div>
            )}
          </div>
        </div>
      </div>
    </GridCard>
  );
}
