import Link from "next/link"
import { Logo } from "@/components/logo"
import { Facebook, Twitter, Instagram, Linkedin } from "lucide-react"

export function Footer() {
  return (
    <footer className="bg-aura-dark py-12 mt-20">
      <div className="container mx-auto px-4">
        <div className="grid grid-cols-1 md:grid-cols-4 gap-8">
          <div className="space-y-4">
            <Link href="/" className="flex items-center space-x-2">
              <Logo size={32} />
              <span className="text-xl font-bold text-gradient">AuraMatch</span>
            </Link>
            <p className="text-gray-400 text-sm">
              Connecting Jedi with The Creed through the power of the Living Force and Cosmic Force. Creating authentic
              partnerships that resonate with audiences.
            </p>
            <div className="flex space-x-4">
              <a href="#" className="text-gray-400 hover:text-white transition-colors">
                <Facebook size={20} />
              </a>
              <a href="#" className="text-gray-400 hover:text-white transition-colors">
                <Twitter size={20} />
              </a>
              <a href="#" className="text-gray-400 hover:text-white transition-colors">
                <Instagram size={20} />
              </a>
              <a href="#" className="text-gray-400 hover:text-white transition-colors">
                <Linkedin size={20} />
              </a>
            </div>
          </div>

          <div>
            <h3 className="text-white font-medium mb-4">The Order</h3>
            <ul className="space-y-2">
              <li>
                <Link href="/about" className="text-gray-400 hover:text-white transition-colors">
                  Our Mission
                </Link>
              </li>
              <li>
                <Link href="/jedi-council" className="text-gray-400 hover:text-white transition-colors">
                  Jedi Council
                </Link>
              </li>
              <li>
                <Link href="/holocron" className="text-gray-400 hover:text-white transition-colors">
                  Holocron
                </Link>
              </li>
              <li>
                <Link href="/careers" className="text-gray-400 hover:text-white transition-colors">
                  Join Our Temple
                </Link>
              </li>
            </ul>
          </div>

          <div>
            <h3 className="text-white font-medium mb-4">The Force</h3>
            <ul className="space-y-2">
              <li>
                <Link href="/living-force" className="text-gray-400 hover:text-white transition-colors">
                  Living Force
                </Link>
              </li>
              <li>
                <Link href="/cosmic-force" className="text-gray-400 hover:text-white transition-colors">
                  Cosmic Force
                </Link>
              </li>
              <li>
                <Link href="/flow" className="text-gray-400 hover:text-white transition-colors">
                  Flow Analysis
                </Link>
              </li>
              <li>
                <Link href="/academy" className="text-gray-400 hover:text-white transition-colors">
                  Jedi Academy
                </Link>
              </li>
            </ul>
          </div>

          <div>
            <h3 className="text-white font-medium mb-4">The Code</h3>
            <ul className="space-y-2">
              <li>
                <Link href="/terms" className="text-gray-400 hover:text-white transition-colors">
                  Jedi Code
                </Link>
              </li>
              <li>
                <Link href="/privacy" className="text-gray-400 hover:text-white transition-colors">
                  Force Privacy
                </Link>
              </li>
              <li>
                <Link href="/cookies" className="text-gray-400 hover:text-white transition-colors">
                  Holocron Storage
                </Link>
              </li>
              <li>
                <Link href="/accessibility" className="text-gray-400 hover:text-white transition-colors">
                  Universal Access
                </Link>
              </li>
            </ul>
          </div>
        </div>

        <div className="border-t border-gray-800 mt-12 pt-8 text-center text-gray-500 text-sm">
          <p>&copy; {new Date().getFullYear()} AuraMatch. May the Force be with you.</p>
        </div>
      </div>
    </footer>
  )
}
