"use client"

import { usePathname } from "next/navigation"
import { ComicHeader } from "./comic-header"
import { useContext } from "react"
import { ComicContext } from "@/app/context/comic-context"

export function ComicHeaderWrapper() {
  const pathname = usePathname()
  const isComicPage = pathname.startsWith("/comic/")
  
  // Always call useContext at the top level - this is required by React Hooks rules
  const comicContext = useContext(ComicContext)
  
  // Determine what props to pass based on comic page status and context availability
  const headerProps = isComicPage && comicContext ? {
    comicTitle: comicContext.currentPage?.title,
    currentPanelIndex: comicContext.currentPanelIndex,
    totalPanels: comicContext.totalPanels
  } : {}

  return <ComicHeader {...headerProps} />
}
