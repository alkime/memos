---
title: "Bubbly T"
date: 2025-12-19
author: James
tags: ["AI Assisted Dev", "Working In The Open", "Go", "TUI", "Bubbletea"]
voiceBased: true
pinned: true
draft: false
---


## Bubbly T

{{< image-caption src="/images/posts/2025-12/demo.gif" alt="Voice CLI terminal UI demo" caption="so bubbly...">}}

This memo is actually pretty special—it's the first one I'm recording with the new terminal UI I've been building over the last few weeks. I'm working on this project implementing a full end-to-end workflow for a voice CLI using a terminal UI library called [Bubbletea](https://github.com/charmbracelet/bubbletea). When I'm recording this, it's more or less in its final form. It's been a minute since the last post—almost exactly a month to the day.

Hopefully in the post I'll include some nice terminal recordings of the UI. It was kind of apropos that the last post was about "[when to do it yourself]({{< ref "posts/2025-11-when-to-build-it-yourself.md" >}})." This piece checked all the boxes for that—all of the things from that post were true here. When you want to do something for quality, when there's no prior art, or when you just don't know enough to guide the LLM, those were basically all true. So I definitely wanted to go out and be highly involved in the pieces, and there was no exception here.

I'll talk about a few things: first, why `bubbletea`, then a bit more about the mode I was in while doing this. It certainly wasn't fully non-LLM mode. And lastly, why good patterns are so important—that idea was really brought home through this part of the project.

## Why Bubbletea?

As with most tech selections, there's a lot out there. Even in Golang, there's a lot of terminal UI libraries. `bubbletea` had crossed my path in the past and was interesting for a few reasons. Whenever I do this now, I go into Claude and use deep research mode to brainstorm. I didn't do full deep research this time—I just used Opus in thinking mode (4.5 was out at this point) to explore some options.

It actually suggested `bubbletea` too. It's got a simple API at its core. The model and message passing architecture is really powerful—you have these simple building blocks that, when you combine and compose them together, create something really powerful. And like most Go libraries, you can read the source code and if needed navigate what's under the hood.

One of the struggles—maybe not for everyone, but for me at least, coming from my server architecture background—is that often UI architectures are different enough that it takes some real elbow grease and focus to get something good.

And also the ecosystem and what they have done with the community seems vibrant to say the least. It's got an opinion on its aesthetic for sure. I think that's a really strong signal for the longevity of an open source project. It's different, at least to a graybeard like me, and there's some fun in that that helps make palatable the inevitable frustration one feels when learning something new.

## The Learning Mode

So I went in headfirst. With anything like this, you have to ask: how much do you want to farm out to the LLM? As I mentioned, this checked the heuristic:

1. Building for quality (perhaps artificially, but everything in this project checks this)
2. No prior art in the codebase. Before, it was all CLI-based, just running sub-commands chained together.
3. I had not done any TUIs in the past, certainly not any `bubbletea`. How would I know if the LLM was slopping me?

So there is another piece that I didn't initially put into the previous post (might go back and update it for completeness): it's fun to learn new things! Sometimes you just want to mess with something to learn about it, and TUIs have been on my list for a very long time. I have this opportunity to learn about it, and I now feel confident I have this tool in my toolbox going forward. But moreover, we as engineers need to still do energizing things for our craft, carving out time to do what we love which is building and crafting and learning and, yes, writing code.

## So How'd It Go?

So there was a bit of a creative loop I got in where I would try out things with the goal of getting some repeatable pattern I could then deploy Claude on flushing out. It wasn't super well planned out. There wasn't a full tech design spec that I would normally do and have reviewed with my team. (This is still a side gig, family stuff at the holidays was also going strong!) It was mostly like, okay, I'm going to try this, I'm going to try that. That's why it took so long because there were some inevitable dead ends.

What was interesting was that my initial take on how to do this was too attached to the old CLI pattern, just running things on the command line. Because that pattern was polluting what I was trying to do, it created a weird scenario where I couldn't pass the baton to the LLM. I did this first piece, it was kind of kludgy and not great. Then I prompted the LLM to do the next parts, and it just wasn't any good. This leads to an interesting thought: not just a heuristic to know "when to do it yourself," but also more of a goal of "how to get to good."

I was also learning about what works and what doesn't with `bubbletea`. Second versions are often much better, and so I was able to do things like build out the `phases` component, which made everything more modular which (duh) makes it easier for the LLM to work with. There were also some mp3 encoding woes I ran into. But once I got past all that, I got to a point where the code was well organized, I had implemented the recording phase and the transcription phase of the workflow on top of the container pieces I'd built, and it all sorta came together, and Claude filled out the rest.

It brought home this idea that good patterns and good code organization is important, combined with this other piece: when you're doing this stuff yourself and being the architect of the feature you're trying to deliver, especially with greenfield stuff, you really want to figure out how to be repeatable for the LLM. How can you get it into a step where you can then farm it out, maybe even in a parallelized fashion? I didn't do any of that this time, but I could see that being a really big piece. Good package organization with loosely coupled interfaces, and single domain ownership within your information architecture. You know, the good hygiene we strive for—turns out it's also foundational for successful LLM output.

---

Everything was basically working, but I wanted to add a component that would display during the recording phase—a little waveform showing the last packet or two from the audio device, so I could really tell data was flowing through the system.

I did a pretty good prompt, and Claude basically went out, did the plan, looked at all the patterns, and delivered a plan that covered all these pieces in basically one shot. For example, it identified it needed a ring buffer to be added to the audio file code, which I hadn't mentioned in my initial prompt other than "we need to worry about memory churn here."

It did the ring buffer, figured out the remote control stuff, really easily added it into the recording phase. I just started it up and saw this little waveform and for the first time I said out loud "oh wow!" because my vision for what I wanted out of this little app had just appeared to me. It had done it with minimal hand holding.

It was a really nice little capstone on where I wanted to get with this project.

---
{{< byline >}}