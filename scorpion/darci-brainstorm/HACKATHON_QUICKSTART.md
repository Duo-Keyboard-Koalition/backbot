# 🚀 DARCI-HACKATHON Quick Start

**From zero to shipped in 15 minutes**

---

## ⚡ 15-Minute Setup

### Minute 0-2: Install
```bash
# Clone or install
pip install darci-hackathon

# Or if from source
cd darci-brainstorm
# (you're already here!)
```

### Minute 2-5: Configure
```bash
# Initialize DARCI-H
darci-h init --mode=redline

# Add API keys (required)
export GEMINI_API_KEY="your-key-here"
export VERCEL_TOKEN="your-token-here"  # for auto-deploy

# Optional but recommended
export SLACK_WEBHOOK="your-webhook"
export GITHUB_TOKEN="your-token"
```

### Minute 5-10: Project Setup
```bash
# Create project
darci-h project create --name="HackProject" --stack="fullstack"

# This creates:
# ✅ GitHub repo
# ✅ Vercel project linked
# ✅ Basic folder structure
# ✅ Package.json / requirements.txt
# ✅ .env template
# ✅ README.md starter
```

### Minute 10-15: First Feature
```bash
# Brainstorm features
darci-h brainstorm --theme="your-hackathon-theme"

# Pick top 3, create tasks
darci-h task create --priority=P0 --estimate=60 "User auth with Google"

# Generate the feature
darci-h code generate --feature="Google OAuth login" --lang=python

# Deploy it
darci-h deploy preview

# 🎉 You're live!
```

---

## 🎯 Common Commands Cheat Sheet

### Project Management
```bash
# Create task
darci-h task create "Build landing page" --priority=P0 --estimate=45

# List tasks
darci-h task list --status=todo --priority=P0

# Update task
darci-h task update TASK_ID --status=done

# Reprioritize
darci-h task prioritize --criteria=impact --demo-focus

# Kill a task
darci-h task kill TASK_ID --reason="no-time"

# Estimate remaining
darci-h task estimate TASK_ID
```

### Build & Deploy
```bash
# Quick build (skip tests)
darci-h build fast

# Full build
darci-h build run --target=all

# Smoke test
darci-h test smoke --critical-only

# Deploy to Vercel
darci-h deploy preview --platform=vercel

# Rollback
darci-h deploy rollback --deployment-id=xxx

# Check status
darci-h build status
```

### Code Generation
```bash
# Generate feature
darci-h code generate --feature="REST API for todos" --lang=python

# Fix error
darci-h code fix --error="ModuleNotFoundError: No module named 'flask'"

# Format code
darci-h code format --path=./src

# Search code
darci-h file search --pattern="def login" --extension=.py
```

### Git Operations
```bash
# Init repo
darci-h git init --name="my-hack" --platform=github

# Commit all changes
darci-h git commit --message="feat: add auth" --all

# Push to main
darci-h git push --branch=main

# Create feature branch
darci-h git branch create --name="feat/auth"

# Merge to main
darci-h git merge fast --branch=feat/auth
```

### Communication
```bash
# Team ping
darci-h notify team --channel="#hackathon" --message="Auth is live!"

# Ask mentor
darci-h notify mentor --mentor="Sarah" --problem="Stuck on OAuth"

# Progress update
darci-h status update --progress=60 --features-done="auth,landing"

# Demo countdown
darci-h demo ping --minutes=5

# EMERGENCY
darci-h panic --issue="Production down!" --severity=critical

# Victory!
darci-h victory post --message="We shipped!" --include-stats
```

### Documentation
```bash
# Generate README
darci-h readme generate --project-name="HackProject"

# Create pitch deck
darci-h pitch deck --slides=10 --theme=startup

# API docs
darci-h api docs quick

# Demo script
darci-h demo script --duration=3 --audience=judges
```

### Quick Actions
```bash
# 1-hour sprint
darci-h sprint start --duration=60 --focus="authentication"

# 30-min fix
darci-h fix --time=30 --issue="Login broken"

# 15-min polish
darci-h polish --time=15 --area="UI"

# 5-min panic
darci-h panic --issue="EVERYTHING IS BROKEN"
```

---

## 🎭 Agent Mode Switching

