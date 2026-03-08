import { create } from 'zustand';
import type { SystemStats } from '../types';

interface DashboardState {
  stats: SystemStats | null;
  loading: boolean;
  error: string | null;
  
  // Actions
  updateStats: (stats: Partial<SystemStats>) => void;
  setLoading: (loading: boolean) => void;
  setError: (error: string | null) => void;
}

export const useDashboardStore = create<DashboardState>((set) => ({
  stats: null,
  loading: false,
  error: null,

  updateStats: (stats) =>
    set((state) => ({
      stats: state.stats ? { ...state.stats, ...stats } : (stats as SystemStats),
    })),

  setLoading: (loading) => set({ loading }),

  setError: (error) => set({ error }),
}));
