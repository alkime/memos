---
title: "Voice CLI Kickoff"
date: "2025-11-10T10:00:00-08:00"
author: James
voiceBased: true
draft: false
pinned: true

tags: ["AI Assisted Dev", "Working In The Open"]
---

# Voice CLI Kickoff

Today I want to talk about the next phase of the memos project. As I'm recording this, I've reached a point where the primary loop—the main V0, basically—is in a pretty solid state. The publishing, hosting, code-based changes, initial Claude Code configuration, initial Hugo and Fly.io configuration are all working well. If I just left it here, I could have a nice blogging website posted on an infrastructure-as-a-service provider, and I would have learned a fair amount about Claude Code.

The end goal was always something more interesting, and I was now ready to start imagining how to do the innovative stuff I set out to do, with Claude as my ally.

As I've mentioned in previous posts, right now I'm creating the initial draft of most posts through a multi-step, very manual process: voice memo recording with automatic transcription, taking that transcript and putting it into Claude, having it do a cleanup pass, and then I do a final edit. That becomes the post.

What I really want to do is add these features into the app itself, making that “top-of-funnel posting process” more optimized so I can actually create posts faster.

## Going Deeper

Adding voice recording & LLM interaction was a major update, so I found myself back in the Claude desktop. For me, when wha the LLM will be doing feels like “market research”—like I was asking Claude to assume the role of “senior technical product manager”—where it’s doing a lot of web searching and deep research—this mode works best.

For example, there was a lot of library selection work: an audio recording component I'd never messed with in Go, and which transcription library to use (OpenAI's or maybe something else), and an overall plan for getting from the very manual process I'm doing now to what it can be.

This was the first time I deployed the Opus 4.1 model for this project. I kicked off the deep research aspect and let Claude go wild on it. It chewed on the problem for about 50 minutes—probably the longest I'd ever seen Claude do research. It came up with a very extensive document with deeply researched information. Emphasis here on ***extensive***.

As I was reading through the resulting artifact, I concluded there was a lot of great stuff in there, but for the things I knew about I wanted to make some tweaks. I wasn’t quite sure about continuing in the Opus chat, so I spun up a new chat on Sonnet 4.5, attached the extensive plan as context, and prompted with the theme of "hey, we want to pare this down and make it more usable for a Claude Code environment."

I'm not sure if that approach was best. It kind of feels like when you're hoarding your potions in an RPG. You're like "ah, I got to keep all of these health potions because I never know if I'm going to use them." Next time I do this, I’ll make a point of sticking with Opus and just push for a final result in-chat.

Anyway, I did get a pretty good result in the followup sonnet 4.5 chat, though it took some back and forth. What I got was a document I could point Claude Code to in-repo, and basically saying “let’s build a plan document only for this piece since everything is so big.” Sticking to a core principle of trying to keep changes small.


## Implementation

***… and the back edge of the “TDD Skill” knife …***

The implementation went okay. There was a lot of code to be written, and Claude deployed the Test Driven Development `superpowers` Skill, which was the first time I'd seen this (check out more about the superpowers plugin and Skills [in this post]({{< ref "2025-10-booting-up-with-claude-part-1.md" >}})). I thought it was quite interesting to see how it worked, but it created a bit of a double-edged sword.

What was interesting was that the initial tests it generated were all somewhat placeholder-y. This really came to the foreground with the `record` subcommand, which included adding the audio recording library the plan had selected and does a bunch of stuff that's outside my wheelhouse—new greenfield stuff for me. Claude created a test that was really just writing files to the file system. Then it went in and implemented the code to make the test pass, and when that was done, it called itself done. It didn't really do anything substantial.

My sense is that this is probably a core challenge with leveraging TDD as a correctness check for LLM changes. Maybe “core challenge” isn’t quite right—but something you have to really prompt around. Like how do you push the system to know when to stop? It can go too far, but it can also struggle with knowing when to go further, especially when instructed that getting the test green is what it should be going after.

Sort of coupled with another idea that I’m stewing on: knowing when to say “no, I’m going to do this myself.” Like this part of the overall code is greenfield enough for me that it probably is worthwhile for *me* to learn about the systems in use so I can guide Claude now and in the future for how to best use the audio recording SDK we’ve chosen. Maybe the guideline is something like “if it’s core to the effort at hand, spend time getting up to speed on it so you can be a better coach.”

So I adjusted the plan, saying we're going to implement only the placeholder here because I want to do a follow-up PR myself.

### Of Interest: Getting Mini-Ambitious

One thing that stood out as Claude was developing this plan: there were some nice file management features Claude ended up proposing—what to do with the resulting files on disk, archiving completed files into different directories, all these things I would have thought of as a stretch for a V0. I’d normally think I’d want to not spend too much time on these pieces, but they are certainly really nice to have, and really do hit a nice sweet spot for an LLM. I was left thinking that this sort of polish gives rise a feeling of being mini-ambitious—we as devs should feel like we can add more small-but-impactful things to what we think are possible in any given motion.

## Code Review as LLM Feedback

As Claude wrapped up what it wanted to do, there were several corrections that needed to be made across many places in the codebase. It was missing key features, using the wrong pieces of structs in places, doing print statements where it should be using the logger—all this stuff.

It’s these kinds of results that really create headwinds for developers to jump into AI-assisted dev. I certainly felt it. I was like: great, now I’m gonna spend hours prompting each of these things to get fixed. But I had a bit of an eureka moment: if I am treating this thing like a smart junior dev that just joined my team, how might I proceed? One thing I’d do (if I couldn’t sit down next to them at a workstation) is leave detailed comments/suggestions on the parts of the PR I’d like to see changed. Maybe I could just do that, and then give my “junior dev” the ability to read, understand, and work on these suggestions.

This spawned a little side quest to build out this ability, which sort of encapsulates what I have a hunch are some powerful ideas around correcting these LLMs but also encoding the learnings for the future, so look for this in the next post.

---
{{< byline >}}
