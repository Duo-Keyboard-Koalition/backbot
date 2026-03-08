"use client"

import { useState, useEffect } from "react"
import Link from "next/link"
import { Github, Twitter } from "lucide-react"
import { usePathname } from "next/navigation"

export function ComicFooter() {
  const [isMinimal, setIsMinimal] = useState(false)
  const pathname = usePathname()

  // Check if we're in a comic page to use minimal footer
  useEffect(() => {
    setIsMinimal(pathname.startsWith("/comic/"))
  }, [pathname])

  if (isMinimal) {
    return (
      <footer className="border-t border-cyber-blue/20 py-2 bg-cyber-dark/80 text-center text-xs text-gray-500">
        &copy; {new Date().getFullYear()} Galatea 2.0 - A Cyberpunk Retelling
      </footer>
    )
  }

  return (
    <footer className="border-t-2 border-cyber-blue/30 py-6 bg-cyber-dark">
      <div className="container flex flex-col md:flex-row items-center justify-between gap-4">
        <div className="flex flex-col items-center md:items-start">
          <p className="text-sm text-muted-foreground">
            &copy; {new Date().getFullYear()} Galatea 2.0 - A Cyberpunk Retelling
          </p>
          <p className="text-xs text-muted-foreground mt-1">Original myth: Pygmalion and Galatea</p>
        </div>
        <div className="flex items-center gap-4">
          <Link href="#" className="text-muted-foreground hover:text-cyber-blue transition-colors">
            <Github className="h-5 w-5" />
            <span className="sr-only">GitHub</span>
          </Link>
          <Link href="#" className="text-muted-foreground hover:text-cyber-blue transition-colors">
            <Twitter className="h-5 w-5" />
            <span className="sr-only">Twitter</span>
          </Link>
        </div>
      </div>
    </footer>
  )
}
