import { LuRefreshCcw } from "react-icons/lu";
import { useCallback, useEffect, useState } from "react";

import { Button } from "@components/Button";
import { GridCard } from "@components/Card";
import { PublicIP } from "@hooks/stores";
import { m } from "@localizations/messages.js";
import { JsonRpcResponse, useJsonRpc } from "@hooks/useJsonRpc";
import notifications from "@/notifications";
import { formatters } from "@/utils";

const TimeAgoLabel = ({ date }: { date: Date }) => {
  const [timeAgo, setTimeAgo] = useState<string | undefined>(formatters.timeAgo(date));
  useEffect(() => {
    const interval = setInterval(() => {
      setTimeAgo(formatters.timeAgo(date));
    }, 1000);
    return () => clearInterval(interval);
  }, [date]);

  return <span className="text-sm text-slate-600 select-none dark:text-slate-400">{timeAgo}</span>;
};

const publicIPRows: PublicIP["family"][] = ["ipv4", "ipv6"];

export default function PublicIPCard() {
  const { send } = useJsonRpc();

  const [isLoading, setIsLoading] = useState(true);
  const [publicIPs, setPublicIPs] = useState<PublicIP[]>([]);
  const refreshPublicIPs = useCallback(
    (refresh = false) => {
      send("getPublicIPAddresses", { refresh }, (resp: JsonRpcResponse) => {
        setIsLoading(false);
        setPublicIPs([]);
        if ("error" in resp) {
          notifications.error(
            m.public_ip_card_refresh_error({ error: resp.error.data || m.unknown_error() }),
          );
          return;
        }
        const publicIPs = resp.result as PublicIP[];
        setPublicIPs(publicIPs);
      });
    },
    [send, setPublicIPs],
  );

  useEffect(() => {
    refreshPublicIPs();
  }, [refreshPublicIPs]);

  return (
    <GridCard>
      <div className="animate-fadeIn p-4 text-black opacity-0 animation-duration-500 dark:text-white">
        <div className="space-y-3">
          <div className="flex items-center justify-between">
            <h3 className="text-base font-bold text-slate-900 dark:text-white">
              {m.public_ip_card_header()}
            </h3>

            <div>
              <Button
                size="XS"
                theme="light"
                type="button"
                className="text-red-500"
                text={m.public_ip_card_refresh()}
                LeadingIcon={LuRefreshCcw}
                onClick={() => {
                  setIsLoading(true);
                  refreshPublicIPs(true);
                }}
              />
            </div>
          </div>
          {isLoading ? (
            <div>
              <div className="space-y-4">
                <div className="animate-pulse space-y-2">
                  <div className="h-4 w-1/4 rounded bg-slate-200 dark:bg-slate-700" />
                  <div className="h-4 w-1/3 rounded bg-slate-200 dark:bg-slate-700" />
                  <div className="h-4 w-1/2 rounded bg-slate-200 dark:bg-slate-700" />
                </div>
              </div>
            </div>
          ) : (
            <div className="flex flex-col gap-y-2">
              <div className="flex-1 space-y-2">
                {publicIPRows.map(family => {
                  const publicIP = publicIPs.find(ip => ip.family === family);
                  const label = family === "ipv4" ? "IPv4" : "IPv6";
                  const value = publicIP?.ip
                    ? publicIP.ip
                    : `no public IP${publicIP?.error ? ` (${publicIP.error})` : ""}`;

                  return (
                    <div
                      key={family}
                      className="flex justify-between border-slate-800/10 pt-2 dark:border-slate-300/20"
                    >
                      <span className="text-sm font-medium">
                        <span className="text-slate-600 dark:text-slate-400">{label}: </span>
                        <span className="font-mono">{value}</span>
                      </span>
                      {publicIP?.last_updated && (
                        <TimeAgoLabel date={new Date(publicIP.last_updated)} />
                      )}
                    </div>
                  );
                })}
              </div>
            </div>
          )}
        </div>
      </div>
    </GridCard>
  );
}
