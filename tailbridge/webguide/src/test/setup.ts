/// <reference types="vitest/globals" />
import '@testing-library/jest-dom';

// Mock environment variables for tests
Object.defineProperty(import.meta, 'env', {
  value: {
    VITE_TAILSCALE_AUTH_KEY: 'tskey-auth-k7Q11ZWj11CNTRL-FbRR2tKLRcPn5L246vsAcP7LP2YCUxWD',
    VITE_TAILA2A_URL: 'http://localhost:8080',
    VITE_TAILFS_URL: 'http://localhost:8081',
    VITE_DEBUG: 'true',
    VITE_REFRESH_INTERVAL: '30',
  },
  writable: true,
  configurable: true,
});

// Mock fetch for API tests
global.fetch = vi.fn();

// Helper to create mock fetch responses
export function createMockResponse(data: unknown, status = 200, ok = true) {
  return {
    status,
    ok,
    json: () => Promise.resolve(data),
    text: () => Promise.resolve(JSON.stringify(data)),
    headers: new Headers(),
  };
}

