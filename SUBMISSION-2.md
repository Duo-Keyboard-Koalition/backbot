# Hack Canada 2026 - Submission Document

## Project: SentinelAI

**Team Name:** [Your Team Name]  
**Track:** [Primary Track] + [Secondary Tracks]  
**Demo Link:** [Live Demo URL]  
**GitHub Repo:** https://github.com/[your-username]/sentinelai  
**Pitch Video:** [3-5 min video link]

---

## 🎯 Recommended Prize Tracks

### PRIMARY TRACK (Highest Priority)

#### 1. **Most Technically Complex AI Hack** 🥇
**Why This Is Your Best Fit:**

Your project demonstrates **agentic AI architecture** with sophisticated multi-layer systems:

- **Agentic Architecture:** SentinelAI implements a full agent loop with:
  - ReAct-style reasoning (Thought → Action → Observation)
  - Tool execution framework (web_search, file I/O, calculate)
  - Conversation history management for context retention

- **Advanced Risk Assessment System:**
  - 4 failure type detectors (Loop, Goal Drift, Low Confidence, Tool Incoherence)
  - Weighted risk scoring algorithm
  - Semantic drift detection using word overlap analysis
  - Hedging language detection for confidence scoring

- **Multi-Strategy Intervention System:**
  - **Reprompt:** Contextual corrective prompts
  - **Rollback:** State recovery mechanism
  - **Goal Decomposition:** Automatic task breakdown into sub-goals
  - **Halt:** Safety mechanism for high-risk scenarios

- **Stateful Execution Management:**
  - Step tracking with failure annotations
  - Intervention history
  - Execution state persistence

**Technical Depth Highlights:**
```
- Custom risk scoring: Weighted multi-factor algorithm
- Semantic analysis: Word overlap + hedge word detection
- Adaptive interventions: Context-aware failure recovery
- Multi-agent potential: Tailbridge integration for distributed agents
```

**Prize:** $500 (1st) | $200 (2nd) | $100 (3rd)  
**Bonus:** 1st place gets interview for $280k/year job at stealth AI lab

---

### SECONDARY TRACKS (Strong Contenders)

#### 2. **Tailscale Integration Challenge** 🥈
**Why You Qualify:**

Your **Tailbridge** subproject is built specifically for Tailscale:

- **Taila2a (Agent-to-Agent Communication):**
  - Tailscale-based secure transport
  - Phone book agent discovery
  - Topic-based pub/sub messaging
  - Buffer-triggered agent activation

- **TailFS (Secure File Transfer):**
  - Chunked file transfers over Tailscale
  - Progress tracking + resume support
  - End-to-end encryption
  - Cross-device file sharing

- **Webguide Dashboard:**
  - React + TypeScript UI
  - Real-time metrics visualization
  - Agent status monitoring
  - File transfer tracking

**What You Need to Verify:**
- [ ] At least 2 devices connected in a tailnet
- [ ] ACL restricting access to one service
- [ ] Either: Tailscale Funnel (public endpoint) OR Tailscale Serve (private service)

**Prize:** Raspberry Pi 4 for each team member

---

#### 3. **Best Use of Gemini API** 🥉
**Why You Qualify:**

Your `.env.example` shows Gemini integration:
```env
GEMINI_API_KEY=your_gemini_api_key_here
GEMINI_FLASH_MODEL=gemini-3-flash
GEMINI_PRO_MODEL=gemini-3-1-pro
```

**What to Highlight:**
- Agent reasoning powered by Gemini models
- Natural language tool selection
- Context-aware response generation
- Multi-turn conversation handling

**Prize:** Google Swag Kits

---

### OPTIONAL TRACKS (If Time Permits)

#### 4. **Backboard.io - Best Use of Backboard**
**Potential Fit If:**
- You integrate Backboard's multi-agent orchestration
- Use stateful memory for long-running agent sessions
- Implement semantic recall across agent conversations

**Your Angle:** SentinelAI's intervention system could leverage Backboard's 17,000+ model hot-swapping for adaptive failure recovery strategies.

**Prize:** $500 Cash

---

#### 5. **MLH x ElevenLabs - Best Project Built with ElevenLabs**
**Potential Fit If:**
- Add voice output to agent responses
- Create audio notifications for interventions
- Build voice-enabled agent interaction

**Your Angle:** "SentinelAI speaks" - Voice alerts when risk scores breach threshold, audio summaries of agent reasoning.

**Prize:** 6 months ElevenLabs Scale tier ($1980 value) + Wireless Earbuds

---

#### 6. **Best Use of Auth0**
**Potential Fit If:**
- Add Auth0 authentication to Webguide dashboard
- Implement social sign-in for agent access
- Use MFA for sensitive file transfers

**Your Angle:** Secure agent-to-agent communication with Auth0 identity verification.

**Prize:** Wireless Headphones

---

