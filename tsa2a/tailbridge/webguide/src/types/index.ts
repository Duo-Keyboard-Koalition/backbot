// Taila2a Types

export interface Agent {
  name: string;
  hostname: string;
  ip: string;
  online: boolean;
  last_seen: string;
  gateways: Gateway[];
}

export interface Gateway {
  port: number;
  protocol: string;
  service: string;
}

export interface AgentsResponse {
  agents: Agent[];
  count: number;
}

export interface Topic {
  name: string;
  consumers: string[];
  message_count: number;
}

export interface Consumer {
  name: string;
  topics: string[];
  status: 'active' | 'idle' | 'offline';
}

export interface BufferStats {
  total_messages: number;
  pending_messages: number;
  failed_messages: number;
  delivered_messages: number;
  oldest_message_age: number;
}

// TailFS Types

export interface Transfer {
  transfer_id: string;
  file_name: string;
  file_size: number;
  destination: string;
  source: string;
  status: 'pending' | 'sending' | 'completed' | 'failed' | 'cancelled';
  bytes_sent: number;
  bytes_total: number;
  percent_complete: string;
  bytes_per_second: number;
  eta_seconds: number;
  created_at: string;
  completed_at?: string;
}

export interface TransferHistory {
  transfers: Transfer[];
  total: number;
}

export interface SendRequest {
  file: string;
  destination: string;
  compress?: boolean;
  encrypt?: boolean;
}

export interface SendResponse {
  transfer_id: string;
  status: string;
  file_size: number;
}

// Dashboard Types

export interface SystemStats {
  total_agents: number;
  online_agents: number;
  active_transfers: number;
  total_topics: number;
  buffer_health: number;
  transfers_today: number;
  total_bytes_sent: number;
}
