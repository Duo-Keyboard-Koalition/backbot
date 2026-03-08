# Prerequisites for Walkie-Talkie Bridge (A2A Communication)

This guide covers everything you need to set up peer-to-peer agent-to-agent communication using Tailscale tsnet bridges.

## 1. System Requirements

### Hardware
- **RAM**: Minimum 512MB per bridge instance
- **Disk**: 100MB for state storage + logs
- **Network**: Internet access for initial Tailscale connection

### Operating System
- Linux (recommended for production)
- macOS
- Windows 10/11

## 2. Software Requirements

### Go Runtime
- **Version**: Go 1.25 or later
- **Installation**:
  ```bash
  # Linux
  wget https://go.dev/dl/go1.25.0.linux-amd64.tar.gz
  sudo tar -C /usr/local -xzf go1.25.0.linux-amd64.tar.gz
  export PATH=$PATH:/usr/local/go/bin

  # macOS
  brew install go@1.25

  # Verify installation
  go version  # Should show go1.25 or later
  ```

### Tailscale Account
- Free account at https://tailscale.com
- Access to admin console for generating auth keys

## 3. Tailscale Setup

### Step 1: Create Tailscale Account
1. Go to https://tailscale.com
2. Sign up with Google, Microsoft, GitHub, or email

### Step 2: Generate Auth Keys
1. Navigate to https://login.tailscale.com/admin/settings/keys
2. Click **Generate auth key**
3. Configure key settings:
   - ✅ **Ephemeral** (recommended for bridges)
   - ✅ **Pre-authorized**
   - Set expiration (or no expiration for testing)
4. Copy the generated key (format: `tskey-auth-...`)

### Step 3: Note Your Tailnet Name
- Your tailnet name appears in the admin console
- Format: `yourname.ts.net`
- Bridge hostnames will be: `bridge-alpha.yourname.ts.net`

## 4. Network Requirements

### Ports
| Port | Purpose | Direction |
|------|---------|-----------|
| 8001 | Bridge inbound (Tailnet) | Bridge-to-Bridge |
| 8080 | Local agent endpoint | Localhost only |
| 9090 | Agent API (example) | Localhost only |

### Firewall Rules
- Allow outbound connections on all ports (Tailscale handles encryption)
- No inbound ports need to be opened (Tailscale uses NAT traversal)

## 5. Configuration Setup

### First-Time Setup (Interactive)

The bridge stores configuration in `~/.tailtalkie/config.json`.

Run the interactive setup:

```bash
cd tailscale-app
go run ./bridge init
```

You'll be prompted for:
1. **Tailscale Auth Key** - From your admin console
2. **Bridge Name** - Unique identifier (e.g., `bridge-alpha`)
3. **Local Agent URL** - Your agent's HTTP endpoint
4. **Inbound Port** - Default: 8001
5. **Local Listen Address** - Default: 127.0.0.1:8080

### Manual Configuration

Alternatively, create the config file manually:

```bash
mkdir -p ~/.tailtalkie
cat > ~/.tailtalkie/config.json <<EOF
{
  "bridge_name": "bridge-alpha",
  "auth_key": "tskey-auth-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
  "local_agent_url": "http://127.0.0.1:9090/api",
  "inbound_port": 8001,
  "peer_inbound_port": 8001,
  "local_listen": "127.0.0.1:8080",
  "state_dir": "/home/username/.tailtalkie/state"
}
EOF
```

### Config File Location

- **Path**: `~/.tailtalkie/config.json`
- **Permissions**: 0600 (owner read/write only)
- **Auto-created**: By `bridge init` command

### Config Schema

```json
{
  "bridge_name": "string (required, unique per bridge)",
  "state_dir": "string (optional, auto-generated)",
  "auth_key": "string (required)",
  "local_agent_url": "string (optional, default: http://127.0.0.1:9090/api)",
  "peer_inbound_port": "integer (optional, default: 8001)",
  "inbound_port": "integer (optional, default: 8001)",
  "local_listen": "string (optional, default: 127.0.0.1:8080)"
}
```

### No Environment Variables

The bridge does **not** use environment variables for configuration. All settings are loaded from `~/.tailtalkie/config.json`.

## 6. Local Agent Requirements

Your local agent (AI service, application, etc.) must:

### HTTP Endpoint
- Accept POST requests at `LOCAL_AGENT_URL`
- Parse JSON payload from bridge
- Return JSON response

