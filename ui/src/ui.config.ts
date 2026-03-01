import { useNativeConfig } from "@/stores/nativeConfigStore";

const toBoolean = (value: string | undefined) => {
  if (!value) return false;
  return ["1", "true", "yes", "y"].includes(value.toLowerCase().trim());
};

export const CLOUD_API = import.meta.env.VITE_CLOUD_API;

export const CLOUD_BACKWARDS_COMPATIBLE_VERSION =
  import.meta.env.VITE_CLOUD_BACKWARDS_COMPATIBLE_VERSION || "0.5.0";

export const CLOUD_ENABLE_VERSIONED_UI = toBoolean(import.meta.env.VITE_CLOUD_ENABLE_VERSIONED_UI);

export const DOWNGRADE_VERSION = import.meta.env.VITE_DOWNGRADE_VERSION || "0.4.8";

// For native mode, get URL from config store dynamically
// This is a getter function that retrieves the current connection URL
export const getDeviceAPI = (): string => {
  if (import.meta.env.MODE === "tauri") {
    try {
      const state = useNativeConfig.getState();
      const url = state.currentConnection?.url || "";
      console.log("[getDeviceAPI] Current connection URL:", url, "State:", state);
      return url;
    } catch (e) {
      console.error("[getDeviceAPI] Failed to get native backend URL:", e);
      return "";
    }
  }
  return ""; // Empty string for device mode (uses window.location)
};

// Legacy export for compatibility - but prefer using getDeviceAPI()
export const DEVICE_API = "";
