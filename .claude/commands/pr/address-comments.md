---
name: pr:address-comments
description: Process PR review comments - fetch, validate, fix issues, update style guide
---

# Address PR Review Comments

This command processes PR review comments through a structured workflow: fetch comments, validate against code, fix issues, and intelligently update the relevant style guide(s).

**Note:** PR comments may come from human reviewers OR AI agents (GitHub Copilot, AI code reviewers, etc.). The user may resolve AI-generated comments they don't want to address, sometimes including responses explaining their decision. These responses are valuable context for style guide updates.

## Style Guides

This project maintains style guides based on the area of the codebase:

| Guide | Path | Applies To |
|-------|------|------------|
| Go | `docs/guides/go-style-guide.md` | `**/*.go` |

## Workflow

### Phase 1: Fetch and Validate Comments

1. **Fetch all PR comments:**
   ```bash
   python3 scripts/format_pr.py
   ```
   This script uses the GitHub API to fetch all review threads and formats them as Markdown, showing unresolved comments first, followed by resolved comments.

2. **Understand the comment landscape:**

   - **Unresolved comments:** Need validation and potential fixing
   - **Resolved comments with user responses:** User intentionally dismissed these, but responses may contain valuable context for style guide updates
   - **Resolved comments without responses:** User silently dismissed, less valuable for learning

3. **Validate each unresolved comment individually:**

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

4. **Create TodoWrite list:**

   After validating all unresolved comments, create a TodoWrite checklist with only the confirmed items that need to be addressed. Include:
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

1. **Determine which guides to analyze:**

   Based on the files touched by PR comments, identify which style guide(s) are relevant:
   - Go files (`*.go`) → `docs/guides/go-style-guide.md`

2. **Read the relevant style guide(s):**

   For each relevant guide, read its current contents.

3. **Analyze for gaps:**
   - Review the PR comments that were addressed (Phase 2 fixes)
   - **IMPORTANT:** Also review resolved comments where the user provided responses
     - User responses to AI reviewers often explain WHY certain patterns were chosen/rejected
     - These explanations are valuable for documenting decision rationale
   - Identify patterns and themes in both the feedback and user responses
   - Compare against existing style guide content
   - Determine if the patterns are already documented

4. **Report findings:**

   **If guide already covers the patterns:**
   ```
   Style guide analysis: All feedback patterns are already documented in the guide(s). No updates needed.
   ```

   **If gaps are found:**
   ```
   Style guide gap analysis:

   ## go-style-guide.md
   - Gap 1: [Describe the pattern from PR feedback] -> Missing from guide
   - Gap 2: [Pattern from user response to AI reviewer] -> User explained why they rejected AI suggestion; rationale should be documented

   Proposed additions:
   [Show the specific text to add and which guide it should go in]
   [Include rationale extracted from user responses when applicable]

   Should I update the style guide(s) with these additions?
   ```

5. **Update if approved:**
   - If user confirms: Use Edit tool to add the new guidelines to the appropriate guide(s)
   - If user declines: End workflow
   - If user wants revisions: Adjust and ask again

## Key Principles

- **Validate before committing:** Always check if comments reflect actual code state
- **User maintains control:** Get confirmation on questionable feedback
- **Intelligent updates only:** Only update style guides when there are clear, valuable additions
- **Right guide for the job:** Match feedback to the appropriate style guide based on file type
- **Leverage existing behavior:** Use standard Claude Code workflows for todo handling and file editing

## Notes

- The `scripts/format_pr.py` script requires the `gh` CLI to be authenticated
- It automatically detects the PR number from the current branch
- Comments are fetched via GitHub GraphQL API for full context (diff hunks, replies, resolved status)
- **AI-generated comments:** Some PR comments come from AI agents (GitHub Copilot, etc.) rather than human reviewers
- **User responses are valuable:** When the user resolves an AI comment with a response explaining their decision, that response often contains valuable insights for the style guide
