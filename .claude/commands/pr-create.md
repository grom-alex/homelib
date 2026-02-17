---
description: Create a pull request for the current branch
---

## User Input

```text
$ARGUMENTS
```

## Goal

Create a well-formatted pull request for the current feature branch.

## Execution Steps

### 1. Pre-flight Checks

Before creating PR, verify:

```bash
# Check current branch (should not be master)
git branch --show-current

# Check for uncommitted changes
git status
```

If on master branch, abort with error.
If there are uncommitted changes, ask user to commit first.

### 2. Gather PR Information

```bash
# Get commits in this branch
git log master..HEAD --oneline

# Get changed files
git diff master..HEAD --stat
```

### 3. Determine PR Title

PR title MUST follow conventional commits (lowercase):
- `feat: description` - for new features
- `fix: description` - for bug fixes
- `refactor: description` - for refactoring
- `docs: description` - for documentation
- `infra: description` - for infrastructure/CI/Docker changes
- `chore: description` - for maintenance tasks

Scope is optional: `fix(ci): description`, `feat(auth): description`

If user provided a title in arguments, validate it matches convention.
Otherwise, derive from branch name and commits.

### 4. Create PR

```bash
gh pr create --title "<TITLE>" --body "$(cat <<'EOF'
## Summary

<bullet points summarizing changes>

## Changes

### Backend
<list backend changes if any>

### Frontend
<list frontend changes if any>

## Test plan

- [ ] <verification steps>

## Staging

<staging URL if deployed>

🤖 Generated with [Claude Code](https://claude.com/claude-code)
EOF
)" --base master
```

### 5. Post-Creation

After PR is created:
1. Output the PR URL
2. Wait 30 seconds for CI to start
3. Check CI status: `gh pr checks <PR_NUMBER>`
4. If CI fails, report which checks failed

## Arguments

User can provide:
- PR title as argument (e.g., `/pr-create feat: add user authentication`)
- `--draft`: Create as draft PR
