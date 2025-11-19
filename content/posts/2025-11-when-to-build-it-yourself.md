---
title: "When to Build It Yourself"
date: 2025-11-18T17:10:15-08:00
author: James
tags: ["AI Assisted Dev", "Working In The Open", "Software Engineering"]
voiceBased: true
pinned: true
draft: false
---

> [!NOTE]
> Before diving into the meat of this post which discusses the idea of *when to build it ourselves*, I wanted to make a small callout. This post will be the first run through of the end-to-end version of the tool I'm calling the "voice CLI"—which is to say that instead of the old "manual" process involving iPhone voice memos + transcription, we're doing this all from a simple (for now) command line interface built into the same code repo that the Hugo site is built from, and more importantly utilizing the same AI Dev Core Principles as that effort.

## When to Build It Yourself

At least for now, it is important that as builders we have some idea about when to *not use* AI agents. This might seem counterintuitive—aren't we trying to *be offloading everything to AI*?—but as it turns out, knowing when to not deploy it has some big benefits.

The underlying idea is actually broadly applicable across most technology decisions, even boring old non-AI ones. When we have an idea of the boundaries of a solution, what we've done is created the necessary conceptual scaffolding about that solution that enables us to adapt it to novel use cases. When we understand the exceptions to the rule, it means we've thought about the rule at a level that gives us a chance at tweaking it when we need to.

So we should do the same with AI Dev, to develop these *when not to* heuristics so that we can be more effective at broadly applying it.

### Finding The Edges

While reserving the right to add or amend this idea, and to attempt to be succinct, I'll propose the following:

> When building for long term quality, you should *do it yourself when*: there is no prior art in the LLM's context for the thing being added, *or*, if you're not sufficiently an expert in the domain of that thing that you could correct the LLM when it makes mistakes.

Just to help anchor on what I mean here:

* "... the thing being added": is actually purposefully vague since I suspect it is applicable to almost anything you'd be asking of yourself or an LLM.
* "sufficiently an expert": this will require some accurate self reflection, and as I stew on it I'm realizing this may be the most difficult part of the entire heuristic for many. How do you know when you're not an expert? *Paging Drs. Dunning & Kruger...*
* Note the "OR"—that is, *even if there is no prior art*, if you are confident you are in fact an expert, why not let the AI loose? You can sort it out when it goes off the rails.
* "When building for long term quality": if you're prototyping, experimenting with UX, learning a new programming language, or building a personal and small internal tool, you'll be less concerned if anti-patterns get into the code. So this is shorthand for "when you're not building for customers" then again feel free to let the AI go nuts.

So the goal, then, is to craft good patterns that the LLM can leverage in the future. Again I'm reminded of the idea of treating the LLM like a smart junior dev who just joined your team. You wouldn't give such a teammate the task of adding some large new complexity to your application until they've gotten up to speed, right? Applying this concept to your requests of AI seems prudent.

#### An Example

I encountered this when working with the audio recording components of the `voice` cli. I don't have a lot of experience with audio libraries, especially in Go. So I actually started by asking Claude to do some "market research" for me (talked about this in a previous post...) to help me find a good library to use if one is out there. We settled on [malgo](https://github.com/gen2brain/malgo), which is a "light binding" to a C library allowing Go to call into [miniaudio](https://github.com/mackron/miniaudio).

I'm sure I could have eventually got to something I liked with the structured dev I'd be employing prior to this work, but let's think about this effort in terms of the above heuristics: Was this "for customers"? Well in all honesty, no, it's just for me, just an "internal tool". Still, the universe I setup for myself for this project sort of parks this fact in the background and tries to image we're creating for real world developers. So at least for now, I'm ok with fudging this one a bit. Was there prior art in the repo or associated context for using this library? No. But, was I an expert enough to be certain that any possible slop Claude would produce I could notice and correct it? Certainly not!

It was time to roll up my sleeves!

### What You Get

What you're doing when you build it yourself is filling gaps in the LLM's context in a way that it can leverage in the future. Basically, you're building in that prior art so that in the future, when applying the above heuristic, you *won't be able to say* there is no prior art anymore since you've added a high quality pattern that can inspire the LLM when it takes it's turn.

It's also fun! This is the kind of building that energizes me, and I suspect is energizing for many engineers. Being creative, figuring out how a new puzzle piece fits into the system in a way that makes the whole more than the sum of its parts, and optimizing that new whole for future building.

Once I got the recording down as I liked it, the remaining pieces of the voice→draft pipeline was easy stuff for the LLM, and was finished up in a flash really, so here I am with pretty usable MVP!

---
{{< byline >}}