---
description: Check the status of the current pull request and CI checks
---

## User Input

```text
$ARGUMENTS
```

## Goal

Display comprehensive status of the current branch's pull request including CI checks, reviews, and merge readiness.

## Execution Steps

### 1. Find Current PR

```bash
# Get current branch
BRANCH=$(git branch --show-current)

# Find PR for this branch
gh pr view --json number,title,state,url,mergeable,mergeStateStatus,reviewDecision,statusCheckRollup,additions,deletions,changedFiles
```

If no PR exists, report that and suggest creating one with `/pr-create`.

### 2. Display PR Overview

Report:
- PR number and title
- PR URL
- State (open/closed/merged)
- Lines changed (+additions/-deletions)
- Files changed

### 3. Check CI Status

```bash
gh pr checks <PR_NUMBER>
```

Display each check with status:
- ✅ pass
- ❌ fail
- ⏳ pending
- ⏸️ skipped

If any checks failed, show the failed check URLs for debugging.

### 4. Check Merge Readiness

Report:
- **Mergeable**: Yes/No (are there conflicts?)
- **Merge State**: BLOCKED/CLEAN/UNSTABLE/etc.
- **Review Decision**: APPROVED/CHANGES_REQUESTED/REVIEW_REQUIRED/none

### 5. Provide Recommendations

Based on status, suggest next actions:

**If CI failing:**
> CI checks are failing. View logs and fix issues before merging.
> Failed checks: <list>

**If conflicts:**
> PR has merge conflicts. Rebase on master:
> ```
> git fetch origin master && git rebase origin/master
> git push --force-with-lease
> ```

**If awaiting review:**
> PR is awaiting review. Request review from team members.

**If all green:**
> ✅ PR is ready to merge! Run `/pr-merge` to merge.

## Arguments

- PR number can be provided: `/pr-status 47`
- `--checks-only`: Only show CI check status
- `--json`: Output raw JSON for scripting
