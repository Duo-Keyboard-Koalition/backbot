/**
 * Debug utilities for Webguide
 * Provides logging, connection info, and debugging tools
 */

const DEBUG_PREFIX = '[Webguide]';
const LOG_PREFIX = `${DEBUG_PREFIX} [LOG]`;
const ERROR_PREFIX = `${DEBUG_PREFIX} [ERROR]`;
const WARN_PREFIX = `${DEBUG_PREFIX} [WARN]`;
const INFO_PREFIX = `${DEBUG_PREFIX} [INFO]`;
const API_PREFIX = `${DEBUG_PREFIX} [API]`;

// Check if debug mode is enabled
export const isDebugEnabled = (): boolean => {
  return import.meta.env.VITE_DEBUG === 'true';
};

// Get Tailscale auth key (masked for security)
export const getTailscaleAuthKey = (): string => {
  return import.meta.env.VITE_TAILSCALE_AUTH_KEY || '';
};

// Get masked auth key for display (shows first and last 4 chars)
export const getMaskedAuthKey = (): string => {
  const key = getTailscaleAuthKey();
  if (!key || key.length < 8) return '****';
  return `${key.substring(0, 8)}...${key.substring(key.length - 4)}`;
};

// Check if auth key is configured
export const isAuthKeyConfigured = (): boolean => {
  const key = getTailscaleAuthKey();
  if (!key || key.length === 0) return false;
  return key.startsWith('tskey-auth-');
};

// Get API URLs
export const getApiUrls = () => ({
  taila2a: import.meta.env.VITE_TAILA2A_URL || 'http://localhost:8080',
  tailfs: import.meta.env.VITE_TAILFS_URL || 'http://localhost:8081',
});

// Get refresh interval
export const getRefreshInterval = (): number => {
  return parseInt(import.meta.env.VITE_REFRESH_INTERVAL || '30', 10) * 1000;
};

// Debug logging
export const debugLog = {
  log: (...args: unknown[]) => {
    if (isDebugEnabled()) {
      console.log(LOG_PREFIX, ...args);
    }
  },
  error: (...args: unknown[]) => {
    if (isDebugEnabled()) {
      console.error(ERROR_PREFIX, ...args);
    } else {
      console.error(ERROR_PREFIX, ...args);
    }
  },
  warn: (...args: unknown[]) => {
    if (isDebugEnabled()) {
      console.warn(WARN_PREFIX, ...args);
    }
  },
  info: (...args: unknown[]) => {
    if (isDebugEnabled()) {
      console.info(INFO_PREFIX, ...args);
    }
  },
  api: (endpoint: string, method: string, data?: unknown) => {
    if (isDebugEnabled()) {
      console.log(`${API_PREFIX} ${method} ${endpoint}`, data ? data : '');
    }
  },
};

// Get connection info
export interface ConnectionInfo {
  debugEnabled: boolean;
  authKeyConfigured: boolean;
  authKeyMasked: string;
  taila2aUrl: string;
  tailfsUrl: string;
  refreshInterval: number;
  userAgent: string;
  timestamp: string;
}

export const getConnectionInfo = (): ConnectionInfo => {
  return {
    debugEnabled: isDebugEnabled(),
    authKeyConfigured: isAuthKeyConfigured(),
    authKeyMasked: getMaskedAuthKey(),
    taila2aUrl: getApiUrls().taila2a,
    tailfsUrl: getApiUrls().tailfs,
    refreshInterval: getRefreshInterval(),
    userAgent: navigator.userAgent,
    timestamp: new Date().toISOString(),
  };
};

// Performance monitoring
export const measurePerformance = {
  marks: new Map<string, number>(),

  start(label: string) {
    this.marks.set(label, performance.now());
    if (isDebugEnabled()) {
      console.log(`${INFO_PREFIX} Performance: ${label} started`);
    }
  },

  end(label: string): number | null {
    const start = this.marks.get(label);
    if (start === undefined) return null;
    const duration = performance.now() - start;
    this.marks.delete(label);
    if (isDebugEnabled()) {
      console.log(`${INFO_PREFIX} Performance: ${label} took ${duration.toFixed(2)}ms`);
    }
    return duration;
  },
};

// Network status monitoring
export interface NetworkStatus {
  online: boolean;
  rtt?: number;
  downlink?: number;
  effectiveType?: string;
}

export const getNetworkStatus = (): NetworkStatus => {
  const connection = (navigator as unknown as { connection?: {
    downlink?: number;
    effectiveType?: string;
  } }).connection;
  
  return {
    online: navigator.onLine,
    downlink: connection?.downlink,
    effectiveType: connection?.effectiveType,
  };
};

// Validate Tailscale auth key format
export const validateAuthKey = (key: string): { valid: boolean; error?: string } => {
  if (!key || key.length === 0) {
    return { valid: false, error: 'Auth key is empty' };
  }
  
  if (!key.startsWith('tskey-auth-')) {
    return { valid: false, error: 'Invalid key format - must start with tskey-auth-' };
  }
  
  if (key.length < 20) {
    return { valid: false, error: 'Auth key is too short' };
  }
  
  return { valid: true };
};

// Test API connectivity
export const testApiConnectivity = async (url: string, timeout = 5000): Promise<{
  success: boolean;
  status?: number;
  error?: string;
  responseTime?: number;
}> => {
  const startTime = performance.now();
  
  try {
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), timeout);
    
    const response = await fetch(url, {
      method: 'GET',
      signal: controller.signal,
      headers: { 'Content-Type': 'application/json' },
    });
    
    clearTimeout(timeoutId);
    const responseTime = performance.now() - startTime;
    
    return {
      success: response.ok,
      status: response.status,
      responseTime,
    };
  } catch (error) {
    const responseTime = performance.now() - startTime;
    return {
      success: false,
      error: error instanceof Error ? error.message : 'Unknown error',
      responseTime,
    };
  }
};

// Log system info
export const logSystemInfo = () => {
  if (!isDebugEnabled()) return;
  
  const info = getConnectionInfo();
  const network = getNetworkStatus();
  
  console.group(`${INFO_PREFIX} System Information`);
  console.log('Debug Enabled:', info.debugEnabled);
  console.log('Auth Key Configured:', info.authKeyConfigured);
  console.log('Auth Key (masked):', info.authKeyMasked);
  console.log('Taila2a URL:', info.taila2aUrl);
  console.log('TailFS URL:', info.tailfsUrl);
  console.log('Refresh Interval:', `${info.refreshInterval / 1000}s`);
  console.log('Network Status:', network.online ? 'Online' : 'Offline');
  console.log('Network Type:', network.effectiveType || 'Unknown');
  console.log('Downlink:', network.downlink ? `${network.downlink} Mbps` : 'Unknown');
  console.log('User Agent:', info.userAgent);
  console.log('Timestamp:', info.timestamp);
  console.groupEnd();
};

// Auto-log on module load if debug is enabled
if (isDebugEnabled()) {
  logSystemInfo();
}
