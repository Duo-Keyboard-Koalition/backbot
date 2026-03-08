import { Button } from "@/components/ui/button"

export function CTASection() {
  return (
    <section className="py-20">
      <div className="container mx-auto px-4">
        <div className="relative overflow-hidden rounded-2xl bg-aura-dark p-8 md:p-12">
          <div className="blob blob-blue" style={{ top: "-150px", left: "-100px", opacity: "0.2" }}></div>
          <div className="blob blob-purple" style={{ bottom: "-150px", right: "-100px", opacity: "0.2" }}></div>

          <div className="relative z-10 max-w-3xl mx-auto text-center">
            <h2 className="text-3xl md:text-4xl font-bold mb-6">
              Ready to Harness the <span className="text-gradient">Force</span>?
            </h2>
            <p className="text-xl text-gray-300 mb-8">
              Join thousands of Jedi and Creeds who have found their perfect match through AuraMatch. May the Force of
              authentic partnerships be with you.
            </p>
            <div className="flex flex-col sm:flex-row justify-center gap-4">
              <Button variant="gradient" size="lg">
                Begin Your Journey
              </Button>
              <Button variant="outline" size="lg">
                Explore the Temple
              </Button>
            </div>
          </div>
        </div>
      </div>
    </section>
  )
}
