---
name: address-pr-comments
description: Process PR review comments - fetch, validate, fix issues, update style guide
---

# Address PR Review Comments

This command processes PR review comments through a structured workflow: fetch comments, validate against code, fix issues, and intelligently update the style guide.

## Workflow

### Phase 1: Fetch and Validate Comments

1. **Fetch unresolved PR comments:**
   ```bash
   python3 scripts/format_pr.py
   ```
   This script uses the GitHub API to fetch all review threads and formats them as Markdown, showing unresolved comments first.

2. **Validate each comment individually:**

   For each unresolved comment:

   a. **Understand the request:**
      - Read the comment text and any replies
      - Note the file path and line number
      - Review the diff hunk provided

   b. **Examine the actual code:**
      - Use the Read tool to examine the file(s) mentioned
      - Look beyond the diff hunk at surrounding context
      - Check related files if referenced

   c. **Critical analysis:**
      - Does the code already address this concern?
      - Has this been fixed in a subsequent commit?
      - Did the reviewer miss implementation details elsewhere?
      - Is the feedback still relevant?

   d. **Decision:**
      - **If code appears to already address the concern:** Present to user: "The comment says X, but the code appears to already Y. Should we still address this?"
      - **If clearly needs fixing:** Note it for the todo list
      - **If ambiguous:** Present findings and ask for user's judgment

3. **Create TodoWrite list:**

   After validating all comments, create a TodoWrite checklist with only the confirmed items that need to be addressed. Include:
   - File path and line number
   - Brief description of what needs to be fixed
   - Reference to the original comment

### Phase 2: Fix Issues

Work through the TodoWrite list systematically, addressing each PR comment. Use your standard todo handling workflow:
- Mark items as in_progress
- Make necessary code changes
- Explain what was changed and why
- Mark completed after each fix

### Phase 3: Style Guide Analysis

After all fixes are complete:

1. **Read the current style guide:**
   ```
   docs/guides/go-style-guide.md
   ```

2. **Analyze for gaps:**
   - Review the PR comments that were addressed
   - Identify patterns and themes in the feedback
   - Compare against existing style guide content
   - Determine if the patterns are already documented

3. **Report findings:**

   **If guide already covers the patterns:**
   ```
   Style guide analysis: All feedback patterns are already documented in the guide. No updates needed.
   ```

   **If gaps are found:**
   ```
   Style guide gap analysis:
   - Gap 1: [Describe the pattern from PR feedback] → Missing from guide
   - Gap 2: [Another pattern] → Should be added to [specific section]

   Proposed additions:
   [Show the specific text to add and where in the guide it should go]

   Should I update the style guide with these additions?
   ```

4. **Update if approved:**
   - If user confirms: Use Edit tool to add the new guidelines
   - If user declines: End workflow
   - If user wants revisions: Adjust and ask again

## Key Principles

- **Validate before committing:** Always check if comments reflect actual code state
- **User maintains control:** Get confirmation on questionable feedback
- **Intelligent updates only:** Only update style guide when there are clear, valuable additions
- **Leverage existing behavior:** Use standard Claude Code workflows for todo handling and file editing

## Notes

- The `scripts/format_pr.py` script requires the `gh` CLI to be authenticated
- It automatically detects the PR number from the current branch
- Comments are fetched via GitHub GraphQL API for full context (diff hunks, replies, resolved status)
