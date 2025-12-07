Purpose:
Prevent whack-a-mole fixes and regressions.

When to use:
Any production bug or flaky test.

Core flow:
- Write a failing test that reproduces the bug.
- Identify the minimal fix.
- Apply fix with no refactors.
- Validate and re-run the failing test.
- Audit for similar failure patterns.