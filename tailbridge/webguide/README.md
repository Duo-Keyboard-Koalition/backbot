# Webguide - Web UI for Tailbridge

**Status:** ✅ Implemented - Ready for Use

---

## Purpose

Webguide provides a modern, responsive web-based user interface for monitoring and managing the Tailbridge system (Taila2a + TailFS).

---

## Features

### Dashboard
- System overview with key metrics
- Real-time agent status (online/offline)
- Active transfers monitoring with progress bars
- Transfer activity charts (24h)
- Agent status distribution
- Message buffer health

### Debug Page (NEW!)
- **Tailscale authentication key status** - Shows if auth key is configured
- **Connection information** - API URLs, refresh interval, network status
- **API connectivity tests** - Test connections to Taila2a and TailFS APIs
- **Debug logs** - Real-time logging with export functionality
- **System information** - User agent, timestamp, network type

### Agents Page
- Phone book with all discovered agents
- Filter by online status
- Search by name, hostname, or IP
- Agent details including gateways/services
- Last seen timestamps
- Visual status indicators

### Transfers Page
- Active transfers with real-time progress
- Transfer history (last 50 transfers)
- Initiate new file transfers via UI
- Progress tracking (speed, ETA, percentage)
- Tab-based navigation (Active/History)

### Topics Page
- Topic list with message counts
- Consumer list with status
- Message buffer statistics
- Topic-consumer relationships
- Visual consumer avatars

### Settings Page
- API endpoint configuration
- Refresh interval settings
- About information
- Quick links to documentation

---

## Technology Stack

| Component | Technology |
|-----------|------------|
| Frontend Framework | React 18 + TypeScript |
| Build Tool | Vite 5 |
| UI Library | Material-UI (MUI) 5 |
| State Management | Zustand 4 |
| Routing | React Router 6 |
| Charts | Recharts 2 |
| Theme | Dark Mode (custom) |
| Testing | Vitest + Testing Library |

---

## Environment Variables

Webguide uses environment variables for configuration. Copy `.env.example` to `.env`:

```bash
cp .env.example .env
```

### Available Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `VITE_TAILSCALE_AUTH_KEY` | (required) | **REAL** Tailscale authentication key |
| `VITE_TAILSCALE_TAILNET` | `tailnet` | Your Tailnet name from Tailscale admin console |
| `VITE_TAILA2A_URL` | `http://localhost:8080` | Taila2a API endpoint |
| `VITE_TAILFS_URL` | `http://localhost:8081` | TailFS API endpoint |
| `VITE_DEBUG` | `false` | Enable debug logging (`true`/`false`) |
| `VITE_REFRESH_INTERVAL` | `30` | Data refresh interval in seconds |

### Example `.env`

```env
# REAL Tailscale Authentication Key (NO MOCKS)
VITE_TAILSCALE_AUTH_KEY=tskey-auth-k7Q11ZWj11CNTRL-FbRR2tKLRcPn5L246vsAcP7LP2YCUxWD

# Your Tailnet name (get from https://login.tailscale.com/admin/settings/general)
VITE_TAILSCALE_TAILNET=tailnet

# API Endpoints
VITE_TAILA2A_URL=http://localhost:8080
VITE_TAILFS_URL=http://localhost:8081

# Debug mode
VITE_DEBUG=true

# Refresh interval
VITE_REFRESH_INTERVAL=30
```

### Getting Your Tailscale Auth Key

