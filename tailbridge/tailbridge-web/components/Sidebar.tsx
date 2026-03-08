'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'
import {
  Home,
  Database,
  Upload,
  FolderOpen,
  ChevronLeft,
  ChevronRight,
  Settings,
  TrendingUp
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { useSidebar } from '@/contexts/SidebarContext'
import { UserSection } from './user-section'

interface NavItem {
  name: string
  href: string
  icon: any
}

const navItems: NavItem[] = [
  { name: 'Dashboard', href: '/user', icon: Home },
  { name: 'Bridges', href: '/bridges', icon: Database },
  { name: 'Messages', href: '/messages', icon: Upload },
  { name: 'My Bridges', href: '/user/bridges', icon: FolderOpen },
  { name: 'Analytics', href: '/user/analytics', icon: TrendingUp },
]

export function Sidebar() {
  const { collapsed, setCollapsed } = useSidebar()
  const pathname = usePathname()

  return (
    <div
      className={cn(
        'fixed left-0 top-0 h-screen bg-black/95 border-r border-cyber-blue/30 backdrop-blur-md z-50 transition-all duration-300 flex flex-col',
        collapsed ? 'w-20' : 'w-64'
      )}
    >
      {/* Header */}
      <div className="flex items-center justify-between p-4 border-b border-cyber-blue/30">
        {!collapsed && (
          <Link href="/" className="flex items-center gap-2">
            <Database className="w-6 h-6 text-cyber-blue" />
            <span className="font-bold bg-gradient-to-r from-cyber-blue to-cyber-purple bg-clip-text text-transparent">
              Walkie-Talkie
            </span>
          </Link>
        )}
        <button
          onClick={() => setCollapsed(!collapsed)}
          className="p-2 rounded hover:bg-cyber-blue/10 text-cyber-blue transition-colors"
        >
          {collapsed ? <ChevronRight className="w-5 h-5" /> : <ChevronLeft className="w-5 h-5" />}
        </button>
      </div>

      {/* Navigation */}
      <nav className="flex-1 p-4 space-y-2 overflow-y-auto">
        {navItems.map((item) => {
          const isActive = pathname === item.href
          const Icon = item.icon

          return (
            <Link
              key={item.href}
              href={item.href}
              className={cn(
                'flex items-center gap-3 px-3 py-2 rounded-lg transition-all',
                isActive
                  ? 'bg-gradient-to-r from-cyber-blue/20 to-cyber-purple/20 border border-cyber-blue text-cyber-blue shadow-[0_0_15px_rgba(0,255,255,0.3)]'
                  : 'text-cyber-light/70 hover:text-cyber-light hover:bg-cyber-blue/10',
                collapsed && 'justify-center'
              )}
              title={collapsed ? item.name : undefined}
            >
              <Icon className="w-5 h-5 flex-shrink-0" />
              {!collapsed && <span className="font-medium">{item.name}</span>}
            </Link>
          )
        })}
      </nav>

      {/* User Section - Fixed to bottom */}
      <div className="mt-auto p-4 border-t border-cyber-blue/30">
        <UserSection collapsed={collapsed} />
      </div>
    </div>
  )
}
