export function HowItWorks() {
  const steps = [
    {
      number: "01",
      title: "Create Your Jedi Profile",
      description: "Influencers share their content style, values, and the essence of their Living Force.",
    },
    {
      number: "02",
      title: "Analyze Your Aura",
      description: "Our algorithm maps your unique energy signature and content flow patterns.",
    },
    {
      number: "03",
      title: "Match with The Creed",
      description: "Discover brands whose Cosmic Force aligns with your authentic influence.",
    },
    {
      number: "04",
      title: "Simulate Engagement",
      description: "Preview how your audience will respond to potential partnerships before committing.",
    },
  ]

  return (
    <section className="py-20 bg-aura-dark">
      <div className="container mx-auto px-4">
        <div className="text-center mb-16">
          <h2 className="text-3xl md:text-4xl font-bold mb-4">
            The <span className="text-gradient">Jedi Path</span>
          </h2>
          <p className="text-xl text-gray-300 max-w-2xl mx-auto">
            Finding your perfect brand match is a journey through the Force
          </p>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-8">
          {steps.map((step, index) => (
            <div key={index} className="relative">
              <div className="text-5xl font-bold text-gradient mb-4">{step.number}</div>
              <h3 className="text-xl font-semibold mb-2">{step.title}</h3>
              <p className="text-gray-400">{step.description}</p>

              {index < steps.length - 1 && (
                <div
                  className="hidden lg:block absolute top-8 left-full w-full h-0.5 bg-gradient-to-r from-aura-blue to-aura-purple transform -translate-x-1/2"
                  style={{ width: "50%" }}
                ></div>
              )}
            </div>
          ))}
        </div>
      </div>
    </section>
  )
}
