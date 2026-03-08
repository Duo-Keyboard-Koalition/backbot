'use client'

import { usePathname } from 'next/navigation'
import { Sidebar } from './Sidebar'
import { SidebarProvider, useSidebar } from '@/contexts/SidebarContext'

function LayoutContent({ children }: { children: React.ReactNode }) {
  const pathname = usePathname()
  const { collapsed } = useSidebar()

  const showSidebar =
    pathname?.startsWith('/user') || pathname?.startsWith('/upload')

  return (
    <div className="relative">
      {showSidebar && <Sidebar />}
      <main
        className="transition-all duration-300 min-h-screen"
        style={{
          marginLeft: showSidebar ? (collapsed ? '80px' : '256px') : '0',
        }}
      >
        {children}
      </main>
    </div>
  )
}

export function LayoutWrapper({ children }: { children: React.ReactNode }) {
  return (
    <SidebarProvider>
      <LayoutContent>{children}</LayoutContent>
    </SidebarProvider>
  )
}
