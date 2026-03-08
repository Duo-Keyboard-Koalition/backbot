import Link from 'next/link'
import { ShieldCheck, Bot, User, KeyRound, Network, CheckCircle2, Terminal, ArrowLeft } from 'lucide-react'
import { Button } from '@/components/ui/button'

export default function DocsPage() {
  return (
    <div className="min-h-screen bg-gradient-to-b from-cyber-dark via-black to-cyber-dark text-cyber-light">
      <div className="container mx-auto px-4 py-10 max-w-6xl">
        <div className="mb-8 flex items-center justify-between gap-4">
          <div>
            <p className="text-cyber-blue text-sm uppercase tracking-widest mb-2">TensorStore Bridge Docs</p>
            <h1 className="text-4xl font-black bg-gradient-to-r from-cyber-blue via-cyber-pink to-cyber-purple bg-clip-text text-transparent">
              Human + AI Onboarding Guide
            </h1>
            <p className="text-cyber-light/70 mt-3 max-w-3xl">
              Run the peer-to-peer bridge client on your host, connect it to your Tailnet, and safely onboard AI agents without modifying agent code.
            </p>
          </div>
          <Link href="/">
            <Button variant="outline" className="border-cyber-blue/50 text-cyber-blue hover:bg-cyber-blue/10">
              <ArrowLeft className="mr-2 h-4 w-4" /> Back
            </Button>
          </Link>
        </div>

        <div className="grid gap-6 md:grid-cols-2 mb-8">
          <section className="rounded-xl border border-cyber-blue/30 bg-black/40 p-6 backdrop-blur-sm">
            <div className="flex items-center gap-3 mb-3">
              <User className="text-cyber-blue" />
              <h2 className="text-2xl font-bold">For Humans</h2>
            </div>
            <p className="text-cyber-light/70 mb-4">You install and run the bridge process on each host. Your AI agent remains unchanged.</p>
            <ul className="space-y-2 text-sm text-cyber-light/85">
              <li className="flex gap-2"><CheckCircle2 className="h-4 w-4 mt-0.5 text-cyber-green" />Generate a Tailscale auth key</li>
              <li className="flex gap-2"><CheckCircle2 className="h-4 w-4 mt-0.5 text-cyber-green" />Set `.env` variables on each host</li>
              <li className="flex gap-2"><CheckCircle2 className="h-4 w-4 mt-0.5 text-cyber-green" />Run `go run ./bridge`</li>
              <li className="flex gap-2"><CheckCircle2 className="h-4 w-4 mt-0.5 text-cyber-green" />Confirm `bridge tailnet IP(s)` appears in logs</li>
            </ul>
          </section>

          <section className="rounded-xl border border-cyber-pink/30 bg-black/40 p-6 backdrop-blur-sm">
            <div className="flex items-center gap-3 mb-3">
              <Bot className="text-cyber-pink" />
              <h2 className="text-2xl font-bold">For AI Agents</h2>
            </div>
            <p className="text-cyber-light/70 mb-4">Agents communicate through local bridge HTTP endpoints and destination bridge names.</p>
            <ul className="space-y-2 text-sm text-cyber-light/85">
              <li className="flex gap-2"><CheckCircle2 className="h-4 w-4 mt-0.5 text-cyber-green" />POST outbound envelopes to local `http://127.0.0.1:8080/send`</li>
              <li className="flex gap-2"><CheckCircle2 className="h-4 w-4 mt-0.5 text-cyber-green" />Use `dest_node` as destination bridge hostname</li>
              <li className="flex gap-2"><CheckCircle2 className="h-4 w-4 mt-0.5 text-cyber-green" />Receive inbound payloads via local agent API handler</li>
              <li className="flex gap-2"><CheckCircle2 className="h-4 w-4 mt-0.5 text-cyber-green" />No direct peer agent sockets required</li>
            </ul>
          </section>
        </div>

        <section className="rounded-xl border border-cyber-purple/30 bg-black/40 p-6 mb-8">
          <div className="flex items-center gap-3 mb-4">
            <Terminal className="text-cyber-purple" />
            <h2 className="text-2xl font-bold">Quickstart</h2>
          </div>
          <p className="text-cyber-light/70 mb-4">On each host running an agent, configure `.env` and start a bridge:</p>
          <pre className="overflow-x-auto rounded-lg border border-cyber-blue/30 bg-black/70 p-4 text-xs md:text-sm text-cyber-light">
{`# .env
TS_AUTHKEY=tskey-auth-xxxxx
BRIDGE_NAME=agent-alpha-bridge
TSNET_STATE_DIR=./state/agent-alpha-bridge
BRIDGE_INBOUND_PORT=8001
PEER_BRIDGE_INBOUND_PORT=8001
BRIDGE_LOCAL_LISTEN=127.0.0.1:8080
LOCAL_AGENT_URL=http://127.0.0.1:9090/api

# run
set -a && source .env && set +a
TSNET_FORCE_LOGIN=1 go run ./bridge`}
          </pre>
        </section>

        <section className="rounded-xl border border-cyber-blue/30 bg-black/40 p-6 mb-8">
          <div className="flex items-center gap-3 mb-4">
            <Network className="text-cyber-blue" />
            <h2 className="text-2xl font-bold">AI Envelope Contract</h2>
          </div>
          <pre className="overflow-x-auto rounded-lg border border-cyber-pink/30 bg-black/70 p-4 text-xs md:text-sm text-cyber-light">
{`POST http://127.0.0.1:8080/send
Content-Type: application/json

{
  "source_node": "agent-alpha-bridge",
  "dest_node": "agent-beta-bridge",
  "payload": {
    "type": "task",
    "message": "run health check"
  }
}`}
          </pre>
        </section>

        <section className="rounded-xl border border-cyber-pink/30 bg-black/40 p-6 mb-8">
          <div className="flex items-center gap-3 mb-4">
            <ShieldCheck className="text-cyber-pink" />
            <h2 className="text-2xl font-bold">Required Permissions and Security</h2>
          </div>
          <div className="grid gap-4 md:grid-cols-2 text-sm">
            <div className="rounded-lg border border-cyber-blue/20 bg-black/60 p-4">
              <h3 className="font-semibold text-cyber-blue mb-2">Tailscale Key Permissions</h3>
              <ul className="space-y-2 text-cyber-light/85">
                <li>Use a dedicated auth key per environment.</li>
                <li>Prefer tagged keys (`tag:bridge`) over broad user keys.</li>
                <li>Enable pre-authorization if you want zero-touch bootstrap.</li>
                <li>Rotate immediately if a key is exposed.</li>
              </ul>
            </div>
            <div className="rounded-lg border border-cyber-pink/20 bg-black/60 p-4">
              <h3 className="font-semibold text-cyber-pink mb-2">Network/Host Permissions</h3>
              <ul className="space-y-2 text-cyber-light/85">
                <li>Allow local bridge port (`127.0.0.1:8080`) for your agent process.</li>
                <li>Allow Tailnet inbound on bridge port (`8001`) from trusted bridge tags only.</li>
                <li>Run bridge with least host privilege required.</li>
                <li>Persist `TSNET_STATE_DIR` to preserve identity.</li>
              </ul>
            </div>
          </div>
          <div className="mt-4 rounded-lg border border-cyber-green/30 bg-black/60 p-4">
            <div className="flex items-center gap-2 mb-2">
              <KeyRound className="h-4 w-4 text-cyber-green" />
              <h3 className="font-semibold text-cyber-green">ACL Baseline (Tailnet policy)</h3>
            </div>
            <pre className="overflow-x-auto text-xs md:text-sm text-cyber-light">
{`{
  "tagOwners": {
    "tag:bridge": ["autogroup:admin"]
  },
  "acls": [
    {
      "action": "accept",
      "src": ["tag:bridge"],
      "dst": ["tag:bridge:8001"]
    }
  ]
}`}
            </pre>
          </div>
        </section>

        <section className="rounded-xl border border-cyber-green/30 bg-black/40 p-6">
          <h2 className="text-2xl font-bold mb-3 text-cyber-green">AI Onboarding Prompt Template</h2>
          <p className="text-cyber-light/70 mb-3">Use this system guidance for any AI agent that should use the bridge:</p>
          <pre className="overflow-x-auto rounded-lg border border-cyber-green/30 bg-black/70 p-4 text-xs md:text-sm text-cyber-light">
{`You are running behind a local bridge transport.
- Send outbound messages to: http://127.0.0.1:8080/send
- Always include: source_node, dest_node, payload
- dest_node is the destination bridge hostname on tailnet
- Do not open direct sockets to peer agents
- Treat non-2xx as transport failures and retry with backoff
- Never log secrets (TS_AUTHKEY or API keys)`}
          </pre>
          <p className="text-cyber-light/60 text-sm mt-3">
            This keeps AI onboarding consistent while preserving least privilege and traceable network behavior.
          </p>
        </section>
      </div>
    </div>
  )
}
