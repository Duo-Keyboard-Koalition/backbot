import { useState } from 'react';
import Link from 'next/link';
import { Menu, X } from 'lucide-react';

export function Navigation() {
  const [isOpen, setIsOpen] = useState(false);

  return (
    <header className="fixed top-0 left-0 right-0 z-50 bg-black/30 backdrop-blur-md">
      <div className="container mx-auto px-4">
        <div className="flex items-center justify-between h-16">
          <Link href="/" className="text-2xl font-cyber text-cyber-pink neon-text-pink">
            Vathsala
          </Link>
          <nav className="hidden md:flex items-center space-x-8">
            <Link href="/" className="font-mono text-cyber-light hover:text-cyber-pink transition-colors">
              Interface
            </Link>
            <Link href="/about" className="font-mono text-cyber-light hover:text-cyber-pink transition-colors">
              About
            </Link>
            <Link href="/debug" className="font-mono text-cyber-light hover:text-cyber-pink transition-colors">
              Debug
            </Link>
          </nav>
          <div className="md:hidden">
            <button onClick={() => setIsOpen(!isOpen)} className="text-cyber-light">
              {isOpen ? <X size={24} /> : <Menu size={24} />}
            </button>
          </div>
        </div>
      </div>
      {isOpen && (
        <div className="md:hidden bg-black/50">
          <nav className="flex flex-col items-center space-y-4 py-4">
            <Link href="/" onClick={() => setIsOpen(false)} className="font-mono text-cyber-light hover:text-cyber-pink transition-colors">
              Interface
            </Link>
            <Link href="/about" onClick={() => setIsOpen(false)} className="font-mono text-cyber-light hover:text-cyber-pink transition-colors">
              About
            </Link>
            <Link href="/debug" onClick={() => setIsOpen(false)} className="font-mono text-cyber-light hover:text-cyber-pink transition-colors">
              Debug
            </Link>
          </nav>
        </div>
      )}
    </header>
  );
}
