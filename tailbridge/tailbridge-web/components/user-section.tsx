'use client'

import Link from 'next/link'
import { User } from 'lucide-react'
import { cn } from '@/lib/utils'

interface UserSectionProps {
  collapsed?: boolean
}

export function UserSection({ collapsed = false }: UserSectionProps) {
  return (
    <Link
      href="/user"
      className={cn(
        'flex items-center gap-3 cursor-pointer rounded-lg p-2 -m-2 transition-all w-full group',
        'hover:bg-cyber-blue/10 hover:border-cyber-blue/30 hover:shadow-[0_0_15px_rgba(0,255,255,0.2)]',
        collapsed && 'justify-center'
      )}
    >
      <div className="flex-shrink-0 w-10 h-10 border-2 border-cyber-blue rounded-full flex items-center justify-center group-hover:shadow-[0_0_10px_rgba(0,255,255,0.4)]">
        <User className="w-5 h-5 text-cyber-blue" />
      </div>
      {!collapsed && (
        <div className="flex-1 min-w-0">
          <p className="text-sm font-medium text-cyber-light truncate transition-colors group-hover:text-cyber-blue">
            Explorer
          </p>
          <p className="text-xs text-cyber-light/50 truncate transition-colors group-hover:text-cyber-light/80">
            Public mode
          </p>
        </div>
      )}
    </Link>
  )
}

