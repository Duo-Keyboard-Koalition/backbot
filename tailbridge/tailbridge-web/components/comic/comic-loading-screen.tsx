"use client"

import { useState, useEffect } from "react"

interface LoadingScreenProps {
  pageNumber: number
}

export function ComicLoadingScreen({ pageNumber }: LoadingScreenProps) {
  const [loadingText, setLoadingText] = useState("Initializing neural interface")
  const [loadingProgress, setLoadingProgress] = useState(0)
  const [randomTip, setRandomTip] = useState("")

  const loadingTips = [
    "Dr. Pyg Malion created 36 failed prototypes before Galatea",
    "Neo-Athens is built on the ruins of the old world",
    "The quantum neural network is the key to true AI consciousness",
    "Galatea's blue hair changes hue based on her emotional state",
    "Pygmalion Industries is the largest tech corporation in Neo-Athens",
    "The Living Force and Cosmic Force are references to another sci-fi universe",
    "The original myth of Pygmalion dates back to ancient Greece",
    "In the original myth, the goddess Aphrodite brings the statue to life",
    "This comic explores themes of creation, identity, and free will",
    "The cyberpunk genre often deals with the blurring of humanity and technology",
  ]

  // Simulate loading progress
  useEffect(() => {
    const interval = setInterval(() => {
      setLoadingProgress((prev) => {
        const next = prev + Math.random() * 15
        return next > 100 ? 100 : next
      })
    }, 500)

    return () => clearInterval(interval)
  }, [])

  // Cycle through loading text messages
  useEffect(() => {
    const messages = [
      "Initializing neural interface",
      "Establishing quantum connection",
      "Downloading consciousness data",
      "Rendering synthetic memories",
      "Calibrating visual cortex",
      "Synchronizing temporal feed",
    ]

    let index = 0
    const interval = setInterval(() => {
      index = (index + 1) % messages.length
      setLoadingText(messages[index])
    }, 2000)

    return () => clearInterval(interval)
  }, [])

  // Set a random tip
  useEffect(() => {
    const randomIndex = Math.floor(Math.random() * loadingTips.length)
    setRandomTip(loadingTips[randomIndex])
  }, [])

  return (
    <div 
      className="w-full h-full flex flex-col items-center justify-center relative overflow-hidden"
      style={{
        backgroundImage: `url("https://hebbkx1anhila5yf.public.blob.vercel-storage.com/0-Galatea-AI-ODlHATevAI2Uf4BGQSPpefgk16KyCH.png")`,
        backgroundSize: "cover",
        backgroundPosition: "center",
        backgroundRepeat: "no-repeat",
      }}
    >
      {/* Dark overlay to make text more readable */}
      <div className="absolute inset-0 bg-cyber-dark/70 z-0"></div>

      {/* Background grid effect */}
      <div
        className="absolute inset-0 z-10"
        style={{
          backgroundImage:
            "linear-gradient(to right, rgba(0, 255, 255, 0.1) 1px, transparent 1px), linear-gradient(to bottom, rgba(0, 255, 255, 0.1) 1px, transparent 1px)",
          backgroundSize: "20px 20px",
        }}
      ></div>

      {/* Animated circles */}
      <div className="absolute top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2 z-10">
        <div className="w-64 h-64 rounded-full border-4 border-cyber-blue/20 animate-pulse"></div>
        <div
          className="absolute top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2 w-48 h-48 rounded-full border-4 border-cyber-blue/30 animate-pulse"
          style={{ animationDelay: "0.5s" }}
        ></div>
        <div
          className="absolute top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2 w-32 h-32 rounded-full border-4 border-cyber-blue/40 animate-pulse"
          style={{ animationDelay: "1s" }}
        ></div>
      </div>

      {/* Loading text */}
      <div className="z-20 text-center">
        <h2 className="text-3xl font-cyber neon-text mb-2">LOADING COMIC</h2>
        <p className="text-cyber-blue mb-6">Page {pageNumber}</p>

        <div className="w-64 h-2 bg-gray-800 rounded-full mb-4 overflow-hidden">
          <div
            className="h-full bg-gradient-to-r from-cyber-blue to-cyber-pink rounded-full transition-all duration-300 ease-out"
            style={{ width: `${loadingProgress}%` }}
          ></div>
        </div>

        <p className="text-gray-400 mb-8 flex items-center">
          <span className="inline-block w-2 h-2 bg-cyber-blue rounded-full mr-2 animate-pulse"></span>
          {loadingText}...
        </p>


      </div>

      {/* Scan line effect */}
      <div className="absolute inset-0 scan-line z-30 pointer-events-none"></div>

      {/* Glitch effect */}
      <div className="absolute top-0 left-0 right-0 h-1 bg-cyber-blue/50 z-30 glitch-line"></div>
    </div>
  )
}
