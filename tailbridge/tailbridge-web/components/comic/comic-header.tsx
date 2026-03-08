"use client"

import { useState, useEffect } from "react"
import Link from "next/link"
import { Button } from "@/components/ui/button"
import { Menu, BookOpen, X, Home } from "lucide-react"
import { usePathname } from "next/navigation"

interface ComicHeaderProps {
  comicTitle?: string
  currentPanelIndex?: number
  totalPanels?: number
}

export function ComicHeader({ comicTitle, currentPanelIndex, totalPanels }: ComicHeaderProps = {}) {
  const [isMenuOpen, setIsMenuOpen] = useState(false)
  const [isMinimal, setIsMinimal] = useState(false)
  const pathname = usePathname()

  // Check if we're in a comic page to use minimal header
  useEffect(() => {
    setIsMinimal(pathname.startsWith("/comic/"))
  }, [pathname])
  return (
    <header
      className={`${isMinimal ? 'fixed' : 'sticky'} top-0 z-50 w-full backdrop-blur-lg border-b-2 border-cyber-blue/30 transition-all duration-300 ${
        isMinimal ? "bg-cyber-dark/80 py-2" : "bg-cyber-dark/90 py-4"
      }`}
    >
      <div className="container flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Link href="/" className="flex items-center gap-2">
            <div className="relative w-8 h-8">
              <div className="absolute inset-0 bg-cyber-pink rounded-full opacity-70 blur-sm"></div>
              <div className="absolute inset-0.5 bg-cyber-dark rounded-full flex items-center justify-center">
                <span className="text-cyber-pink font-bold text-xs">G2.0</span>
              </div>
            </div>
            <span className={`font-cyber tracking-wider neon-text ${isMinimal ? "text-lg" : "text-xl"}`}>
              GALATEA 2.0
            </span>
          </Link>
        </div>        {isMinimal ? (
          // Minimal navigation for comic pages with comic title
          <div className="flex items-center gap-4 flex-1 justify-between">
            {/* Comic Title in center */}
            {comicTitle && (
              <div className="flex-1 text-center">
                <h1 className="text-lg font-cyber neon-text">{comicTitle}</h1>
                {currentPanelIndex !== undefined && totalPanels !== undefined && (
                  <div className="text-xs text-cyber-blue mt-1">
                    Panel {currentPanelIndex + 1} of {totalPanels}
                  </div>
                )}
              </div>
            )}
            
            {/* Navigation buttons */}
            <div className="flex items-center gap-2">
              <Link href="/">
                <Button variant="ghost" size="sm" className="text-gray-400 hover:text-white">
                  <Home className="h-4 w-4 mr-1" />
                  Home
                </Button>
              </Link>
              <Link href="/chapters">
                <Button variant="ghost" size="sm" className="text-gray-400 hover:text-white">
                  Chapters
                </Button>
              </Link>
            </div>
          </div>
        ) : (
          // Full navigation for other pages
          <>
            <nav className="hidden md:flex items-center gap-6">
              <Link
                href="/comic/1"
                className="text-sm font-medium text-cyber-blue hover:text-cyber-light transition-colors"
              >
                READ
              </Link>
              <Link
                href="/chapters"
                className="text-sm font-medium text-muted-foreground hover:text-cyber-light transition-colors"
              >
                CHAPTERS
              </Link>
              <Link
                href="/about"
                className="text-sm font-medium text-muted-foreground hover:text-cyber-light transition-colors"
              >
                ABOUT
              </Link>
            </nav>
            <div className="flex items-center gap-2">
              <Button variant="ghost" size="icon" className="md:hidden" onClick={() => setIsMenuOpen(!isMenuOpen)}>
                {isMenuOpen ? <X className="h-5 w-5" /> : <Menu className="h-5 w-5" />}
                <span className="sr-only">Toggle menu</span>
              </Button>
              <Link href="/comic/1" className="hidden md:block">
                <Button
                  variant="default"
                  size="sm"
                  className="bg-cyber-blue text-black hover:bg-cyber-light flex items-center gap-2"
                >
                  <BookOpen className="h-4 w-4" />
                  <span>READ NOW</span>
                </Button>
              </Link>
            </div>

            {/* Mobile menu */}
            {isMenuOpen && (
              <div className="absolute top-full left-0 right-0 bg-cyber-dark/95 border-b border-cyber-blue/30 py-4 md:hidden">
                <div className="container flex flex-col gap-4">
                  <Link
                    href="/comic/1"
                    className="text-sm font-medium text-cyber-blue hover:text-cyber-light transition-colors"
                    onClick={() => setIsMenuOpen(false)}
                  >
                    READ
                  </Link>
                  <Link
                    href="/chapters"
                    className="text-sm font-medium text-muted-foreground hover:text-cyber-light transition-colors"
                    onClick={() => setIsMenuOpen(false)}
                  >
                    CHAPTERS
                  </Link>
                  <Link
                    href="/about"
                    className="text-sm font-medium text-muted-foreground hover:text-cyber-light transition-colors"
                    onClick={() => setIsMenuOpen(false)}
                  >
                    ABOUT
                  </Link>
                </div>
              </div>
            )}
          </>
        )}
      </div>
    </header>
  )
}
