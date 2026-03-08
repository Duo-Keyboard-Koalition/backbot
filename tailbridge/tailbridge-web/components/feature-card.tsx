import type { ReactNode } from "react"

interface FeatureCardProps {
  icon: ReactNode
  title: string
  description: string
}

export default function FeatureCard({ icon, title, description }: FeatureCardProps) {
  return (
    <div className="bg-[#1a1025] rounded-lg p-8 flex flex-col items-center text-center">
      <div className="text-purple-500 mb-4">{icon}</div>
      <h3 className="text-xl font-semibold text-purple-500 mb-3">{title}</h3>
      <p className="text-white/70">{description}</p>
    </div>
  )
}
