# Webguide Quick Start Guide

## Prerequisites

Before running Webguide, ensure you have:

1. **Node.js 18+** installed ([Download](https://nodejs.org/))
2. **Taila2a service** running on port 8080
3. **TailFS service** running on port 8081

## Installation

```bash
# Navigate to webguide directory
cd webguide

# Install dependencies
npm install
```

## Start Tailbridge Services

In separate terminal windows:

### Terminal 1 - Start Taila2a
```bash
cd ..\taila2a
go run ./cmd/taila2a
```

### Terminal 2 - Start TailFS
```bash
cd ..\tailfs
go run ./cmd/tailfs
```

### Terminal 3 - Start Webguide
```bash
cd webguide
npm run dev
```

## Access the Dashboard

Open your browser to: **http://localhost:3000**

## What You'll See

### Dashboard Page (`/`)
- System statistics (agents, transfers, topics)
- Transfer activity chart
- Agent status distribution
- Recent active transfers

### Agents Page (`/agents`)
- All discovered agents
- Search and filter capabilities
- Agent details and services

### Transfers Page (`/transfers`)
- Active transfers with progress bars
- Transfer history
- Initiate new file transfers

### Topics Page (`/topics`)
- Topics and message counts
- Consumers and their status
- Buffer statistics

## Common Issues

### "Cannot connect to API"
Make sure Taila2a and TailFS services are running before starting Webguide.

### Port already in use
If port 3000 is busy, Vite will automatically use port 3001 or another available port. Check the terminal output.

### Dependencies fail to install
Try clearing npm cache:
```bash
npm cache clean --force
rm -rf node_modules package-lock.json
npm install
```

## Production Build

```bash
# Build for production
npm run build

# Preview production build
npm run preview
```

## Next Steps

1. Configure your agents in Taila2a
2. Start sending files with TailFS
3. Monitor everything from the Webguide dashboard!

---

For detailed documentation, see [README.md](README.md).
