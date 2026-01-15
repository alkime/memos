---
title: "Your LLM Is Always Onboarding"
date: 2026-01-15
author: James
tags: ["AI Assisted Dev", "Claude Code", "LLM", "Developer Experience"]
voiceBased: false
pinned: true
draft: false
---

# Your LLM Is Always Onboarding

When you onboard a new dev to your team, what happens? One common cornerstone is that you probably point them at your documentation: architecture docs, code style guides, testing guides, deployment guides.

If you're like me, every time you're supporting a new engineer as they go through this process, there's a small pit in your stomach because these docs are woefully out of date. There's an old saying: "a new car loses 20% of its value the minute you drive it off the lot." There's a corollary about your technical docs, which is "your docs are stale the minute they're published."

Yet we always point the new dev at them. Why? Well, they're hopefully _mostly_ correct, and the knowledge transferred sets the foundation where the real knowledge transfer happens: the code walkthrough pairing sessions. If we're really ambitious, we can also tidy up those docs as we go. These onboarding sessions are expensive, but well worth it. A fully onboarded teammate is highly functioning and independent, greatly increasing what the team as a whole can achieve.

Now, let's imagine our documentation is of the highest quality. It's super fresh, reflecting the ground truth with high fidelity, and constructed in a way that the new dev reading it absorbs it like Neo learning kung fu. Our new teammate gets immediately to high-impact after one read-through—no costly pairing with engineering leads required. Wow, what a world that would be!

When your LLM starts up a session—before it's loaded any context—every time is its first day on the job. The context you give it—Memory in Claude Code (and others) parlance—is analogous to the onboarding docs you give your new dev teammate to get them to high-functioning ASAP. The quality and fidelity of these "onboarding docs" have an outsized impact on just how well it ends up functioning.

Your LLM is always onboarding.

This analogy has been a bit of a eureka moment for me as I try to onboard (yuk yuk) Claude Code. Claude Code is flexible, purposefully so, and this flexibility means there is _a lot_ to get a handle on, and it can be hard to know what to focus on. This analogy has given me a heuristic on where to direct that focus.

For example, I haven't spent a lot of cycles working with [subagents](https://code.claude.com/docs/en/sub-agents). Should I shift into learning about them, or should I expand my learning loop modalities (such as my process for turning PR feedback into LLM feedback as described in this post on [turning PR comments into LLM learning]({{< ref "posts/2025-11-pr-comments-as-training-loop.md" >}}))? 

From the frame that your LLM is always onboarding, it's pretty clear that I should invest in strengthening the learning loops. I'll do that next.

---
{{< byline >}}
