#!/usr/bin/env python3
"""Format GitHub PR inline comments as Markdown.

Meant to be used with the gh CLI, like so:


gh api repos/:owner/:repo/pulls/$(gh pr view --json number -q .number)/comments | ./format_pr.py
"""

import json
import sys

def main():
    comments = json.load(sys.stdin)

    print(f"# PR Inline Code Comments ({len(comments)})\n")

    for comment in comments:
        author = comment.get('user', {}).get('login', 'Unknown')
        path = comment.get('path', 'unknown file')
        line = comment.get('line') or comment.get('original_line', '?')
        body = comment.get('body', '')

        print(f"## {author} on `{path}:{line}`\n")

        # Include the code diff context if available
        if comment.get('diff_hunk'):
            print(f"```diff\n{comment['diff_hunk']}\n```\n")

        print(f"{body}\n")
        print("---\n")

if __name__ == '__main__':
    main()
