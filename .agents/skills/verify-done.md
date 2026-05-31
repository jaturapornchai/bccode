---
name: verify-done
description: Trigger when about to claim task complete. Forces evidence-based verification.
---

# Verification Protocol

## "Done" requires evidence

### Code
- [ ] `<test cmd>` — paste output
- [ ] `<build cmd>` — 0 errors
- [ ] Lint passes

### Endpoint
- [ ] curl example + expected response
- [ ] Test row added & passing

### Bug fix
- [ ] Failing test BEFORE fix
- [ ] Passing test AFTER fix

### UI
- [ ] Screenshot or describe rendered state
- [ ] Mobile + desktop checked
- [ ] Dark mode (if supported)

### Bash output (all models — read carefully)
- [ ] Re-read terminal output literally
- [ ] Check exit code, not just last line
- [ ] Watch for syntax edge cases

## FORBIDDEN
- ❌ "should work now"
- ❌ "should fix it"
- ❌ "implementation complete"

## REQUIRED
- ✅ "ตรวจแล้ว: <output>"
- ✅ "ยังไม่ verify เพราะ <reason>"
