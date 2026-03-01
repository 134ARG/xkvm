import { useMemo } from "react";
import { getDeviceAPI } from "@/ui.config";
import { isNative } from "@/main";

/**
 * Hook to get the device API URL
 * For native mode, this retrieves the URL from the config store
 * For device mode, returns empty string (uses window.location)
 */
export function useDeviceAPI(): string {
  return useMemo(() => {
    if (isNative) {
      return getDeviceAPI();
    }
    return "";
  }, []);
}
