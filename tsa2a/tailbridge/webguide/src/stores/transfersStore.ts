import { create } from 'zustand';
import type { Transfer } from '../types';
import { tailfsApi } from '../api/client';

interface TransfersState {
  activeTransfers: Transfer[];
  history: Transfer[];
  loading: boolean;
  error: string | null;
  lastUpdated: Date | null;
  
  // Actions
  fetchActiveTransfers: () => Promise<void>;
  fetchHistory: () => Promise<void>;
  refreshTransfer: (transferId: string) => Promise<void>;
  clearError: () => void;
}

export const useTransfersStore = create<TransfersState>((set) => ({
  activeTransfers: [],
  history: [],
  loading: false,
  error: null,
  lastUpdated: null,

  fetchActiveTransfers: async () => {
    set({ loading: true, error: null });
    try {
      const data = await tailfsApi.getTransfers();
      set({ 
        activeTransfers: data, 
        loading: false, 
        lastUpdated: new Date() 
      });
    } catch (err) {
      set({ 
        error: err instanceof Error ? err.message : 'Failed to fetch active transfers',
        loading: false 
      });
    }
  },

  fetchHistory: async () => {
    set({ loading: true, error: null });
    try {
      const data = await tailfsApi.getHistory();
      set({ 
        history: data.transfers, 
        loading: false, 
        lastUpdated: new Date() 
      });
    } catch (err) {
      set({ 
        error: err instanceof Error ? err.message : 'Failed to fetch transfer history',
        loading: false 
      });
    }
  },

  refreshTransfer: async (transferId: string) => {
    try {
      const transfer = await tailfsApi.getProgress(transferId);
      set((state) => ({
        activeTransfers: state.activeTransfers.map((t) =>
          t.transfer_id === transferId ? transfer : t
        ),
      }));
    } catch (err) {
      console.error('Failed to refresh transfer:', err);
    }
  },

  clearError: () => set({ error: null }),
}));
