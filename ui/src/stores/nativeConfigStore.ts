import { create } from "zustand";
import type { invoke as TauriInvoke } from "@tauri-apps/api/core";

export interface Connection {
  id: string;
  name: string;
  url: string;
  is_default: boolean;
  last_connected?: string;
}

export interface AppConfig {
  version: string;
  connections: Connection[];
  settings: {
    auto_connect: boolean;
    remember_last_connection: boolean;
  };
}

interface ConfigStore {
  config: AppConfig | null;
  currentConnection: Connection | null;
  isLoading: boolean;
  error: string | null;

  loadConfig: () => Promise<void>;
  saveConfig: (config: AppConfig) => Promise<void>;
  updateSettings: (settings: Partial<AppConfig["settings"]>) => Promise<void>;
  addConnection: (name: string, url: string) => Promise<Connection>;
  updateConnection: (id: string, name: string, url: string) => Promise<Connection>;
  testConnection: (url: string) => Promise<{ ok: boolean; error?: string }>;
  switchConnection: (id: string) => Promise<void>;
  removeConnection: (id: string) => Promise<void>;
  setDefaultConnection: (id: string) => Promise<void>;
  setCurrentConnection: (connection: Connection) => void;
  updateLastConnected: (id: string) => Promise<void>;
}

// Session-scoped id of the connection the user explicitly switched to during
// this app run. Survives a window reload but is cleared when the app closes,
// so cold starts fall back to the startup preference below.
const ACTIVE_CONNECTION_KEY = "xkvm.activeConnectionId";

// Decide which profile is the active connection, independently of which is the
// user's chosen default:
//   1. the profile explicitly connected to this session (sessionStorage), else
//   2. the startup preference — newest last_connected when "remember last" is on,
//      otherwise the default, else
//   3. the default, else the first profile.
function resolveActiveConnection(config: AppConfig): Connection | null {
  const conns = config.connections;
  if (conns.length === 0) return null;

  const sessionId = sessionStorage.getItem(ACTIVE_CONNECTION_KEY);
  const sessionConn = sessionId ? conns.find(c => c.id === sessionId) : undefined;
  if (sessionConn) return sessionConn;

  const defaultConn = conns.find(c => c.is_default) ?? conns[0];

  if (config.settings.remember_last_connection) {
    const lastUsed = conns
      .filter(c => c.last_connected)
      .sort((a, b) => (a.last_connected! < b.last_connected! ? 1 : -1))[0];
    return lastUsed ?? defaultConn;
  }

  return defaultConn;
}

// Lazy load the Tauri API to avoid issues in production builds
let invokeCache: typeof TauriInvoke | null = null;
async function getInvoke(): Promise<typeof TauriInvoke> {
  if (!invokeCache) {
    const { invoke } = await import("@tauri-apps/api/core");
    invokeCache = invoke;
  }
  return invokeCache;
}

export const useNativeConfig = create<ConfigStore>((set, get) => ({
  config: null,
  currentConnection: null,
  isLoading: false,
  error: null,

  loadConfig: async () => {
    set({ isLoading: true, error: null });
    try {
      const invoke = await getInvoke();
      const config = await invoke<AppConfig>("get_config");

      // Only resolve the active connection on a cold load. Later reloads (after
      // set-default / add / edit / remove in settings) must NOT re-pick it, or
      // editing the default would move the "Connected" badge without an actual
      // reconnect. Re-read the live profile by id to pick up edits; if it was
      // removed, fall back to a fresh resolution.
      const existing = get().currentConnection;
      const currentConn = existing
        ? (config.connections.find(c => c.id === existing.id) ??
          resolveActiveConnection(config))
        : resolveActiveConnection(config);

      set({
        config,
        currentConnection: currentConn,
        isLoading: false,
      });
    } catch (error) {
      console.error("Failed to load config:", error);
      set({ error: String(error), isLoading: false });
      throw error;
    }
  },

  saveConfig: async (config: AppConfig) => {
    try {
      const invoke = await getInvoke();
      await invoke("save_config", { config });
      set({ config });
    } catch (error) {
      console.error("Failed to save config:", error);
      set({ error: String(error) });
      throw error;
    }
  },

  updateSettings: async (settings: Partial<AppConfig["settings"]>) => {
    const current = get().config;
    if (!current) return;
    const next: AppConfig = {
      ...current,
      settings: { ...current.settings, ...settings },
    };
    await get().saveConfig(next);
  },

  addConnection: async (name: string, url: string) => {
    try {
      const invoke = await getInvoke();
      const connection = await invoke<Connection>("add_connection", { name, url });
      // Note: adding a connection does not switch the active one. Callers that
      // want to activate it (e.g. first-run setup) call switchConnection().
      await get().loadConfig();

      return connection;
    } catch (error) {
      console.error("Failed to add connection:", error);
      set({ error: String(error) });
      throw error;
    }
  },

  updateConnection: async (id: string, name: string, url: string) => {
    try {
      const invoke = await getInvoke();
      const connection = await invoke<Connection>("update_connection", { id, name, url });
      await get().loadConfig();
      return connection;
    } catch (error) {
      console.error("Failed to update connection:", error);
      set({ error: String(error) });
      throw error;
    }
  },

  testConnection: async (url: string) => {
    try {
      const invoke = await getInvoke();
      await invoke("test_connection", { url });
      return { ok: true };
    } catch (error) {
      return { ok: false, error: String(error) };
    }
  },

  switchConnection: async (id: string) => {
    try {
      const invoke = await getInvoke();
      // Connecting marks this profile as "last used" but does NOT change the
      // user's chosen default. Record it as the active connection for this app
      // session so it survives the reload below (cleared when the app closes).
      await invoke("update_last_connected", { id });
      sessionStorage.setItem(ACTIVE_CONNECTION_KEY, id);
      // Full reload so the boot config-check and device reconnect run cleanly.
      window.location.href = "/";
    } catch (error) {
      console.error("Failed to switch connection:", error);
      set({ error: String(error) });
      throw error;
    }
  },

  removeConnection: async (id: string) => {
    try {
      const invoke = await getInvoke();
      await invoke("remove_connection", { id });
      await get().loadConfig();
    } catch (error) {
      console.error("Failed to remove connection:", error);
      set({ error: String(error) });
      throw error;
    }
  },

  setDefaultConnection: async (id: string) => {
    try {
      const invoke = await getInvoke();
      await invoke("set_default_connection", { id });
      await get().loadConfig();
    } catch (error) {
      console.error("Failed to set default connection:", error);
      set({ error: String(error) });
      throw error;
    }
  },

  setCurrentConnection: (connection: Connection) => {
    set({ currentConnection: connection });
  },

  updateLastConnected: async (id: string) => {
    try {
      const invoke = await getInvoke();
      await invoke("update_last_connected", { id });
    } catch (error) {
      console.error("Failed to update last connected:", error);
    }
  },
}));
