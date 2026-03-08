import { Button } from "@/components/ui/button"
import Image from "next/image"

export function HeroSection() {
  return (
    <div className="relative pt-24 pb-16 overflow-hidden hero-gradient">
      <div className="blob blob-blue"></div>
      <div className="blob blob-purple"></div>

      <div className="container mx-auto px-4">
        <div className="flex flex-col lg:flex-row items-center">
          <div className="lg:w-1/2 mb-12 lg:mb-0">
            <h1 className="text-4xl md:text-5xl lg:text-6xl font-bold mb-6">
              Where <span className="text-gradient">Jedi</span> Meet <span className="text-gradient">The Creed</span>
            </h1>
            <p className="text-xl text-gray-300 mb-8 max-w-lg">
              AuraMatch uses the Living Force and Cosmic Force to connect influencers with brands they truly resonate
              with, creating authentic partnerships that flow naturally.
            </p>
            <div className="flex flex-col sm:flex-row gap-4">
              <Button variant="gradient" size="lg">
                Join the Order
              </Button>
              <Button variant="outline" size="lg">
                Explore the Force
              </Button>
            </div>
          </div>
          <div className="lg:w-1/2 relative">
            <div className="relative w-full h-[400px] md:h-[500px] glow rounded-lg overflow-hidden">
              <Image
                src="/placeholder.svg?height=500&width=500"
                alt="AuraMatch visualization of the Living Force"
                fill
                className="object-cover"
                priority
              />
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
