"use client"

import { useState } from "react"
import { Button } from "@/components/ui/button"
import { ThumbsUp, MessageCircle, Share2 } from "lucide-react"
import Image from "next/image"

export function SimulationPreview() {
  const [activeTab, setActiveTab] = useState("engagement")

  return (
    <section className="py-20">
      <div className="container mx-auto px-4">
        <div className="text-center mb-16">
          <h2 className="text-3xl md:text-4xl font-bold mb-4">
            Foresee the <span className="text-gradient">Future</span>
          </h2>
          <p className="text-xl text-gray-300 max-w-2xl mx-auto">
            Our Force Vision technology simulates how your audience will react to potential partnerships
          </p>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-2 gap-12 items-center">
          <div className="aura-card">
            <div className="flex items-center mb-4">
              <div className="relative w-12 h-12 rounded-full overflow-hidden mr-3">
                <Image src="/placeholder.svg?height=48&width=48" alt="Jedi Profile" fill className="object-cover" />
              </div>
              <div>
                <h4 className="font-medium">Rey Skywalker</h4>
                <p className="text-sm text-gray-400">Fitness & Wellness Jedi</p>
              </div>
            </div>

            <p className="mb-4">
              I've been using @RebelFitness gear for my morning training sessions and the quality is incredible! The
              Force flows so naturally when I'm wearing their performance line. #ad #RebelFit
            </p>

            <div className="relative h-64 w-full rounded-lg overflow-hidden mb-4">
              <Image src="/placeholder.svg?height=256&width=512" alt="Post content" fill className="object-cover" />
            </div>

            <div className="flex justify-between text-gray-400 mb-2">
              <div className="flex items-center">
                <ThumbsUp size={18} className="mr-1" />
                <span>4.2K</span>
              </div>
              <div className="flex items-center">
                <MessageCircle size={18} className="mr-1" />
                <span>156</span>
              </div>
              <div className="flex items-center">
                <Share2 size={18} className="mr-1" />
                <span>89</span>
              </div>
            </div>

            <div className="text-sm text-gray-500">Predicted engagement: 38% higher than average</div>
          </div>

          <div>
            <div className="flex mb-6 bg-aura-dark rounded-lg p-1">
              <button
                className={`flex-1 py-2 px-4 rounded-md transition ${activeTab === "engagement" ? "bg-gradient-to-r from-aura-blue to-aura-purple text-white" : "text-gray-400"}`}
                onClick={() => setActiveTab("engagement")}
              >
                Engagement
              </button>
              <button
                className={`flex-1 py-2 px-4 rounded-md transition ${activeTab === "sentiment" ? "bg-gradient-to-r from-aura-blue to-aura-purple text-white" : "text-gray-400"}`}
                onClick={() => setActiveTab("sentiment")}
              >
                Sentiment
              </button>
              <button
                className={`flex-1 py-2 px-4 rounded-md transition ${activeTab === "conversion" ? "bg-gradient-to-r from-aura-blue to-aura-purple text-white" : "text-gray-400"}`}
                onClick={() => setActiveTab("conversion")}
              >
                Conversion
              </button>
            </div>

            <div className="aura-card">
              <h3 className="text-xl font-semibold mb-4">Force Alignment Score: 92%</h3>
              <p className="text-gray-300 mb-6">
                This partnership has strong alignment with your Living Force signature and your audience's Cosmic Force.
              </p>

              <div className="space-y-4">
                <div>
                  <div className="flex justify-between mb-1">
                    <span className="text-sm text-gray-400">Authenticity</span>
                    <span className="text-sm text-gray-400">95%</span>
                  </div>
                  <div className="w-full bg-gray-700 rounded-full h-2">
                    <div
                      className="bg-gradient-to-r from-aura-blue to-aura-purple h-2 rounded-full"
                      style={{ width: "95%" }}
                    ></div>
                  </div>
                </div>

                <div>
                  <div className="flex justify-between mb-1">
                    <span className="text-sm text-gray-400">Audience Resonance</span>
                    <span className="text-sm text-gray-400">88%</span>
                  </div>
                  <div className="w-full bg-gray-700 rounded-full h-2">
                    <div
                      className="bg-gradient-to-r from-aura-blue to-aura-purple h-2 rounded-full"
                      style={{ width: "88%" }}
                    ></div>
                  </div>
                </div>

                <div>
                  <div className="flex justify-between mb-1">
                    <span className="text-sm text-gray-400">Brand Alignment</span>
                    <span className="text-sm text-gray-400">94%</span>
                  </div>
                  <div className="w-full bg-gray-700 rounded-full h-2">
                    <div
                      className="bg-gradient-to-r from-aura-blue to-aura-purple h-2 rounded-full"
                      style={{ width: "94%" }}
                    ></div>
                  </div>
                </div>
              </div>

              <div className="mt-6">
                <Button variant="gradient" className="w-full">
                  View Detailed Force Analysis
                </Button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>
  )
}
