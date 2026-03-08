import { create } from 'zustand';
import type { Agent } from '../types';
import { taila2aApi } from '../api/client';

interface AgentsState {
  agents: Agent[];
  loading: boolean;
  error: string | null;
  lastUpdated: Date | null;
  
  // Actions
  fetchAgents: () => Promise<void>;
  fetchOnlineAgents: () => Promise<void>;
  clearError: () => void;
}

export const useAgentsStore = create<AgentsState>((set) => ({
  agents: [],
  loading: false,
  error: null,
  lastUpdated: null,

  fetchAgents: async () => {
    set({ loading: true, error: null });
    try {
      const data = await taila2aApi.getAgents();
      set({ 
        agents: data.agents, 
        loading: false, 
        lastUpdated: new Date() 
      });
    } catch (err) {
      set({ 
        error: err instanceof Error ? err.message : 'Failed to fetch agents',
        loading: false 
      });
    }
  },

  fetchOnlineAgents: async () => {
    set({ loading: true, error: null });
    try {
      const data = await taila2aApi.getOnlineAgents();
      set({ 
        agents: data.agents, 
        loading: false, 
        lastUpdated: new Date() 
      });
    } catch (err) {
      set({ 
        error: err instanceof Error ? err.message : 'Failed to fetch online agents',
        loading: false 
      });
    }
  },

  clearError: () => set({ error: null }),
}));
