import Link from 'next/link'
import { Button } from '@/components/ui/button'

export default function Page() {
  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-b from-cyber-dark via-black to-cyber-dark">
      <div className="text-center space-y-4">
        <h1 className="text-3xl font-bold text-cyber-light">Authentication Removed</h1>
        <p className="text-cyber-light/70">This project now runs without sign-in.</p>
        <Link href="/">
          <Button className="bg-cyber-blue/20 border border-cyber-blue text-cyber-blue hover:bg-cyber-blue/30">
            Go to Home
          </Button>
        </Link>
      </div>
    </div>
  )
}
