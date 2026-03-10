// TODO: Remove this component - it's orphaned and no longer used.
// Replaced by inline implementation in devices.$id.settings.network.tsx

import { GridCard } from "@components/Card";
import { NetworkSettings } from "@hooks/stores";
import { m } from "@localizations/messages.js";

interface StaticIpv6CardProps {
  ipv6Static?: NetworkSettings["ipv6_static"];
}

export default function StaticIpv6Card({ ipv6Static }: StaticIpv6CardProps) {
  if (!ipv6Static) {
    return (
      <GridCard>
        <div className="p-4">
          <h3 className="text-base font-bold text-slate-900 dark:text-white">
            {m.network_static_ipv6_header()}
          </h3>
          <p className="mt-2 text-sm text-slate-600 dark:text-slate-400">
            No static IPv6 configuration available
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
              {m.network_static_ipv6_header()}
            </h3>
            <div className="text-xs text-slate-500 dark:text-slate-400">Read-only</div>
          </div>

          <div className="space-y-3 text-sm">
            {ipv6Static.prefix && (
              <div>
                <span className="text-slate-600 dark:text-slate-400">
                  {m.network_ipv6_prefix()}:
                </span>
                <span className="ml-2 font-mono">{ipv6Static.prefix}</span>
              </div>
            )}

            {ipv6Static.gateway && (
              <div>
                <span className="text-slate-600 dark:text-slate-400">
                  {m.network_ipv6_gateway()}:
                </span>
                <span className="ml-2 font-mono">{ipv6Static.gateway}</span>
              </div>
            )}

            {ipv6Static.dns && ipv6Static.dns.length > 0 && (
              <div>
                <span className="text-slate-600 dark:text-slate-400">{m.network_ipv6_dns()}:</span>
                <div className="ml-2 font-mono">
                  {ipv6Static.dns
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