1. Go to [Tailscale Admin Console](https://login.tailscale.com/admin/settings/keys)
2. Click "Generate auth key"
3. Copy the key (starts with `tskey-auth-`)
4. Paste it in your `.env` file

**Important:** The auth key is used to make **REAL API calls** to Tailscale's management API at `https://api.tailscale.com/api/v2`. No mocks - this is the real thing!

---

## Quick Start

### Prerequisites

- Node.js 18+ 
- npm or yarn
- Taila2a service running on `http://localhost:8080`
- TailFS service running on `http://localhost:8081`

### Installation

```bash
cd webguide

# Install dependencies
npm install

# Start development server
npm run dev
```

The application will be available at `http://localhost:3000`

### Build for Production

```bash
# Build optimized production bundle
npm run build

# Preview production build
npm run preview
```

---

## Directory Structure

```
webguide/
├── src/
│   ├── api/              # API client layer
│   │   ├── client.ts     # Taila2a & TailFS API clients
│   │   └── index.ts
│   ├── components/       # Reusable UI components
│   │   ├── Layout.tsx    # Main app layout with sidebar
│   │   ├── StatCard.tsx  # Statistics cards
│   │   ├── DataTable.tsx # Data table with pagination
│   │   └── index.ts
│   ├── pages/            # Page components
│   │   ├── Dashboard.tsx # Main dashboard
│   │   ├── Agents.tsx    # Agents phone book
│   │   ├── Transfers.tsx # Transfer monitoring
│   │   ├── Topics.tsx    # Topics & consumers
│   │   └── Settings.tsx  # Settings page
│   ├── stores/           # Zustand state stores
│   │   ├── agentsStore.ts
│   │   ├── transfersStore.ts
│   │   ├── topicsStore.ts
│   │   ├── dashboardStore.ts
│   │   └── index.ts
│   ├── types/            # TypeScript type definitions
│   │   └── index.ts
│   ├── hooks/            # Custom React hooks
│   ├── App.tsx           # Main app component
│   └── main.tsx          # Entry point
├── public/               # Static assets
├── index.html
├── package.json
├── tsconfig.json
├── tsconfig.node.json
├── vite.config.ts
└── README.md
```

---

## API Integration

### Proxy Configuration

The Vite dev server proxies API requests to the backend services:

```typescript
// vite.config.ts
server: {
  proxy: {
    '/api/taila2a': {
      target: 'http://localhost:8080',
      changeOrigin: true,
      rewrite: (path) => path.replace(/^\/api\/taila2a/, ''),
    },
    '/api/tailfs': {
      target: 'http://localhost:8081',
      changeOrigin: true,
      rewrite: (path) => path.replace(/^\/api\/tailfs/, ''),
    },
  },
}
```

### Taila2a Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/agents` | GET | Get all agents from phone book |
| `/agents/online` | GET | Get online agents only |
| `/topics` | GET | Get all topics |
| `/consumers` | GET | Get all consumers |
| `/buffer/stats` | GET | Get message buffer statistics |
| `/trigger/status` | GET | Get trigger status |
| `/send` | POST | Send message to agent |

### TailFS Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/transfers` | GET | Get active transfers |
| `/history` | GET | Get transfer history |
| `/progress` | GET | Get transfer progress |
| `/send` | POST | Initiate file transfer |
| `/agents` | GET | Get agents with file capability |

---

## State Management

Webguide uses Zustand for lightweight, fast state management:

### Stores

- **agentsStore**: Agent phone book data
- **transfersStore**: Active and historical transfers
- **topicsStore**: Topics, consumers, and buffer stats
- **dashboardStore**: Aggregated dashboard statistics

### Auto-Refresh

Data is automatically refreshed at intervals:
- Dashboard: Every 30 seconds
- Agents: Every 30 seconds
- Transfers: Every 5 seconds (active), 30 seconds (history)
- Topics: Every 30 seconds

---

## Customization

### Theme

The application uses a custom dark theme. To customize:

```typescript
// src/components/Layout.tsx
const theme = createTheme({
  palette: {
    mode: 'dark',
    primary: { main: '#90caf9' },
    secondary: { main: '#f48fb1' },
    // ... customize colors
  },
});
```

### API URLs

For production deployments, update the proxy configuration in `vite.config.ts` or set environment variables:

```bash
# .env
VITE_TAILA2A_URL=http://your-taila2a-host:8080
VITE_TAILFS_URL=http://your-tailfs-host:8081
```

---

## Development

### Commands

```bash
# Install dependencies
npm install

# Start dev server with hot reload
npm run dev

# Build for production
npm run build

# Type check
npx tsc --noEmit

# Lint
npm run lint

# Preview production build
npm run preview

# Run tests
npm run test

# Run tests (single run, no watch mode)
npm run test:run
```

### Testing

Webguide includes a comprehensive test suite using Vitest and Testing Library:

```bash
# Run tests in watch mode
npm run test

# Run tests once
npm run test:run

# Run tests with UI
npm run test:ui
```

#### Test Coverage

The test suite covers:
- **Tailscale Authentication** - Key validation and format checking
- **Taila2a API** - Agent discovery, buffer stats, messaging
- **TailFS API** - File transfers, progress tracking, history
- **Debug Utilities** - Connection info, network status

#### Writing Tests

Tests are located in `src/test/`. Example:

```typescript
import { describe, it, expect, vi } from 'vitest';
import { taila2aApi } from '../api/client';

describe('Taila2a API', () => {
  it('should fetch agents successfully', async () => {
    vi.mocked(fetch).mockResolvedValueOnce({
      ok: true,
      status: 200,
      json: () => Promise.resolve({ agents: [], count: 0 }),
    } as Response);

    const result = await taila2aApi.getAgents();
    expect(result.count).toBe(0);
  });
});
```

### Adding New Pages

1. Create page component in `src/pages/`
2. Add route in `src/App.tsx`
3. Add menu item in `src/components/Layout.tsx`
4. Create store if needed in `src/stores/`

---

## Deployment

### Docker (Optional)

```dockerfile
FROM node:18-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

### Nginx Configuration

```nginx
server {
    listen 80;
    server_name webguide.yourdomain.com;
    root /usr/share/nginx/html;
    index index.html;

    location / {
        try_files $uri $uri/ /index.html;
    }

    location /api/taila2a/ {
        proxy_pass http://localhost:8080/;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
    }

    location /api/tailfs/ {
        proxy_pass http://localhost:8081/;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
    }
}
```

---

## Debug Logging

Webguide includes comprehensive debug logging capabilities:

### Enabling Debug Mode

Set `VITE_DEBUG=true` in your `.env` file to enable debug logging.

### Debug Features

- **API Logging** - All API requests and responses are logged
- **System Information** - Connection info, network status, auth key status
- **Performance Monitoring** - Measure operation durations
- **Error Tracking** - Detailed error messages with context

### Debug Page

Access the debug page at `/debug` to:
- View Tailscale auth key status (masked for security)
- Test API connectivity
- View and export debug logs
- Monitor network status

### Debug Utilities

```typescript
import { 
  debugLog, 
  getConnectionInfo, 
  getNetworkStatus,
  isAuthKeyConfigured,
  validateAuthKey,
} from './utils/debug';

// Check if debug is enabled
if (isDebugEnabled()) {
  debugLog.info('Custom debug message');
  debugLog.api('/endpoint', 'GET', data);
}

// Get connection info
const info = getConnectionInfo();
console.log('Auth key configured:', info.authKeyConfigured);
console.log('Taila2a URL:', info.taila2aUrl);

// Validate auth key
const result = validateAuthKey('tskey-auth-xxx');
if (!result.valid) {
  console.error(result.error);
}
```

### Log Output

When debug mode is enabled, logs appear in the browser console with prefixes:
- `[Webguide] [LOG]` - General logs
- `[Webguide] [INFO]` - Information messages
- `[Webguide] [WARN]` - Warnings
- `[Webguide] [ERROR]` - Errors
- `[Webguide] [API]` - API requests/responses

---

## Screenshots

### Dashboard
- System stats cards (agents, transfers, topics, buffer health)
- Transfer activity line chart
- Agent status bar chart
- Recent active transfers table

### Agents
- Searchable agent cards
- Online/offline filter
- Gateway/service chips
- Last seen timestamps

### Transfers
- Active transfer cards with progress bars
- Speed and ETA indicators
- Transfer history table
- New transfer dialog

---

## Troubleshooting

### API Connection Errors

1. Ensure Taila2a and TailFS services are running
2. Check proxy configuration in `vite.config.ts`
3. Verify CORS settings on backend services

### Build Errors

```bash
# Clear node_modules and reinstall
rm -rf node_modules package-lock.json
npm install
```

### Type Errors

```bash
# Run TypeScript check
npx tsc --noEmit
```

### Tailscale Auth Key Errors

**"Auth key is missing or invalid"**

1. Ensure `VITE_TAILSCALE_AUTH_KEY` is set in `.env`
2. Key must start with `tskey-auth-`
3. Key should be at least 20 characters long
4. Restart the dev server after changing `.env`

**"API Error: 401"**

The auth key may be invalid or expired. Verify:
- Key format: `tskey-auth-XXXXXXXXXX...`
- Key has not been revoked in Tailscale admin console
- Key has appropriate permissions

**Debug Page Shows Auth Key Issues**

Visit `/debug` to:
- Check if auth key is configured
- View masked auth key (first 8 + last 4 chars)
- Test API connectivity
- View detailed error logs

---

## Future Enhancements

- [ ] WebSocket support for real-time updates
- [ ] Authentication/Authorization
- [ ] Agent control (start/stop/restart)
- [ ] Topic management UI
- [ ] Transfer scheduling
- [ ] Notifications/Alerts
- [ ] Export data (CSV, JSON)
- [ ] Mobile responsive improvements
- [ ] Multi-language support

---

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests and linting
5. Submit a pull request

---

## License

Same as the parent Tailbridge project.

---

*Last updated: March 8, 2026*
*Version: 1.1.0 - Debug & Testing Added*
