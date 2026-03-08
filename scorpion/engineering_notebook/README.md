# Engineering Notebook - Index and Guidelines

## 🎯 Purpose

This engineering notebook serves as the **single source of truth** for all development work on the Scorpion project by Qwen Code agents.

---

## 📁 Notebook Structure

The engineering notebook is organized to match the project structure:

```
engineering_notebook/
├── README.md                 # This file - guidelines and index
├── python/                   # Python-specific implementation notes
│   └── (Python-specific entries)
├── go/                       # Go-specific implementation notes
│   └── 2026-03-07_go_compilation_fix.md
└── shared/                   # Cross-cutting concerns (both implementations)
    ├── 2026-03-07_telegram_migration_status.md
    └── 2026-03-07_tui_agent_test_results.md
```

### Directory Guide

| Directory | Purpose | Example Topics |
|-----------|---------|----------------|
| `python/` | Python implementation (`scorpion-python/`) | TUI tests, channel implementations, provider updates |
| `go/` | Go implementation (`scorpion-go/`) | Compilation fixes, ADK integration, tool implementations |
| `shared/` | Both implementations | Migration status, config changes, architecture decisions |

---

## 📝 Quick Links

### Python Implementation Notes
- [TUI Agent Test Results](python/2026-03-07_tui_agent_test_results.md) - Testing Python agent with `~/.scorpion-python/` config

### Go Implementation Notes
- [Compilation Fix Guide](go/2026-03-07_go_compilation_fix.md) - How we fixed ADK API mismatches

### Shared/Cross-Platform Notes
- [Telegram Migration Status](shared/2026-03-07_telegram_migration_status.md) - Python vs Go Telegram bot comparison

---

## ⚠️ Why This Matters

### 1. **Continuity Across Sessions**

Qwen agents may work on this project across multiple separate sessions. The engineering notebook ensures:
- No work is duplicated
- Previous decisions are respected
- Context is preserved between sessions
- New agents can quickly understand the current state

### 2. **Accountability & Traceability**

Every change should be documented so we can answer:
- **What** was changed?
- **Why** was it changed?
- **When** was it changed?
- **Who** (which agent session) made the change?

### 3. **Knowledge Preservation**

The notebook captures:
- ✅ Successful implementations
- ❌ Failed attempts (so we don't repeat mistakes)
- 🔧 Workarounds and temporary solutions
- 📋 Test results and verification steps
- 🎯 Design decisions and their rationale

---

## 📝 When to Update

**Update the engineering notebook for EVERY significant action:**

| Action | Required? |
|--------|-----------|
| Starting a new task | ✅ Yes |
| Completing a task | ✅ Yes |
| Running tests | ✅ Yes |
| Making code changes | ✅ Yes |
| Fixing bugs | ✅ Yes |
| Adding dependencies | ✅ Yes |
| Configuration changes | ✅ Yes |
| Build succeeds/fails | ✅ Yes |
| User feedback received | ✅ Yes |

**Rule of thumb:** If you spent more than 5 minutes on something, document it.

---

## 📋 What to Document

### For Each Session

1. **Date and Session Goal**
   ```markdown
   **Date:** YYYY-MM-DD
   **Task:** [Brief description]
   ```

2. **Changes Made**
   - Files modified
   - Commands run
   - Configuration changes

3. **Test Results**
   - Commands used to test
   - Output (success or failure)
   - Screenshots or logs if relevant

4. **Issues Encountered**
   - Error messages
   - Root cause analysis
   - Resolution (or why unresolved)

5. **Next Steps**
   - What remains undone
   - Known issues
   - Recommendations for future sessions

---

## 📁 File Naming Convention

```
YYYY-MM-DD_topic.md
```

**Examples:**
- `2026-03-07_telegram_migration_status.md`
- `2026-03-07_tui_agent_test_results.md`
- `2026-03-08_gateway_implementation.md`

---

## 📖 Template for New Entries

```markdown
# Engineering Notebook - [Topic]

**Date:** YYYY-MM-DD  
**Engineer:** Qwen Code  
**Task:** [Brief task description]

---

## Objective

[What are we trying to accomplish?]

---

## Changes Made

### Files Modified

| File | Change |
|------|--------|
| `path/to/file.py` | Description of change |

### Commands Run

```bash
command1
command2
```

---

## Test Results

### Test 1: [Name]

**Command:**
```bash
test command
```

**Result:** ✅ PASSED / ❌ FAILED

**Output:**
```
[relevant output]
```

---

## Issues Encountered

| Issue | Status | Resolution |
|-------|--------|------------|
| Description | Resolved/Open | How fixed or why blocked |

---

## Conclusion

[Summary of what was accomplished]

---

## Next Steps

- [ ] Task 1
- [ ] Task 2
- [ ] Task 3
```

---

## 🚫 Common Mistakes to Avoid

| Mistake | Why It's Bad | Fix |
|---------|--------------|-----|
| Not documenting failures | Future agents repeat the same mistakes | Document what didn't work and why |
| Vague descriptions | "Fixed stuff" is meaningless | Be specific: "Fixed null pointer in X when Y" |
| No timestamps | Can't track session progress | Include date/time for each entry |
| Skipping test results | Don't know if changes work | Always record test commands and output |
| One giant entry | Hard to find specific info | Break into logical sections |

---

## ✅ Quality Checklist

Before ending a session, verify:

- [ ] New entry created (or existing updated)
- [ ] Date and task clearly stated
- [ ] All modified files listed
- [ ] Test results documented (pass or fail)
- [ ] Issues and resolutions recorded
- [ ] Next steps defined
- [ ] File named correctly (YYYY-MM-DD_topic.md)
- [ ] File placed in correct directory (python/, go/, or shared/)

---

## 🔗 Related Documentation

| Document | Purpose |
|----------|---------|
| `../scorpion-python/README.md` | Python implementation docs |
| `../scorpion-go/README.md` | Go implementation docs |
| `../COMMUNICATION.md` | Project communication channels |
| `../QUICKSTART.md` | Quick start guide |

---

## 📞 For Questions

If you're a Qwen agent working on this project and need clarification:
1. Check existing notebook entries first
2. Look for related documentation
3. Ask the user if documentation is unclear

---

**Remember:** A well-maintained engineering notebook is the difference between a professional, maintainable project and a chaotic codebase. Future you (and future agents) will thank you!

---

*Last updated: 2026-03-07*
