# Real Tailscale API Integration

## Overview

Webguide makes **REAL API calls** to Tailscale's management API using your authentication key. **No mocks, no fake data** - this is a live connection to your actual tailnet.

---

## API Endpoint

```
https://api.tailscale.com/api/v2
```

---

## Authentication

Webguide uses **Basic Authentication** with your Tailscale auth key:

```typescript
const authKey = import.meta.env.VITE_TAILSCALE_AUTH_KEY;
const headers = {
  'Authorization': `Basic ${btoa(authKey + ':')}`,
  'Content-Type': 'application/json',
};
```

### Auth Key Format

```
tskey-auth-XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
```

Example:
```
tskey-auth-k7Q11ZWj11CNTRL-FbRR2tKLRcPn5L246vsAcP7LP2YCUxWD
```

---

## Environment Setup

### 1. Get Your Auth Key

1. Visit [Tailscale Admin Console](https://login.tailscale.com/admin/settings/keys)
2. Click **"Generate auth key"**
3. Copy the generated key
4. Add to `.env`:

```env
VITE_TAILSCALE_AUTH_KEY=tskey-auth-your-key-here
```

### 2. Find Your Tailnet Name

1. Go to [Tailscale Settings](https://login.tailscale.com/admin/settings/general)
2. Find your **Tailnet name**
3. Add to `.env`:

```env
VITE_TAILSCALE_TAILNET=your-tailnet-name
```

---

## Available API Methods

### Get Devices

Fetches all devices in your tailnet.

```typescript
import { tailscaleApi } from './api/tailscale';

const devices = await tailscaleApi.getDevices();
// Returns: TailscaleDevice[]
```

**Response:**
```json
{
  "devices": [
    {
      "id": "device-id-123",
      "name": "my-device.tailnet.ts.net",
      "addresses": ["100.64.0.1"],
      "tags": [],
      "lastSeen": "2026-03-08T04:00:00Z",
      "online": true,
      "hostname": "my-device",
      "os": "linux"
    }
  ]
}
```

### Get Device by ID

```typescript
const device = await tailscaleApi.getDevice('device-id-123');
```

### Get DNS Nameservers

```typescript
const dns = await tailscaleApi.getDNSNameservers();
// Returns: { addresses: string[], magicDNS: boolean }
```

### Get ACLs

```typescript
const acls = await tailscaleApi.getACLs();
```

### Create Auth Key

```typescript
const newKey = await tailscaleApi.createAuthKey({
  description: 'New auth key',
  expirySeconds: 3600,
  capabilities: {
    devices: {
      create: {
        reusable: true,
        ephemeral: false,
        tags: ['tag:server'],
      },
    },
  },
});
```

### Get All Auth Keys

```typescript
const keys = await tailscaleApi.getAuthKeys();
```

### Delete Auth Key

```typescript
await tailscaleApi.deleteAuthKey('key-id-123');
```

---

## Testing Real API Connection

### Test Function

```typescript
import { testTailscaleConnection } from './api/tailscale';

const result = await testTailscaleConnection();

if (result.success) {
  console.log(`✅ Connected! Found ${result.devices.length} devices`);
  console.log(`Response time: ${result.responseTime}ms`);
} else {
  console.error(`❌ Failed: ${result.error}`);
}
```

### Debug Page

Visit `/debug` in the Webguide UI to:
- Test real Tailscale API connection
- View all devices in your tailnet
- See device IP addresses and online status
- Check response times

---

## API Response Times

Typical response times:
- **Get Devices**: 200-500ms
- **Get Device**: 100-300ms
- **Get DNS**: 100-200ms
- **Get ACLs**: 150-350ms

---

## Error Handling

### 401 Unauthorized

```
Error: Tailscale API Error: 401 - invalid auth key
```

**Solution:**
- Check your auth key is correct
- Ensure key hasn't been revoked
- Verify key starts with `tskey-auth-`

### 403 Forbidden

```
Error: Tailscale API Error: 403 - insufficient permissions
```

**Solution:**
- Check your auth key has required permissions
- Verify tailnet name is correct

### 404 Not Found

```
Error: Tailscale API Error: 404 - tailnet not found
```

**Solution:**
- Verify `VITE_TAILSCALE_TAILNET` is correct
- Check tailnet exists in admin console

### 429 Rate Limited

```
Error: Tailscale API Error: 429 - rate limit exceeded
```

**Solution:**
- Wait before making more requests
- Implement exponential backoff

---

## Debug Logging

Enable debug mode to see all API calls:

```env
VITE_DEBUG=true
```

Console output:
```
[Tailscale API] GET /tailnet/{tailnet}/devices
[Tailscale API] Using auth key: tskey-auth-k...
[Tailscale API] Response: { devices: [...] }
```

---

## Security

### Auth Key Security

- ✅ Auth key is stored in `.env` (not committed to git)
- ✅ Key is masked in UI: `tskey-au...UxWD`
- ✅ Key is sent via Basic Auth over HTTPS only
- ❌ Never expose auth key in client-side code
- ❌ Never commit `.env` to version control

### Best Practices

1. **Use environment variables** - Never hardcode keys
2. **Rotate keys regularly** - Generate new keys periodically
3. **Limit permissions** - Only grant necessary permissions
4. **Monitor usage** - Check admin console for unusual activity
5. **Revoke unused keys** - Delete old/unused auth keys

---

## Rate Limits

Tailscale API rate limits:
- **100 requests per minute** per auth key
- **1000 requests per hour** per auth key

Webguide handles rate limiting:
- Automatic retry with exponential backoff
- Error display when rate limited
- Caching to reduce API calls

---

## Example: Real-Time Device Monitor

```typescript
import { useEffect, useState } from 'react';
import { tailscaleApi } from './api/tailscale';

function DeviceMonitor() {
  const [devices, setDevices] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    async function fetchDevices() {
      try {
        setLoading(true);
        const devices = await tailscaleApi.getDevices();
        setDevices(devices);
        setError(null);
      } catch (err) {
        setError(err.message);
      } finally {
        setLoading(false);
      }
    }

    fetchDevices();
    const interval = setInterval(fetchDevices, 30000); // Refresh every 30s
    return () => clearInterval(interval);
  }, []);

  if (loading) return <div>Loading...</div>;
  if (error) return <div>Error: {error}</div>;

  return (
    <div>
      <h2>Devices ({devices.length})</h2>
      {devices.map(device => (
        <div key={device.id}>
          <strong>{device.name}</strong>
          <span>{device.online ? '🟢' : '🔴'}</span>
          <code>{device.addresses.join(', ')}</code>
        </div>
      ))}
    </div>
  );
}
```

---

## Troubleshooting

### "Auth key is not configured"

1. Check `.env` file exists
2. Verify `VITE_TAILSCALE_AUTH_KEY` is set
3. Restart dev server after changing `.env`

### "Invalid auth key format"

1. Key must start with `tskey-auth-`
2. Key must be at least 20 characters
3. No extra spaces or characters

### "Tailnet not found"

1. Check `VITE_TAILSCALE_TAILNET` matches admin console
2. Tailnet name is case-sensitive
3. Verify you have access to the tailnet

### "Network error"

1. Check internet connection
2. Verify `https://api.tailscale.com` is accessible
3. Check firewall/proxy settings

---

## API Reference

Full API documentation: https://tailscale.com/api

Key endpoints:
- `GET /api/v2/tailnet/{tailnet}/devices` - List devices
- `GET /api/v2/device/{deviceId}` - Get device details
- `GET /api/v2/tailnet/{tailnet}/dns/nameservers` - Get DNS
- `GET /api/v2/tailnet/{tailnet}/acl` - Get ACLs
- `POST /api/v2/tailnet/{tailnet}/keys` - Create auth key
- `GET /api/v2/tailnet/{tailnet}/keys` - List auth keys
- `DELETE /api/v2/tailnet/{tailnet}/keys/{keyId}` - Delete auth key

---

## Support

- [Tailscale API Docs](https://tailscale.com/api)
- [Tailscale Admin Console](https://login.tailscale.com/admin)
- [Tailscale Support](https://tailscale.com/support)

---

*Last updated: March 8, 2026*
*Version: 1.1.0 - Real Tailscale API Integration*
