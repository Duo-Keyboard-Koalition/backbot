'use client'

import { useEffect, useState } from 'react'
import Link from 'next/link'
import {
  Network,
  MessageSquare,
  TrendingUp,
  Download,
  Upload,
  Plus,
  ArrowRight,
  Activity,
  Server,
  Shield
} from 'lucide-react'

interface Bridge {
  id: string
  name: string
  status: 'online' | 'offline'
  lastSeen: string
  messagesSent: number
  messagesReceived: number
  peerCount: number
}

interface BridgeStats {
  totalBridges: number
  onlineBridges: number
  totalMessagesSent: number
  totalMessagesReceived: number
  avgMessagesPerBridge: number
  mostActiveBridge: Bridge | null
}

export function DashboardHome() {
  const [bridges, setBridges] = useState<Bridge[]>([])
  const [stats, setStats] = useState<BridgeStats | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    fetchDashboardData()
  }, [])

  async function fetchDashboardData() {
    try {
      // Mock data for now - replace with actual API call
      const mockBridges: Bridge[] = [
        {
          id: '1',
          name: 'bridge-alpha',
          status: 'online',
          lastSeen: new Date().toISOString(),
          messagesSent: 42,
          messagesReceived: 38,
          peerCount: 3
        },
        {
          id: '2',
          name: 'bridge-beta',
          status: 'online',
          lastSeen: new Date().toISOString(),
          messagesSent: 38,
          messagesReceived: 42,
          peerCount: 3
        }
      ]

      setBridges(mockBridges)

      const totalBridges = mockBridges.length
      const onlineBridges = mockBridges.filter(b => b.status === 'online').length
      const totalMessagesSent = mockBridges.reduce((acc, b) => acc + b.messagesSent, 0)
      const totalMessagesReceived = mockBridges.reduce((acc, b) => acc + b.messagesReceived, 0)
      const avgMessagesPerBridge = totalBridges > 0 ? totalMessagesSent / totalBridges : 0

      const mostActiveBridge = mockBridges.length > 0
        ? mockBridges.reduce((prev, curr) =>
            (curr.messagesSent + curr.messagesReceived) > (prev.messagesSent + prev.messagesReceived) ? curr : prev
          )
        : null

      setStats({
        totalBridges,
        onlineBridges,
        totalMessagesSent,
        totalMessagesReceived,
        avgMessagesPerBridge,
        mostActiveBridge
      })
    } catch (error) {
      console.error('Error fetching dashboard data:', error)
    } finally {
      setLoading(false)
    }
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-cyber-light">Loading dashboard...</div>
      </div>
    )
  }

  return (
    <div className="space-y-8">
      {/* Welcome */}
      <div className="bg-gradient-to-r from-cyber-blue/10 to-cyber-purple/10 border border-cyber-blue/30 rounded-lg p-6">
        <h1 className="text-3xl font-bold text-cyber-light mb-2">
          Welcome to Walkie-Talkie for Bots
        </h1>
        <p className="text-cyber-light/70">
          Peer-to-peer communication system for AI agents using Tailscale tsnet bridges
        </p>
      </div>

      {/* Stats Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        <StatCard
          icon={Network}
          label="Total Bridges"
          value={stats?.totalBridges || 0}
          color="cyber-blue"
        />
        <StatCard
          icon={Server}
          label="Online Bridges"
          value={stats?.onlineBridges || 0}
          color="cyber-green"
        />
        <StatCard
          icon={MessageSquare}
          label="Messages Sent"
          value={stats?.totalMessagesSent || 0}
          color="cyber-pink"
        />
        <StatCard
          icon={Download}
          label="Messages Received"
          value={stats?.totalMessagesReceived || 0}
          color="cyber-purple"
        />
      </div>

      {/* Quick Actions */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <QuickActionCard
          title="Deploy Bridge"
          description="Start a new bridge node on your host"
          icon={Plus}
          href="/bridges/new"
          color="cyber-blue"
        />
        <QuickActionCard
          title="Send Message"
          description="Send a message to another agent via bridge"
          icon={MessageSquare}
          href="/messages/send"
          color="cyber-green"
        />
        <QuickActionCard
          title="View Logs"
          description="Monitor bridge activity and message logs"
          icon={Activity}
          href="/logs"
          color="cyber-pink"
        />
      </div>

      {/* Bridge Status */}
      <div className="bg-black/50 border border-cyber-blue/30 rounded-lg p-6">
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-2xl font-bold text-cyber-light">Your Bridges</h2>
          <Link href="/bridges">
            <Button variant="ghost" className="text-cyber-blue hover:text-cyber-blue">
              View All <ArrowRight className="w-4 h-4 ml-2" />
            </Button>
          </Link>
        </div>

        {bridges.length === 0 ? (
          <div className="text-center py-12 text-cyber-light/60">
            <h3 className="text-xl font-bold text-cyber-light mb-2">No bridges yet</h3>
            <p className="mb-4">Deploy your first bridge to start peer-to-peer communication</p>
            <Link href="/bridges/new">
              <Button className="bg-gradient-to-r from-cyber-blue to-cyber-purple text-white">
                <Plus className="w-4 h-4 mr-2" />
                Deploy Bridge
              </Button>
            </Link>
          </div>
        ) : (
          <div className="space-y-4">
            {bridges.map((bridge) => (
              <Link key={bridge.id} href={`/bridges/${bridge.id}`}>
                <div className="bg-black/50 border border-cyber-blue/30 rounded-lg p-4 hover:border-cyber-blue transition-colors">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-4">
                      <div className={`w-3 h-3 rounded-full ${
                        bridge.status === 'online' ? 'bg-cyber-green animate-pulse' : 'bg-cyber-pink'
                      }`} />
                      <div>
                        <h3 className="text-lg font-bold text-cyber-light">{bridge.name}</h3>
                        <p className="text-sm text-cyber-light/60">
                          {bridge.peerCount} peers • {bridge.messagesSent} sent • {bridge.messagesReceived} received
                        </p>
                      </div>
                    </div>
                    <div className="text-sm text-cyber-light/60">
                      {bridge.status === 'online' ? 'Active' : 'Offline'}
                    </div>
                  </div>
                </div>
              </Link>
            ))}
          </div>
        )}
      </div>

      {/* Documentation */}
      <div className="bg-black/50 border border-cyber-purple/30 rounded-lg p-6">
        <h3 className="text-xl font-bold text-cyber-light mb-2">Documentation</h3>
        <p className="text-cyber-light/70 mb-4">
          Learn how to set up and configure your bridge network
        </p>
        <div className="flex gap-4">
          <Link href="/docs/quickstart">
            <Button variant="outline" className="border-cyber-purple text-cyber-purple hover:bg-cyber-purple/10">
              Quick Start Guide
            </Button>
          </Link>
          <Link href="/docs/architecture">
            <Button variant="outline" className="border-cyber-blue text-cyber-blue hover:bg-cyber-blue/10">
              Architecture
            </Button>
          </Link>
          <Link href="/docs/security">
            <Button variant="outline" className="border-cyber-green text-cyber-green hover:bg-cyber-green/10">
              <Shield className="w-4 h-4 mr-2" />
              Security
            </Button>
          </Link>
        </div>
      </div>
    </div>
  )
}

function StatCard({ icon: Icon, label, value, color }: {
  icon: any
  label: string
  value: number
  color: string
}) {
  return (
    <div className={`bg-black/50 border border-${color}/30 rounded-lg p-6`}>
      <div className="flex items-center justify-between mb-2">
        <Icon className={`w-6 h-6 text-${color}`} />
        <TrendingUp className={`w-4 h-4 text-${color}/60`} />
      </div>
      <div className={`text-3xl font-bold text-${color}`}>{value}</div>
      <div className="text-sm text-cyber-light/70">{label}</div>
    </div>
  )
}

function QuickActionCard({
  title,
  description,
  icon: Icon,
  href,
  color
}: {
  title: string
  description: string
  icon: any
  href: string
  color: string
}) {
  return (
    <Link href={href}>
      <div className={`bg-black/50 border border-${color}/30 rounded-lg p-6 hover:border-${color} transition-colors cursor-pointer`}>
        <Icon className={`w-8 h-8 text-${color} mb-4`} />
        <h3 className="text-lg font-bold text-cyber-light mb-2">{title}</h3>
        <p className="text-sm text-cyber-light/70">{description}</p>
      </div>
    </Link>
  )
}
