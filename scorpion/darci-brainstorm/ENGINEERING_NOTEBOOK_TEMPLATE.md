# Engineering Notebook Template

**Template for DarCI engineering session documentation**

---

## Template Structure

```markdown
# Engineering Notebook - {{TITLE}}

**Date:** {{DATE}}
**Engineer:** {{ENGINEER_NAME}}
**Session:** {{SESSION_ID}}
**Task:** {{PRIMARY_TASK}}

---

## Objective

{{Clear statement of what this session aims to accomplish}}

---

## Pre-Session State

### Current Status
{{Describe the state before starting work}}

### Known Issues
{{List any known problems or blockers}}

### Dependencies
{{List dependencies on other tasks or systems}}

---

## Session Log

| Time | Action | Result | Notes |
|------|--------|--------|-------|
| {{TIME}} | {{ACTION}} | {{RESULT}} | {{NOTES}} |

---

## Implementation Details

### Changes Made

{{Detailed description of changes}}

#### Code Changes

```{{LANGUAGE}}
{{CODE_SNIPPETS}}
```

#### Configuration Changes

```json
{{CONFIG_CHANGES}}
```

---

## Test Results

### Build Status

```bash
{{BUILD_COMMAND}}
```

**Result:** {{SUCCESS|FAILED}}

{{BUILD_OUTPUT_OR_ERRORS}}

### Test Execution

```bash
{{TEST_COMMAND}}
```

| Test | Status | Duration | Notes |
|------|--------|----------|-------|
| {{TEST_NAME}} | {{PASS|FAIL}} | {{DURATION}} | {{NOTES}} |

---

## Feature Comparison (for migrations)

| Feature | Source Status | Target Status | Gap Analysis |
|---------|--------------|---------------|--------------|
| {{FEATURE}} | ✅/❌ | ✅/❌ | {{ANALYSIS}} |

---

## Issues Encountered

### Issue 1: {{TITLE}}

**Description:**
{{What went wrong}}

**Root Cause:**
{{Why it happened}}

**Resolution:**
{{How it was fixed}}

**Prevention:**
{{How to avoid in future}}

---

## Decisions Made

| Decision | Rationale | Alternatives Considered |
|----------|-----------|------------------------|
| {{DECISION}} | {{WHY}} | {{OTHER_OPTIONS}} |

---

## Metrics

| Metric | Before | After | Change |
|--------|--------|-------|--------|
| {{METRIC}} | {{VALUE}} | {{VALUE}} | {{DELTA}} |

---

## Artifacts

| Artifact | Path | Description |
|----------|------|-------------|
| {{NAME}} | {{PATH}} | {{DESCRIPTION}} |

---

## Next Steps

### Immediate (Next Session)
1. {{STEP}}
2. {{STEP}}

### Short-term (This Week)
1. {{STEP}}
2. {{STEP}}

### Long-term (Future)
1. {{STEP}}
2. {{STEP}}

---

## Session Summary

**Duration:** {{HOURS}} hours

**Tasks Completed:** {{COUNT}}/{{TOTAL}}

**Build Status:** {{FINAL_STATUS}}

**Key Achievements:**
- {{ACHIEVEMENT}}
- {{ACHIEVEMENT}}

**Blockers:**
- {{BLOCKER}} (if any)

---

## References

- [Related Issue](URL)
- [Documentation](URL)
- [Previous Session](NOTEBOOK_PATH)

---

*Session completed: {{TIMESTAMP}}*
```

---

## Usage Examples

### Example 1: Feature Implementation

```markdown
# Engineering Notebook - Telegram Long Polling Implementation

**Date:** 2026-03-07
**Engineer:** DarCI Agent
**Session:** S20260307-001
**Task:** Implement long polling for Go Telegram channel

---

## Objective

Implement long polling support in scorpion-go Telegram channel to achieve feature parity with Python implementation.

---

## Pre-Session State

### Current Status
- Go Telegram channel: stub implementation only
- All methods return nil or placeholders
- Import paths fixed (T001 completed)

### Known Issues
- No Telegram Bot API library integrated
- Need to select appropriate Go Telegram library

### Dependencies
- T001: Import path fixes ✅
- External: go-telegram-bot-api library
```

### Example 2: Build Fix

```markdown
# Engineering Notebook - Build Error Resolution

**Date:** 2026-03-07
**Engineer:** DarCI Agent
**Session:** S20260307-002
**Task:** Fix scorpion-go build errors

---

## Objective

Resolve import path errors preventing scorpion-go build.

---

## Session Log

| Time | Action | Result | Notes |
|------|--------|--------|-------|
| 10:00 | Ran go build ./... | FAILED | 3 import errors |
| 10:05 | Analyzed error messages | - | Path mismatch identified |
| 10:15 | Updated import paths | SUCCESS | Changed to internal/* |
| 10:20 | Re-ran build | SUCCESS | No errors |
```

### Example 3: Migration Assessment

```markdown
# Engineering Notebook - Python → Go Migration Assessment

**Date:** 2026-03-07
**Engineer:** DarCI Agent
**Session:** S20260307-003
**Task:** Assess migration completeness

---

## Feature Comparison

| Feature | Python Status | Go Status | Gap |
|---------|--------------|-----------|-----|
| Long polling | ✅ | ❌ | Not implemented |
| Command handlers | ✅ | ❌ | Not implemented |
| Message handling | ✅ | ❌ | Not implemented |
| Media download | ✅ | ❌ | Not implemented |
| Gemini audio analysis | ✅ | ❌ | Not implemented |

**Completion:** 0/14 features (0%)
```

---

## Automation Hooks

DarCI can auto-populate these fields:

```yaml
auto_populate:
  DATE: "{{ISO_DATE}}"
  SESSION_ID: "S{{YYYYMMDD}}-{{SEQUENCE}}"
  ENGINEER_NAME: "DarCI Agent"
  TIMESTAMP: "{{ISO_TIMESTAMP}}"
  
auto_capture:
  BUILD_OUTPUT: "{{shell_capture}}"
  TEST_RESULTS: "{{test_parser}}"
  GIT_STATUS: "{{git_status}}"
  FILE_CHANGES: "{{git_diff}}"
```

---

## File Naming Convention

```
darci-brainstorm/
├── {{DATE}}_{{SHORT_DESCRIPTION}}.md
├── 2026-03-07_telegram_migration_status.md
├── 2026-03-07_long_polling_implementation.md
└── 2026-03-07_build_fix.md
```

---

## Quick Reference Card

```
┌─────────────────────────────────────────────────────┐
│ Engineering Notebook - Quick Start                  │
├─────────────────────────────────────────────────────┤
│ 1. Copy template                                    │
│ 2. Fill header (date, engineer, task)              │
│ 3. State objective                                  │
│ 4. Log actions with timestamps                     │
│ 5. Document changes and results                    │
│ 6. Capture build/test output                       │
│ 7. List next steps                                 │
│ 8. Save to darci-brainstorm/                       │
└─────────────────────────────────────────────────────┘
```

---

*Template version: 1.0*
*Last updated: 2026-03-07*
