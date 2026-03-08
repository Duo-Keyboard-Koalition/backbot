# Webguide Testing Guide

## Overview

Webguide includes a comprehensive test suite for validating API connections, Tailscale authentication, and debug utilities.

---

## Test Setup

### Environment Variables

Tests require the following environment variables to be set in `.env`:

```env
VITE_TAILSCALE_AUTH_KEY=tskey-auth-k7Q11ZWj11CNTRL-FbRR2tKLRcPn5L246vsAcP7LP2YCUxWD
VITE_TAILA2A_URL=http://localhost:8080
VITE_TAILFS_URL=http://localhost:8081
VITE_DEBUG=true
VITE_REFRESH_INTERVAL=30
```

### Test Dependencies

- **Vitest** - Fast unit test framework
- **Testing Library** - React testing utilities
- **jsdom** - Browser environment simulation

---

## Running Tests

### Watch Mode (Development)

```bash
npm run test
```

Runs tests in watch mode - automatically re-runs when files change.

### Single Run (CI/CD)

```bash
npm run test:run
```

Runs tests once and exits.

### With UI

```bash
npm run test:ui
```

Opens the Vitest UI for interactive test exploration.

---

## Test Suites

### 1. Tailscale Authentication Tests

Location: `src/test/api.test.ts`

Tests validate the Tailscale auth key configuration and format:

```typescript
// Test: Auth key is configured
expect(isAuthKeyConfigured()).toBe(true);

// Test: Auth key has correct value
const key = getTailscaleAuthKey();
expect(key).toBe('tskey-auth-k7Q11ZWj11CNTRL-FbRR2tKLRcPn5L246vsAcP7LP2YCUxWD');

// Test: Valid key format
const result = validateAuthKey('tskey-auth-XXXXX');
expect(result.valid).toBe(true);

// Test: Invalid key format
const result = validateAuthKey('invalid-key');
expect(result.valid).toBe(false);
expect(result.error).toContain('Invalid key format');
```

**What's Tested:**
- ✅ Auth key is present in environment
- ✅ Auth key starts with `tskey-auth-`
- ✅ Auth key has minimum length
- ✅ Invalid keys are rejected with appropriate error messages

---

### 2. Taila2a API Tests

Location: `src/test/api.test.ts`

Tests the Taila2a API client:

```typescript
// Test: Fetch agents
const mockAgents = {
  agents: [{
    name: 'test-agent-1',
    hostname: 'test-agent-1.tailnet.ts.net',
    ip: '100.64.0.1',
    online: true,
    last_seen: new Date().toISOString(),
    gateways: [{ port: 8001, protocol: 'tcp', service: 'bridge-inbound' }],
  }],
  count: 1,
};

vi.mocked(fetch).mockResolvedValueOnce({
  ok: true,
  status: 200,
  json: () => Promise.resolve(mockAgents),
} as Response);

const result = await taila2aApi.getAgents();
expect(result.count).toBe(1);
```

**What's Tested:**
- ✅ GET /agents - Fetch all agents
- ✅ GET /agents/online - Fetch online agents only
- ✅ GET /buffer/stats - Fetch buffer statistics
- ✅ Error handling for API failures (500, 404, etc.)

---

### 3. TailFS API Tests

Location: `src/test/api.test.ts`

Tests the TailFS API client:

```typescript
// Test: Send file
const result = await tailfsApi.sendFile({
  file: '/path/to/file.pdf',
  destination: 'tailfs-beta',
  compress: true,
});

expect(fetch).toHaveBeenCalledWith('/api/tailfs/send', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    file: '/path/to/file.pdf',
    destination: 'tailfs-beta',
    compress: true,
  }),
});
```

**What's Tested:**
- ✅ GET /transfers - Fetch active transfers
- ✅ GET /history - Fetch transfer history
- ✅ POST /send - Initiate file transfer
- ✅ Request body format and headers

---

### 4. Debug Utilities Tests

Location: `src/test/api.test.ts`

Tests the debug utilities:

```typescript
// Test: Get connection info
const info = getConnectionInfo();
expect(info.debugEnabled).toBe(true);
expect(info.authKeyConfigured).toBe(true);
expect(info.taila2aUrl).toBe('http://localhost:8080');
expect(info.tailfsUrl).toBe('http://localhost:8081');

// Test: Get network status
const status = getNetworkStatus();
expect(typeof status.online).toBe('boolean');
```

