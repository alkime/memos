---
name: pr:open
description: Create a PR with auto-generated description, or update an existing PR's description
---

# Open Pull Request

Create a pull request for the current branch with an auto-generated description, or update an existing PR's description.

## Workflow

### 1. Gather Branch Information

Run these commands to understand what's in the branch:

```bash
# Get current branch name
git branch --show-current

# List commits on this branch vs main
git log main..HEAD --oneline

# Show detailed diff stats
git diff main...HEAD --stat
```

### 2. Check for Uncommitted Changes

Before creating/updating a PR, ensure all changes are committed:

```bash
git status --short
```

**If there are uncommitted changes:**
- Ask the user: "There are uncommitted changes. Would you like to commit them before creating the PR?"
- If user confirms: stage the relevant files, create a commit with an appropriate message, and continue
- If user declines: proceed without committing (changes won't be in the PR)

**Important:** The PR will only include committed and pushed changes. Uncommitted work won't appear in the PR diff.

### 3. Check for Existing PR

Check if a PR already exists for this branch:

```bash
gh pr view --json number,title,body,state 2>/dev/null
```

**If PR exists:**
- Check if the body is empty or minimal
- If body has content, ask user: "This PR already has a description. Would you like to overwrite it?"
- If user confirms or body is empty, proceed to update flow (Step 6b)
- If user declines, stop

**If no PR exists:**
- Check if branch has been pushed: `git rev-parse --abbrev-ref @{upstream} 2>/dev/null`
- If not pushed, push first: `git push -u origin $(git branch --show-current)`
- Proceed to create flow (Step 6a)

**Important:** After creating a PR (Step 6a), verify tracking is set up:
```bash
# Verify tracking exists, set it if not
git rev-parse --abbrev-ref @{upstream} 2>/dev/null || git branch --set-upstream-to=origin/$(git branch --show-current)
```

### 4. Analyze Changes

Review the commits and changes to understand:
- What was added, removed, or modified
- The overall theme/purpose of the changes
- Any notable statistics (files changed, lines added/removed)

### 5. Generate Description

Write the PR description to a temp file (avoids sandbox/heredoc issues):

```bash
cat > /tmp/claude/pr-body.md << 'PREOF'
## Summary

- Bullet point 1
- Bullet point 2
- Bullet point 3

## Test plan

- [ ] Test item 1
- [ ] Test item 2

🤖 Generated with [Claude Code](https://claude.com/claude-code)
PREOF
```

### 6a. Create New PR

```bash
gh pr create --title "your title here" --body-file /tmp/claude/pr-body.md
```

### 6b. Update Existing PR

```bash
gh pr edit --body-file /tmp/claude/pr-body.md
```

Optionally update the title too if it needs improvement:

```bash
gh pr edit --title "improved title" --body-file /tmp/claude/pr-body.md
```

### 7. Report Result

Output the PR URL so the user can review it.

## PR Description Guidelines

- **Title:** Use conventional commit format when appropriate (feat:, fix:, chore:, docs:, refactor:)
- **Summary:** 3-5 bullet points covering the key changes
- **Stats:** Include net line changes if significant (e.g., "Net change: -2,333 lines")
- **Test plan:** List verification steps as checkboxes

## Notes

- Always use `/tmp/claude/` for temp files (sandbox-safe)
- The `--body-file` flag avoids shell quoting issues with complex markdown
- Use `gh pr view` to check existing PR state before deciding create vs edit
- Ask before overwriting existing descriptions that have content
- **Always verify upstream tracking** after creating a PR - `gh pr create` doesn't set local tracking automatically
