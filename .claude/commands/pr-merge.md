---
description: Merge the current pull request after CI passes
---

## User Input

```text
$ARGUMENTS
```

## Goal

Safely merge the current branch's pull request after verifying all checks pass.

## Execution Steps

### 1. Find Current PR

```bash
# Get current branch
BRANCH=$(git branch --show-current)

# Find PR for this branch
gh pr list --head "$BRANCH" --json number,title,state,mergeable,mergeStateStatus
```

If no PR exists for current branch, abort with instructions to create one.

### 2. Verify PR Status

Check that:
- PR state is `OPEN`
- All CI checks have passed
- PR is mergeable (no conflicts)

```bash
gh pr checks <PR_NUMBER>
```

If any check is failing or pending:
- Report which checks failed
- Do NOT proceed with merge
- Suggest fixing the issues first

### 3. Verify No Conflicts

```bash
gh pr view <PR_NUMBER> --json mergeable,mergeStateStatus
```

If there are merge conflicts:
- Report the conflict
- Suggest rebasing: `git fetch origin master && git rebase origin/master`
- Do NOT proceed with merge

### 4. Merge PR

Use regular merge (NOT squash) as per project convention:

```bash
gh pr merge <PR_NUMBER> --merge --delete-branch
```

### 5. Post-Merge Cleanup

```bash
# Switch to master and pull
git checkout master
git pull origin master

# Verify merge
git log -1 --oneline
```

### 6. Report Success

Output:
- Confirmation that PR was merged
- The merge commit hash
- Suggest deploying to staging if needed

## Safety Checks

- NEVER merge if CI is failing
- NEVER merge if there are conflicts
- NEVER use `--force` or `--admin` flags
- ALWAYS use regular merge (not squash) per project convention

## Arguments

- `--force`: Skip CI check (DANGEROUS - requires explicit confirmation)
- PR number can be provided if not on the branch: `/pr-merge 47`