```bash
# Default: Unicorn mode (hyper-optimistic)
darci-h mode set --mode=unicorn

# Focus: Robot mode (code only)
darci-h mode set --mode=robot

# Polish: Designer mode (UI focus)
darci-h mode set --mode=designer

# Data: Analyst mode (metrics)
darci-h mode set --mode=analyst

# Pitch: Pitchman mode (presentation)
darci-h mode set --mode=pitchman

# Check current mode
darci-h mode get
```

---

## 📊 Dashboard Commands

```bash
# Live dashboard
darci-h dashboard

# Burndown chart
darci-h metrics burndown

# Feature completion
darci-h metrics features

# Team velocity
darci-h metrics velocity

# Time remaining
darci-h metrics countdown
```

---

## 🆘 Emergency Commands

### Build is Broken
```bash
# Auto-rollback
darci-h emergency rollback

# Find breaking commit
darci-h git bisect start

# Fix and redeploy
darci-h emergency fix-and-deploy
```

### Critical Bug
```bash
# War room mode
darci-h emergency war-room

# All hands on deck
darci-h panic --issue="Critical bug in auth" --severity=critical

# Generate fix
darci-h code fix --error="[error message]" --auto-apply
```

### Demo in 5 Minutes
```bash
# Full demo prep sequence
darci-h emergency demo-ready --minutes=5

# This runs:
# ✅ Smoke tests
# ✅ Build check
# ✅ Deploy if needed
# ✅ Opens demo tabs
# ✅ Posts countdown
```

### Team Exhausted
```bash
# Motivational mode
darci-h motivate --level=maximum

# Auto-take over boring tasks
darci-h automate --tasks="testing,docs,formatting"

# Order virtual coffee
darci-h coffee --for=team
```

---

## 🎯 Sprint Templates

### 1-Hour Feature Sprint
```bash
darci-h sprint start --duration=60

# Auto-creates:
# 1. Task breakdown (5 min)
# 2. Code generation (30 min)
# 3. Build & test (15 min)
# 4. Deploy (10 min)

# Sprint complete!
```

### 4-Hour MVP Sprint
```bash
darci-h sprint start --duration=240 --goal="MVP"

# Auto-creates:
# 1. Core features only (P0 tasks)
# 2. Basic UI
# 3. Working backend
# 4. Deployed and demoable
```

### 12-Hour Alpha Sprint
```bash
darci-h sprint start --duration=720 --goal="alpha"

# Includes:
# 1. All core features
# 2. Basic error handling
# 3. User testing round
# 4. Iteration based on feedback
```

### 24-Hour Beta Sprint
```bash
darci-h sprint start --duration=1440 --goal="beta"

# Includes:
# 1. Feature complete
# 2. Bug fixes
# 3. Polish pass
# 4. Demo rehearsal
```

---

## 🏆 Winning Workflows

### Workflow 1: Idea → Live in 30 Min
```bash
# 1. Brainstorm (5 min)
darci-h brainstorm --theme="sustainability" --count=10

# 2. Pick best idea (2 min)
# (you decide)

# 3. Create MVP tasks (5 min)
darci-h task create --priority=P0 "Landing page"
darci-h task create --priority=P0 "User signup"
darci-h task create --priority=P0 "Core feature"

# 4. Generate code (15 min)
darci-h code generate --feature="Landing page with Tailwind"
darci-h code generate --feature="User auth with email"

# 5. Deploy (3 min)
darci-h deploy preview

# 🎉 LIVE!
```

### Workflow 2: Feature Factory (Per Feature: 20 Min)
```bash
# 1. Create task (1 min)
darci-h task create --priority=P1 "Add dark mode" --estimate=20

# 2. Generate (15 min)
darci-h code generate --feature="Dark mode toggle with localStorage"

# 3. Test & deploy (4 min)
darci-h test smoke
darci-h deploy preview

# ✅ Feature shipped!
```

### Workflow 3: Crisis → Fix in 10 Min
```bash
# 1. Panic! (0 min)
darci-h panic --issue="Users can't login" --severity=critical

# 2. Diagnose (2 min)
darci-h error monitor --since=1h

# 3. Fix (5 min)
darci-h code fix --error="[specific error]"

# 4. Deploy (3 min)
darci-h build fast && darci-h deploy preview

# 🔥 Fire put out!
```

### Workflow 4: Demo Prep in 30 Min
```bash
# 1. Generate script (10 min)
darci-h demo script --duration=3 --audience=judges

# 2. Create deck (10 min)
darci-h pitch deck --slides=10

# 3. Fake metrics (5 min)
darci-h demo metrics --metric-type=users --exaggeration=2.0

# 4. Rehearse (5 min)
# (you practice)

# 🎤 Ready to win!
```

