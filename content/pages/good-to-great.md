---
title: "From Good To Great DevX"
date: 2026-01-18
author: James
tags: ["AI Assisted Dev", "Working In The Open", "DevX", "Claude Code", "LLM", "Workflows"]
voiceBased: true
pinned: false
draft: false
---

# From Good To Great DevX

Up until this [Memos project](https://memos.alki.me/pages/readme/), my experience with AI coding assistants was what I'd call "AI pairing"—reaching for agent mode when I needed help exploring unfamiliar parts of a codebase. That was good, even magical at times. But I sensed there was *more* to fully agentic workflows, and I wanted to understand what separated good from great. Through building Memos, I've landed on some answers that work for me: great AI DevX comes from LLM *independence*, and independence requires high-quality Memory plus orchestration of planning context. Your mileage may vary—but this is a little bit about of how I got there.

## Your LLM is Always Onboarding

Let's start with high-quality Memory. 

There's a core insight that is best expressed as this analogy: [your LLM is always onboarding](https://memos.alki.me/posts/2026/01/your-llm-is-always-onboarding/). Every session, you're essentially handing context to a new team member who's smart but has no memory of yesterday's conversations.

This reframes a lot of what I like to call the "eat your veggies" engineering practices—good docs, clean architecture, clear commit messages, tests. We know we *should* do these things, but they often slip down the priority list. The reason they matter is that they help teammates understand complex, opaque systems when they come to them fresh.

When your LLM is always onboarding, that list of "teammates who need support" gets a lot longer—it's every LLM session you spin up! High-quality, high-fidelity documentation is no longer something to backburner. It's table stakes for agentic workflows.

This insight leads to two practical focus areas: **learning loops** that keep your Memory files accurate, and **clean codebases** that help the LLM plan effectively.

## Learning Loops

We've all been there. "Dude, I've told you this a thousand times! Please make sure you…" whatever it may be. Once I understood a bit more about Memory, I actually went looking for ways to hook in learning loops into my process, where the session isn't just about producing a solution, but it's a meta-level step of capturing gaps in the LLM's Memory files and updating those files to make them higher quality for future sessions.

- [Turn PR Comments Into LLM Feedback](https://memos.alki.me/posts/2025/11/pr-comments-as-a-training-loop/): This was my first experiment with this idea, and it felt like a big unlock when I started doing it. Lots of details in the linked post, but in short: I built a system where I'd create inline comments on LLM-generated PRs (see code reviews below), then gave Claude the ability to read these comments. I'd prompt the LLM to address them, and then when finished, consider what, if anything, needed to be updated in the language style guides in the Memory files.
- [Update Architecture Command](https://github.com/alkime/memos/blob/main/.claude/commands/docs/update-architecture.md): Where the PR comments flow updates the LLM's understanding of coding language or structure, the architecture doc portion of the Memory files naturally drifts from the ground truth of the real architecture of the app. The command looks at the architecture doc, finding which commit last updated it, and then inspects all the commits and PRs that have happened since then. This is where commit and PR messages are important because this kind of learning loop is much more time-consuming without them!

***Up Next***

I suspect there may be another loop leveraging hooks to sort of create a log of corrections and preferences that can be detected with periodic semantic analysis over chat transcripts. Each could then be reflected into the Memory files in the same sessions that the architecture is updated. I haven't had a chance to explore this idea in depth yet, but it's in the wheelhouse of the kinds of mechanisms I think are important for keeping the Memory fresh.

### Clean Codebases

I've noticed that more or less every time I'm in planning mode with Claude Code, the first thing it does is grep around the codebase for patterns it sees (the latest incarnation deploys the built-in [Explore subagent](https://code.claude.com/docs/en/sub-agents#built-in-subagents)). It's expanding the context—layering on top of its Memory files so that it has a better understanding of how the problem it's been presented should be solved.

Complex and chaotically organized codebases will naturally confuse it. If packages are overwrought, or if there is more than one way to do the same thing, or a given interface has tons of methods—this all pollutes what the LLM knows and how it plans to solve the problem you've tasked it with. LLMs are also notoriously bad at knowing when to NOT do something when their context is confusing. This is where the analogy that this is an onboarding dev sort of falls down, because at least the dev will be like "WTF mate?"

This really struck home for me when I was struggling to get the [voice CLI onto the Bubbletea TUI framework](https://memos.alki.me/posts/2025/12/bubbly-t/). The first pass at the TUI was, frankly, garbage and needed to be scrapped. It wasn't until I pulled back, [decided to do it myself](https://memos.alki.me/posts/2025/11/when-to-build-it-yourself/), and introduced some good patterns that things improved. I learned a lot about Bubbletea in the process, which enabled me to understand what was good and where corrections needed to be made, further feeding the learning loops. The final structure separates concerns and codes to a simple structure, while also separating out "widgets" from "containers" for a bit more clarity that allowed the LLM to more or less add phases without any handholding.

## Continue Doing Code Reviews

I'm still for code reviews. This is a primary mechanism I've deployed to really give the LLM batched feedback—there are many times when the LLM needs a lot of corrections. Code review systems are *already a core process* every dev team knows about, so why not continue to leverage this? Setting up the learning loop approach above created a positive feedback loop that pushed the system closer to agentic independence. With this learning setup, there may be a point where my Memory files have been cultivated in such a way that most of the reviews are rubber stamps. Wouldn't that be something!

This isn't to say an AI pairing motion isn't warranted here. The LLMs can review code too, and I do think an interesting setup exists where you have *another* LLM provider do the review. For a bit I was paying for both GH Copilot and Claude Code, so I'd have Copilot tag in for reviews. I thought it was high quality, and it being GH, it did inline comments—wiring up into the core feedback loop.

## Scaling With Parallelism

The idea of parallelism feels like a slightly less spoken-about idea in AI DevX, so it's worth level-setting here, I think. What I'm talking about here is the most basic idea where your setup is doing more than one thing at the same time. There are many approaches and tools to build this kind of workflow, some native like Claude Code's Background Tasks and Cursor's Background Agents, to DIY setups with multiple checkouts of the same repo all tmux'd out.

You pay a cost in coordination—isn't that always the case with parallelism??—but my sense is that agentic systems need some kind of independence built into them, which naturally empowers the system to scale up and really deliver on its promise. If AI pairing is getting to ***cruising altitude***, agentic dev is blasting into ***orbit***.

The coordination I'm referring to exists in at least two levels. The lowest level has to do with making sure updates happen in isolation. Classic write contention problem. If you're not careful and several LLM instances or background tasks are operating on the same repo, the chances are high they will clobber one another ("clobber" being the technical term for data write contention).

There are solutions to this, from the obvious (multi-repo checkouts) to some little-known (at least to me) git features (see `git-worktree`). This level is interesting, but I won't discuss it too much right now. I think what's more interesting is coordination of Memory, particularly in the form of memory related to planning.

### Planning

The most successful sessions will have a planning phase—where you and the LLM brainstorm on what the feature or task it will be working on, and the plan is somehow recorded and added to accessible Memory. This Memory is then used to implement the plan.

The planning phase is when the LLM's Memory gets augmented with targeted exploration of the codebase, where good (or not so good) patterns are loaded up and used as kindling for a plan document. This document then becomes context, becomes Memory, that the current or future sessions can work off of.

{{< image-caption src="/images/pages/good-to-great/single_session.png" alt="Single Sessions" caption="Simplest. Single session of planning and implementation" >}}

This diagram shows how planning context—the `plan.md`—is passed from planning phase to implementation phase. In the simplest frame, it's all in the same session.

Saving the plan to disk has a nice benefit. Let's imagine we get to the end of our context window and we haven't yet done any implementation (for the sake of argument, let's pretend compaction is not a thing). Because the plan is on disk, we can simply start a new session and instruct Claude to execute the plan doc, and off it goes.

{{< image-caption src="/images/pages/good-to-great/multi_session_simple.png" alt="Simple Multi Session" caption="Simple but with more robust context management..." >}}

So the final idea to express here requires one more leap of imagination. As the feature set gets more complicated, the plan can get larger and more complex. Let's say there is a frontend piece and a backend piece, and a complicated testing part of the plan. You can break these up into more than one planning file—a `frontend_plan.md`, a `backend_plan.md`, and a `testing_plan.md`.

{{< image-caption src="/images/pages/good-to-great/multi_session_complex.png" alt="Full Multi Session" caption="Enable agentic parallelism with many implementation instances..." >}}

Here, we have a planning session which generates a detailed and fairly complicated plan, then we spin up implementation sessions, each one executing a different plan. Sequencing may be necessary—the testing plan execution might need some or all of the other two to have completed before it can start executing—but the theory allows for some impressive parallelism.

### Advanced Planning Systems

Writing a plan file to disk as detailed above is the native, basic form of planning that happens in Claude Code. Project planning apps are plentiful though, and through mechanisms like MCP, the LLM can be given access to them. Linear has an [official MCP integration](https://linear.app/integrations/claude), as does [Atlassian](https://www.atlassian.com/blog/announcements/remote-mcp-server). My sense is that teams might have a bit of use-case creep here since the context generated for a useful LLM plan might be overly verbose for these tools.

There is a tool named [`beads`](https://github.com/steveyegge/beads) by Steve Yegge that delivers a lot of the features you'd see in some of these systems but all locally synced with your repo through use of SQLite & JSONL data. The issues stored have more structure and are linked. The structure gives them bead titles, descriptions, types (bug, feature, task, epic, …), key-value labels, and importantly dependency linkages, as well as a simple set of commands to get ready work.

I recently moved the Memos project onto beads. This included building out two subagent definitions: an `architect` which handles the requirements gathering and technical design pieces in the "planning" box above, and an `artificer` which is the generic "pick up this task and implement it" agent.

I implemented this *using* beads. First step after initializing the beads db was to generate the architect, ensuring it was "beads aware," then I used the architect to flush out the new "Setup Beads Based Workflow" epic whose task was to plan out and generate the artificer definition. Quite meta, but also quite useful.

{{< image-caption src="/images/pages/good-to-great/bees_question_mark.png" alt="beads" caption="Bees? ... BEADS!!" >}}

This exercise was my first time exploring them, and I'm just reminded again that everything comes back to context. The agent definition is a way to steer the LLM to great by layering on specific, domain-specific context to what it's doing in that session, increasing the autonomy and chances of a good outcome. It all comes back to context.

# Final Thoughts

Most of my AI workflow before this project was what I would call "AI pairing," where the LLM helped me explore ideas or even increased my capabilities. This was *good*, but I knew there was a *great* somewhere, and it turns out that *great* AI DevX comes from LLM *independence*. In order for the LLM to be independent, it needs high-quality *ground truth* Memory, and systems to empower parallelism through orchestration of planning *memory*.

As I worked on Memos, I was able to dig deeper on the why and the how of powerful AI-based workflows, getting closer to a root conceptualization of AI DevX. Astonishing about covers it, and I feel energized for what's next.

---
{{< byline >}}
