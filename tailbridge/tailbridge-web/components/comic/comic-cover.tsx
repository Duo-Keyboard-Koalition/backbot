"use client"

import { useEffect } from "react"
import Link from "next/link"
import Image from "next/image"
import { Button } from "@/components/ui/button"
import { ChevronRight } from "lucide-react"
import { useComicContext } from "@/app/context/comic-context"

export function ComicCover() {
  const { coverData, isCoverLoading, fetchCoverData } = useComicContext()

  useEffect(() => {
    fetchCoverData()
  }, [fetchCoverData])

  if (isCoverLoading) {
    return (
      <div className="max-w-4xl w-full h-[600px] flex items-center justify-center">
        <div className="text-cyber-blue animate-pulse">Loading...</div>
      </div>
    )
  }

  if (!coverData) {
    return (
      <div className="max-w-4xl w-full">
        <div className="text-center text-red-500">Failed to load comic cover</div>
      </div>
    )
  }

  return (
    <div className="max-w-4xl w-full">
      <div className="relative aspect-[2/3] w-full overflow-hidden rounded-lg border-2 border-cyber-blue/50 shadow-lg">
        <div className="absolute inset-0 bg-gradient-to-b from-transparent via-transparent to-cyber-dark/90 z-10"></div>
        <Image src={coverData.image || "/placeholder.svg"} alt="Galatea 2.0" fill className="object-cover" priority />
        <div className="absolute inset-0 flex flex-col justify-end p-6 z-20">
          <div className="mb-4">
            <h1 className="font-cyber text-4xl md:text-6xl mb-2 glitch" data-text={coverData.title}>
              <span className="neon-text">{coverData.title}</span>
            </h1>
            <p className="text-lg md:text-xl text-cyber-blue max-w-lg">{coverData.subtitle}</p>
          </div>
          <div className="flex flex-wrap gap-3 mb-6">
            {coverData.tags.map((tag, index) => (
              <span
                key={index}
                className={`px-3 py-1 ${index % 2 === 0 ? "bg-cyber-blue/20 border-cyber-blue/50" : "bg-cyber-pink/20 border-cyber-pink/50"} border rounded-full text-xs`}
              >
                {tag}
              </span>
            ))}
          </div>
          <Link href="/comic/1">
            <Button className="bg-cyber-blue hover:bg-cyber-light text-black font-bold group">
              BEGIN READING
              <ChevronRight className="ml-2 h-4 w-4 transition-transform group-hover:translate-x-1" />
            </Button>
          </Link>
        </div>
      </div>
      <div className="mt-8 text-center">
        <h2 className="text-xl font-bold mb-4 text-cyber-blue">SYNOPSIS</h2>
        {coverData.synopsis.map((paragraph, index) => (
          <p key={index} className="text-gray-300 mb-4 last:mb-0">
            {paragraph}
          </p>
        ))}
      </div>
    </div>
  )
}
