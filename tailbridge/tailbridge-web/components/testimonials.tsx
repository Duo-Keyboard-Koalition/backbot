import Image from "next/image"

const testimonials = [
  {
    quote:
      "AuraMatch helped me find brands that truly align with my values. My audience immediately noticed the authenticity in my partnerships.",
    name: "Leia S.",
    role: "Lifestyle Jedi | 500K followers",
    avatar: "/placeholder.svg?height=80&width=80",
  },
  {
    quote:
      "The engagement predictions were spot on. I've seen a 40% increase in conversion rates since I started partnering only with Creeds that match my Living Force.",
    name: "Han T.",
    role: "Tech Jedi | 1.2M followers",
    avatar: "/placeholder.svg?height=80&width=80",
  },
  {
    quote:
      "As a brand, finding influencers who genuinely love our products has transformed our marketing. The Force is strong with AuraMatch's algorithm.",
    name: "Ahsoka R.",
    role: "Marketing Director at Rebel Fitness",
    avatar: "/placeholder.svg?height=80&width=80",
  },
]

export function TestimonialsSection() {
  return (
    <section className="py-20">
      <div className="container mx-auto px-4">
        <div className="text-center mb-16">
          <h2 className="text-3xl md:text-4xl font-bold mb-4">
            Voices from the <span className="text-gradient">Galaxy</span>
          </h2>
          <p className="text-xl text-gray-300 max-w-2xl mx-auto">Hear from Jedi and Creeds who found their match</p>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
          {testimonials.map((testimonial, index) => (
            <div key={index} className="aura-card">
              <div className="mb-6 text-gray-300">
                <svg className="w-8 h-8 text-aura-purple opacity-50 mb-2" fill="currentColor" viewBox="0 0 24 24">
                  <path d="M14.017 21v-7.391c0-5.704 3.731-9.57 8.983-10.609l.995 2.151c-2.432.917-3.995 3.638-3.995 5.849h4v10h-9.983zm-14.017 0v-7.391c0-5.704 3.748-9.57 9-10.609l.996 2.151c-2.433.917-3.996 3.638-3.996 5.849h3.983v10h-9.983z" />
                </svg>
                <p className="text-lg">{testimonial.quote}</p>
              </div>
              <div className="flex items-center">
                <div className="relative w-12 h-12 rounded-full overflow-hidden mr-4">
                  <Image
                    src={testimonial.avatar || "/placeholder.svg"}
                    alt={testimonial.name}
                    fill
                    className="object-cover"
                  />
                </div>
                <div>
                  <h4 className="font-medium">{testimonial.name}</h4>
                  <p className="text-sm text-gray-400">{testimonial.role}</p>
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>
    </section>
  )
}