### Example Agent (Python)
```python
from flask import Flask, request, jsonify

app = Flask(__name__)

@app.route('/api', methods=['POST'])
def handle_message():
    payload = request.json
    # Process the message
    response = {"status": "received", "data": payload}
    return jsonify(response)

if __name__ == '__main__':
    app.run(host='127.0.0.1', port=9090)
```

### Example Agent (Node.js)
```javascript
const express = require('express');
const app = express();

app.use(express.json());

app.post('/api', (req, res) => {
  const payload = req.body;
  // Process the message
  res.json({ status: 'received', data: payload });
});

app.listen(9090, '127.0.0.1');
```

## 7. Quick Verification Checklist

Before running bridges, verify:

- [ ] Go 1.25+ installed: `go version`
- [ ] Tailscale account created
- [ ] Auth key generated (ephemeral recommended)
- [ ] Ports 8001, 8080, 9090 available
- [ ] Local agent running and accessible
- [ ] State directory exists or can be created

## 8. Installation Steps

### Clone and Build
```bash
cd tailscale-app

# Download dependencies
go mod download

# Build the bridge
go build -o bridge ./bridge
```

### Initialize Configuration
```bash
# Interactive setup (creates ~/.tailtalkie/config.json)
go run ./bridge init

# Or manually create config (see section 5)
```

### Verify Setup
```bash
# Check config exists
cat ~/.tailtalkie/config.json

# Start the bridge
go run ./bridge
```

## 9. Running Multiple Bridges

Each bridge needs a unique `bridge_name` in its config.

### Option 1: Multiple Config Files

```bash
# Create config for bridge-alpha
cat > ~/.tailtalkie/config-alpha.json <<EOF
{
  "bridge_name": "bridge-alpha",
  "auth_key": "tskey-xxx",
  "inbound_port": 8001,
  "local_listen": "127.0.0.1:8080"
}
EOF

# Create config for bridge-beta
cat > ~/.tailtalkie/config-beta.json <<EOF
{
  "bridge_name": "bridge-beta",
  "auth_key": "tskey-xxx",
  "inbound_port": 8002,
  "local_listen": "127.0.0.1:8081"
}
EOF
```

### Option 2: Edit Config Between Runs

Edit `~/.tailtalkie/config.json` to change `bridge_name` and ports before each run.

## 10. Security Considerations

### Before Production

1. **Tailscale ACLs**: Configure access control lists
   - See `docs/tailscale-acl.example.json`
   - Restrict which bridges can communicate

2. **HMAC Authentication**: Enable envelope signing
   - Set `BRIDGE_HMAC_SECRET` on all trusted bridges
   - Validate signatures on inbound messages

3. **Tags**: Use Tailscale tags for bridge identification
   - Tag bridges with `tag:bridge` in ACLs
   - Enforce tag-based access policies

4. **Key Rotation**: Rotate auth keys periodically
   - Use ephemeral keys for automatic cleanup
   - Monitor key usage in admin console

## 11. Troubleshooting

### Bridge Won't Start
```bash
# Check Go version
go version  # Must be 1.25+

# Check if port is in use
lsof -i :8001
lsof -i :8080

# Check config file
cat ~/.tailtalkie/config.json

# Check state directory permissions
ls -la ~/.tailtalkie/
```

### Config File Not Found
```bash
# Run interactive setup
go run ./bridge init

# Or verify config exists
ls -la ~/.tailtalkie/config.json
```

### Can't Connect to Peer
```bash
# Verify both bridges are online
# Check bridge logs for "Listening on" messages

# Verify bridge names are unique in config
cat ~/.tailtalkie/config.json | grep bridge_name
```

### Auth Key Issues
- Ensure key is not expired
- Check key has not been revoked
- Verify key has correct permissions (ephemeral vs. regular)

### Reset Configuration
```bash
# Remove config and re-run init
rm ~/.tailtalkie/config.json
go run ./bridge init
```

## 12. Next Steps

After completing prerequisites:

1. **Quick Start**: See `README.md` for running bridges
2. **Architecture**: Read `docs/agent-communication.md` for message flow
3. **Security**: Configure ACLs in `docs/tailscale-acl.example.json`
4. **Engineering Notes**: Check `../engineering-notebook/README.md`

## Support

- Tailscale Docs: https://tailscale.com/kb/
- tsnet Reference: https://pkg.go.dev/tailscale.com/tsnet
- Issues: GitHub repository issues
