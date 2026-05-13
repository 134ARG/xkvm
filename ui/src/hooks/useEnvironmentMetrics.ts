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
  const [metrics, setMetrics] = useState<EnvironmentMetrics>({
    caseTemperatureC: null,
    caseHumidityPercent: null,
    socTemperatureC: null,
    updatedAt: 0,
  });
  const [supported, setSupported] = useState(true);
  const [now, setNow] = useState(0);

  const fetchMetrics = useCallback(() => {
    if (!supported) return;

    send("getEnvironmentMetrics", {}, (resp: JsonRpcResponse) => {
      if ("error" in resp) {
        if (resp.error.code === RpcMethodNotFound) setSupported(false);
        return;
      }
      setMetrics({
        caseTemperatureC: null,
        caseHumidityPercent: null,
        socTemperatureC: null,
        updatedAt: 0,
        ...(resp.result as Partial<EnvironmentMetrics>),
      });
      setNow(Date.now());
    });
  }, [send, supported]);

  useEffect(fetchMetrics, [fetchMetrics]);
  useInterval(fetchMetrics, supported ? 2000 : null);
  useInterval(() => setNow(Date.now()), 2000);

  const stale = metrics.updatedAt > 0 ? now - metrics.updatedAt > 15000 : false;

  return { metrics, supported, stale };
}
