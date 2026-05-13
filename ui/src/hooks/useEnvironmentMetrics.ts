import { useCallback, useEffect, useState } from "react";
import { useInterval } from "usehooks-ts";

import { JsonRpcResponse, RpcMethodNotFound, useJsonRpc } from "@hooks/useJsonRpc";

export interface EnvironmentMetrics {
  caseTemperatureC: number | null;
  caseHumidityPercent: number | null;
  socTemperatureC: number | null;
  updatedAt: number;
}

export function useEnvironmentMetrics() {
  const { send } = useJsonRpc();
  const [metrics, setMetrics] = useState<EnvironmentMetrics | null>(null);
  const [supported, setSupported] = useState(true);
  const [now, setNow] = useState(0);

  const fetchMetrics = useCallback(() => {
    if (!supported) return;

    send("getEnvironmentMetrics", {}, (resp: JsonRpcResponse) => {
      if ("error" in resp) {
        if (resp.error.code === RpcMethodNotFound) setSupported(false);
        return;
      }
      setMetrics(resp.result as EnvironmentMetrics);
      setNow(Date.now());
    });
  }, [send, supported]);

  useEffect(fetchMetrics, [fetchMetrics]);
  useInterval(fetchMetrics, supported ? 2000 : null);
  useInterval(() => setNow(Date.now()), metrics ? 2000 : null);

  const stale = metrics ? now - metrics.updatedAt > 15000 : false;
  const available =
    supported &&
    !!metrics &&
    (metrics.caseTemperatureC != null ||
      metrics.caseHumidityPercent != null ||
      metrics.socTemperatureC != null);

  return { metrics, available, stale };
}
