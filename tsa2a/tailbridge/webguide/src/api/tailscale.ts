/**
 * Tailscale API Client
 * Makes real API calls to Tailscale's management API
 */

const TAILSCALE_API_BASE = 'https://api.tailscale.com/api/v2';

// Get the auth key from environment
const getAuthKey = (): string => {
  const key = import.meta.env.VITE_TAILSCALE_AUTH_KEY;
  if (!key || !key.startsWith('tskey-auth-')) {
    throw new Error(
      'VITE_TAILSCALE_AUTH_KEY is not configured or invalid. ' +
      'Please set it in your .env file.'
    );
  }
  return key;
};

// Get tailnet name from auth key or config
const getTailnet = (): string => {
  return import.meta.env.VITE_TAILSCALE_TAILNET || 'tailnet';
};

interface TailscaleDevice {
  id: string;
  name: string;
  addresses: string[];
  tags?: string[];
  lastSeen?: string;
  online: boolean;
  hostname: string;
  os: string;
}

interface TailscaleDNSNameServers {
  addresses: string[];
  magicDNS: boolean;
}

interface TailscaleACL {
  acls: Array<{
    action: string;
    users: string[];
    ports: string[];
  }>;
}

export const tailscaleApi = {
  /**
   * Get all devices in the tailnet
   */
  getDevices: async (): Promise<TailscaleDevice[]> => {
    const authKey = getAuthKey();
    const tailnet = getTailnet();
    
    console.log('[Tailscale API] GET /tailnet/{tailnet}/devices');
    console.log('[Tailscale API] Using auth key:', `${authKey.substring(0, 12)}...`);
    
    const response = await fetch(
      `${TAILSCALE_API_BASE}/tailnet/${tailnet}/devices`,
      {
        method: 'GET',
        headers: {
          'Authorization': `Basic ${btoa(authKey + ':')}`,
          'Content-Type': 'application/json',
        },
      }
    );

    if (!response.ok) {
      const error = await response.text();
      console.error('[Tailscale API] Error:', response.status, error);
      throw new Error(`Tailscale API Error: ${response.status} - ${error}`);
    }

    const data = await response.json();
    console.log('[Tailscale API] Response:', data);
    return data.devices || [];
  },

  /**
   * Get device by ID
   */
  getDevice: async (deviceId: string): Promise<TailscaleDevice> => {
    const authKey = getAuthKey();
    
    const response = await fetch(
      `${TAILSCALE_API_BASE}/device/${deviceId}`,
      {
        method: 'GET',
        headers: {
          'Authorization': `Basic ${btoa(authKey + ':')}`,
          'Content-Type': 'application/json',
        },
      }
    );

    if (!response.ok) {
      const error = await response.text();
      throw new Error(`Tailscale API Error: ${response.status} - ${error}`);
    }

    return response.json();
  },

  /**
   * Get DNS nameservers
   */
  getDNSNameservers: async (): Promise<TailscaleDNSNameServers> => {
    const authKey = getAuthKey();
    const tailnet = getTailnet();
    
    const response = await fetch(
      `${TAILSCALE_API_BASE}/tailnet/${tailnet}/dns/nameservers`,
      {
        method: 'GET',
        headers: {
          'Authorization': `Basic ${btoa(authKey + ':')}`,
          'Content-Type': 'application/json',
        },
      }
    );

    if (!response.ok) {
      const error = await response.text();
      throw new Error(`Tailscale API Error: ${response.status} - ${error}`);
    }

    return response.json();
  },

  /**
   * Get ACLs
   */
  getACLs: async (): Promise<TailscaleACL> => {
    const authKey = getAuthKey();
    const tailnet = getTailnet();
    
    const response = await fetch(
      `${TAILSCALE_API_BASE}/tailnet/${tailnet}/acl`,
      {
        method: 'GET',
        headers: {
          'Authorization': `Basic ${btoa(authKey + ':')}`,
          'Content-Type': 'application/json',
        },
      }
    );

    if (!response.ok) {
      const error = await response.text();
      throw new Error(`Tailscale API Error: ${response.status} - ${error}`);
    }

    return response.json();
  },

  /**
   * Create auth key
   */
  createAuthKey: async (options: {
    capabilities?: {
      devices?: {
        create?: {
          reusable?: boolean;
          ephemeral?: boolean;
          tags?: string[];
        };
      };
    };
    expirySeconds?: number;
    description?: string;
  }): Promise<{
    id: string;
    key: string;
    description: string;
    created: string;
    expires: string;
  }> => {
    const authKey = getAuthKey();
    const tailnet = getTailnet();
    
    const response = await fetch(
      `${TAILSCALE_API_BASE}/tailnet/${tailnet}/keys`,
      {
        method: 'POST',
        headers: {
          'Authorization': `Basic ${btoa(authKey + ':')}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(options),
      }
    );

    if (!response.ok) {
      const error = await response.text();
      throw new Error(`Tailscale API Error: ${response.status} - ${error}`);
    }

    return response.json();
  },

  /**
   * Get all auth keys
   */
  getAuthKeys: async (): Promise<Array<{
    id: string;
    description: string;
    created: string;
    expires: string;
    revoked: boolean;
    invalid: boolean;
  }>> => {
    const authKey = getAuthKey();
    const tailnet = getTailnet();
    
    const response = await fetch(
      `${TAILSCALE_API_BASE}/tailnet/${tailnet}/keys`,
      {
        method: 'GET',
        headers: {
          'Authorization': `Basic ${btoa(authKey + ':')}`,
          'Content-Type': 'application/json',
        },
      }
    );

    if (!response.ok) {
      const error = await response.text();
      throw new Error(`Tailscale API Error: ${response.status} - ${error}`);
    }

    const data = await response.json();
    return data.keys || [];
  },

  /**
   * Delete an auth key
   */
  deleteAuthKey: async (keyId: string): Promise<void> => {
    const authKey = getAuthKey();
    const tailnet = getTailnet();
    
    const response = await fetch(
      `${TAILSCALE_API_BASE}/tailnet/${tailnet}/keys/${keyId}`,
      {
        method: 'DELETE',
        headers: {
          'Authorization': `Basic ${btoa(authKey + ':')}`,
        },
      }
    );

    if (!response.ok) {
      const error = await response.text();
      throw new Error(`Tailscale API Error: ${response.status} - ${error}`);
    }
  },
};

/**
 * Test Tailscale API connectivity
 */
export const testTailscaleConnection = async (): Promise<{
  success: boolean;
  devices?: TailscaleDevice[];
  error?: string;
  responseTime?: number;
}> => {
  const startTime = performance.now();
  
  try {
    console.log('[Tailscale Test] Testing connection...');
    const devices = await tailscaleApi.getDevices();
    const responseTime = performance.now() - startTime;
    
    console.log(`[Tailscale Test] Success! Found ${devices.length} devices in ${responseTime.toFixed(0)}ms`);
    
    return {
      success: true,
      devices,
      responseTime,
    };
  } catch (error) {
    const responseTime = performance.now() - startTime;
    const errorMessage = error instanceof Error ? error.message : 'Unknown error';
    
    console.error('[Tailscale Test] Failed:', errorMessage);
    
    return {
      success: false,
      error: errorMessage,
      responseTime,
    };
  }
};
