import { create } from 'zustand';
import type { Topic, Consumer, BufferStats } from '../types';
import { taila2aApi } from '../api/client';

interface TopicsState {
  topics: Topic[];
  consumers: Consumer[];
  bufferStats: BufferStats | null;
  loading: boolean;
  error: string | null;
  lastUpdated: Date | null;
  
  // Actions
  fetchTopics: () => Promise<void>;
  fetchConsumers: () => Promise<void>;
  fetchBufferStats: () => Promise<void>;
  clearError: () => void;
}

export const useTopicsStore = create<TopicsState>((set) => ({
  topics: [],
  consumers: [],
  bufferStats: null,
  loading: false,
  error: null,
  lastUpdated: null,

  fetchTopics: async () => {
    set({ loading: true, error: null });
    try {
      const data = await taila2aApi.getTopics();
      set({ 
        topics: data, 
        loading: false, 
        lastUpdated: new Date() 
      });
    } catch (err) {
      set({ 
        error: err instanceof Error ? err.message : 'Failed to fetch topics',
        loading: false 
      });
    }
  },

  fetchConsumers: async () => {
    set({ loading: true, error: null });
    try {
      const data = await taila2aApi.getConsumers();
      set({ 
        consumers: data, 
        loading: false, 
        lastUpdated: new Date() 
      });
    } catch (err) {
      set({ 
        error: err instanceof Error ? err.message : 'Failed to fetch consumers',
        loading: false 
      });
    }
  },

  fetchBufferStats: async () => {
    set({ loading: true, error: null });
    try {
      const data = await taila2aApi.getBufferStats();
      set({ 
        bufferStats: data, 
        loading: false, 
        lastUpdated: new Date() 
      });
    } catch (err) {
      set({ 
        error: err instanceof Error ? err.message : 'Failed to fetch buffer stats',
        loading: false 
      });
    }
  },

  clearError: () => set({ error: null }),
}));