**What's Tested:**
- ✅ Connection info retrieval
- ✅ Network status detection
- ✅ Debug mode flag
- ✅ API URL configuration

---

## Debug Logging in Tests

When tests run, debug logs are output to the console:

```
stdout | src/test/api.test.ts > Taila2a API > should fetch agents successfully
[Webguide] [API] GET /agents 
[Webguide] [API] RESPONSE /agents {
  agents: [
    {
      name: 'test-agent-1',
      hostname: 'test-agent-1.tailnet.ts.net',
      ip: '100.64.0.1',
      online: true,
      last_seen: '2026-03-08T04:18:56.546Z',
      gateways: [Array]
    }
  ],
  count: 1
}
```

**Log Prefixes:**
- `[Webguide] [API]` - API requests and responses
- `[Webguide] [INFO]` - Informational messages
- `[Webguide] [ERROR]` - Error messages
- `[Webguide] [WARN]` - Warning messages

---

## Writing New Tests

### Test Structure

```typescript
import { describe, it, expect, beforeEach, vi } from 'vitest';
import { yourModule } from '../path/to/module';

describe('Module Name', () => {
  beforeEach(() => {
    // Reset mocks before each test
    vi.clearAllMocks();
  });

  it('should do something', async () => {
    // Arrange - setup mocks
    vi.mocked(fetch).mockResolvedValueOnce({
      ok: true,
      status: 200,
      json: () => Promise.resolve({ data: 'test' }),
    } as Response);

    // Act - call the function
    const result = await yourModule.function();

    // Assert - check results
    expect(result).toEqual({ data: 'test' });
    expect(fetch).toHaveBeenCalledWith('/expected/endpoint');
  });

  it('should handle errors', async () => {
    // Arrange - mock error response
    vi.mocked(fetch).mockResolvedValueOnce({
      ok: false,
      status: 500,
      text: () => Promise.resolve('Internal Error'),
    } as Response);

    // Act & Assert - should throw
    await expect(yourModule.function()).rejects.toThrow('API Error: 500');
  });
});
```

### Mocking Fetch

```typescript
// Success response
vi.mocked(fetch).mockResolvedValueOnce({
  ok: true,
  status: 200,
  json: () => Promise.resolve({ key: 'value' }),
  text: () => Promise.resolve('text response'),
  headers: new Headers(),
} as Response);

// Error response
vi.mocked(fetch).mockResolvedValueOnce({
  ok: false,
  status: 404,
  text: () => Promise.resolve('Not Found'),
  headers: new Headers(),
} as Response);

// Network error
vi.mocked(fetch).mockRejectedValueOnce(new Error('Network error'));
```

---

## Test Coverage

To check test coverage (future enhancement):

```bash
npm run test -- --coverage
```

---

## Troubleshooting Tests

### "vi is not defined"

Ensure `src/test/setup.ts` includes:
```typescript
/// <reference types="vitest/globals" />
```

And `tsconfig.json` has:
```json
{
  "compilerOptions": {
    "types": ["vitest/globals"]
  }
}
```

### "Cannot find module"

Check that:
1. Import paths are correct
2. Module is exported properly
3. File extension is `.ts` or `.tsx`

### Environment Variables Not Loading

1. Ensure `.env` file exists in project root
2. Variables must start with `VITE_` prefix
3. Restart test runner after changing `.env`

### Fetch Mock Not Working

Ensure you're using `vi.mocked(fetch)` and not just `fetch.mock`:
```typescript
// ✅ Correct
vi.mocked(fetch).mockResolvedValueOnce(...)

// ❌ Incorrect
fetch.mockResolvedValueOnce(...)
```

---

## Continuous Integration

Add to your CI/CD pipeline:

```yaml
# Example GitHub Actions
- name: Install dependencies
  run: npm ci

- name: Run tests
  run: npm run test:run

- name: Build
  run: npm run build
```

---

## Test Files

```
src/
├── test/
│   ├── setup.ts          # Test setup and mocks
│   └── api.test.ts       # API and auth tests
├── utils/
│   └── debug.ts          # Debug utilities (tested)
└── api/
    └── client.ts         # API client (tested)
```

---

## Additional Resources

- [Vitest Documentation](https://vitest.dev/)
- [Testing Library](https://testing-library.com/)
- [Vitest Mocking Guide](https://vitest.dev/guide/mocking.html)

---

*Last updated: March 8, 2026*
*Version: 1.1.0*