## 📋 Project Overview

### Problem Statement
AI agents often fail silently, repeating mistakes, drifting from goals, or producing incoherent outputs. SentinelAI provides **real-time risk assessment and adaptive intervention** to keep agents on track.

### Solution
A monitoring layer that:
1. **Scores** each agent step for risk (0.0 - 1.0)
2. **Detects** 4 failure types automatically
3. **Intervenes** with context-aware strategies
4. **Tracks** all interventions for auditability

### Target Users
- Developers building agentic AI systems
- Teams deploying autonomous agents in production
- Researchers studying agent failure modes

### Technical Architecture
```
┌─────────────────────────────────────────────────────────┐
│                    SentinelAI Core                       │
├─────────────────────────────────────────────────────────┤
│  Agent Loop (ReAct)  →  Risk Scorer  →  Intervention    │
│  - Thought/Action    - Loop Detection   - Reprompt      │
│  - Tool Execution    - Goal Drift       - Rollback      │
│  - Observation       - Confidence       - Decompose     │
│                      - Coherence        - Halt          │
└─────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────┐
│                    Tailbridge Layer                      │
├─────────────────────────────────────────────────────────┤
│  Taila2a (A2A)  |  TailFS (Files)  |  Webguide (UI)    │
│  Tailscale-secured agent communication                   │
└─────────────────────────────────────────────────────────┘
```

### Key Features
- **Weighted Risk Algorithm:** Multi-factor scoring (Loop 35%, Drift 30%, Confidence 20%, Coherence 15%)
- **Semantic Drift Detection:** Word overlap analysis between goal and steps
- **Hedge Word Analysis:** Confidence scoring via uncertainty language detection
- **Adaptive Interventions:** 4 strategies matched to failure severity
- **Distributed Agent Support:** Tailscale-secured multi-agent communication
- **Real-time Dashboard:** React + Material-UI monitoring interface

---

## 🏆 Judging Criteria Alignment

### For "Most Technically Complex AI Hack"

| Criterion | How SentinelAI Delivers |
|-----------|------------------------|
| **Technical Complexity** | Multi-layer risk assessment, semantic analysis, adaptive interventions |
| **Agentic Architecture** | Full ReAct agent loop with tool execution, state management, intervention system |
| **Innovation** | Proactive failure detection + recovery (not just monitoring) |
| **Execution** | Working backend + dashboard + distributed agent protocol |
| **Scalability** | Tailscale enables cross-device, cross-network agent coordination |

---

## 📝 Submission Checklist

### Required Deliverables
- [ ] Live demo link (deployed Webguide dashboard)
- [ ] 3-5 minute pitch video
- [ ] GitHub repository with README
- [ ] This SUBMISSION.md file

### For Technical Complexity Track
- [ ] Highlight agentic architecture in pitch
- [ ] Demo risk scoring in action (show failure detection)
- [ ] Show intervention strategies working
- [ ] Explain weighted algorithm design

### For Tailscale Track
- [ ] Verify 2+ devices in tailnet
- [ ] Show ACL configuration
- [ ] Demo Funnel (public) OR Serve (private) endpoint
- [ ] Record cross-device communication

### For Gemini API Track
- [ ] Show Gemini-powered agent reasoning
- [ ] Highlight model selection (Flash vs Pro)
- [ ] Demo natural language tool selection

---

## 🎤 Pitch Script Outline (3-5 minutes)

**[0:00-0:30] Hook**
> "AI agents fail. They loop, drift from goals, and make mistakes. What if agents had a safety net?"

**[0:30-1:30] Problem + Solution**
> Show agent failure example → Introduce SentinelAI's risk scoring → Demo intervention in action

**[1:30-2:30] Technical Deep Dive**
> Explain the 4 failure detectors → Show weighted algorithm → Demo goal decomposition

**[2:30-3:30] Tailbridge Integration**
> Show multi-agent communication over Tailscale → Demo file transfer → Show dashboard

**[3:30-4:00] Impact + Future**
> "SentinelAI makes agents production-ready. Next: integrate Backboard for persistent memory."

**[4:00-4:30] Call to Action**
> "Try it at [demo link]. Build with us at [GitHub]."

---

## 🔗 Links & Resources

- **GitHub:** https://github.com/[your-username]/sentinelai
- **Live Demo:** [Deployed Webguide URL]
- **Pitch Video:** [YouTube/Loom link]
- **Documentation:**
  - [Backend README](backend/README.md)
  - [Tailbridge README](tailbridge/README.md)
  - [Engineering Notebook](scorpion/engineering_notebook/)

---

## 📊 Team Information

| Role | Team Member | LinkedIn |
|------|-------------|----------|
| Lead Developer | [Name] | [URL] |
| AI/ML Engineer | [Name] | [URL] |
| Full-Stack Dev | [Name] | [URL] |

---

## 🇨🇦 Canadian Impact Statement

