import { create } from 'zustand';
import { invoke } from '@tauri-apps/api/core';

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
  addConnection: (name: string, url: string) => Promise<Connection>;
  removeConnection: (id: string) => Promise<void>;
  setDefaultConnection: (id: string) => Promise<void>;
  setCurrentConnection: (connection: Connection) => void;
  updateLastConnected: (id: string) => Promise<void>;
}

export const useNativeConfig = create<ConfigStore>((set, get) => ({
  config: null,
  currentConnection: null,
  isLoading: false,
  error: null,
  
  loadConfig: async () => {
    set({ isLoading: true, error: null });
    try {
      const config = await invoke<AppConfig>('get_config');
      const defaultConn = config.connections.find(c => c.is_default);
      const currentConn = defaultConn || config.connections[0] || null;
      
      set({ 
        config, 
        currentConnection: currentConn,
        isLoading: false 
      });
    } catch (error) {
      console.error('Failed to load config:', error);
      set({ error: String(error), isLoading: false });
      throw error;
    }
  },
  
  saveConfig: async (config: AppConfig) => {
    try {
      await invoke('save_config', { config });
      set({ config });
    } catch (error) {
      console.error('Failed to save config:', error);
      set({ error: String(error) });
      throw error;
    }
  },
  
  addConnection: async (name: string, url: string) => {
    try {
      const connection = await invoke<Connection>('add_connection', { name, url });
      await get().loadConfig(); // Reload to get updated config
      
      // Set as current connection
      set({ currentConnection: connection });
      
      return connection;
    } catch (error) {
      console.error('Failed to add connection:', error);
      set({ error: String(error) });
      throw error;
    }
  },
  
  removeConnection: async (id: string) => {
    try {
      await invoke('remove_connection', { id });
      await get().loadConfig(); // Reload to get updated config
    } catch (error) {
      console.error('Failed to remove connection:', error);
      set({ error: String(error) });
      throw error;
    }
  },
  
  setDefaultConnection: async (id: string) => {
    try {
      await invoke('set_default_connection', { id });
      await get().loadConfig(); // Reload to get updated config
    } catch (error) {
      console.error('Failed to set default connection:', error);
      set({ error: String(error) });
      throw error;
    }
  },
  
  setCurrentConnection: (connection: Connection) => {
    set({ currentConnection: connection });
  },
  
  updateLastConnected: async (id: string) => {
    try {
      await invoke('update_last_connected', { id });
    } catch (error) {
      console.error('Failed to update last connected:', error);
    }
  },
}));
