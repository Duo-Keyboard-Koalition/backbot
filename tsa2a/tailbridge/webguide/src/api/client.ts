import type {
  AgentsResponse,
  Topic,
  Consumer,
  BufferStats,
  Transfer,
  TransferHistory,
  SendRequest,
  SendResponse,
} from '../types';
import { debugLog } from '../utils/debug';

const TAILA2A_BASE = '/api/taila2a';
const TAILFS_BASE = '/api/tailfs';

async function handleResponse<T>(response: Response, endpoint?: string): Promise<T> {
  if (!response.ok) {
    const error = await response.text().catch(() => 'Unknown error');
    debugLog.api(endpoint || 'unknown', 'ERROR', { status: response.status, error });
    throw new Error(`API Error: ${response.status} - ${error}`);
  }
  const data = await response.json();
  debugLog.api(endpoint || 'unknown', 'RESPONSE', data);
  return data;
}

// Taila2a API Client

export const taila2aApi = {
  // Get all agents from phone book
  getAgents: async (): Promise<AgentsResponse> => {
    debugLog.api('/agents', 'GET');
    const response = await fetch(`${TAILA2A_BASE}/agents`);
    return handleResponse<AgentsResponse>(response, '/agents');
  },

  // Get online agents only
  getOnlineAgents: async (): Promise<AgentsResponse> => {
    debugLog.api('/agents/online', 'GET');
    const response = await fetch(`${TAILA2A_BASE}/agents/online`);
    return handleResponse<AgentsResponse>(response, '/agents/online');
  },

  // Get topics
  getTopics: async (): Promise<Topic[]> => {
    debugLog.api('/topics', 'GET');
    const response = await fetch(`${TAILA2A_BASE}/topics`);
    return handleResponse<Topic[]>(response, '/topics');
  },

  // Get consumers
  getConsumers: async (): Promise<Consumer[]> => {
    debugLog.api('/consumers', 'GET');
    const response = await fetch(`${TAILA2A_BASE}/consumers`);
    return handleResponse<Consumer[]>(response, '/consumers');
  },

  // Get buffer stats
  getBufferStats: async (): Promise<BufferStats> => {
    debugLog.api('/buffer/stats', 'GET');
    const response = await fetch(`${TAILA2A_BASE}/buffer/stats`);
    return handleResponse<BufferStats>(response, '/buffer/stats');
  },

  // Get trigger status
  getTriggerStatus: async (): Promise<unknown> => {
    debugLog.api('/trigger/status', 'GET');
    const response = await fetch(`${TAILA2A_BASE}/trigger/status`);
    return handleResponse<unknown>(response, '/trigger/status');
  },

  // Send message to agent
  sendMessage: async (
    sourceNode: string,
    destNode: string,
    payload: unknown
  ): Promise<unknown> => {
    const body = JSON.stringify({
      source_node: sourceNode,
      dest_node: destNode,
      payload,
    });
    debugLog.api('/send', 'POST', body);
    const response = await fetch(`${TAILA2A_BASE}/send`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body,
    });
    return handleResponse<unknown>(response, '/send');
  },
};

// TailFS API Client

export const tailfsApi = {
  // Get active transfers
  getTransfers: async (): Promise<Transfer[]> => {
    debugLog.api('/transfers', 'GET');
    const response = await fetch(`${TAILFS_BASE}/transfers`);
    return handleResponse<Transfer[]>(response, '/transfers');
  },

  // Get transfer history
  getHistory: async (): Promise<TransferHistory> => {
    debugLog.api('/history', 'GET');
    const response = await fetch(`${TAILFS_BASE}/history`);
    return handleResponse<TransferHistory>(response, '/history');
  },

  // Get transfer progress
  getProgress: async (transferId: string): Promise<Transfer> => {
    debugLog.api(`/progress`, 'GET', { transferId });
    const response = await fetch(
      `${TAILFS_BASE}/progress?transfer_id=${transferId}`
    );
    return handleResponse<Transfer>(response, '/progress');
  },

  // Send file
  sendFile: async (request: SendRequest): Promise<SendResponse> => {
    debugLog.api('/send', 'POST', request);
    const response = await fetch(`${TAILFS_BASE}/send`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(request),
    });
    return handleResponse<SendResponse>(response, '/send');
  },

  // Get agents with file receive capability
  getReceiveAgents: async (): Promise<AgentsResponse> => {
    debugLog.api('/agents', 'GET', { capability: 'file_receive' });
    const response = await fetch(`${TAILFS_BASE}/agents?capability=file_receive`);
    return handleResponse<AgentsResponse>(response, '/agents');
  },
};