### Project Vision: Abolishing Monostandardism in AI Systems

SentinelAI is not just a technical solution—it is a philosophical intervention into how we build and manage AI systems. Our project seeks to **abolish monostandardism** in AI development by rethinking how we interact with AI and developing novel management principles that recognize diversity, context, and individual needs.

### The Problem: Monostandardism in AI Management

Traditional AI systems operate on rigid, one-size-fits-all principles:
- **Single failure model:** All errors treated identically regardless of context
- **Uniform risk thresholds:** Same standards applied to diverse use cases
- **Centralized control:** Top-down management ignoring agent specialization
- **Static evaluation:** No adaptation to changing circumstances or agent capabilities

This mirrors broader societal issues where "one size fits all" approaches fail to account for human diversity—whether in disability access, workplace equity, or cultural inclusion.

### Our Solution: Contextual, Equitable AI Management

SentinelAI implements **ethical pluralism** in agent management through the **DarCI framework** (Darcy's AI Agent):

#### 1. **Differentiated Risk Assessment**
Just as we recognize that a wheelchair user needs a ramp while others need stairs, SentinelAI recognizes that different agents and tasks require different monitoring approaches:

| Agent Type | Risk Sensitivity | Intervention Style |
|------------|------------------|-------------------|
| **OpenClaw** (creative tasks) | High coherence, low loop risk | Gentle reprompting |
| **ZeroClaw** (research tasks) | High drift detection | Goal decomposition |
| **Nanobot** (production tasks) | High confidence thresholds | Immediate rollback |
| **Human collaborators** | Contextual understanding | Collaborative adjustment |

#### 2. **Adaptive Interventions**
Our 4-strategy intervention system (Reprompt → Rollback → Decompose → Halt) mirrors the equity model in social policy:
- **Not identical treatment:** Each agent receives interventions matched to its failure mode
- **Context-aware recovery:** Interventions consider the agent's role, history, and current task
- **Proportional response:** Low-risk issues get gentle correction; high-risk issues trigger safety halts

#### 3. **Transferable Management Principles**
The **DarCI agent** (available in both Python and Go) manages a diverse array of AI agents (OpenClaw, ZeroClaw, Nanobot) using the same principles that effective human managers use with diverse teams:

| AI Management Principle | Human Management Parallel |
|-------------------------|---------------------------|
| Recognize agent specialization | Value diverse skill sets |
| Adapt communication style | Adjust leadership approach per team member |
| Provide contextual support | Offer individualized accommodations |
| Monitor without micromanaging | Trust with verification |
| Intervene proportionally | Escalate appropriately |

### The Broader Impact: From AI to Society

**The Curb-Cut Effect for AI Management:**
Just as curb cuts designed for wheelchair users also help parents with strollers, travelers with suitcases, and cyclists, the equitable management principles we develop for AI agents transfer directly to human team management:

1. **Flexibility benefits everyone:** Flexible schedules that accommodate caregiving (often affecting women disproportionately) also help disabled workers, neurodivergent employees, and anyone with non-standard life circumstances.

2. **Contextual evaluation is fairer:** Measuring output rather than hours worked helps working parents, disabled employees, and remote workers—groups often disadvantaged by monostandard "face time" cultures.

3. **Diverse standards produce better outcomes:** Just as designing for the "average pilot" fit no one (the US Air Force cockpit study), managing all agents or humans by a single standard systematically excludes those who differ from the dominant norm.

### Canadian Values Alignment

This project embodies Canadian values of:
- **Inclusivity:** Recognizing and accommodating diversity rather than forcing assimilation
- **Equity:** Treating people (and agents) fairly by acknowledging different needs, not identically
- **Innovation:** Developing novel technical solutions to social problems
- **Collaboration:** Building systems that enhance human-AI partnership rather than replacement

### Long-Term Vision

**DarCI's** management principles—developed through managing diverse AI agents—provide a **proof of concept** for abolishing monostandardism in broader organizational contexts:

```
AI Agent Management  →  Human Team Management  →  Organizational Policy  →  Societal Norms
     (DarCI)             (Workplace)                (HR Systems)           (Cultural Change)
```

By demonstrating that equitable, contextual management produces better outcomes for AI agents, we provide a model for treating humans equitably—recognizing that **fairness requires flexibility, not rigidity**.

---

## 🚀 Next Steps

1. **Deploy Webguide** to Vercel/Netlify for live demo
2. **Record pitch video** (3-5 minutes)
3. **Test Tailscale setup** (verify 2+ devices, ACL, Funnel/Serve)
4. **Polish README** with setup instructions
5. **Submit to tracks:**
   - Primary: Most Technically Complex AI Hack
   - Secondary: Tailscale Integration Challenge
   - Tertiary: Best Use of Gemini API

---

*Good luck at Hack Canada 2026! 🇨🇦*
