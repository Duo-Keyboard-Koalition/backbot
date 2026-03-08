"use client"

import { useState } from "react"
import Link from "next/link"
import { Logo } from "@/components/logo"
import { Button } from "@/components/ui/button"
import { Menu, X } from "lucide-react"

export function Navbar() {
  const [isMenuOpen, setIsMenuOpen] = useState(false)

  return (
    <nav className="fixed top-0 left-0 right-0 z-50 bg-aura-darker/80 backdrop-blur-md">
      <div className="container mx-auto px-4 py-4">
        <div className="flex items-center justify-between">
          <Link href="/" className="flex items-center space-x-2">
            <Logo />
            <span className="text-xl font-bold text-gradient">AuraMatch</span>
          </Link>

          {/* Desktop Navigation */}
          <div className="hidden md:flex items-center space-x-8">
            <Link href="/jedi" className="text-gray-300 hover:text-white transition-colors">
              For Jedi
            </Link>
            <Link href="/creed" className="text-gray-300 hover:text-white transition-colors">
              For The Creed
            </Link>
            <Link href="/force" className="text-gray-300 hover:text-white transition-colors">
              The Force
            </Link>
            <Link href="/academy" className="text-gray-300 hover:text-white transition-colors">
              Academy
            </Link>
          </div>

          <div className="hidden md:flex items-center space-x-4">
            <Button variant="ghost" className="text-gray-300 hover:text-white">
              Enter Temple
            </Button>
            <Button variant="gradient">Join the Order</Button>
          </div>

          {/* Mobile Menu Button */}
          <button className="md:hidden text-gray-300 hover:text-white" onClick={() => setIsMenuOpen(!isMenuOpen)}>
            {isMenuOpen ? <X size={24} /> : <Menu size={24} />}
          </button>
        </div>

        {/* Mobile Navigation */}
        {isMenuOpen && (
          <div className="md:hidden pt-4 pb-2 space-y-4">
            <Link
              href="/jedi"
              className="block py-2 text-gray-300 hover:text-white transition-colors"
              onClick={() => setIsMenuOpen(false)}
            >
              For Jedi
            </Link>
            <Link
              href="/creed"
              className="block py-2 text-gray-300 hover:text-white transition-colors"
              onClick={() => setIsMenuOpen(false)}
            >
              For The Creed
            </Link>
            <Link
              href="/force"
              className="block py-2 text-gray-300 hover:text-white transition-colors"
              onClick={() => setIsMenuOpen(false)}
            >
              The Force
            </Link>
            <Link
              href="/academy"
              className="block py-2 text-gray-300 hover:text-white transition-colors"
              onClick={() => setIsMenuOpen(false)}
            >
              Academy
            </Link>
            <div className="pt-2 flex flex-col space-y-2">
              <Button variant="ghost" className="justify-center text-gray-300 hover:text-white">
                Enter Temple
              </Button>
              <Button variant="gradient" className="justify-center">
                Join the Order
              </Button>
            </div>
          </div>
        )}
      </div>
    </nav>
  )
}