---

## 💡 Pro Tips

### Tip 1: Use Aliases
```bash
# Add to ~/.bashrc or ~/.zshrc
alias dh="darci-h"
alias dh-task="darci-h task"
alias dh-build="darci-h build"
alias dh-deploy="darci-h deploy"
alias dh-panic="darci-h panic"

# Now you can type:
dh task create "Fix bug" -p P0
```

### Tip 2: Keyboard Shortcuts
```bash
# Create shell functions
dh-quick() {
  darci-h task create "$1" --priority=P0 --estimate=30
  darci-h code generate --feature="$1"
  darci-h deploy preview
}

# Usage: dh-quick "user profile page"
```

### Tip 3: Preset Configs
```bash
# Save common configurations
darci-h config save --name="fullstack" --stack="nextjs+fastapi+supabase"
darci-h config save --name="ml" --stack="python+streamlit+huggingface"

# Use later
darci-h project create --preset="fullstack"
```

### Tip 4: Team Sync
```bash
# Set up team notifications
darci-h config notifications --slack="#hack-team" --discord="Team Channel"

# Auto-post updates
darci-h config auto-post --interval=60 --channels=slack,discord
```

### Tip 5: Demo Mode
```bash
# 1 hour before demo
darci-h mode set --mode=pitchman
darci-h demo prep --minutes=60

# Auto-mutes notifications, preps demo environment
```

---

## 📈 Progress Tracking

```bash
# Check progress
darci-h progress

# Output:
┌─────────────────────────────────────────┐
│  🚀 HACKATHON PROGRESS - Hour 18/48    │
├─────────────────────────────────────────┤
│  Features: 6/10 [██████░░░░] 60%       │
│  Bugs: 3 🔴                            │
│  Build: ✅ Green                       │
│  Deployments: 12                       │
│  Commits: 47                           │
│  Time Left: 30 hours                   │
├─────────────────────────────────────────┤
│  Next: Fix login bug (P0)              │
│  ETA: 25 min                           │
└─────────────────────────────────────────┘
```

---

## 🎯 Hour-by-Hour Guide

### Hour 0-1: Setup ✅
```bash
darci-h init
darci-h project create
darci-h brainstorm
darci-h task create (x5)
```

### Hour 1-4: Core Features 🔨
```bash
darci-h sprint start --duration=180
# Repeat: task → code → build → deploy
```

### Hour 4-12: More Features 🚀
```bash
darci-h sprint start --duration=480
# Feature factory mode
```

### Hour 12-24: Complete MVP 🎯
```bash
darci-h task prioritize --criteria=completeness
darci-h task kill (low-value tasks)
darci-h sprint finish
```

### Hour 24-36: Polish & Bugs 🐛
```bash
darci-h mode set --mode=designer
darci-h polish --time=120
darci-h fix --all
```

### Hour 36-42: Demo Prep 🎬
```bash
darci-h mode set --mode=pitchman
darci-h demo script
darci-h pitch deck
darci-h demo metrics
```

### Hour 42-46: Rehearsal 🎤
```bash
darci-h demo ping --minutes=30
darci-h test smoke
darci-h build status
# Practice 3x
```

### Hour 46-48: Final Prep 🏁
```bash
darci-h mode set --mode=unicorn
darci-h victory post --pre-game
darci-h demo ready
# WIN!
```

---

## 🆘 Troubleshooting

### "Build failing!"
```bash
darci-h build status
darci-h error monitor
darci-h code fix --error="[error from above]"
```

### "Can't deploy!"
```bash
darci-h deploy preview --platform=netlify  # try alternative
darci-h config check vercel  # diagnose
```

### "Running out of time!"
```bash
darci-h task kill --all --except=P0
darci-h sprint start --duration=120 --focus=core
darci-h panic --issue="Time crisis"  # motivates team
```

### "Team is tired!"
```bash
darci-h motivate
darci-h automate --tasks="boring"
darci-h coffee --virtual --for=all
```

### "Demo in 5 minutes, IS IT READY?!"
```bash
darci-h emergency demo-ready --minutes=5
# Deep breaths. You got this.
```

---

*DARCI-HACKATHON Quick Start v1.0*
*15 minutes to setup. 48 hours to glory.*
*🚀 Now go win that hackathon!*
