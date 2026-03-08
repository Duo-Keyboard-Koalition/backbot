// This is a mock implementation for demonstration purposes
// In a real application, this would connect to an AI service

interface BuildResult {
  response: string
  buildImage: string
}

// Sample build responses
const buildResponses = [
  {
    keywords: ["castle", "medieval", "fortress"],
    response:
      "I've created a medieval castle for you! It features stone walls, four towers with flags, a central keep, and a moat with a working drawbridge. The castle has multiple interior rooms including a great hall, bedrooms, and a dungeon. I've used stone bricks for the main structure, oak wood for details, and blue wool for the flags.",
    buildImage: "/generated-castle.jpg",
  },
  {
    keywords: ["house", "modern", "home"],
    response:
      "Here's your modern house! It has a sleek design with large glass windows, a flat roof with a garden, and a minimalist interior. I've included a swimming pool in the backyard, a garage, and an open-concept living area. The main materials are concrete, glass, and quartz for that contemporary look.",
    buildImage: "/generated-house.jpg",
  },
  {
    keywords: ["tree", "forest", "nature"],
    response:
      "I've built a magical treehouse for you! It's constructed around a massive oak tree with multiple platforms connected by spiral staircases. The house features wooden walls with large windows, hanging gardens with vines and flowers, and glowstone lanterns for lighting. There's even a small waterfall flowing from one of the branches!",
    buildImage: "/generated-treehouse.jpg",
  },
  {
    keywords: ["sci-fi", "futuristic", "spaceship", "space"],
    response:
      "Your futuristic spaceship is ready! It has a streamlined hull with glowing blue engines, a bridge with control panels, crew quarters, and a cargo bay. I've used a combination of quartz, iron blocks, and blue stained glass for the design, with sea lanterns and redstone lamps for the lighting effects.",
    buildImage: "/generated-spaceship.jpg",
  },
  {
    keywords: ["underwater", "ocean", "sea"],
    response:
      "I've created an underwater base as requested! It features multiple glass domes connected by tunnels, with a central hub and specialized rooms. There's a submarine dock for entry, research labs, living quarters, and an observation dome. The structure is built with prismarine, sea lanterns, and glass, with coral and seaweed gardens surrounding it.",
    buildImage: "/generated-underwater.jpg",
  },
]

// Default fallback response
const defaultResponse = {
  response:
    "I've created your build based on your description! The structure features a balanced design with appropriate materials and details to match your request. I've included interior spaces and decorative elements to make it feel complete.",
  buildImage: "/generated-default.jpg",
}

export async function generateBuild(prompt: string): Promise<BuildResult> {
  // Simulate processing time
  await new Promise((resolve) => setTimeout(resolve, 2000))

  // Convert prompt to lowercase for easier matching
  const promptLower = prompt.toLowerCase()

  // Find a matching response based on keywords
  const matchedResponse = buildResponses.find((response) =>
    response.keywords.some((keyword) => promptLower.includes(keyword)),
  )

  // Return matched response or default
  return matchedResponse || defaultResponse
}
