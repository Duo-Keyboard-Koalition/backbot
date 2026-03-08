"use client"

import { useState, useEffect, useCallback } from "react"
import Link from "next/link"
import Image from "next/image"
import { Button } from "@/components/ui/button"
import { ChevronLeft, ChevronRight, ArrowLeft, ArrowRight, Maximize2, Minimize2 } from "lucide-react"
import type { ComicPanel } from "@/lib/.env"
import { useRouter } from "next/navigation"
import { useComicContext } from "@/app/context/comic-context"

interface ComicReaderProps {
  currentPageNumber: number
}

export function ComicReader({ currentPageNumber }: ComicReaderProps) {
  const router = useRouter()
  const [isFullscreen, setIsFullscreen] = useState(false)
  const [screenSize, setScreenSize] = useState<'mobile' | 'tablet' | 'desktop'>('desktop')
  
  // Use comic context for state management
  const {
    currentPage: page,
    currentPageNumber: pageNumber,
    currentPanelIndex,
    totalPages,
    totalPanels,
    currentPanel,
    prevPageLink,
    nextPageLink,
    isLoading,
    error,
    setCurrentPanelIndex,
  } = useComicContext()
  
  // Screen size detection
  useEffect(() => {
    const checkScreenSize = () => {
      if (window.innerWidth < 768) {
        setScreenSize('mobile')
      } else if (window.innerWidth < 1024) {
        setScreenSize('tablet')
      } else {
        setScreenSize('desktop')
      }
    }

    checkScreenSize()
    window.addEventListener('resize', checkScreenSize)
    return () => window.removeEventListener('resize', checkScreenSize)
  }, [])

  // Function to get responsive aspect ratio
  const getAspectRatio = (panel: ComicPanel): string => {
    switch (screenSize) {
      case 'mobile':
        return panel.mobileAspectRatio || panel.aspectRatio || '4/3'
      case 'tablet':
        return panel.tabletAspectRatio || panel.aspectRatio || '3/2'
      case 'desktop':
      default:
        return panel.desktopAspectRatio || panel.aspectRatio || '16/9'
    }
  }

  // Custom navigation functions that handle both panels and pages
  const handleNextPanel = useCallback(() => {
    if (!page) return
    
    if (currentPanelIndex < page.panels.length - 1) {
      setCurrentPanelIndex(currentPanelIndex + 1)
    } else if (pageNumber < totalPages) {
      // Navigate to next page
      router.push(`/comic/${pageNumber + 1}`)
    }
  }, [page, currentPanelIndex, setCurrentPanelIndex, pageNumber, totalPages, router])

  const handlePrevPanel = useCallback(() => {
    if (currentPanelIndex > 0) {
      setCurrentPanelIndex(currentPanelIndex - 1)
    } else if (pageNumber > 1) {
      // Navigate to previous page
      router.push(`/comic/${pageNumber - 1}`)
    }
  }, [currentPanelIndex, setCurrentPanelIndex, pageNumber, router])

  // Keyboard navigation
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (!page) return

      if (e.key === "ArrowLeft") {
        handlePrevPanel()
      } else if (e.key === "ArrowRight") {
        handleNextPanel()
      } else if (e.key === "f") {
        toggleFullscreen()
      }
    }

    window.addEventListener("keydown", handleKeyDown)
    return () => window.removeEventListener("keydown", handleKeyDown)
  }, [page, handlePrevPanel, handleNextPanel])

  const toggleFullscreen = () => {
    if (!document.fullscreenElement) {
      document.documentElement.requestFullscreen().catch((err) => {
        console.error(`Error attempting to enable fullscreen: ${err.message}`)
      })
      setIsFullscreen(true)
    } else {
      if (document.exitFullscreen) {
        document.exitFullscreen()
        setIsFullscreen(false)
      }
    }
  }

  const renderPanel = (panel: ComicPanel) => {
    try {
      const aspectRatio = getAspectRatio(panel)
      
      return (
        <div 
          className="relative w-full h-full flex items-center justify-center"
          style={{ aspectRatio }}
        >
          <div className="relative w-full h-full max-w-full max-h-full">
            {panel.image && (
              <Image
                src={panel.image || "/placeholder.svg"}
                alt={`Panel ${currentPanelIndex + 1}`}
                fill
                className="object-contain"
                priority
              />
            )}
            {panel.background && (
              <div
                className="absolute inset-0"
                style={{
                  background: panel.background,
                }}
              ></div>
            )}
            {panel.content && (
              <div className="absolute inset-0 flex items-center justify-center p-4">
                <div className="max-w-full">{panel.content}</div>
              </div>
            )}
            {panel.dialogue && (
              <div className="absolute bottom-4 md:bottom-8 right-4 md:right-8 max-w-[80%] z-10">
                <div className="comic-text">
                  <p className="text-sm md:text-base lg:text-lg">{panel.dialogue}</p>
                  {panel.speaker && <p className="text-xs md:text-sm lg:text-base text-cyber-blue mt-1">— {panel.speaker}</p>}
                </div>
              </div>
            )}
            {panel.narration && (
              <div className="absolute top-4 md:top-8 left-4 md:left-8 max-w-[80%] z-10">
                <div className="comic-narration">
                  <p className="text-sm md:text-base lg:text-lg">{panel.narration}</p>
                </div>
              </div>
            )}
          </div>
        </div>
      )
    } catch (error) {
      console.error(`Error rendering panel:`, error)
      return (
        <div className="relative w-full h-full bg-cyber-darker flex items-center justify-center">
          <p className="text-red-500">Error rendering panel</p>
        </div>
      )
    }
  }

  if (error) {
    return (
      <div className="w-full text-center py-16">
        <div className="text-red-500 text-xl mb-4">Error loading page</div>
        <div className="text-gray-400 mb-8">{error}</div>
        <div className="flex justify-center gap-4">
          <Link href="/">
            <Button className="bg-cyber-blue text-black hover:bg-cyber-light">
              Return Home
            </Button>
          </Link>
          {pageNumber > 1 && (
            <Button
              variant="outline"
              className="border-cyber-blue/50 hover:border-cyber-blue"
              onClick={() => router.push(`/comic/${pageNumber - 1}`)}
            >
              Previous Page
            </Button>
          )}
        </div>
      </div>
    )
  }

  if (!page || !page.panels || page.panels.length === 0) {
    return (
      <div className="w-full text-center py-16">
        <div className="text-red-500 text-xl">Error loading page. Please try again.</div>
      </div>
    )
  }

  return (
    <div className="w-full h-screen flex flex-col">
      {/* Comic Panel Slideshow - Takes full viewport height minus header */}
      <div className="relative bg-cyber-dark flex-1 pt-16 pb-24">
        {/* Full-screen panel - Uses remaining space with bottom padding for navigation */}
        <div className="relative h-full w-full overflow-hidden">
          <div className="absolute inset-0 transition-opacity duration-300 ease-in-out">
            {currentPanel && renderPanel(currentPanel)}
          </div>
        </div>

        {/* Navigation Controls */}
        <div className="absolute inset-0 flex items-center justify-between pointer-events-none">
          <Button
            variant="ghost"
            size="icon"
            className="h-16 w-16 rounded-full bg-black/30 text-white hover:bg-black/50 pointer-events-auto ml-4 nav-button"
            onClick={handlePrevPanel}
            disabled={currentPanelIndex === 0 && pageNumber === 1}
          >
            <ArrowLeft className="h-8 w-8" />
            <span className="sr-only">Previous panel</span>
          </Button>

          <Button
            variant="ghost"
            size="icon"
            className="h-16 w-16 rounded-full bg-black/30 text-white hover:bg-black/50 pointer-events-auto mr-4 nav-button"
            onClick={handleNextPanel}
            disabled={currentPanelIndex === totalPanels - 1 && pageNumber === totalPages}
          >
            <ArrowRight className="h-8 w-8" />
            <span className="sr-only">Next panel</span>
          </Button>
        </div>

        {/* Fullscreen Toggle */}
        <Button
          variant="ghost"
          size="icon"
          className="absolute top-4 right-4 h-10 w-10 rounded-full bg-black/30 text-white hover:bg-black/50 nav-button"
          onClick={toggleFullscreen}
        >
          {isFullscreen ? <Minimize2 className="h-5 w-5" /> : <Maximize2 className="h-5 w-5" />}
          <span className="sr-only">{isFullscreen ? "Exit fullscreen" : "Enter fullscreen"}</span>
        </Button>

        {/* Bottom Navigation - Absolutely positioned */}
        <div className="absolute bottom-0 left-0 right-0 bg-cyber-darker/90 backdrop-blur-sm border-t-2 border-cyber-blue/30 py-4">
          <div className="container mx-auto px-4">
            <div className="flex justify-between items-center">
              <div className="flex items-center gap-4">
                {pageNumber > 1 && (
                  <Button 
                    variant="outline" 
                    className="border-cyber-blue/50 hover:border-cyber-blue"
                    onClick={() => router.push(`/comic/${pageNumber - 1}`)}
                  >
                    <ChevronLeft className="mr-2 h-4 w-4" />
                    Previous Page
                  </Button>
                )}
              </div>

              <div className="text-center text-cyber-blue">
                Page {pageNumber} of {totalPages}
              </div>

              <div className="flex items-center gap-4">
                {pageNumber < totalPages && (
                  <Button 
                    variant="outline" 
                    className="border-cyber-blue/50 hover:border-cyber-blue"
                    onClick={() => router.push(`/comic/${pageNumber + 1}`)}
                  >
                    Next Page
                    <ChevronRight className="ml-2 h-4 w-4" />
                  </Button>
                )}
              </div>
            </div>
          </div>

          {/* Panel Indicator */}
          <div className="py-2 border-t border-cyber-blue/10">
            <div className="container mx-auto px-4">
              <div className="flex justify-center gap-2">
                {page.panels.map((_, index) => (
                  <button
                    key={index}
                    className={`w-3 h-3 rounded-full transition-colors ${
                      index === currentPanelIndex
                        ? "bg-cyber-blue"
                        : "bg-gray-600 hover:bg-gray-500"
                    }`}
                    onClick={() => setCurrentPanelIndex(index)}
                    aria-label={`Go to panel ${index + 1}`}
                  />
                ))}
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
