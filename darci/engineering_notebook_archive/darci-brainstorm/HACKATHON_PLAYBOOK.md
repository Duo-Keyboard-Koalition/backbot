# 🏆 DARCI-HACKATHON Scenario Playbook

**Real-world hackathon scenarios and winning strategies**

---

## 📖 Table of Contents

1. [Scenario 1: The Last-Minute Pivot](#scenario-1-the-last-minute-pivot)
2. [Scenario 2: The Broken API](#scenario-2-the-broken-api)
3. [Scenario 3: The Demo Disaster](#scenario-3-the-demo-disaster)
4. [Scenario 4: The All-Nighter](#scenario-4-the-all-nighter)
5. [Scenario 5: The Feature Creep](#scenario-5-the-feature-creep)
6. [Scenario 6: The Mentor Save](#scenario-6-the-mentor-save)
7. [Scenario 7: The Winning Pitch](#scenario-7-the-winning-pitch)
8. [Scenario 8: The Technical Debt](#scenario-8-the-technical-debt)

---

## Scenario 1: The Last-Minute Pivot

### 📋 Situation
**Hour 6.** Your team has been building a blockchain-based supply chain tracker. Another team already has an almost-identical project. You need to pivot FAST.

### ⏱️ Time Available: 42 hours

### 🎯 Objective
Pivot to a unique angle while preserving core tech

### 🚀 DARCI-H Game Plan

```bash
# Hour 6:00 - Emergency assessment
darci-h task list --status=done
darci-h task list --status=todo

# Hour 6:15 - Brainstorm pivot ideas
darci-h brainstorm --theme="supply-chain-but-different" --count=20

# Hour 6:30 - Team vote on best pivot
# (you decide)

# Hour 7:00 - Kill old tasks
darci-h task kill TASK_OLD_1 --reason="pivot"
darci-h task kill TASK_OLD_2 --reason="pivot"
darci-h task kill --all --filter="blockchain"

# Hour 7:30 - Create new tasks
darci-h task create --priority=P0 "AI-powered supplier risk prediction" --estimate=120
darci-h task create --priority=P0 "Carbon footprint calculator" --estimate=90
darci-h task create --priority=P1 "Supplier verification badges" --estimate=60

# Hour 8:00 - Reuse existing code
darci-h code generate --feature="Risk prediction UI using existing components"

# Hour 12:00 - First pivot demo
darci-h demo ping --minutes=5
```

### 💡 Strategy
```
✅ Keep: Backend infrastructure, database schema
❌ Drop: Blockchain integration (too slow, not unique)
🔄 Pivot: Focus on AI predictions + sustainability
🎯 New angle: "Green supply chain optimization"
```

### 📊 Timeline
```
Hour 6:  Pivot decision
Hour 8:  New tasks created, coding resumes
Hour 14: Core pivot features working
Hour 24: Full pivot complete
Hour 36: Polish & demo prep
Hour 48: WIN
```

### ⚠️ Pitfalls to Avoid
```
❌ Keeping too much old code
❌ Not communicating pivot to judges
❌ Trying to build both old + new
✅ DO: Commit fully to new direction
```

### 🏆 Winning Angle
> "We started building supply chain tracking, but realized the REAL problem is predicting disruptions before they happen. Our AI predicts supplier risks 30 days in advance."

---

## Scenario 2: The Broken API

### 📋 Situation
**Hour 18.** Your entire app depends on a third-party API (e.g., Stripe, Twilio, Google Maps). It's down or rate-limited. Demo is in 30 hours.

### ⏱️ Time Available: 30 hours

### 🎯 Objective
Keep building without the critical dependency

### 🚀 DARCI-H Game Plan

```bash
# Hour 18:00 - Confirm API is down
darci-h http request --url="https://api.example.com/health"
# Returns: 503 Service Unavailable

# Hour 18:05 - Panic! (briefly)
darci-h panic --issue="Critical API down!" --severity=critical

# Hour 18:15 - Assess impact
darci-h task list --filter="api-dependent"

# Hour 18:30 - Create mocks
darci-h api mock --endpoint="/payments" --response='{"success":true,"id":"mock_123"}'
darci-h api mock --endpoint="/sms" --response='{"sent":true,"message_id":"mock_456"}'
darci-h api mock --endpoint="/maps" --response='{"lat":40.7,"lng":-74.0}'

# Hour 19:00 - Update code to use mocks
darci-h code fix --error="API unavailable, switch to mock"

# Hour 20:00 - Continue building with mock data
# Build continues!

# Hour 36:00 - API is back, swap to real
darci-h config update --use-real-api=true

# Hour 38:00 - Test with real API
darci-h test smoke --critical-only
```

### 💡 Strategy
```
✅ Create mock API immediately
✅ Build entire frontend with mock data
✅ Add feature flag to swap real/mock
✅ Test with real API only at the end
🎯 Bonus: Keep mocks for demo reliability
```

### 🛠️ Mock Templates

```python
# DARCI-H generates these automatically:

class MockAPI:
    def charge_payment(self, amount):
        return {"success": True, "id": f"mock_{random_id()}"}
    
    def send_sms(self, to, message):
        return {"sent": True, "message_id": f"mock_{random_id()}"}
    
    def geocode(self, address):
        return {"lat": 40.7128, "lng": -74.0060}
```

### ⚠️ Pitfalls to Avoid
```
❌ Waiting for API to come back
❌ Building around broken API
❌ Not testing with real API before demo
✅ DO: Build with mocks, swap at the end
```

### 🏆 Winning Angle
> "We built our system to be API-agnostic. When our payment provider went down, we switched to mocks in 30 minutes and kept shipping. Our architecture is resilient."

---

## Scenario 3: The Demo Disaster

### 📋 Situation
**Hour 47.** Final demo in 1 hour. You just found a critical bug. Login is broken. The demo flow requires login.

### ⏱️ Time Available: 60 minutes

### 🎯 Objective
Have a working demo, no matter what

### 🚀 DARCI-H Game Plan

```bash
# Hour 47:00 - Discover bug
darci-h test smoke --critical-only
# Result: Login failing

# Hour 47:05 - PANIC
darci-h panic --issue="LOGIN BROKEN! DEMO IN 1 HOUR!" --severity=critical

# Hour 47:10 - Diagnose
darci-h error monitor --since=1h
# Found: Auth token expiration bug

# Hour 47:15 - Assess options
# Option A: Fix properly (45 min, risky)
# Option B: Hardcode demo user (10 min, safe)
# Option C: Record video backup (15 min, safe)

# Hour 47:20 - Choose Option B + C
darci-h code fix --error="Bypass auth for demo user" --auto-apply
# Creates hardcoded demo account

# Hour 47:35 - Test fix
darci-h test smoke --critical-only
# Result: ✅ Login works (for demo user)

# Hour 47:40 - Create backup video
darci-h demo record --flow="login → dashboard → feature"

# Hour 47:55 - Final check
darci-h build status
darci-h demo ping --minutes=5

# Hour 48:00 - DEMO TIME
# Use hardcoded demo account
# If it breaks: play video
```

### 💡 Strategy
```
✅ Hardcode demo credentials
✅ Create video backup
✅ Test demo flow 3x before presenting
✅ Have multiple backup plans
🎯 Judges care about vision, not edge cases
```

### 🎬 Backup Plan Hierarchy
```
Plan A: Live demo (hardcoded demo account)
Plan B: Screen recording (made 1 hour before)
Plan C: Screenshots + narration
Plan D: "Imagine if..." (pure pitch)
```

### ⚠️ Pitfalls to Avoid
```
❌ Trying to properly fix complex bugs
❌ Using real auth in demo
❌ Not having backups
✅ DO: Fake it till you make it
```

### 🏆 Winning Angle
> "For demo purposes, we're using a pre-authenticated session. In production, we'd have full OAuth flow. Now let me show you the magic..."

---

## Scenario 4: The All-Nighter

### 📋 Situation
**Hour 2 (3 AM).** Your team pulled an all-nighter. Everyone is exhausted. Code quality is suffering. Morale is low.

### ⏱️ Time Available: 46 hours remaining

### 🎯 Objective
Survive the night, maintain momentum

### 🚀 DARCI-H Game Plan

```bash
# Hour 2:00 - Detect exhaustion
# (DARCI notices low commit activity)

# Hour 2:05 - Motivational mode
darci-h motivate --level=maximum
# Output: "You've built 8 features! Only 4 more to MVP! Let's go!"

# Hour 2:15 - Automate boring stuff
darci-h automate --tasks="testing,formatting,docs"

# Hour 2:30 - Order virtual coffee
darci-h coffee --for=team --emoji=☕☕☕

# Hour 2:45 - Simplify scope
darci-h task list --priority=P2
darci-h task kill --all --priority=P2
# Drops 10 low-impact tasks

# Hour 3:00 - Power sprint
darci-h sprint start --duration=180 --focus="core-features-only"

# Hour 6:00 - Sunrise, morale boost
darci-h status update --progress=70
# "Look how far we've come!"

# Hour 8:00 - Team nap time
darci-h config quiet-mode --duration=60
# Auto-handles builds while team naps
```

### 💡 Strategy
```
✅ Automate everything boring
✅ Drop P2/P3 tasks ruthlessly
✅ Take 20-min power naps
✅ Celebrate small wins
✅ Set smaller, achievable goals
🎯 Survival > perfection
```

### 🌙 Hour-by-Hour Survival Guide
```
Hour 2-4:  Darkest hour. Automate, simplify
Hour 4-6:  Push through. Small wins
Hour 6-8:  Sunrise! Morale boost
Hour 8-9:  Power nap (20 min each)
Hour 9-12: Second wind. Feature sprint
Hour 12:   Breakfast! Celebrate MVP
```

### ⚠️ Pitfalls to Avoid
```
❌ Coding for 6+ hours without break
❌ Perfectionism at 3 AM
❌ Skipping meals
❌ No naps
✅ DO: 20-min power naps, hydrate, snack
```

### 🏆 Winning Angle
> "We built an MVP in 12 hours through the night. Our dedication shows our commitment to solving this problem. Here's what we shipped..."

---

## Scenario 5: The Feature Creep

### 📋 Situation
**Hour 20.** Your team keeps adding "cool ideas." Scope has exploded. You have 47 features planned. At this rate, you'll finish 3 of them poorly.

### ⏱️ Time Available: 28 hours

### 🎯 Objective
Ruthlessly cut scope, ship core features well

### 🚀 DARCI-H Game Plan

```bash
# Hour 20:00 - Reality check
darci-h task list
# Output: 47 tasks, only 28 hours left

# Hour 20:15 - Impact analysis
darci-h task prioritize --criteria=impact --demo-focus

# Hour 20:30 - Ruthless cuts
darci-h task kill TASK_47 --reason="no-time"
darci-h task kill TASK_46 --reason="no-time"
# ... repeat 20 times ...

# Hour 21:00 - Final scope
darci-h task list --priority=P0
# Output: 8 core features

# Hour 21:15 - Team alignment
# "We're doing THESE 8 things, and doing them WELL"

# Hour 21:30 - Sprint on core
darci-h sprint start --duration=420 --focus="core-eight"

# Hour 36:00 - Core features done
darci-h task list --status=done
# Output: 8/8 ✅

# Hour 36:30 - IF TIME: Add ONE nice-to-have
darci-h task list --priority=P1 --limit=1
```

### 💡 Strategy
```
✅ 8 features done > 20 features half-baked
✅ Cut anything that doesn't demo well
✅ Focus on user flow, not edge cases
✅ One "wow" feature > ten "meh" features
🎯 Judges see the demo, not your task list
```

### ✂️ Cutting Framework
```
Keep if:
✅ Core to value prop
✅ Visually impressive in demo
✅ Can be built in <2 hours

Cut if:
❌ "Would be cool to have"
❌ Requires complex backend
❌ Only matters for edge cases
❌ Another team already did it
```

### ⚠️ Pitfalls to Avoid
```
❌ "Just one more feature"
❌ Building for scale (you have 3 users)
❌ Edge case handling
❌ Nice-to-have features
✅ DO: Ship 8 amazing things
```

### 🏆 Winning Angle
> "We focused on doing 8 things exceptionally well, rather than 40 things poorly. Each feature in our demo is production-quality."

---

## Scenario 6: The Mentor Save

### 📋 Situation
**Hour 14.** You're stuck on a technical problem. It's blocking 3 other tasks. The mentor is a senior engineer from a sponsoring company.

### ⏱️ Time Available: 34 hours

### 🎯 Objective
Get expert help, unblock team

### 🚀 DARCI-H Game Plan

```bash
# Hour 14:00 - Identify blocker
darci-h task blocker TASK_15
# Output: "Technical: OAuth callback not working"

# Hour 14:15 - Prepare question
darci-h notify mentor --mentor="Sarah (Google)" \
  --problem="OAuth callback returning 400" \
  --context="Using Google OAuth, callback URL: https://our-app.vercel.app/auth/callback"

# Hour 14:30 - Mentor arrives
# Sarah: "Let me take a look..."

# Hour 14:35 - Show Sarah the code
darci-h file read --path="./src/auth.py" --max-lines=50

# Hour 14:45 - Sarah identifies issue
# "You need to register the callback URL in Google Cloud Console"

# Hour 14:50 - Fix applied
darci-h code fix --error="Add callback URL to allowed origins"

# Hour 15:00 - Test
darci-h test smoke --critical-only
# Result: ✅ OAuth working!

# Hour 15:05 - Unblock dependent tasks
darci-h task update TASK_15 --status=done
darci-h task list --status=blocked
# 3 tasks unblocked!

# Hour 15:15 - Thank mentor
darci-h notify team --channel="#general" \
  --message="Huge thanks to Sarah for the OAuth save! 🙏"
```

### 💡 Strategy
```
✅ Ask for help EARLY
✅ Be specific about the problem
✅ Show what you've tried
✅ Have code ready to show
✅ Thank them publicly
🎯 Mentors want to help you win
```

### 📝 Mentor Ask Template
```
Hey [Name]! We're stuck on [specific problem].

What we're trying to do: [1 sentence]
What's happening: [error/behavior]
What we've tried: [list 2-3 things]

Can you take a look when you have 5 min?

[Link to repo/code]
```

### ⚠️ Pitfalls to Avoid
```
❌ "Nothing works" (too vague)
❌ Asking without trying first
❌ Not having code ready
❌ Wasting mentor's time
✅ DO: Be specific, prepared, grateful
```

### 🏆 Winning Angle
> "We leveraged our mentor's expertise to unblock a critical authentication issue. This is why mentorship matters!"

---

## Scenario 7: The Winning Pitch

### 📋 Situation
**Hour 46.** Your project is done. Now you need to win over the judges. You have 3 minutes to pitch + 2 min Q&A.

### ⏱️ Time Available: 2 hours

### 🎯 Objective
Create a pitch that wins

### 🚀 DARCI-H Game Plan

```bash
# Hour 46:00 - Generate pitch deck
darci-h pitch deck --slides=10 --theme=startup

# Hour 46:30 - Review deck
# Slide 1: Problem (painful, relatable)
# Slide 2: Solution (your product)
# Slide 3: Market size (big number)
# Slide 4: Product demo (screenshots)
# Slide 5: Technology (AI! Blockchain! Cloud!)
# Slide 6: Business model (how you make $)
# Slide 7: Competition (you're different)
# Slide 8: Traction (fake metrics ok for hackathon)
# Slide 9: Team (you're awesome)
# Slide 10: Ask / Vision (change the world)

# Hour 47:00 - Generate demo script
darci-h demo script --duration=3 --audience=judges

# Hour 47:30 - Generate metrics
darci-h demo metrics --metric-type=users --exaggeration=2.0

# Hour 48:00 - Practice pitch
# Run through 3 times

# Hour 48:30 - Final prep
darci-h demo ping --minutes=5

# Hour 49:00 - PITCH TIME
# Kill it. Win stuff.
```

### 💡 Strategy
```
✅ Start with painful problem
✅ Show, don't tell (demo!)
✅ Big market number
✅ Clear differentiation
✅ Memorable tagline
🎯 Judges remember stories, not features
```

### 🎤 Pitch Structure (3 Minutes)
```
0:00-0:30  Problem (make it hurt)
0:30-1:00  Solution (your product)
1:00-2:00  Demo (the magic!)
2:00-2:30  Market + Business
2:30-3:00  Vision + Ask
```

### 🏆 Winning Pitch Examples

**Bad:**
> "We built a blockchain-based supply chain tracker with AI and IoT integration..."

**Good:**
> "Every year, $40B of food spoils in transit. Farmers lose their livelihoods. Families go hungry. We fix this. Our AI predicts spoilage 7 days before it happens, rerouting shipments automatically. We saved 1000 tons of food in our test. We're the early warning system for the global food supply."

### ⚠️ Pitfalls to Avoid
```
❌ Reading slides
❌ Too much tech jargon
❌ No clear problem
❌ Boring demo
❌ No ask/vision
✅ DO: Tell a story, show the magic
```

---

## Scenario 8: The Technical Debt

### 📋 Situation
**Hour 40.** You shipped 15 features fast. Code is a mess. No tests. Hardcoded values everywhere. It works, but barely. Do you refactor?

### ⏱️ Time Available: 8 hours

### 🎯 Objective
Balance polish with stability

### 🚀 DARCI-H Game Plan

```bash
# Hour 40:00 - Assess debt
darci-h code analyze --path=./src
# Output: 47 issues, 234 hardcoded values, 0 tests

# Hour 40:15 - Triage
# Critical (will break demo): Fix NOW
# Important (should fix): If time
# Nice-to-have: Ignore

# Hour 40:30 - Fix critical only
darci-h code fix --error="Hardcoded API keys" --auto-apply
darci-h code fix --error="Missing error handling" --auto-apply

# Hour 42:00 - Add minimal tests
darci-h test create --critical-only
# Creates 5 smoke tests

# Hour 43:00 - Test demo flow
darci-h demo run-through --count=3

# Hour 44:00 - Polish visible parts
darci-h polish --time=60 --area="UI"

# Hour 45:00 - Final stability check
darci-h build status
darci-h error monitor

# Hour 46:00 - CODE FREEZE
# No more changes!

# Hour 48:00 - Demo with confidence
```

### 💡 Strategy
```
✅ Fix only what breaks demo
✅ Leave "ugly but working" code
✅ Add minimal error handling
✅ Polish UI (judges see this)
✅ CODE FREEZE 2 hours before demo
🎯 Ugly code that works > pretty code that doesn't
```

### 🔧 Fix Priority
```
P0 (Fix now):
- Crashes on demo path
- Security vulnerabilities (API keys)
- Data loss bugs

P1 (If time):
- Error messages
- Loading states
- Edge cases

P2 (Ignore):
- Code duplication
- Missing tests
- Architecture issues
- Performance optimization
```

### ⚠️ Pitfalls to Avoid
```
❌ Refactoring working code
❌ Adding tests at the end
❌ "Cleaning up" everything
❌ No code freeze
✅ DO: Fix demo-critical only
```

### 🏆 Winning Angle
> "We prioritized shipping working features over perfect code. Our demo is rock-solid because we focused on stability for the critical path."

---

## 🎯 General Scenario Playbook

### For Any Crisis:
```
1. Pause. Breathe.
2. Assess: What's actually broken?
3. Options: What can we do?
4. Decide: Pick fastest path to working demo
5. Execute: All hands on deck
6. Learn: Post-mortem AFTER hackathon
```

### For Any Decision:
```
Does this help us win?
  ✅ Yes → Do it
  ❌ No → Don't do it
  
Will judges notice?
  ✅ Yes → Do it
  ❌ No → Skip it
  
Can we do this in time?
  ✅ Yes → Great
  ❌ No → Simplify or skip
```

### For Any Feature:
```
Is it core to the demo?
  ✅ Yes → Build it
  ❌ No → Cut it
  
Does it look impressive?
  ✅ Yes → Prioritize
  ❌ No → Deprioritize
  
Can we build it in <2 hours?
  ✅ Yes → Add to sprint
  ❌ No → Simplify or skip
```

---

*DARCI-HACKATHON Scenario Playbook v1.0*
*Real scenarios. Real solutions. Real wins.*
*🏆 Now go write your own winning story!*
