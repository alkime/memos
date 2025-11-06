---
title: "Booting Up With Claude Part 2 - Getting to V0"
date: "2025-11-06T12:15:54-08:00"
author: James
voiceBased: true
pinned: true
draft: false

tags: ["AI Assisted Dev", "Working In The Open"]
---

# Getting To v0

Okay, this is the second part of a two-part series about getting this site to what I'd call v0—basically happy with all the plumbing and really set up for adding posts and content. You can check out the first part ([Booting Up With Claude]({{< ref "2025-11-booting-up-with-claude-part-1.md" >}})), which covers getting the placeholder website up on Fly and securing it with good security headers.

If you read that first post, you can see there's obviously a lot left to do to get to v0. You could call this a grab bag of chores, and this post isn't going to go too deeply into each individual task. They're small things, but they all add up to this larger effort. What's more interesting is *how* Claude decided to go about the work, which was kind of awesome. It was my first experience watching an LLM execute things in parallel using worktrees and Tasks (capital T—anyone who uses Claude Code knows the Task Tool is a proper noun). It was really pretty awesome to see.

## The Grab Bag

So what needed to get done? A few things worth calling out:

First, I wanted to move to a ***different theme***—specifically the Bear Blog theme for Hugo. (It's linked in the footer of the site, and you should definitely check it out). It's great, very simple, and my favorite: easy to reason about when you need to make customizations, which I ended up doing quite a bit. The initial theme needed to be cleaned up and removed. It was a submodule, which I'm not super thrilled about to be honest. Git submodules are kind of a pain, and this really bit me when working with worktrees, which I'll talk about later.

I also got the ***TLS cert*** going for the final domain **memos.alki.me**. This wasn't something Claude did for me too much—I asked for guidance, but I just wanted to give props again to the Let's Encrypt effort. I'm old enough to remember when every two to four years your site would go down because someone forgot to add a to-do item due *YEARS IN THE FUTURE* to renew the cert. Thankfully that's a thing of the past now. It's so easy to get TLS and make sure your site is secure—there's just no excuse not to have HTTPS anymore. That's not a Claude thing, just an awesome thing that deserves a mention.

Other pieces like ***adding `testify` tests and Go linter*** changes are what I'd consider table stakes for any AI-assisted dev work. You need to be able to systematically and deterministically allow the AI to run commands that verify what it's doing, so you—as the one managing the AI—don't have to. The better your tests and semantic checking, the more the system can hook into deterministic ways of verifying correctness, and the better the outcomes will be.

Finally there was a change to use the ***gin `contrib/static` package*** for static files and a few ***global style changes***.

## Parallel (capital “T”) Tasks

The really interesting piece here is how Claude went about it.

The `brainstorming` Skill and `writing-plans` Skill executed again—I talk a lot more about this in the previous post. But just to reiterate, this is where I use the `superpowers` Skills I added to Claude Code to help generate a tech design docs and plans that Claude can execute toward. What's interesting is that once Claude got to the plan documented and was ready to execute, it noticed these were kind of a grab bag of disparate efforts and asked if we could maybe do this in parallel (using the AskUserQuestions UX of course).

{{< image-caption src="/images/posts/2025-11/claude_asks_parallel_tasks.png" alt="claude asks" caption="Claude Recognized We Could Parallelize" >}}

It called out benefits and drawbacks of each approach. The choice for “separate PRs” noted “parallel work” that was “more reviewable.” Yes, please. More overhead? Ok ok, but that’s a Future Me problem. I’ve been wanting to see Claude tackle something in parallel and it sorta fell in my lap here, so that’s what I chose.

And it was awesome.

Basically, I had a project branch I was working in. Claude made worktree branches for each of the efforts—ended up making four of them—and implemented each one in parallel, with a PR opened for each branch back into the project branch.

At some point Claude needed permission to run one of the test commands—I think to build the Hugo site—but we have a make target for this so I needed to correct Claude’s behavior. Now it had these four Tasks going in the background, and as far as I could tell it “stopped the world” when it needed my input, because when I answered it said: “Okay, got it. Let me dispatch this correction to the four agents.”

Not totally sure how the handoff is meant to go, but something to keep in mind for the future when running background, parallel Tasks.

## Bringing It All Down

When I finished, I had four PRs with varying degrees of complexity. One was very simple from a code perspective—just the removal of the initial Hugo theme. But because it was a submodule, that was a real pain to actually get working, especially when that PR was specifically just to remove the submodule. In worktrees, submodules don't get initiated automatically, so if you have a submodule, you need to make sure you're tracking that. (did I mention submodules blek?)

To get everything merged back into the project branch—this being my first time working with worktrees—I didn't quite have the process down. My merge process usually squashes commits into one so it's easily revertible. That's kind of my go-to for things going into main, though I probably don't need to squash into a project branch. But I went with my main habit and squashed each commit, which required a fair amount of rebasing as each successive PR came down. Claude helped with that, but each time I had to rebase, then go back into the worktree branch, refresh it, rebase it again, etc.

Trying to review all the changes for this effort would have been manageable, but probably bordered on painful, and because each effort was sort of thematically separate, it’s easier to miss something. It was really nice to have everything broken down into small, easy-to-review, easy-to-understand pieces that I could give a green checkmark when they looked good.

So I ended up with a good result that didn’t need any corrections, and more importantly was implemented in something closer to the “final form” of how I want my AI assistant to work, in parallel and producing good stuff.

---
