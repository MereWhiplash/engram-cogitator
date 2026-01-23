---
name: tdd
description: Guides test-driven development with red-green-refactor cycle. Stores successful test patterns in EC. Use when implementing new features or fixing bugs with tests.
---

# Test-Driven Development

Write the test first. Watch it fail. Write minimal code to pass.

**Announce:** "I'm using the tdd skill for this implementation."

## Step 0: Load Context

Get project config and relevant patterns:

```
ec_search:
  query: project config
  type: config

ec_search:
  query: test pattern
  type: pattern
```

Use the configured `test_command` for running tests.

## The Cycle

```
RED → GREEN → REFACTOR → REPEAT
```

### RED: Write Failing Test

Search for similar test patterns first:

```
ec_search:
  query: [feature area] test
  type: pattern
```

Write one minimal test for the behavior you want:

```typescript
test('rejects empty email', () => {
  const result = validateEmail('');
  expect(result.valid).toBe(false);
  expect(result.error).toBe('Email required');
});
```

### Verify RED: Watch It Fail @verifying

Run tests with configured command:

```bash
{test_command} path/to/test
```

Confirm:
- Test fails (not errors)
- Fails because feature is missing
- Failure message is expected

**If test passes:** You're testing existing behavior. Fix the test.

### GREEN: Minimal Code

Write the simplest code to make the test pass:

```typescript
function validateEmail(email: string) {
  if (!email?.trim()) {
    return { valid: false, error: 'Email required' };
  }
  return { valid: true };
}
```

Don't add features beyond what the test requires.

### Verify GREEN: Watch It Pass @verifying

```bash
{test_command} path/to/test
```

Confirm:
- Test passes
- Other tests still pass

### REFACTOR: Clean Up

After green only:
- Remove duplication
- Improve names
- Extract helpers

Keep tests green. Don't add behavior.

### REPEAT

Next failing test for next behavior.

## Store Patterns

When you establish a useful test pattern:

```
ec_add:
  type: pattern
  area: [component]
  content: [Description of test pattern and when to use it]
  rationale: Established during TDD implementation
```

Good patterns to store:
- How to test async operations in this codebase
- Mocking conventions used
- Test data factories or fixtures
- Integration test setup patterns

## Good Tests

- **One behavior** - If test name has "and", split it
- **Clear name** - Describes what it tests
- **Real code** - Minimize mocks

## When Stuck

| Problem | Solution |
|---------|----------|
| Don't know how to test | Write the assertion first, then figure out setup |
| Test too complicated | Design too complicated - simplify the interface |
| Must mock everything | Code too coupled - use dependency injection |

Search EC for prior solutions:

```
ec_search:
  query: test difficulty [problem area]
  type: learning
```

## The Rule

```
No production code without a failing test first.
```

Code before test? Delete it. Start over.

## Store Learnings

If you discover a testing gotcha:

```
ec_add:
  type: learning
  area: testing
  content: [What the issue was and how to handle it]
```
