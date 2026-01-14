# Your LLM Is Always Onboarding

When you onboard a new dev to your team, what happens? One common cornerstone is that you probably point them at your documentation. Architecture Docs, code style guides, testing guides, deployment guides.

If you're like me, every time you're supporting a new engineer as they go through this process, there is a small pit in your stomach because these docs are woefully out of date. There's an old saying: "a new car looses 20% of its value the minute you drive it off the lot." There's a corollary about your technical docs, which is "your docs are stale the minute they're published."

Yet we always point the new dev at them. Why? Well, they're hopefully _mostly_ correct, and the knowledge transferred sets the foundation where the real knowledge transfer happens: the code walkthrough pairing sessions. If we're really ambitious, we can also tidy up those docs as we go. 

These onboarding sessions are expensive, but well worth it. A fully onboarded teammate is highly functioning and independent, greatly increasing what the team as a whole can achieve. It's truly awesome and probably the thing I love most about being a team--like Voltron coming together, the whole is greater then the sum of it's parts.

Now, let's imagine our documentation is of the highest quality. It's super fresh, reflecting the ground truth with high fidelity, and constructed in a way that the new dev reading it absorbs it like Neo learning kung fu. Our new teammate gets immediately to high-impact after one read through--no costly pairing with eng leads required. Wow, what a world that would be! 

When your LLM starts up a session--before it's loaded any context--every time is it's first day on the job. The context you give it--Memory in Claude Code (and others) parlance--are analogous to the onboarding docs you give your new dev teammate to get them to high-functioning ASAP. The quality and fidelity of these "onboarding docs" have an outsized impact on just how well it ends up functioning.

Your LLM is always onboarding.

As I've onboarded (yuk yuk) to Claude Code, this analogy--as overwrought as it may be--has been for me a slightly deeper insight then all the pro-tip sorts of posts you see on r/claudeai (don't get me wrong, I loves me a good "how I finally got Claude to hack the planet!" post).

## The Importance of Learning

This is all to say that Memory[^1] is important. It's no surprise that the first thing you do when you initial Claude Code[^2] is build its initial Memory (`/init` ). It's also important to build in processes that curate and expand this ground truth--we want our "onboarding docs" to be as high quality and useful as possible, allowing our LLM to get to a higher functioning state without a lot of in-session prompting. 

This is the core reason why I'm a strong proponent of building in ways for Claude to access code reviews [[link-to /posts/2025/11/pr-comments-as-a-training-loop/]. When I review the LLM's code, I'm giving it feedback that can be captured into it's Memory then there's a higher likelihood (tho not 100% of course!) that it won't make that mistake again. We could call this a form of learning. 

Look for places to introduce learning loops into your LLM workflows. It's not just on code changes that I try to close Memory gaps. I've put together a [custom slash command](https://github.com/alkime/memos/blob/main/.claude/commands/docs/update-architecture.md) that is in a similar vein to what I do with addressing PR comments, but meant to fill gaps in the architecture docs portion of Claud's Memory. I'll do it less frequently and more outside the normal change merging cadence, but is important way to help keep the onboarding docs fresh.

--



---

#### Footnotes

[^1]: See https://code.claude.com/docs/en/memory for more on Claude Code Memory. It's more then just `CLAUDE.md` and here when I reference Memory it _all_ of the context loaded up at session startup. 
[^2]: I've been using Claude Code, so I use the parlance of that platform. My sense is that the lexicon is transferable, sometime directly, but in the least probably can be swapped for others between the SOTA coding assistants. Memory would be an example of direct terms, whereas maybe "planning mode" might mean something different for Claude Code than it does for Codex, etc.

