import { Zap, Shield, BarChart, Users } from "lucide-react"

const features = [
  {
    icon: <Zap className="h-8 w-8 text-aura-blue" />,
    title: "Living Force Analysis",
    description: "Our algorithm analyzes an influencer's aura to find brands that align with their true essence.",
  },
  {
    icon: <BarChart className="h-8 w-8 text-aura-purple" />,
    title: "Engagement Predictions",
    description: "Simulate how followers will react to potential partnerships before they happen.",
  },
  {
    icon: <Users className="h-8 w-8 text-aura-blue" />,
    title: "Authentic Connections",
    description: "Connect Jedi with Creeds that truly resonate with their personal values and audience vibe.",
  },
  {
    icon: <Shield className="h-8 w-8 text-aura-purple" />,
    title: "Force Protection",
    description: "Our platform ensures all partnerships maintain balance in the Force, preserving authenticity.",
  },
]

export function FeaturesSection() {
  return (
    <section className="py-20">
      <div className="container mx-auto px-4">
        <div className="text-center mb-16">
          <h2 className="text-3xl md:text-4xl font-bold mb-4">
            The Power of the <span className="text-gradient">Force</span>
          </h2>
          <p className="text-xl text-gray-300 max-w-2xl mx-auto">
            Our platform goes beyond superficial metrics to create partnerships that flow with the cosmic energy.
          </p>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-8">
          {features.map((feature, index) => (
            <div key={index} className="aura-card">
              <div className="mb-4">{feature.icon}</div>
              <h3 className="text-xl font-semibold mb-2">{feature.title}</h3>
              <p className="text-gray-400">{feature.description}</p>
            </div>
          ))}
        </div>
      </div>
    </section>
  )
}
