import { useCallback, useEffect, useState } from "react";
import { useInterval } from "usehooks-ts";

import { JsonRpcResponse, RpcMethodNotFound, useJsonRpc } from "@hooks/useJsonRpc";

export interface VFDHostMetrics {
  cpuUtil: number | null;
  ramUtil: number | null;
  gpuUtil: number | null;
  cpuTemp: number | null;
  gpuTemp: number | null;
  netRxBytes: number | null;
  netTxBytes: number | null;
  uptimeSec: number | null;
  failedUnits: number | null;
  connected: boolean;
  updatedAt: number;
}

const emptyMetrics: VFDHostMetrics = {
  cpuUtil: null,
  ramUtil: null,
  gpuUtil: null,
  cpuTemp: null,
  gpuTemp: null,
  netRxBytes: null,
  netTxBytes: null,
  uptimeSec: null,
  failedUnits: null,
  connected: false,
  updatedAt: 0,
};

export function useVFDHostMetrics() {
  const { send } = useJsonRpc();
  const [metrics, setMetrics] = useState<VFDHostMetrics>(emptyMetrics);
  const [supported, setSupported] = useState(true);

  const fetchMetrics = useCallback(() => {
    if (!supported) return;

    send("getVFDHostMetrics", {}, (resp: JsonRpcResponse) => {
      if ("error" in resp) {
        if (resp.error.code === RpcMethodNotFound) setSupported(false);
        return;
      }
      setMetrics({ ...emptyMetrics, ...(resp.result as Partial<VFDHostMetrics>) });
    });
  }, [send, supported]);

  useEffect(fetchMetrics, [fetchMetrics]);
  useInterval(fetchMetrics, supported ? 2000 : null);

  return { metrics, supported };
}
